package asupayme

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayme/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/shopspring/decimal"
)

type gatewayMethod struct {
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
	ApiKey     string `json:"api_key"`
	MerchantId int    `json:"merchant_id"`
	SecretKey  string `json:"secret_key"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams ChannelParams
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, channelParams.SecretKey, gatewayParams.Transport.Timeout),
		dbClient:      db,
		channelParams: channelParams,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	requestData, err := a.fillPayoutRequest(ctx, txn)
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

	gtwTxnId := response.Id

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &gtwTxnId,
	}

	return tr, nil
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// getGatewayMethod
func getGatewayMethod( /* methodInternalId string, methods []gatewayMethod */ ) (*gatewayMethod, error) {
	/* methodIdxInArray := slices.IndexFunc(methods, func(method gatewayMethod) bool {
		return methodInternalId == method.ID
	})

	if methodIdxInArray < 0 {
		return nil, errors.New("such method not found")
	}

	return &methods[methodIdxInArray], nil */

	return &gatewayMethod{}, nil
}

// fillPayoutRequest
func (a *Acquirer) fillPayoutRequest(ctx context.Context, txn *models.Transaction /* , method *gatewayMethod */) (*api.PayoutRequest, error) {
	fullName := txn.Customer.FullName
	if len(fullName) == 0 {
		return nil, errors.New("customer full name is required data")
	}

	request := api.PayoutRequest{
		MerchantId: strconv.Itoa(a.channelParams.MerchantId),
		WithdrawId: fmt.Sprintf("%d", txn.TxnId),
		Amount:     fmt.Sprintf("%d", txn.TxnAmountSrc),
		ApiKey:     a.channelParams.ApiKey,
		CardData: api.CardData{
			Owner:      fullName,
			CardNumber: txn.PaymentData.Object.Credentials,
			ExpMonth:   txn.PaymentData.Object.ExpMonth,
			ExpYear:    txn.PaymentData.Object.ExpYear,
		},
	}

	return &request, nil
}

// fillPaymentRequest
func (a *Acquirer) fillPaymentRequest( /* ctx context.Context, txn *models.Transaction, method *gatewayMethod, currID, preferId int */ ) (*api.PayoutRequest, error) {
	request := api.PayoutRequest{}
	return &request, nil
}

// handlePayment
func (a *Acquirer) handlePayment( /* ctx context.Context, txn *models.Transaction,  */ response *api.Response /* mapOutputKey string */) (*acquirer.TransactionStatus, error) {

	return &acquirer.TransactionStatus{
		Status: acquirer.REJECTED,
		Info: map[string]string{
			"ps_error_code": response.Error,
		},
	}, nil

	/* if len(response.Error) != 0 {
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

	return &tr, nil */
}
