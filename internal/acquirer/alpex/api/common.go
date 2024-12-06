package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Request struct {
	FiatSymbol      string  `json:"fiat_symbol"`
	FiatAmount      float64 `json:"fiat_amount"`
	CustomerName    string  `json:"customer_name"`
	CustomerAddress *string `json:"customer_address"`
	Direction       string  `json:"direction"`
	GateId          *string `json:"gate_id"`
	ExternalId      string  `json:"external_id"`
	WebhookUrl      string  `json:"webhook_url"`
}

type Response struct {
	Id              string         `json:"_id"`
	PaymentMethod   *PaymentMethod `json:"payment_method,omitempty"`
	Direction       string         `json:"direction"`
	Amount          float64        `json:"amount"`
	AmountFiat      float64        `json:"amount_fiat"`
	Status          string         `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	OverdueAt       time.Time      `json:"overdue_at"`
	StatusChangedAt time.Time      `json:"status_changed_at"`
	Fee             int            `json:"fee"`
	FeeExternal     int            `json:"fee_external"`
	ApproveCode     string         `json:"approve_code"`
	TimeDiffSeconds int            `json:"time_diff_seconds"`
	IsExternal      bool           `json:"is_external"`
	BeneficiaryName string         `json:"beneficiary_name"`
	UnitCost        int            `json:"unit_cost"`
	Signature       *string        `json:"signature"`
	ExternalId      string         `json:"external_id"`

	Error string `json:"error"`
}

type PaymentMethod struct {
	Id          string `json:"_id"`
	Gate        Gate   `json:"gate"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Person      string `json:"person"`
	IsTemporary bool   `json:"is_temporary"`
}

type Gate struct {
	Id   string `json:"_id"`
	Name string `json:"name"`
}

func GenerateSignature(id, status, signatureKey string) *string {
	offerString := fmt.Sprintf("id=%s\nstatus=%s", id, status)

	h := hmac.New(sha256.New, []byte(signatureKey))
	h.Write([]byte(offerString))

	signature := hex.EncodeToString(h.Sum(nil))
	return &signature
}
