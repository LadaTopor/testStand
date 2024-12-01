package api

const P2Pmethod = 1

type Request struct {
	WithdrawId string   `json:"withdraw_id"`
	MerchId    string   `json:"merchant"`
	Amount     string   `json:"amount"`
	CardData   CardData `json:"card_data"`
	Sign       string   `json:"signature"`
}

type CardData struct {
	CardNumber string `json:"card_number"`
}

type Response struct {
	OK     bool   `json:"ok"`
	Status string `json:"status"`
	Id     string `json:"id"`
	//if error
	Message string `json:"message"`
	Code    string `json:"code"`
}

type Status struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
