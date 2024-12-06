package service

import (
	"time"

	"github.com/shopspring/decimal"
)

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

type Alpex struct {
	Id            string `json:"_id"`
	PaymentMethod struct {
		Id   string `json:"_id"`
		Gate struct {
			Id   string `json:"_id"`
			Name string `json:"name"`
		} `json:"gate"`
		Name        string `json:"name"`
		Address     string `json:"address"`
		Person      string `json:"person"`
		IsTemporary bool   `json:"is_temporary"`
	} `json:"payment_method,omitempty"`
	Direction       string    `json:"direction"`
	Amount          float64   `json:"amount"`
	AmountFiat      float64   `json:"amount_fiat"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	OverdueAt       time.Time `json:"overdue_at"`
	StatusChangedAt time.Time `json:"status_changed_at"`
	Fee             int       `json:"fee"`
	FeeExternal     int       `json:"fee_external"`
	ApproveCode     string    `json:"approve_code"`
	TimeDiffSeconds int       `json:"time_diff_seconds"`
	IsExternal      bool      `json:"is_external"`
	BeneficiaryName string    `json:"beneficiary_name"`
	UnitCost        int       `json:"unit_cost"`
	Signature       *string   `json:"signature"`
	ExternalId      string    `json:"external_id"`

	Error string `json:"error"`
}
