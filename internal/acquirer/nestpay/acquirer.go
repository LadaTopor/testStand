package nestpay

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

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
}

type ChannelParams struct {
	ClientId    string `json:"client_id"`
	Currency    string `json:"currency"`
	ApiName     string `json:"api_name"`
	ApiPassword string `json:"api_password"`
	StoreKey    string `json:"store_key"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(ctx, channelParams.StoreKey, gatewayParams.Transport.BaseAddress),
		dbClient:      db,
	}
}

func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	request := &api.CC5Request{
		OrderId:   fmt.Sprintf("%d", txn.TxnId),
		Name:      a.channelParams.ApiName,
		Password:  a.channelParams.ApiPassword,
		ClientId:  a.channelParams.ClientId,
		Currency:  a.channelParams.Currency,
		Type:      "Auth",
		Amount:    strconv.FormatInt(txn.TxnAmountSrc, 10),
		Pan:       txn.PaymentData.Object.Credentials,
		IPAddress: "193.243.172.214",
	}
	data := url.Values{
		"clientid":                        {request.ClientId},
		"storetype":                       {"3d_pay"},
		"trantype":                        {request.Type},
		"oid":                             {request.OrderId},
		"Amount":                          {request.Amount},
		"currency":                        {request.Currency},
		"pan":                             {request.Pan},
		"Ecom_Payment_Card_ExpDate_Year":  {txn.PaymentData.Object.ExpYear},
		"Ecom_Payment_Card_ExpDate_Month": {txn.PaymentData.Object.ExpMonth},
		"cv2":                             {txn.PaymentData.Object.Cvv},
		"encoding":                        {"utf-8"},
		"lang":                            {"en"},
		"hashAlgorithm":                   {"ver3"},
		"rnd":                             {GenerateRndNumb(20)},
	}
	response, err := a.api.MakePayment(ctx, request, data)
	if err != nil {
		return nil, err
	}

	if response.Response != api.Approve {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_message": response.ErrMsg,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		GtwTxnId: &response.OrderId,
	}

	return handleStatus(tr, response.Response)
}

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {

	switch status {
	case api.Approve:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.Error:
		tr.Status = acquirer.REJECTED
		return tr, nil
	default:
		tr.Status = acquirer.REJECTED
		return tr, nil
	}
}
func GenerateRndNumb(length int) string {
	if length > 20 || length <= 0 {
		length = 20
	}
	rand.Seed(time.Now().UnixNano())
	randomNumber := ""
	for i := 0; i < length; i++ {
		randomDigit := rand.Intn(10)
		randomNumber += strconv.Itoa(randomDigit)
	}
	return randomNumber
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
