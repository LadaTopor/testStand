package asupay

import (
	"context"                                // встроенный пакет
	"fmt"                                    // встроенный пакет
	"log"                                    // встроенный пакет
	"strconv"                                // встроенный пакет
	"testStand/internal/acquirer"            // наш импорт
	"testStand/internal/acquirer/asupay/api" // наш импорт
	"testStand/internal/acquirer/helper"     // наш импорт
	"testStand/internal/models"              // наш импорт
	"testStand/internal/repos"               // наш импорт

	"github.com/shopspring/decimal" // внешний импорт
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
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
	callbackUrl   string
}

const (
	approved = "success"
	rejected = "decline"
)

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	log.Println(channelParams.ApiKey)
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, gatewayParams.Transport.Timeout),
		dbClient:      db,
		callbackUrl:   callbackUrl,
	}

}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.Request{
		MerchId:    a.channelParams.MerchId,
		Amount:     strconv.FormatInt(txn.TxnAmountSrc, 10),
		WithdrawId: fmt.Sprintf("%d", txn.TxnId),
		CardData: api.CardData{
			CardNumber: txn.PaymentData.Object.Credentials,
		},
	}
	secretKey := a.channelParams.SecretKey
	sign := api.CreateSign(request.MerchId + request.CardData.CardNumber + request.Amount + secretKey)

	request.Sign = sign

	response, err := a.api.MakePayout(ctx, request, a.channelParams.SecretKey)
	if err != nil {
		return nil, err
	}

	if response.Status != approved {
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

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
	switch status {
	case approved:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case rejected:
		tr.Status = acquirer.REJECTED
		return tr, nil
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
