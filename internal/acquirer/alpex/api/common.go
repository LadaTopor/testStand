package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	Pending    = "PENDING"
	Reconciled = "RELEASED"
	Decline    = "DECLINED"
)

const (
	DirectPayout  = "SELL"
	DirectPayment = "BUY"
)

type Request struct {
	FiatSymbol      string          `json:"fiat_symbol,omitempty"`
	FiatAmount      decimal.Decimal `json:"fiat_amount,omitempty"`
	CustomerName    string          `json:"customer_name,omitempty"`
	Directions      string          `json:"direction,omitempty"`
	GateId          string          `json:"gate_id,omitempty"`
	CustomerAddress string          `json:"customer_address,omitempty"`
	WebhookUrl      string          `json:"webhook_url,omitempty"`
	ExternalId      string          `json:"external_id,omitempty"`
}

type Gate struct {
	Name string `json:"name"`
}

type PaymentMethod struct {
	Gate    *Gate  `json:"gate"`
	Address string `json:"address"`
	Person  string `json:"person"`
}

type Response struct {
	Error         string         `json:"error"`
	PaymentMethod *PaymentMethod `json:"payment_method"`
	ExternalId    string         `json:"external_id"`
	Id            string         `json:"_id"`
	Status        string         `json:"status"`
}

type Callback struct {
	Id        string `json:"_id"`
	Status    string `json:"status"`
	Signature string `json:"signature"`
}

func CreateSign(id, status, signatureKey string) string {
	offerString := fmt.Sprintf("id=%s\nstatus=%s", id, status)
	hashSign := hmac.New(sha256.New, []byte(signatureKey))
	hashSign.Write([]byte(offerString))

	return hex.EncodeToString(hashSign.Sum(nil))
}
