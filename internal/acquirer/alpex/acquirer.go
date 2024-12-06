package alpex

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"testStand/internal/acquirer"
	"testStand/internal/acquirer/alpex/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/gommon/log"
)

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
}

type ChannelParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	GateId   string `json:"gate_id"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
	callbackUrl   string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(gatewayParams.Transport.BaseAddress, channelParams.Email, channelParams.Password),
		dbClient:      db,
		callbackUrl:   callbackUrl,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		FiatSymbol:   txn.TxnCurrencySrc,
		FiatAmount:   float64(txn.TxnAmountSrc),
		CustomerName: txn.Customer.FullName,
		Direction:    "BUY",
		ExternalId:   strconv.Itoa(int(txn.TxnId)),
		WebhookUrl:   "https://webhook.site/67d91371-72ef-4128-ba38-3a85afba084f",
	}

	response, err := a.api.MakeOffer(request)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" {
		return nil, errors.New("bad status")
	}

	return &acquirer.TransactionStatus{
		Status: acquirer.PENDING,
		Outputs: map[string]string{
			"credentials": response.PaymentMethod.Address,
			"bank":        response.PaymentMethod.Gate.Name,
			"description": response.PaymentMethod.Person,
		},
		GtwTxnId: &response.Id,
	}, nil
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		FiatSymbol:      txn.TxnCurrencySrc,
		FiatAmount:      float64(txn.TxnAmountSrc),
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: &txn.PaymentData.Object.Credentials,
		GateId:          &a.channelParams.GateId,
		Direction:       "SELL",
		ExternalId:      strconv.Itoa(int(txn.TxnId)),
		WebhookUrl:      "https://webhook.site/67d91371-72ef-4128-ba38-3a85afba084f",
	}

	response, err := a.api.MakeOffer(request)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" {
		return nil, errors.New("bad status")
	}

	return &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &response.Id,
	}, nil
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("dev")

	callbackBody, ok := txn.TxnInfo["callback"]
	if !ok {
		return nil, errors.New("callback body is missing")
	}

	callback := api.Response{}
	err := json.Unmarshal([]byte(callbackBody), &callback)
	if err != nil {
		logger.Error("Error unmarshalling callback body - ", callbackBody)
		return nil, err
	}

	signKey, err := a.api.GetSignatureKey()
	if err != nil {
		return nil, err
	}

	if *callback.Signature != *api.GenerateSignature(callback.Id, callback.Status, signKey) {
		logger.Error("Invalid callback - ", callbackBody)
		return nil, err
	}

	switch callback.Status {
	case "RELEASED":
		return &acquirer.TransactionStatus{
			Status: acquirer.APPROVED,
		}, nil
	case "CANCELED":
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
		}, nil
	default:
		return nil, errors.New("unexpected status")
	}
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
