package asupayme

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	ApiKey string `json:"api_key"`
	ShopId int    `json:"shop_id"`
}
