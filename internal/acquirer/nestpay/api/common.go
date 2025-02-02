package api

import "encoding/xml"

const (
	StatusApproved = "Approve"
	StatusDeclined = "Declined"
	StatusError    = "Error"
)

type Request struct {
	ClientId                        string `form:"ClientId"`
	StoreType                       string `form:"storeType"`
	TranType                        string `form:"tranType"`
	Amount                          int64  `form:"amount"`
	Currency                        string `form:"Currency"`
	Lang                            string `form:"lang"`
	Oid                             string `form:"oid"`
	Rnd                             string `form:"rnd"`
	Pan                             string `form:"pan"`
	Ecom_Payment_Card_ExpDate_Month string `form:"Ecom_Payment_Card_ExpDate_Month"`
	Ecom_Payment_Card_ExpDate_Year  string `form:"Ecom_Payment_Card_ExpDate_Year"`
	Cv2                             string `form:"cv2"`
	HashAlgorithm                   string `form:"hashAlgorithm"`
	Encoding                        string `form:"encoding"`
	Hash                            string `form:"hash"`
}

type PaymentResponse struct {
	ClientID      string `form:"clientId"`
	ClientIP      string `form:"clientIp"`
	Oid           string `form:"oid"`
	Md            string `form:"md"`
	Amount        int    `form:"amount"`
	Currency      string `form:"currency"`
	ErrMsg        string `form:"ErrMsg"`
	TransID       string `form:"TransID"`
	HashAlgorithm string `form:"hashAlgorithm"`
	MDStatus      int    `form:"mdStatus"`
	MerchantID    string `form:"merchantID"`
	SId           string `form:"sID"`
	XId           string `form:"xid"`
	Eci           string `form:"eci"`
	Cavv          string `form:"cavv"`
}

type StatusRequest struct {
	XMLName                 xml.Name `xml:"CC5Request"`
	Name                    string   `xml:"Name"`
	Password                string   `xml:"Password"`
	ClientId                string   `xml:"ClientId"`
	IpAddress               string   `xml:"IpAddress"`
	Oid                     string   `xml:"oid"`
	Type                    string   `xml:"Type"`
	Number                  string   `xml:"Number"`
	Amount                  string   `xml:"Amount"`
	Currency                string   `xml:"Currency"`
	PayerTxnId              string   `xml:"PayerTxnId"`
	PayerSecurityLevel      string   `xml:"PayerSecurityLevel"`
	PayerAuthenticationCode string   `xml:"PayerAuthenticationCode"`
}

type StatusResponse struct {
	XMLName        xml.Name `xml:"CC5Response"`
	Response       string   `xml:"Response"`
	OrderID        string   `xml:"OrderID"`
	ProcReturnCode string   `xml:"ProcReturnCode"`
	ErrMsg         string   `xml:"ErrMsg"`
}
