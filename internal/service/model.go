package service

import (
	"testStand/internal/models"
)

type Request struct {
	Id          int64              `json:"id"`
	Customer    models.Customer    `json:"customer"`
	PaymentData models.PaymentData `json:"payment_data"`
	Amount      models.Amount      `json:"amount"`
	GtwName     string             `json:"gtw_name"`
	ChnName     string             `json:"chn_name"`
}

type Response struct {
	TxnId     int64   `json:"txn_id"`
	TxnStatus string  `json:"txn_status"`
	Result    *Result `json:"result,omitempty"`
}

type Result struct {
	Credentials string `json:"credentials"`
	Bank        string `json:"bank"`
	Description string `json:"description"`
}
