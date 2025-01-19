package nestpay

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

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
	ApiName     string `json:"api_name"`
	Currency    string `json:"currency"`
	ClientId    string `json:"client_id"`
	StoreKey    string `json:"store_key"`
	ApiPassword string `json:"api_password"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
	callbackUrl   string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, gatewayParams.Transport.BaseAddress, gatewayParams.Transport.Timeout),
		dbClient:      db,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	formData := url.Values{}
	formData.Set("clientid", a.channelParams.ClientId)
	formData.Set("oid", fmt.Sprintf("%d", txn.TxnId))
	formData.Set("trantype", "Auth")
	formData.Set("pan", txn.PaymentData.Object.Credentials)
	formData.Set("Ecom_Payment_Card_ExpDate_Year", txn.PaymentData.Object.ExpYear)
	formData.Set("Ecom_Payment_Card_ExpDate_Month", txn.PaymentData.Object.ExpMonth)
	formData.Set("cv2", txn.PaymentData.Object.Cvv)
	formData.Set("amount", fmt.Sprintf("%.2f", float64(txn.TxnAmountSrc)/100))
	formData.Set("currency", a.channelParams.Currency)
	formData.Set("encoding", "utf-8")
	formData.Set("storetype", "3d_pay")
	formData.Set("hashAlgorithm", "ver3")
	formData.Set("hash", api.GenerateHash(formData, a.channelParams.StoreKey))

	response, err := a.api.MakePayment(ctx, formData)
	if err != nil {
		return nil, err
	}

	request := &api.Request{
		Name:                    a.channelParams.ApiName,
		Password:                a.channelParams.ApiPassword,
		ClientId:                a.channelParams.ClientId,
		Oid:                     strconv.FormatInt(txn.TxnId, 10),
		Type:                    "Auth",
		Number:                  response.Get("md"),
		Amount:                  fmt.Sprintf("%.2f", float64(txn.TxnAmountSrc)/100),
		Currency:                a.channelParams.Currency,
		PayerTxnId:              response.Get("xid"),
		PayerSecurityLevel:      response.Get("eci"),
		PayerAuthenticationCode: response.Get("cavv"),
	}

	finishResp, err := a.api.MakeFinish3ds(ctx, request)

	if finishResp.Response != "Approve" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code":    finishResp.ProcReturnCode,
				"ps_error_message": finishResp.ErrMsg,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		GtwTxnId: &finishResp.OrderId,
	}

	return handleStatus(tr, finishResp.Response)
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
	case api.Reconciled:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.Decline:
		tr.Status = acquirer.REJECTED
		return tr, nil
	case api.Pending:
		fallthrough
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
