package api

import (
	"crypto/sha1"
	"encoding/base64"
	"testStand/internal/acquirer/helper"
)

const (
	Pending    = "new"
	Reconciled = "executed"
	Decline    = "cancelled"
)

type Request struct {
	MerchId         string `json:"merch_id"`
	Extra           string `json:"extra"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	NotificationUrl string `json:"notification_url"`
	UserId          string `json:"user_id,omitempty"`
	UserRef         string `json:"user_ref,omitempty"`
	UserIp          string `json:"user_ip,omitempty"`
	FinishUrl       string `json:"finish_url"`
	//Sign            string `json:"sign"`
}

type Response struct {
	OK             bool   `json:"ok"`
	Status         string `json:"status"`
	Id             string `json:"id"`
	Url            string `json:"url"`
	P2PDestination string `json:"p2p_destination"`
	P2PBank        string `json:"p2p_bank"`
	P2PName        string `json:"p2p_name"`
	Error          string `json:"error"`
	Amount         string `json:"amount"`
}

type Callback struct {
	Id               string `json:"id"`
	UserRef          string `json:"user_ref"`
	Status           string `json:"status"`
	Description      string `json:"description"`
	TimestampUpdated string `json:"timestamp_updated"`
	Amount           string `json:"amount"`
	Sign             string `json:"sign"`
}

type StatusRequest struct {
	Id      string `json:"id"`
	MerchId string `json:"merch_id"`
	UserRef string `json:"user_ref,omitempty"`
}

func createSign(input, apiKey string) string {
	sum := helper.GenerateHMAC(sha1.New, []byte(input), apiKey)
	return base64.StdEncoding.EncodeToString(sum)
}
