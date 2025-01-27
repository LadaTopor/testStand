package aplex

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"testStand/internal/acquirer"
	"testStand/internal/acquirer/aplex/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
)

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	Login        string `json:"login"`
	Password     string `json:"password"`
	WebhookUrl   string `json:"webhook_url"`
	PayoutGateId string `json:"payout_gate_id"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams

	callbackUrl string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.Login, channelParams.Password, gatewayParams.Transport.Timeout),
		dbClient:      db,
		callbackUrl:   callbackUrl,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	requestData, err := a.fillPaymentRequest(txn)
	if err != nil {
		return nil, err
	}

	response, err := a.api.CreateOffer(ctx, requestData)
	if err != nil {
		return nil, err
	}

	if response.Status == "DECLINED" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Outputs: map[string]string{
			"credentials": response.PaymentMethod.Address,
			"bank":        response.PaymentMethod.Gate.Name,
			"description": response.PaymentMethod.Person,
		},
		GtwTxnId: &response.Id,
	}

	return handleStatus(tr, response.Status)
}

// fillPaymentRequest
func (a *Acquirer) fillPaymentRequest(txn *models.Transaction) (*api.Request, error) {
	request := &api.Request{
		FiatSymbol:   txn.TxnCurrencySrc,
		FiatAmount:   decimal.NewFromInt(txn.TxnAmountSrc),
		CustomerName: txn.Customer.FullName,
		Directions:   api.PaymentDirection,
		WebhookUrl:   a.channelParams.WebhookUrl,
		ExternalId:   strconv.FormatInt(txn.TxnId, 10),
	}

	return request, nil
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	requestData, err := a.fillPayoutRequest(txn)
	if err != nil {
		return nil, err
	}

	response, err := a.api.CreateOffer(ctx, requestData)
	if err != nil {
		return nil, err
	}

	if response.Status == "DECLINED" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		GtwTxnId: &response.Id,
	}

	return handleStatus(tr, response.Status)
}

// fillPayoutRequest
func (a *Acquirer) fillPayoutRequest(txn *models.Transaction) (*api.Request, error) {
	request := &api.Request{
		FiatSymbol:      txn.TxnCurrencySrc,
		FiatAmount:      decimal.NewFromInt(txn.TxnAmountSrc),
		CustomerName:    txn.Customer.FullName,
		Directions:      api.PayoutDirection,
		GateId:          a.channelParams.PayoutGateId,
		CustomerAddress: txn.PaymentData.Object.Credentials,
		WebhookUrl:      a.channelParams.WebhookUrl,
		ExternalId:      strconv.FormatInt(txn.TxnId, 10),
	}

	return request, nil
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

	isVerify, err := a.api.Sign(callback)
	if err != nil {
		return nil, err
	}

	if !isVerify {
		logger.Error("Invalid callback - ", callbackBody)
		return nil, errors.New("invalid callback")
	}

	tr := &acquirer.TransactionStatus{}

	return handleStatus(tr, callback.Status)
}

// handleStatus
func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
	switch status {
	case api.StatusReconciled:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.StatusDecline:
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
