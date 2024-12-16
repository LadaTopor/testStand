package api

const (
	Pending    = "PENDING"
	Reconciled = "RELEASED"
	Decline    = "CANCELLED"
)

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Gate struct {
	Id   string `json:"gate_id"`
	Name string `json:"name"`
}

type Request struct {
	GateId          string `json:"gate_id,omitempty"`
	CustomerName    string `json:"customer_name,omitempty"`
	CustomerAddress string `json:"customer_address,omitempty"`
	FiatSymbol      string `json:"fiat_symbol"`
	FiatAmount      int64  `json:"fiat_amount"`
	Direction       string `json:"direction"`
	WebhookUrl      string `json:"webhook_url"`
	ExternalId      string `json:"external_id"`
}

type ResponsePaymentMethod struct {
	Id              string `json:"_id"`
	Gate            Gate   `json:"gate"`
	CustomerName    string `json:"person"`
	CustomerAddress string `json:"address"`
	PayMethod       string `json:"name"`
}

type Response struct {
	Status        string                `json:"status"`
	PaymentMethod ResponsePaymentMethod `json:"payment_method"`
	Id            string                `json:"id"`
	Url           string                `json:"url"`
	Error         string                `json:"error"`
	Amount        string                `json:"amount"`
	Direction     string                `json:"direction"`
	Sign          string                `json:"signature"`
}

type Callback struct {
	Id               string `json:"_id"`
	UserRef          string `json:"user_ref"`
	Status           string `json:"status"`
	Description      string `json:"description"`
	TimestampUpdated string `json:"timestamp_updated"`
	Amount           string `json:"amount"`
	Sign             string `json:"signature"`
}
