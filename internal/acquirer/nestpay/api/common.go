package api

const (
	Error   = "Error"
	Approve = "Approve"
)

type CC5Request struct {
	Name          string `xml:"Name"`
	Password      string `xml:"Password"`
	ClientId      string `xml:"ClientId"`
	OrderId       string `xml:"OrderId"`
	Email         string `xml:"Email"`
	StoreType     string `xml:"storetype"`
	Type          string `xml:"Type"`
	CardNumber    string `xml:"Number"`
	CardMonth     string `xml:"Ecom_Payment_Card_ExpDate_Month"`
	CardYear      string `xml:"Ecom_Payment_Card_ExpDate_Year"`
	CVV           string `xml:"Cvv2Val"`
	Total         int64  `xml:"Amount"`
	Currency      string `xml:"Currency"`
	PayerTxnId    string `xml:"PayerTxnId"`
	ECI           string `xml:"PayerSecurityLevel"`
	CAVV          string `xml:"PayerAuthenticationCode"`
	HashAlgorithm string `xml:"HashAlgorithm"`
	Encoding      string `xml:"encoding"`
}

type CC5Response struct {
	OrderId        string `xml:"OrderId"`
	Response       string `xml:"Response"`
	ProcReturnCode string `xml:"ProcReturnCode"`
	ErrMsg         string `xml:"ErrMsg,omitempty"`
}
