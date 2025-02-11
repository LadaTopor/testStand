package nestpay

import (
	"context"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/helper"
	"testStand/internal/acquirer/nestpay/api"
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
	ClientId    string `json:"client_id"`
	Currency    string `json:"currency"`
	StoreKey    string `json:"store_key"`
	ApiName     string `json:"api_name"`
	ApiPassword string `json:"api_password"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams ChannelParams
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.StoreKey, channelParams.Currency, gatewayParams.Transport.Timeout),
		dbClient:      db,
		channelParams: channelParams,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	requestData, err := a.fillPaymentRequest(ctx, txn)
	if err != nil {
		return nil, err
	}

	response, err := a.api.MakePayment(ctx, requestData)
	if err != nil {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.ErrMsg,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		GtwTxnId: &response.TransId,
	}

	return handleStatus(tr, response.ResponseStatus)
}

// fillPaymentRequest
func (a *Acquirer) fillPaymentRequest(ctx context.Context, txn *models.Transaction) (*api.CC5Request, error) {
	request := &api.CC5Request{
		ApiName:     a.channelParams.ApiName,
		ApiPassword: a.channelParams.ApiPassword,
		ClientId:    a.channelParams.ClientId,
		Currency:    a.channelParams.Currency,
		TxnId:       strconv.Itoa(int(txn.TxnId)),
		TransType:   "Auth",
		Amount:      txn.TxnAmountSrc,
		CardNumber:  txn.PaymentData.Object.Credentials,
		CVV:         txn.PaymentData.Object.Cvv,
		ExpYear:     txn.PaymentData.Object.ExpYear,
		ExpMonth:    txn.PaymentData.Object.ExpMonth,
	}

	return request, nil
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
	case api.StatusApproved:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.StatusDeclined, api.StatusError:
		tr.Status = acquirer.REJECTED
		return tr, nil
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
