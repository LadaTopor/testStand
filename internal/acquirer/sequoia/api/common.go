package api

import (
	"github.com/shopspring/decimal"
)

const P2Pmethod = 1

type PaymentRequest struct {
	OrderID           string          `json:"order_id"`
	Amount            decimal.Decimal `json:"amount,omitempty"`
	PaymentMethod     int             `json:"payment_method,omitempty"`
	Token             string          `json:"token"`
	Currency          string          `json:"currency,omitempty"`
	CallbackURL       string          `json:"callback_url,omitempty"`
	BackToMerchantURL string          `json:"back_to_merchant_url,omitempty"`
}

type CardDetails struct {
	TargetCardNumber string `json:"target_card_number"`
	Holder           string `json:"holder"`
	BankName         string `json:"bank_name"`
	ValidTill        int    `json:"valid_till"`
}

type Response struct {
	Status string      `json:"status"`
	Data   CardDetails `json:"data"`
	//if error
	Message string `json:"message"`
	Code    string `json:"code"`
}

type Callback struct {
	OrderId     string          `json:"order_id"`
	Date        string          `json:"date"`
	Amount      decimal.Decimal `json:"amount"`
	NewAmount   string          `json:"new_amount"`
	CardNumber  string          `json:"card_number"`
	PaymentType int             `json:"payment_type"`
	Status      string          `json:"status"`
}

type Status struct {
	Status  string     `json:"status"`
	Data    StatusData `json:"data"`
	Message string     `json:"message"`
	Code    int        `json:"code"`
}

type StatusData struct {
	OrderID    string          `json:"order_id"`
	Status     string          `json:"status"`
	Currency   string          `json:"currency"`
	Amount     decimal.Decimal `json:"amount"`
	NewAmount  string          `json:"new_amount"`
	CardNumber string          `json:"card_number"`
	InternalID int             `json:"internal_id"`
}
