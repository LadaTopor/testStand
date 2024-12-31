package nestpay

import (
	"context"
	"fmt"

	"testStand/internal/acquirer"
	"testStand/internal/acquirer/helper"
	"testStand/internal/acquirer/nestpay/api"
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
	ClientId    string `json:"client_id"`
	Currency    string `json:"currency"`
	ApiName     string `json:"api_name"`
	ApiPassword string `json:"api_password"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, gatewayParams.Transport.Timeout),
		dbClient:      db,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.CC5Request{
		OrderId:       fmt.Sprintf("%d", txn.TxnId),
		Name:          a.channelParams.ApiName,
		Password:      a.channelParams.ApiPassword,
		ClientId:      a.channelParams.ClientId,
		Type:          "Auth",
		Total:         txn.TxnAmountSrc,
		Currency:      txn.TxnCurrencySrc,
		StoreType:     "3d_pay",
		CardNumber:    txn.PaymentData.Object.Credentials,
		CardYear:      txn.PaymentData.Object.ExpYear,
		CardMonth:     txn.PaymentData.Object.ExpMonth,
		CVV:           txn.PaymentData.Object.Cvv,
		HashAlgorithm: "ver3",
	}

	response, err := a.api.MakePayment(ctx, request)
	if err != nil {
		return nil, err
	}

	if response.Response != api.Approve {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.ErrMsg,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		GtwTxnId: &response.OrderId,
	}

	return handleStatus(tr, response.Response)
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
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

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {

	switch status {
	case api.Approve:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.Error:
		fallthrough
	default:
		tr.Status = acquirer.REJECTED
		return tr, nil
	}
}
