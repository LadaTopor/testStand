package api

import (
	"github.com/shopspring/decimal"
)

const P2Pmethod = 1

type Request struct {
	WithdrawId      string   `json:"withdraw_id"`
	MerchId         string   `json:"merchant"`
	Extra           string   `json:"extra"`
	Amount          string   `json:"amount"`
	Currency        string   `json:"currency"`
	NotificationUrl string   `json:"notification_url"`
	UserId          string   `json:"user_id,omitempty"`
	UserRef         string   `json:"user_ref,omitempty"`
	UserIp          string   `json:"user_ip,omitempty"`
	FinishUrl       string   `json:"finish_url"`
	CardData        CardData `json:"card_data"`
	Sign            string   `json:"signature"`
}

type CardData struct {
	CardNumber string `json:"card_number"`
}

type CardDetails struct {
	TargetCardNumber string `json:"target_card_number"`
	Holder           string `json:"holder"`
	BankName         string `json:"bank_name"`
	ValidTill        int    `json:"valid_till"`
}

type Response struct {
	OK     bool        `json:"ok"`
	Status string      `json:"status"`
	Id     string      `json:"id"`
	Data   CardDetails `json:"data"`
	//if error
	Message string `json:"message"`
	Code    string `json:"code"`
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
