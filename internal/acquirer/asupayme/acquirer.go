package asupayme

import (
	"context"
	"fmt"
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
	ApiKey     string `json:"api_key"`
	SecretKey  string `json:"secret_key"`
	MerchantID string `json:"merchant_id"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, channelParams.MerchantID, channelParams.SecretKey),
		dbClient:      db,
	}
}

func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	cardData := txn.PaymentData.Object.Credentials
	signature := a.api.GenerateSignature(cardData, fmt.Sprintf("%d", txn.TxnAmountSrc))

	request := &api.WithdrawRequest{
		Merchant:   a.api.MerchantID,
		WithdrawID: fmt.Sprintf("txn-%d", txn.TxnId),
		CardData: api.CardData{
			OwnerName:    txn.Customer.FullName,
			CardNumber:   txn.PaymentData.Object.Credentials,
			ExpiredMonth: "12",
			ExpiredYear:  "26",
		},
		Amount:    fmt.Sprintf("%d", txn.TxnAmountSrc),
		Signature: signature,
	}

	response, err := a.api.MakeWithdraw(ctx, request)
	if err != nil {
		return nil, err
	}

	return &acquirer.TransactionStatus{
		Status: acquirer.APPROVED,
		Outputs: map[string]string{
			"withdraw_id": response.ID,
		},
	}, nil
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
