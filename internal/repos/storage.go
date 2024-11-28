package repos

import (
	"encoding/json"
	"errors"

	"github.com/shopspring/decimal"
)

var ErrGtwNotFound = errors.New("gateway not found")
var ErrChnNotFound = errors.New("channel not found")

type Currency struct {
	Code     string
	NumCode  string
	Name     string
	Exponent int32
}

// DecimalAmount
func (c *Currency) DecimalAmount(amount int64) decimal.Decimal {
	if c == nil {
		panic("nil currency")
	}

	return decimal.New(amount, -c.Exponent)
}

// Int64AmountFromString
func (c *Currency) Int64AmountFromString(amount string) (int64, error) {
	if c == nil {
		return 0, errors.New("currency is nil")
	}

	originalDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		return 0, err
	}
	if c.Exponent == 0 {
		return originalDecimal.IntPart(), nil
	}
	finalDecimal, err := decimal.NewFromString(originalDecimal.StringFixed(c.Exponent))
	if err != nil {
		return 0, err
	}

	return finalDecimal.CoefficientInt64(), nil
}

type Gateway struct {
	Adapter    string
	ParamsJson string
}

type Channel struct {
	Id       int32
	Name     string
	IsActive bool
	GtwId    int32
	Params   Params
}

type Params struct {
	Credentials json.RawMessage `json:"credentials"`
}
