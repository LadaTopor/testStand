package asupayme

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayme/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"
)

type gatewayMethod struct {
	Id          string         `json:"id"`
	GtwId       map[string]int `json:"gtw_id"`
	PreferId    int            `json:"prefer_id"`
	MapInputId  string         `json:"map_input_id"`
	MapOutputId string         `json:"map_output_id"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport            Transport        `json:"transport"`
	PayoutMethods        []gatewayMethod  `json:"payout_methods"`
	PercentageDifference *decimal.Decimal `json:"percentage_difference"`
}

type ChannelParams struct {
	ApiKey     string `json:"api_key"`
	MerchantId string `json:"merch_id"`
	SecretKey  string `json:"secret_key"`
}

type Acquirer struct {
	api      *api.Client
	dbClient *repos.Repo

	channelParams ChannelParams
	payoutMethods []gatewayMethod

	percentageDifference *decimal.Decimal
	callbackUrl          string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, channelParams.SecretKey),
		dbClient:             db,
		channelParams:        channelParams,
		percentageDifference: gatewayParams.PercentageDifference,
		callbackUrl:          callbackUrl,
		payoutMethods:        gatewayParams.PayoutMethods,
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

	if response.Status != "success" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Code,
			},
		}, nil
	}

	gtwTxnId := response.Id

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.APPROVED,
		GtwTxnId: &gtwTxnId,
	}

	return tr, nil
}

// fillPayoutRequest
func (a *Acquirer) fillPayoutRequest(ctx context.Context, txn *models.Transaction) (*api.Request, error) {

	request := api.Request{
		Merchant:   a.channelParams.MerchantId,
		WithdrawId: strconv.FormatInt(txn.TxnId, 10),
		Amount:     strconv.FormatInt(txn.TxnAmountSrc, 10),
		CardData: api.CardData{
			OwnerName:  txn.Customer.FullName,
			CardNumber: txn.PaymentData.Object.Credentials,
		},
	}

	if txn.TxnCurrencySrc != "TRY" {
		return &request, nil
	}

	fullName := txn.Customer.FullName
	if len(fullName) == 0 {
		return nil, errors.New("customer full name is required data")
	}

	request.CardData.OwnerName = fullName

	return &request, nil

}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
