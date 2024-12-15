package alpex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

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
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	GateId   string `json:"gate_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
	callbackUrl   string
}

const (
	BUY  = "BUY"
	SELL = "SELL"
)

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, channelParams.Email, channelParams.Password, gatewayParams.Transport.BaseAddress, gatewayParams.Transport.Timeout),
		dbClient:      db,
		callbackUrl:   "https://webhook.site/88d71697-ff27-49e8-8887-02faeeb1a166",
	}

}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		FiatAmount: txn.TxnAmountSrc,
		FiatSymbol: txn.TxnCurrencySrc,
		ExternalId: fmt.Sprintf("%d", txn.TxnId),
		WebhookUrl: a.callbackUrl,
		Direction:  BUY,
	}

	response, err := a.api.MakePayment(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" || len(response.Error) != 0 {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Outputs: map[string]string{
			"credentials": response.PaymentMethod.CustomerAddress,
			"bank":        response.PaymentMethod.Gate.Name,
			"description": response.PaymentMethod.CustomerName,
		},
		GtwTxnId: &response.Id,
	}

	return handleStatus(tr, response.Status)
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	request := &api.Request{
		FiatAmount:      txn.TxnAmountSrc,
		FiatSymbol:      txn.TxnCurrencySrc,
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.PaymentData.Object.Credentials,
		GateId:          a.channelParams.GateId,
		ExternalId:      fmt.Sprintf("%d", txn.TxnId),
		WebhookUrl:      a.callbackUrl,
		Direction:       SELL,
	}

	response, err := a.api.MakePayment(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Status != "PENDING" || len(response.Error) != 0 {
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

func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("dev")

	sign, err := a.api.GetSign()
	if err != nil {
		return nil, err
	}

	callbackBody, ok := txn.TxnInfo["callback"]

	if !ok {
		return nil, errors.New("callback body is missing")
	}

	callback := api.Callback{}
	err = json.Unmarshal([]byte(callbackBody), &callback)
	if err != nil {
		logger.Error("Error unmarshalling callback body - ", callbackBody)
		return nil, err
	}

	if callback.Sign != hex.EncodeToString(helper.GenerateHMAC(sha256.New, []byte(fmt.Sprintf("id=%s\nstatus=%s", callback.Id, callback.Status)), sign)) {
		logger.Error("Invalid callback - ", callbackBody)
		return nil, err
	}

	tr := &acquirer.TransactionStatus{}
	if len(callback.Description) != 0 {
		tr.Info = map[string]string{
			"ps_error_code": callback.Description,
		}
	}

	return handleStatus(tr, callback.Status)
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {

	switch status {
	case api.Reconciled:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.Decline:
		tr.Status = acquirer.REJECTED
		return tr, nil
	case api.Pending:
		fallthrough
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
