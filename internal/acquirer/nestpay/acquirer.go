package nestpay

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
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
	requestData, authParams, err := a.fillPaymentRequest(ctx, txn)
	if err != nil {
		return nil, err
	}

	authResponse, err := a.api.MakeAuth(ctx, authParams)
	if err != nil {
		return nil, err
	}

	respErr := handleAuthResponse(authResponse)
	if respErr != nil {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": respErr.Error(),
			},
		}, nil
	}

	requestData.PayerCAVV = authResponse.Get("cavv")
	requestData.PayerECI = authResponse.Get("eci")
	requestData.PayerXID = authResponse.Get("xid")
	requestData.CardNumber = authResponse.Get("md")
	requestData.TransId = authResponse.Get("TransID")

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
func (a *Acquirer) fillPaymentRequest(ctx context.Context, txn *models.Transaction) (*api.CC5Request, url.Values, error) {
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
	authParams := url.Values{
		"clientid":                        {request.ClientId},
		"storetype":                       {"3d_pay"},
		"oid":                             {request.TxnId},
		"trantype":                        {request.TransType},
		"amount":                          {strconv.Itoa(int(request.Amount))},
		"currency":                        {a.channelParams.Currency},
		"lang":                            {"en"},
		"rnd":                             {randomString(10)},
		"pan":                             {request.CardNumber},
		"Ecom_Payment_Card_ExpDate_Year":  {request.ExpYear},
		"Ecom_Payment_Card_ExpDate_Month": {request.ExpMonth},
		"cv2":                             {request.CVV},
		"encoding":                        {"utf-8"},
		"hashAlgorithm":                   {"ver3"},
	}
	hash := api.CreateHash(authParams, a.channelParams.StoreKey)
	if len(hash) == 0 {
		return nil, nil, errors.New("hash is empty")
	}
	authParams.Add("hash", hash)

	return request, authParams, nil
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

func handleAuthResponse(authResponse url.Values) error {
	if len(authResponse.Get("ErrMsg")) != 0 {
		return errors.New(authResponse.Get("ErrMsg"))
	}

	status, err := strconv.Atoi(authResponse.Get("mdStatus"))
	if err != nil {
		return err
	}

	if status < api.AuthStatusSuccessFull || status > api.AuthStatusOkHalf {
		switch status {
		case api.AuthStatusRejected, api.AuthStatusFailed:
			return errors.New("3d secure authentication failed")
		case api.AuthStatusNotAvailable, api.AuthStatusError, api.AuthStatusSystemError:
			return errors.New("mpi fallback")
		default:
			return errors.New("status is not full or half 3d secure")
		}
	}
	return nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
