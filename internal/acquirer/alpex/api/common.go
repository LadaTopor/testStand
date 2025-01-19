package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	StatusPending  = "PENDING"
	StatusApproved = "RELEASED"
	StatusDeclined = "DECLINED"
)

type Request struct {
	Id              string `json:"_id"`
	ExternalId      string `json:"external_id"`
	Symbol          string `json:"fiat_symbol"`
	Amount          int64  `json:"fiat_amount"`
	Status          string `json:"status"`
	FullName        string `json:"fullName"`
	CustomerAddress string `json:"customer_address"`
	Credentials     string `json:"credentials"`
	Direction       string `json:"direction"`
	GateId          string `json:"gate_id"`
	WebhookUrl      string `json:"webhook_url"`
	Sign            string `json:"sign"`
}

type Response struct {
	Id              string          `json:"_id"`
	ExternalId      string          `json:"external_id"`
	Direction       string          `json:"direction"`
	Amount          decimal.Decimal `json:"amount"`
	AmountFiat      decimal.Decimal `json:"amount_fiat"`
	Status          string          `json:"status"`
	ApproveCode     string          `json:"approve_code"`
	BeneficiaryName string          `json:"beneficiary_name"`
	Sign            string          `json:"sign"`
}

type PaymentResponse struct {
	Id            string          `json:"_id"`
	ExternalId    string          `json:"external_id"`
	PaymentMethod PaymentMethod   `json:"payment_method"`
	Amount        decimal.Decimal `json:"amount"`
	AmountFiat    decimal.Decimal `json:"amount_fiat"`
	Status        string          `json:"status"`
	ApproveCode   string          `json:"approve_code"`
	Sign          string          `json:"sign"`
}

type PaymentMethod struct {
	Address string `json:"address"`
	Person  string `json:"person"`
	Gate    Gate   `json:"gate"`
}

type Gate struct {
	Id   string `json:"_id"`
	Name string `json:"name"`
}

type Callback struct {
	Id     string `json:"_id"`
	Status string `json:"status"`
	Sign   string `json:"sign"`
}

type User struct {
	email    string `json:"email"`
	password string `json:"password"`
}

type UserSign struct {
	SignatureKey string `json:"signature_key"`
}

func createSign(request *Request, key string) string {
	offerString := fmt.Sprintf(`id=%d\nstatus=%s`, request.Id, request.Status) //до \ убрал >
	hmac := hmac.New(sha256.New, []byte(key))
	hmac.Write([]byte(offerString))
	sign := hex.EncodeToString(hmac.Sum(nil))

	return sign
}
