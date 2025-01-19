package api

import (
	"crypto/sha256"
	"encoding/hex"
)

type CardData struct {
	CardNumber string `json:"card_number"`
}

type Request struct {
	Merchant   string   `json:"merchant"`
	WithdrawId string   `json:"withdraw_id"`
	CardData   CardData `json:"card_data"`
	Amount     string   `json:"amount"`
	Signature  string   `json:"signature"`
}

type Response struct {
	Status string `json:"status"`
	Id     string `json:"uuid"`
	Error  string `json:"error"`
}

func CreateSign(request *Request, key string) {
	hashString := request.Merchant + request.CardData.CardNumber + request.Amount + key
	sign := sha256.Sum256([]byte(hashString))
	request.Signature = hex.EncodeToString(sign[:])
}
