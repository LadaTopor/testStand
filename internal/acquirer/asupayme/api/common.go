package api

import (
	"crypto/sha256"
	"fmt"
	"testStand/internal/acquirer/helper"
)

const (
	Pending    = "new"
	Reconciled = "executed"
	Decline    = "cancelled"
)

type CardData struct {
	Owner      string `json:"owner_name"`
	CardNumber string `json:"card_number"`
	ExpMonth   string `json:"expired_month"`
	ExpYear    string `json:"expired_year"`
}

type PayoutRequest struct {
	MerchantId string   `json:"merchant"`
	WithdrawId string   `json:"withdraw_id"`
	CardData   CardData `json:"card_data"`
	Amount     string   `json:"amount"`
	Sign       string   `json:"signature"`
	ApiKey     string   `json:"api_key"`
	Payload    any      `json:"payload"`
}

type Response struct {
	Status string `json:"status"`
	Id     string `json:"id"`
	Error  string `json:"error"`
}

type Callback struct {
	Id               string `json:"id"`
	UserRef          string `json:"user_ref"`
	Status           string `json:"status"`
	Description      string `json:"description"`
	TimestampUpdated string `json:"timestamp_updated"`
	Amount           string `json:"amount"`
	Sign             string `json:"sign"`
}

type StatusRequest struct {
	Id      string `json:"id"`
	MerchId string `json:"merch_id"`
	UserRef string `json:"user_ref,omitempty"`
}

func createSign(input string) string {
	sum := helper.GenerateHash(sha256.New(), []byte(input))
	return fmt.Sprintf("%x", sum)
}
