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

type gatewayMethod struct {
	Id          string         `json:"id"`
	GtwId       map[string]int `json:"gtw_id"`
	PreferId    int            `json:"prefer_id"`
	MapInputId  string         `json:"map_input_id"`
	MapOutputId string         `json:"map_output_id"`
}

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	GateId      string `json:"gate_id"`
	BaseAddress string `json:"base_address"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	ChannelParams *ChannelParams
	callbackUrl   string
}

const (
	BUY  = "BUY"
	SELL = "SELL"
)

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {

	return &Acquirer{
		ChannelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, api.GetApi(gatewayParams.Transport.BaseAddress), gatewayParams.Transport.Timeout),
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
		GateId:     a.ChannelParams.GateId,
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
		GateId:          a.ChannelParams.GateId,
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

	if callback.Sign != hex.EncodeToString(helper.GenerateHMAC(sha256.New, []byte(fmt.Sprintf("id=%s\nstatus=%s", callback.Id, callback.Status)), api.GetSign(a.ChannelParams.BaseAddress))) {
		logger.Error("Invalid callback - ", callbackBody)
		return nil, err
	}

	tr := &acquirer.TransactionStatus{}
	if len(callback.Description) != 0 {
		tr.Info = map[string]string{
			"ps_error_code": callback.Description,
		}
	}
	callback.Webhook_url = a.callbackUrl

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
