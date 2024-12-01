package asupay

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupay/api"
	"testStand/internal/acquirer/helper"
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
	MerchId   string `json:"merchant"`
	SecretKey string `json:"secret_key"`
	ApiKey    string `json:"api_key"`
}

type Acquirer struct {
	MerchId              *api.Request
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        *ChannelParams
	callbackUrl          string
	percentageDifference *decimal.Decimal
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	log.Println(channelParams.ApiKey)
	return &Acquirer{
		channelParams:        channelParams,
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, gatewayParams.Transport.Timeout),
		dbClient:             db,
		callbackUrl:          callbackUrl,
		percentageDifference: gatewayParams.PercentageDifference,
	}

}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		MerchId:         a.channelParams.MerchId,
		Extra:           txn.PayMethodId,
		Amount:          strconv.FormatInt(txn.TxnAmountSrc, 10),
		Currency:        txn.TxnCurrencySrc,
		NotificationUrl: a.callbackUrl,
		UserRef:         fmt.Sprintf("%d", txn.TxnId),
		WithdrawId:      fmt.Sprintf("%d", txn.TxnId),
		CardData: api.CardData{
			CardNumber: txn.PaymentData.Object.Credentials,
		},
	}
	secretKey := a.channelParams.SecretKey
	sign := api.CreateSign(request.MerchId + request.CardData.CardNumber + request.Amount + secretKey)
	fmt.Println("---------so: ", request.MerchId, request.CardData.CardNumber, request.Amount, secretKey)

	request.Sign = sign

	response, err := a.api.MakePayout(ctx, request, a.channelParams.SecretKey)
	if err != nil {
		return nil, err
	}

	fmt.Println("------status", response.Status)

	if response.Status != "success" {
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

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func handleFinalStatus(status string, newAmount decimal.Decimal, currency string) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
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

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
	switch status {
	case "success":
		tr.Status = acquirer.APPROVED
		return tr, nil
	case "decline":
		tr.Status = acquirer.REJECTED
		return tr, nil
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
