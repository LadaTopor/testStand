package api

type Response struct {
	Status string `json:"status"`
	ID     string `json:"id,omitempty"`
	Error  string `json:"error,omitempty"`
}

type WithdrawRequest struct {
	Merchant   string            `json:"merchant"`
	WithdrawID string            `json:"withdraw_id"`
	Amount     string            `json:"amount"`
	Signature  string            `json:"signature"`
	Payload    map[string]string `json:"payload,omitempty"`
	CardData   CardData          `json:"card_data"`
}

type CardData struct {
	OwnerName    string `json:"owner_name"`
	CardNumber   string `json:"card_number"`
	ExpiredMonth string `json:"expired_month"`
	ExpiredYear  string `json:"expired_year"`
}
