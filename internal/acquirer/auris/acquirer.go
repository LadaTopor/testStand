package auris

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/auris/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type gatewayMethod struct {
	Id          string         `json:"id"`
	GtwId       map[string]int `json:"gtw_id"`
	PreferId    int            `json:"prefer_id"`
	MapInputId  string         `json:"map_input_id"`
	MapOutputId string         `json:"map_output_id"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport            Transport        `json:"transport"`
	PaymentMethods       []gatewayMethod  `json:"payment_methods"`
	PayoutMethods        []gatewayMethod  `json:"payout_methods"`
	PercentageDifference *decimal.Decimal `json:"percentage_difference"`
}

type ChannelParams struct {
	ApiKey string `json:"api_key"`
	ShopId int    `json:"shop_id"`
}

type Acquirer struct {
	api      *api.Client
	dbClient *repos.Repo

	channelParams  ChannelParams
	paymentMethods []gatewayMethod
	payoutMethods  []gatewayMethod

	percentageDifference *decimal.Decimal
	callbackUrl          string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, gatewayParams.Transport.Timeout),
		dbClient:             db,
		channelParams:        channelParams,
		percentageDifference: gatewayParams.PercentageDifference,
		callbackUrl:          callbackUrl,
		paymentMethods:       gatewayParams.PaymentMethods,
		payoutMethods:        gatewayParams.PayoutMethods,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	method, err := getGatewayMethod(txn.PayMethodId, a.paymentMethods)
	if err != nil {
		return nil, err
	}

	currID, ok := method.GtwId[txn.TxnCurrencySrc]
	if !ok {
		return nil, errors.New("no currID for this currency")
	}

	requestData, err := a.fillPaymentRequest(ctx, txn, method, currID, method.PreferId)
	if err != nil {
		return nil, err
	}

	response, err := a.api.MakeDeposit(ctx, *requestData)
	if err != nil {
		return nil, err
	}

	return a.handlePayment(ctx, txn, response, method.MapOutputId)
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	method, err := getGatewayMethod(txn.PayMethodId, a.payoutMethods)
	if err != nil {
		return nil, err
	}

	currID, ok := method.GtwId[txn.TxnCurrencySrc]
	if !ok {
		return nil, errors.New("no currID for this currency")
	}

	requestData, err := a.fillPayoutRequest(ctx, txn, method, currID)
	if err != nil {
		return nil, err
	}

	response, err := a.api.MakeWithdraw(ctx, *requestData)
	if err != nil {
		return nil, err
	}

	if len(response.Error) != 0 {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	gtwTxnId := strconv.Itoa(response.ID)

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &gtwTxnId,
	}

	return tr, nil
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger, _ := zap.NewDevelopment()

	callbackBody, ok := txn.TxnInfo["callback"]
	if !ok {
		return nil, errors.New("callback body is missing")
	}

	callback := api.Callback{}
	err := json.Unmarshal([]byte(callbackBody), &callback)
	if err != nil {
		logger.Error("Error unmarshalling callback body", zap.String("callback", callbackBody))
		return nil, err
	}

	txnStatus := acquirer.TransactionStatus{}

	if txn.GtwTxnId == nil && callback.ID != 0 {
		gtwTxnId := strconv.Itoa(callback.ID)
		txnStatus.GtwTxnId = &gtwTxnId
	}

	switch callback.Status {
	case api.StatusApproved:
		txnStatus.SetApproved()

		if !txn.IsPayment() {
			return &txnStatus, nil
		}
	case api.StatusCancelled, api.StatusExpired:
		txnStatus.SetRejected()
	default:
		logger.Info("transaction is still in the pending status")
		txnStatus.SetPending()
	}

	return &txnStatus, nil
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := api.StatusRequest{
		ShopID: a.channelParams.ShopId,
		ID:     *txn.GtwTxnId,
	}

	response, err := a.api.CheckStatus(ctx, request)
	if err != nil {
		return nil, err
	}

	switch response.Status {
	case api.StatusApproved:
		return &acquirer.TransactionStatus{
			Status: acquirer.APPROVED,
		}, nil
	case api.StatusCancelled:
		status := &acquirer.TransactionStatus{Status: acquirer.REJECTED}
		if len(response.Error) > 0 {
			status.Info = map[string]string{
				"ps_error_message": helper.DecodeUnicode(response.Error),
			}
		} else {
			status.Info = map[string]string{
				"ps_error_message": response.StatusText,
			}
		}
		return status, nil
	case api.StatusExpired:
		return &acquirer.TransactionStatus{
			Status:   acquirer.REJECTED,
			TxnError: acquirer.NewTxnError(5003, response.StatusText),
		}, nil
	default:
		return &acquirer.TransactionStatus{
			Status: acquirer.PENDING,
		}, nil
	}
}

// getGatewayMethod
func getGatewayMethod(methodInternalId string, methods []gatewayMethod) (*gatewayMethod, error) {
	methodIdxInArray := slices.IndexFunc(methods, func(method gatewayMethod) bool {
		return methodInternalId == method.Id
	})

	if methodIdxInArray < 0 {
		return nil, errors.New("such method not found")
	}

	return &methods[methodIdxInArray], nil
}

// fillPayoutRequest
func (a *Acquirer) fillPayoutRequest(ctx context.Context, txn *models.Transaction, method *gatewayMethod, currID int) (*api.Request, error) {
	var number string
	var bankCode string

	switch txn.PayMethodId {
	case "p2piban", "p2pm10", "p2pemanat", "banktransfer":
		iban := txn.PaymentData.Object.Credentials
		if len(iban) == 0 {
			return nil, errors.New("iban is missing")
		}
		number = iban
	case "sbp":
		customerPhone := txn.Customer.Phone
		if len(customerPhone) == 0 {
			return nil, errors.New("customer phone is missing")
		}

		bank := "ru_sberbank"

		number = customerPhone
		bankCode = bank
	}

	request := api.Request{
		ShopID:    a.channelParams.ShopId,
		UniqID:    fmt.Sprintf("%d", txn.TxnId),
		UserID:    fmt.Sprintf("%d", txn.Customer.AccountId),
		CurrID:    currID,
		Amount:    txn.TxnAmountSrc,
		Label:     fmt.Sprintf("%d", txn.TxnId),
		Memo:      "Test Order",
		Number:    number,
		StatusURL: a.callbackUrl,
		BankCode:  bankCode,
	}

	if txn.TxnCurrencySrc != "TRY" {
		return &request, nil
	}

	fullName := txn.Customer.FullName
	if len(fullName) == 0 {
		return nil, errors.New("customer full name is required data")
	}

	request.ExtraInfo = fullName

	return &request, nil

}

// fillPaymentRequest
func (a *Acquirer) fillPaymentRequest(ctx context.Context, txn *models.Transaction, method *gatewayMethod, currID, preferId int) (*api.Request, error) {
	request := api.Request{
		ShopID:    a.channelParams.ShopId,
		UniqID:    fmt.Sprintf("%d", txn.TxnId),
		UserID:    fmt.Sprintf("%s", txn.Customer.AccountId),
		CurrID:    currID,
		Amount:    txn.TxnAmountSrc,
		Label:     fmt.Sprintf("%d", txn.TxnId),
		Memo:      "Test Order",
		StatusURL: a.callbackUrl,
	}

	if txn.PayMethodId == "p2pcarduzcard" || txn.PayMethodId == "p2pcardhumo" {
		request.PreferId = preferId
	}

	return &request, nil
}

// handlePayment
func (a *Acquirer) handlePayment(ctx context.Context, txn *models.Transaction, response *api.Response, mapOutputKey string) (*acquirer.TransactionStatus, error) {

	if len(response.Error) != 0 {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	gtwTxnId := strconv.Itoa(response.ID)

	tr := acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &gtwTxnId,
		Outputs: map[string]string{
			"credentials": response.Number,
			"qr_data":     response.PaymentLink,
		},
	}

	var mappingBank string
	switch txn.PayMethodId {
	case "p2psbp", "banktransfer":
		mappingBank = response.Nspk
	case "p2pcard":
		if len(response.BankID) != 0 {
			mappingBank = response.BankID
		} else {
			mappingBank = response.Bank
		}

	default:
		return &tr, nil
	}

	tr.Outputs["bank"] = mappingBank

	return &tr, nil
}
