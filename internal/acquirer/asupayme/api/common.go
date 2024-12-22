package api

import (
	"crypto/sha256"
	"encoding/hex"
)

type Request struct {
	Merchant   string   `json:"merchant"`
	WithdrawId string   `json:"withdraw_id"`
	Amount     string   `json:"amount"`
	Sign       string   `json:"signature"`
	CardData   CardData `json:"card_data"`
}

type Response struct {
	Status  string `json:"status"`
	Id      string `json:"id"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type CardData struct {
	OwnerName    string `json:"owner_name"`
	CardNumber   string `json:"card_number"`
	ExpiredMonth string `json:"expired_month"`
	ExpiredYear  string `json:"expired_year"`
}

func createSign(request *Request, key string) string {
	hashString := request.Merchant + request.CardData.CardNumber + request.Amount + key
	sum := sha256.Sum256([]byte(hashString))
	sign := hex.EncodeToString(sum[:])
	return sign
}
