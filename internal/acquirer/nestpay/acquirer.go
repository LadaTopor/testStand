package nestpay

import (
	"context"
	"strconv"

	"testStand/internal/acquirer"
	"testStand/internal/acquirer/helper"
	"testStand/internal/acquirer/nestpay/api"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/shopspring/decimal"
)

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport            Transport        `json:"transport"`
	PercentageDifference *decimal.Decimal `json:"percentage_difference"`
}

type ChannelParams struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	StoreKey string `json:"store_key"`
}

type Acquirer struct {
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        ChannelParams
	percentageDifference *decimal.Decimal
	callbackUrl          string
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.Name, channelParams.Password, channelParams.StoreKey),
		dbClient:             db,
		channelParams:        channelParams,
		percentageDifference: gatewayParams.PercentageDifference,
		callbackUrl:          callbackUrl,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestBody := &api.Request{
		ClientId:                        a.channelParams.Id,
		StoreType:                       "3d_pay",
		Amount:                          txn.TxnAmountSrc,
		Currency:                        txn.TxnCurrencySrc,
		TranType:                        "Auth",
		Lang:                            "tr",
		Rnd:                             "randomString",
		Pan:                             txn.PaymentData.Object.Credentials,
		Ecom_Payment_Card_ExpDate_Month: txn.PaymentData.Object.ExpMonth,
		Ecom_Payment_Card_ExpDate_Year:  txn.PaymentData.Object.ExpYear,
		Cv2:                             txn.PaymentData.Object.Cvv,
		Oid:                             strconv.FormatInt(txn.TxnId, 10),
		HashAlgorithm:                   "ver3",
		Encoding:                        "utf-8",
		Hash:                            "",
	}

	hash := a.api.CreateSign(requestBody, a.channelParams.StoreKey)
	requestBody.Hash = hash

	paymentResponse, err := a.api.MakePayment(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	statusRequest := &api.StatusRequest{
		Name:                    a.channelParams.Name,
		Password:                a.channelParams.Password,
		ClientId:                a.channelParams.Id,
		IpAddress:               paymentResponse.ClientIP,
		Oid:                     paymentResponse.Oid,
		Type:                    "Auth",
		Number:                  paymentResponse.Md,
		Amount:                  strconv.FormatInt(txn.TxnAmountSrc, 10),
		Currency:                txn.TxnCurrencySrc,
		PayerTxnId:              paymentResponse.XId,
		PayerSecurityLevel:      paymentResponse.Eci,
		PayerAuthenticationCode: paymentResponse.Cavv,
	}

	statusResponse, err := a.api.CheckStatus(ctx, statusRequest)
	if err != nil {
		return nil, err
	}

	switch statusResponse.Response {
	case api.StatusApproved:
		return &acquirer.TransactionStatus{
			Status: acquirer.APPROVED,
		}, nil
	case api.StatusDeclined:
		status := &acquirer.TransactionStatus{Status: acquirer.REJECTED}
		if len(statusResponse.ErrMsg) > 0 {
			status.Info = map[string]string{
				"ps_error_message": helper.DecodeUnicode(statusResponse.ErrMsg),
			}
		}
		return status, nil
	case api.StatusError:
		return &acquirer.TransactionStatus{
			Status:   acquirer.UNSPECIFIED,
			TxnError: acquirer.NewTxnError(5003, statusResponse.ErrMsg),
		}, nil
	default:
		return &acquirer.TransactionStatus{
			Status: acquirer.PENDING,
		}, nil
	}
}

func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
