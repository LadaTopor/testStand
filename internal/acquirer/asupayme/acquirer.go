package asupayme

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayme/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"
)

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type ChannelParams struct {
	ApiKey     string `json:"api_key"`
	MerchantId int    `json:"merchant_id"`
	SecretKey  string `json:"secret_key"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams ChannelParams
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, channelParams.SecretKey, gatewayParams.Transport.Timeout),
		dbClient:      db,
		channelParams: channelParams,
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

	response, err := a.api.MakeWithdraw(ctx, *requestData, a.channelParams.ApiKey)
	if err != nil {
		return nil, err
	}

	if len(response.Error) != 0 {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Status:   acquirer.APPROVED,
		GtwTxnId: &response.Id,
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

// fillPayoutRequest
func (a *Acquirer) fillPayoutRequest(ctx context.Context, txn *models.Transaction) (*api.PayoutRequest, error) {
	fullName := txn.Customer.FullName
	if len(fullName) == 0 {
		return nil, errors.New("customer full name is required data")
	}

	request := api.PayoutRequest{
		MerchantId: strconv.Itoa(a.channelParams.MerchantId),
		WithdrawId: fmt.Sprintf("%d", txn.TxnId),
		Amount:     fmt.Sprintf("%d", txn.TxnAmountSrc),
		CardData: api.CardData{
			Owner:      fullName,
			CardNumber: txn.PaymentData.Object.Credentials,
			ExpMonth:   txn.PaymentData.Object.ExpMonth,
			ExpYear:    txn.PaymentData.Object.ExpYear,
		},
	}

	return &request, nil
}
