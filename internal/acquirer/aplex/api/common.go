package api

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/shopspring/decimal"

	"testStand/internal/acquirer/helper"
)

const (
	StatusPending    = "PENDING"
	StatusReconciled = "RELEASED"
	StatusDecline    = "DECLINED"
)

const (
	PayoutDirection  = "SELL"
	PaymentDirection = "BUY"
)

type Request struct {
	FiatSymbol      string          `json:"fiat_symbol"`
	FiatAmount      decimal.Decimal `json:"fiat_amount"`
	CustomerName    string          `json:"customer_name"`
	Directions      string          `json:"direction"`
	GateId          string          `json:"gate_id"`
	CustomerAddress string          `json:"customer_address"`
	WebhookUrl      string          `json:"webhook_url"`
	ExternalId      string          `json:"external_id"`
}

type Gate struct {
	Id           string `json:"_id"`
	Name         string `json:"name"`
	IsRandomPool bool   `json:"is_random_pool"`
}
type PaymentMethod struct {
	Id          string `json:"_id"`
	Gate        *Gate  `json:"gate"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Person      string `json:"person"`
	IsTemporary bool   `json:"is_temporary"`
}
type Response struct {
	Error         string         `json:"error"`
	Id            string         `json:"_id"`
	PaymentMethod *PaymentMethod `json:"payment_method"`
	Status        string         `json:"status"`
	CreatedAt     string         `json:"created_at"`
	ExternalId    string         `json:"external_id"`
	Signature     string         `json:"signature"`
}

type Callback struct {
	OfferId   string `json:"_id"`
	Status    string `json:"status"`
	Signature string `json:"signature"`
}

func createSign(callback *Callback, signKey string) string {
	offerString := "id=" + callback.OfferId + "\nstatus=" + callback.Status
	sum := helper.GenerateHash(sha256.New(), []byte(offerString+signKey))
	return hex.EncodeToString(sum)
}
