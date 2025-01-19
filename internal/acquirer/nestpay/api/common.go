package api

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/xml"
	"net/url"
	"sort"
	"strings"
)

const (
	Pending    = "new"
	Reconciled = "Approve"
	Decline    = "Error"
)

type Request struct {
	XMLName                 xml.Name `xml:"CC5Request"`
	Name                    string   `xml:"Name"`
	Password                string   `xml:"Password"`
	ClientId                string   `xml:"ClientId"`
	Oid                     string   `xml:"oid"`
	Type                    string   `xml:"Type"`
	Number                  string   `xml:"Number"`
	Amount                  string   `xml:"amount"`
	Currency                string   `xml:"Currency"`
	PayerTxnId              string   `xml:"PayerTxnId"`
	PayerSecurityLevel      string   `xml:"PayerSecurityLevel"`
	PayerAuthenticationCode string   `xml:"PayerAuthenticationCode"`
}

type Response struct {
	ErrMsg         string `xml:"ErrMsg"`
	ProcReturnCode string `xml:"ProcReturnCode"`
	Response       string `xml:"Response"`
	OrderId        string `xml:"OrderId"`
}

func GenerateHash(formData url.Values, storeKey string) string {
	keys := make([]string, 0, len(formData))
	for key := range formData {
		lowerKey := strings.ToLower(key)
		if lowerKey != "encoding" && lowerKey != "hash" && lowerKey != "countdown" {
			keys = append(keys, key)
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})

	var hashVal strings.Builder
	for _, key := range keys {
		value := formData.Get(key)
		escapedValue := strings.ReplaceAll(value, `\`, `\\`)
		escapedValue = strings.ReplaceAll(escapedValue, `|`, `\|`)
		hashVal.WriteString(escapedValue + "|")
	}

	escapedStoreKey := strings.ReplaceAll(storeKey, `\`, `\\`)
	escapedStoreKey = strings.ReplaceAll(escapedStoreKey, `|`, `\|`)
	hashVal.WriteString(escapedStoreKey)

	hash := sha512.Sum512([]byte(hashVal.String()))
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])

	return hashBase64
}
