package api

import (
	"crypto/sha256"
	"fmt"
	"testStand/internal/acquirer/helper"
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
	Payload    any      `json:"payload"`
}

type Response struct {
	Status string `json:"status"`
	Id     string `json:"id"`
	Error  string `json:"error"`
}

func createSign(req PayoutRequest, secretKey string) string {
	sum := helper.GenerateHash(sha256.New(), []byte(req.MerchantId+req.CardData.CardNumber+req.Amount+secretKey))
	return fmt.Sprintf("%x", sum)
}
