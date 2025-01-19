package asupayme

import (
	"context"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayme/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"
)

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	ApiKey    string `json:"api_key"`
	MerchId   string `json:"merch_id"`
	SecretKey string `json:"secret_key"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, gatewayParams.Transport.Timeout),
		dbClient:      db,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	request := &api.Request{
		Merchant:   a.channelParams.MerchId,
		WithdrawId: strconv.FormatInt(txn.TxnId, 10),
		CardData: api.CardData{
			CardNumber: txn.PaymentData.Object.Credentials,
		},
		Amount: strconv.FormatInt(txn.TxnAmountSrc, 10),
	}

	api.CreateSign(request, a.channelParams.SecretKey)

	response, err := a.api.MakePayout(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Status != "success" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		GtwTxnId: &response.Id,
		Status:   acquirer.APPROVED,
	}

	return tr, nil
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
