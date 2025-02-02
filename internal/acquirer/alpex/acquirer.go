package alpex

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/labstack/gommon/log"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/alpex/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/shopspring/decimal"
)

const (
	DirectionBuy  = "BUY"
	DirectionSell = "SELL"
)

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport            Transport        `json:"transport"`
	PercentageDifference *decimal.Decimal `json:"percentage_difference"`
}

type ChannelParams struct {
	Id       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Acquirer struct {
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        ChannelParams
	percentageDifference *decimal.Decimal
	callbackUrl          string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.Login, channelParams.Password),
		dbClient:             db,
		channelParams:        channelParams,
		percentageDifference: gatewayParams.PercentageDifference,
		callbackUrl:          callbackUrl,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestBody := &api.Request{
		ExternalId: strconv.FormatInt(txn.TxnId, 10),
		Id:         strconv.FormatInt(txn.TxnId, 10),
		Symbol:     txn.TxnCurrencySrc,
		Amount:     txn.TxnAmountSrc,
		Direction:  DirectionBuy,
		FullName:   txn.Customer.FullName,
		WebhookUrl: "https://webhook.site/f4368fb0-9d4d-4400-90fe-506b7271996c",
	}

	response, err := a.api.MakePayment(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.ApproveCode,
			},
		}, nil
	}

	outputs := map[string]string{"credentials": response.PaymentMethod.Address}
	if len(response.PaymentMethod.Gate.Name) != 0 {
		outputs["bank"] = response.PaymentMethod.Gate.Name
	}
	if len(response.PaymentMethod.Person) != 0 {
		outputs["description"] = response.PaymentMethod.Person
	}

	return &acquirer.TransactionStatus{
		Status:  acquirer.PENDING,
		Outputs: outputs,
	}, nil
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestData, err := a.fillPayoutRequest(ctx, txn)
	if err != nil {
		return nil, err
	}

	response, err := a.api.MakePayout(ctx, requestData)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.ApproveCode,
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

// fillPayoutRequest
func (a *Acquirer) fillPayoutRequest(ctx context.Context, txn *models.Transaction) (*api.Request, error) {
	fullName := txn.Customer.FullName
	if len(fullName) == 0 {
		return nil, errors.New("customer full name is required data")
	}

	request := api.Request{
		ExternalId:      strconv.FormatInt(txn.TxnId, 10),
		Id:              strconv.FormatInt(txn.TxnId, 10),
		Status:          txn.TxnStatusId,
		Symbol:          txn.TxnCurrencySrc,
		Amount:          txn.TxnAmountSrc,
		Credentials:     txn.PaymentData.Object.Credentials,
		FullName:        fullName,
		CustomerAddress: txn.Customer.Address,
		Direction:       DirectionSell,
		GateId:          a.channelParams.Id,
		WebhookUrl:      "https://webhook.site/b8e353ac-c5b9-4938-863f-3d23a1285130",
	}

	return &request, nil

}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("dev")

	callbackBody, ok := txn.TxnInfo["callback"]
	if !ok {
		return nil, errors.New("callback body is missing")
	}

	callback := api.Callback{}
	err := json.Unmarshal([]byte(callbackBody), &callback)
	if err != nil {
		logger.Error("Error unmarshalling callback body - ", callbackBody)
		return nil, err
	}

	tr := &acquirer.TransactionStatus{}

	Key, err := a.api.MakeSignatureKey(ctx)
	Sign := a.api.CreateSign(callback.Id, callback.Status, Key.SignatureKey)

	if callback.Signature != Sign {
		logger.Error("Invalid Callback - ", err)
		return tr, nil
	}

	return handleStatus(tr, callback.Status)
}

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
	switch status {
	case api.StatusApproved:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.StatusDeclined:
		tr.Status = acquirer.REJECTED
		return tr, nil
	case api.StatusPending:
		fallthrough
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
