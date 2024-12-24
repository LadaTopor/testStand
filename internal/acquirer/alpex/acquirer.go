package alpex

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	WebhookUrl  string `json:"webhook_url"`
	BaseAddress string `json:"base_address"`
}

type ChannelParams struct {
	GateId string `json:"gate_id"`
	Login  api.Login
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
	callbackUrl   string
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {

	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, channelParams.Login, gatewayParams.Transport.BaseAddress),
		dbClient:      db,
		callbackUrl:   gatewayParams.Transport.WebhookUrl,
	}
}

func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		FiatAmount: txn.TxnAmountSrc,
		FiatSymbol: txn.TxnCurrencySrc,
		ExternalId: strconv.FormatInt(txn.TxnId, 10),
		WebhookUrl: a.callbackUrl,
		Direction:  api.BUY,
	}

	response, err := a.api.MakePayment(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Status != api.Pending {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code":    response.Code,
				"ps_error_message": response.Message,
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

func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	request := &api.Request{
		FiatAmount:      txn.TxnAmountSrc,
		FiatSymbol:      txn.TxnCurrencySrc,
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.PaymentData.Object.Credentials,
		GateId:          a.channelParams.GateId,
		ExternalId:      strconv.FormatInt(txn.TxnId, 10),
		WebhookUrl:      a.callbackUrl,
		Direction:       api.SELL,
	}

	response, err := a.api.MakePayment(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Status != api.Pending {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Code,
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

	sign, err := a.api.Signature()
	if err != nil {
		return nil, err
	}

	callbackBody, ok := txn.TxnInfo["callback"]
	if !ok || callbackBody == "" {
		logger.Error("callbackBody is missing or empty")
		return nil, errors.New("callback body is missing or empty")
	}

	callback := api.Callback{}
	err = json.Unmarshal([]byte(callbackBody), &callback)
	if err != nil {
		logger.Error("Error unmarshalling callback body - ", callbackBody)
		return nil, err
	}

	dataToSign := fmt.Sprintf("id=%s\nstatus=%s", callback.Id, callback.Status)
	expectedSignature := hex.EncodeToString(helper.GenerateHMAC(sha256.New, []byte(dataToSign), sign))
	if callback.Id == "" {
		logger.Error("Callback ID is missing")
		return nil, errors.New("callback ID is missing")
	}

	if callback.Signature != expectedSignature {
		logger.Error(fmt.Sprintf("Invalid signature. Expected: %s, Got: %s", expectedSignature, callback.Signature))
		return nil, errors.New("invalid callback signature")
	}

	tr := &acquirer.TransactionStatus{}
	if len(callback.Description) != 0 {
		tr.Info = map[string]string{
			"ps_error_code": callback.Description,
		}
	}
	logger.Info(fmt.Sprintf("Processing callback status: %s", callback.Status))
	return handleStatus(tr, callback.Status)
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
		tr.Status = acquirer.PENDING
		return tr, nil
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
