package sequoia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/helper"
	"testStand/internal/acquirer/sequoia/api"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type GatewayParams struct {
	Transport            Transport        `json:"transport"`
	PercentageDifference *decimal.Decimal `json:"percentage_difference"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	SecretKey      string `json:"secret_key"`
	CallbackSecret string `json:"callback_secret"`
}

type Acquirer struct {
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        *ChannelParams
	callbackUrl          string
	percentageDifference *decimal.Decimal
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams:        channelParams,
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.SecretKey, gatewayParams.Transport.Timeout),
		dbClient:             db,
		callbackUrl:          callbackUrl,
		percentageDifference: gatewayParams.PercentageDifference,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestBody := &api.PaymentRequest{
		OrderID:       fmt.Sprintf("%d", txn.TxnId),
		Amount:        decimal.NewFromInt(txn.TxnAmountSrc),
		PaymentMethod: api.P2Pmethod,
		Currency:      txn.TxnCurrencySrc,
		CallbackURL:   a.callbackUrl,
	}
	requestBody.Token = generateHash(requestBody.OrderID, a.channelParams.SecretKey)

	response, err := a.api.MakePayment(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	if response.Status != "ok" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code":    response.Code,
				"ps_error_message": response.Message,
			},
		}, nil
	}

	outputs := map[string]string{"credentials": response.Data.TargetCardNumber}
	if len(response.Data.BankName) != 0 {
		outputs["bank"] = response.Data.BankName
	}
	if len(response.Data.Holder) != 0 {
		outputs["description"] = response.Data.Holder
	}

	return &acquirer.TransactionStatus{
		Status:  acquirer.PENDING,
		Outputs: outputs,
	}, nil
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
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

	newAmount, err := a.convertAmount(ctx, callback.Amount, callback.NewAmount, txn.TxnCurrencySrc)
	if err != nil {
		return nil, err
	}

	return handleFinalStatus(callback.Status, newAmount, txn.TxnCurrencySrc)
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	rq := &api.PaymentRequest{
		OrderID: fmt.Sprintf("%d", txn.TxnId),
	}
	rq.Token = generateHash(rq.OrderID, a.channelParams.SecretKey)

	response, err := a.api.CheckStatus(ctx, rq)
	if err != nil {
		return nil, err
	}

	newAmount, err := a.convertAmount(ctx, response.Data.Amount, response.Data.NewAmount, response.Data.Currency)
	if err != nil {
		return nil, err
	}

	return handleFinalStatus(response.Data.Status, newAmount, response.Data.Currency)
}

func handleFinalStatus(status string, newAmount decimal.Decimal, currency string) (*acquirer.TransactionStatus, error) {
	tr := &acquirer.TransactionStatus{}

	if !newAmount.IsZero() {
		tr.ConvertedAmount = &acquirer.TransactionAmount{
			Amount:   newAmount.IntPart(),
			Currency: currency,
		}
	}

	switch status {
	case "success":
		tr.Status = acquirer.APPROVED
	case "fail", "expired":
		tr.Status = acquirer.REJECTED
		tr.Info = map[string]string{
			"ps_error_message": status,
		}
	case "pending":
		fallthrough
	default:
		tr.Status = acquirer.PENDING
	}
	return tr, nil
}

func (a *Acquirer) convertAmount(ctx context.Context, amount decimal.Decimal, newAmount, currencySrc string) (decimal.Decimal, error) {
	logger, _ := zap.NewDevelopment()

	if len(newAmount) != 0 {
		newAmount, err := decimal.NewFromString(newAmount)
		if err != nil {
			return decimal.Decimal{}, err
		}

		if a.percentageDifference != nil {
			delta := amount.Mul(a.percentageDifference.Div(decimal.NewFromInt(100)))

			if amount.Sub(newAmount).Abs().GreaterThanOrEqual(delta) {
				logger.Warn("Skip callback: new amount is too low", zap.String("new_amount", newAmount.String()))
				return decimal.Decimal{}, errors.New("callback amount is too low")
			}
		}

		return newAmount, nil
	}
	return decimal.Decimal{}, nil
}

func generateHash(txnId, secretKey string) string {
	return helper.GenerateMD5Hash(fmt.Sprintf("%s%s", txnId, secretKey))
}
