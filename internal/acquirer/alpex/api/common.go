package api

const (
	BUY  = "BUY"
	SELL = "SELL"
)
const (
	Pending    = "PENDING"
	Reconciled = "RELEASED"
	Decline    = "CANCELLED"
)

type PaymentMethod struct {
	Gate    Gate   `json:"gate"`
	Person  string `json:"person"`
	Address string `json:"address"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type SignResponse struct {
	SignatureKey string `json:"signature_key"`
}

type Gate struct {
	Id   string `json:"gate_id"`
	Name string `json:"name"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Request struct {
	FiatSymbol      string `json:"fiat_symbol"`
	FiatAmount      int64  `json:"fiat_amount"`
	CustomerName    string `json:"customer_name,omitempty"`
	CustomerAddress string `json:"customer_address,omitempty"`
	Direction       string `json:"direction"`
	GateId          string `json:"gate_id,omitempty"`
	ExternalId      string `json:"external_id"`
	WebhookUrl      string `json:"webhook_url"`
}

type Response struct {
	Id            string        `json:"id"`
	Status        string        `json:"status"`
	Signature     string        `json:"signature"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	//if error
	Message string `json:"message"`
	Code    string `json:"code"`
}

type Callback struct {
	Id          string `json:"_id"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Signature   string `json:"signature"`
}
