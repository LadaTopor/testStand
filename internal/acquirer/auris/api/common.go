package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testStand/internal/acquirer/helper"

	"github.com/jeremywohl/flatten"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/maps"
)

// Statuses
const (
	StatusPending   = 0
	StatusConfirmed = 2
	StatusApproved  = 3
	StatusExpired   = 4
	StatusCancelled = 5
)

type Request struct {
	ShopID    int    `json:"shopID"`
	UniqID    string `json:"uniqID"`
	CurrID    int    `json:"currID"`
	Amount    int64  `json:"amount"`
	Label     string `json:"label"`
	UserID    string `json:"userID"`
	Memo      string `json:"memo"`
	BankCode  string `json:"bank_code,omitempty"`
	Number    string `json:"number,omitempty"`
	StatusURL string `json:"statusURL,omitempty"`
	Sign      string `json:"sign"`
	PreferId  int    `json:"prefer,omitempty"`
	ExtraInfo string `json:"info,omitempty"`
}

type StatusRequest struct {
	ShopID int    `json:"shopID"`
	ID     string `json:"id"`
	Sign   string `json:"sign"`
}

type Response struct {
	ID          int             `json:"id"`
	CurrID      int             `json:"currID"`
	Curr        string          `json:"curr"`
	Amount      decimal.Decimal `json:"amount"`
	Number      string          `json:"number"`
	Info        string          `json:"info"`
	BankTitle   string          `json:"bankTitle"`
	BankID      string          `json:"bankID"`
	Bank        string          `json:"bank"`
	Page        string          `json:"page"`
	Card        string          `json:"card"`
	FIO         string          `json:"fio"`
	Holder      string          `json:"holder"`
	Error       string          `json:"error"`
	Nspk        string          `json:"nspk"`
	PaymentLink string          `json:"paymentLink"`
}

type StatusResponse struct {
	Type       string          `json:"type"`
	ID         int             `json:"Id"`
	CurrID     int             `json:"currID"`
	Amount     decimal.Decimal `json:"amount"`
	Label      string          `json:"label"`
	Memo       string          `json:"memo"`
	Status     int             `json:"status"`
	StatusText string          `json:"statusText"`
	Error      string          `json:"error"`
}

type Callback struct {
	Version    string          `json:"mychanger.icu"`
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
	Error      string          `json:"error"`
	Sign       string          `json:"sign"`
}

type UserInfo struct {
	Name       string `json:"name,omitempty"`
	Surname    string `json:"surname,omitempty"`
	Patronomic string `json:"patronomic,omitempty"`
}

func createSign(s any, apiKey, decimalAmount string) (string, error) {
	signString, err := getSignaturePayload(s, decimalAmount)
	if err != nil {
		return "", err
	}
	signString += fmt.Sprintf(":%s", apiKey)

	sign := fmt.Sprintf("%x", helper.GenerateSHA1Hash(signString))
	return sign, nil
}

func getSignaturePayload(request any, decimalAmount string) (string, error) {
	fieldsMap, ok := request.(map[string]any)
	if !ok {
		rBytes, err := json.Marshal(request)
		if err != nil {
			return "", err
		}
		json.Unmarshal(rBytes, &fieldsMap)
	}

	// after unmarshal amount is float64
	// we keep as decimal string
	fieldsMap["amount"] = decimalAmount

	flatFields, err := flatten.Flatten(fieldsMap, "", flatten.DotStyle)
	if err != nil {
		return "", err
	}

	if len(flatFields) == 0 {
		return "", errors.New("signature payload is empty")
	}

	names := maps.Keys(flatFields)
	names = slices.DeleteFunc(names, func(s string) bool { return s == "sign" })

	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})

	resultBuilder := strings.Builder{}
	for i, v := range names {
		if v == "sign" {
			continue
		}

		fieldValue := ""
		if flatFields[v] != nil {
			fieldValue = fmt.Sprintf("%v", flatFields[v])
		}
		resultBuilder.WriteString(fieldValue)
		if i != len(names)-1 {
			resultBuilder.WriteString(":")
		}
	}
	return resultBuilder.String(), nil
}
