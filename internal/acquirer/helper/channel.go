package helper

type channelPayMethodMode struct {
	DefaultBank string `json:"default_bank"`
}

type channelPayCurrency map[string]channelPayMethodMode

type ChannelPayMethod map[string]channelPayCurrency

// method -> currency -> value
// TODO:
// ? add to GetDefaultBank key to get not only default_bank
// ? add payment/payount before method to get specific data
// ? maybe concatinate method + paymethodId + currency, value still map

// GetDefaultBank
func GetDefaultBank(method ChannelPayMethod, payMethod, currency string) string {

	if method == nil {
		return ""
	}

	methodData, methodExists := method[payMethod]
	if !methodExists {
		return ""
	}

	currencyData, currencyExists := methodData[currency]
	if !currencyExists {
		return ""
	}

	return currencyData.DefaultBank
}
