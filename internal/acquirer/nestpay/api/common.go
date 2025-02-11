package api

import (
	"crypto/sha512"
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
	"testStand/internal/acquirer/helper"
)

const (
	StatusApproved = "Approved"
	StatusDeclined = "Declined"
	StatusError    = "Error"
)

type CC5Request struct {
	ApiName     string `xml:"Name"`
	ApiPassword string `xml:"Password"`
	ClientId    string `xml:"ClientId"`
	Currency    string `xml:"Currency"`

	TxnId string `xml:"OrderId"`

	TransType  string `xml:"Type"`
	TransId    string `xml:"TransId"`
	Amount     int64  `xml:"Total"`
	CardNumber string `xml:"Number"`
	CVV        string `xml:"Cvv2Val"`
	ExpYear    string `xml:"ExpYear"`
	ExpMonth   string `xml:"ExpMonth"`

	// Payer
	PayerXID  string `xml:"PayerTxnId"`
	PayerCAVV string `xml:"PayerAuthenticationCode"`
	PayerECI  string `xml:"PayerSecurityLevel"`
}
type CC5Response struct {
	TxnId          string `xml:"OrderId"`
	GroupId        string `xml:"GroupId"`
	ResponseStatus string `xml:"Response"`
	AuthCode       string `xml:"AuthCode"`
	HostRefNum     string `xml:"HostRefNum"`
	ProcReturnCode string `xml:"ProcReturnCode"`
	TransId        string `xml:"TransId"`
	ErrMsg         string `xml:"ErrMsg"`
	Extra          struct {
		SettleId   string `xml:"SETTLEID"`
		TransDate  string `xml:"TRXDATE"`
		ErrCode    string `xml:"ERRORCODE"`
		CardNumber string `xml:"HOSTMSG"`
		EndErrCode string `xml:"NUMCODE"`
	} `xml:"Extra"`
}

func createHash(params url.Values, storeKey string) string {
	keys := make([]string, 0, len(params))

	for key := range params {
		if key == "encoding" || key == "hash" {
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)
	sort.Slice(keys, func(i, j int) bool { return strings.ToLower(keys[i]) < strings.ToLower(keys[j]) })

	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, params.Get(key))
	}
	sum := helper.GenerateHash(sha512.New(), []byte(strings.Join(values, "|")+"|"+storeKey))
	return base64.StdEncoding.EncodeToString(sum)
}
