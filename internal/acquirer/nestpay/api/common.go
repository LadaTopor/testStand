package api

const (
	Approve = "Approve"
	Error   = "Error"
)

type CC5Request struct {
	Name       string `xml:"Name"`
	Password   string `xml:"Password"`
	ClientId   string `xml:"ClientId"`
	IPAddress  string `xml:"IPAddress"`
	OrderId    string `xml:"oid"`
	Type       string `xml:"Type"`
	Pan        string `xml:"Number"`
	Amount     string `xml:"Amount"`
	Currency   string `xml:"Currency"`
	PayerTxnId string `xml:"PayerTxnId"`
	ECI        string `xml:"PayerSecurityLevel"`
	CAVV       string `xml:"PayerAuthenticationCode"`
}

type PaymentResponse struct {
	Response       string `xml:"response"`
	ProcReturnCode string `xml:"procReturnCode"`
	ErrMsg         string `xml:"errMsg"`
	OrderId        string `xml:"orderId"`
}
