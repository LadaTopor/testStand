package service

import "github.com/shopspring/decimal"

type Paylink struct {
	Id               string `json:"id"`
	UserRef          string `json:"user_ref"`
	Status           string `json:"status"`
	TimestampUpdated string `json:"timestamp_updated"`
	Amount           string `json:"amount"`
	Sign             string `json:"sign"`
}

type Auris struct {
	TimeStamp  int64           `json:"timeStamp"`
	ShopID     int             `json:"shopID"`
	Type       string          `json:"type"`
	ID         int             `json:"id"`
	CurrID     int             `json:"currID"`
	Curr       string          `json:"curr"`
	InitAmount decimal.Decimal `json:"initAmount,omitempty"`
	Amount     decimal.Decimal `json:"amount"`
	Label      string          `json:"label"`
	UserID     string          `json:"userID"`
	Memo       string          `json:"memo"`
	Info       string          `json:"info,omitempty"`
	Status     int             `json:"status"`
	StatusText string          `json:"statusText"`
	StatusInfo string          `json:"statusInfo,omitempty"`
	Way        int             `json:"way"`
	Attempts   int             `json:"attempts,omitempty"`
	Sign       string          `json:"sign"`
}

type Sequoia struct {
	OrderId     string          `json:"order_id"`
	Date        string          `json:"date"`
	Amount      decimal.Decimal `json:"amount"`
	NewAmount   string          `json:"new_amount"`
	CardNumber  string          `json:"card_number"`
	PaymentType int             `json:"payment_type"`
	Status      string          `json:"status"`
}
