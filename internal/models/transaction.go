package models

import (
	"time"
)

type Transaction_Type int32

const (
	Transaction_TYPE_UNSPECIFIED Transaction_Type = 0
	Transaction_PAYMENT          Transaction_Type = 1
	Transaction_REFUND           Transaction_Type = 2
	Transaction_REVERSE          Transaction_Type = 3
	Transaction_PAYOUT           Transaction_Type = 4
	Transaction_CALLBACK         Transaction_Type = 5
	Transaction_TRANSFER         Transaction_Type = 5
	Transaction_TECHREVERSE      Transaction_Type = 6
	Transaction_INTERNAL         Transaction_Type = 7
)

// Transaction - объект транзакции.
type Transaction struct {
	TxnId          int64             `json:"txn_id,omitempty"`
	ParentTxn      *Transaction      `json:"parent_txn,omitempty"`
	TxnTypeId      Transaction_Type  `json:"txn_type_id,omitempty"`
	PayMethodId    string            `json:"pay_method_id,omitempty"`
	PaymentData    PaymentData       `json:"payment_data,omitempty"`
	Customer       *Customer         `json:"customer,omitempty"`
	ChnName        *string           `json:"chn_name,omitempty"`
	GtwName        *string           `json:"gtw_name,omitempty"`
	GtwTxnId       *string           `json:"gtw_txn_id,omitempty"`
	TxnAmountSrc   int64             `json:"txn_amount_src,omitempty"`
	TxnCurrencySrc string            `json:"txn_currency_src,omitempty"`
	TxnAmount      int64             `json:"txn_amount,omitempty"`
	TxnCurrency    string            `json:"txn_currency,omitempty"`
	TxnInfo        map[string]string `json:"txn_info,omitempty"`
	TxnStatusId    string            `json:"txn_status_id,omitempty"`
	TxnUpdatedAt   time.Time         `json:"txn_updated_at,omitempty"`
	Err            *TxnError         `json:"txn_error,omitempty"`
	Outputs        map[string]string `json:"outputs,omitempty"`
}

type TxnError struct {
	Code        int    `json:"code,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Amount struct {
	Value    int64  `json:"value"`
	Currency string `json:"currency"`
}

type PaymentData struct {
	Type   string `json:"type"`
	Object struct {
		Credentials string  `json:"credentials"`
		Bank        *string `json:"bank"`
		ExpYear     string  `json:"exp_year"`
		ExpMonth    string  `json:"exp_month"`
		Cvv         string  `json:"cvv"`
	} `json:"object"`
}

type Customer struct {
	AccountId    string `json:"accountId"`
	Fingerprint  string `json:"fingerprint"`
	Ip           string `json:"ip"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	FullName     string `json:"fullName"`
	Country      string `json:"country"`
	Address      string `json:"address"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postalCode"`
	Neighborhood string `json:"neighborhood"`
	Birthdate    string `json:"birthdate"`
	BrowserData  struct {
		AcceptHeader string `json:"acceptHeader"`
		ColorDepth   int    `json:"colorDepth"`
		Language     string `json:"language"`
		ScreenHeight int    `json:"screenHeight"`
		ScreenWidth  int    `json:"screenWidth"`
		Timezone     int    `json:"timezone"`
		UserAgent    string `json:"userAgent"`
		JavaEnabled  bool   `json:"javaEnabled"`
		WindowHeight int    `json:"windowHeight"`
		WindowWidth  int    `json:"windowWidth"`
	} `json:"browserData"`
}

type Transaction_Status int32

const (
	Transaction_STATUS_UNSPECIFIED Transaction_Status = 0
	Transaction_NEW                Transaction_Status = 1
	Transaction_PENDING            Transaction_Status = 2
	Transaction_DECLINED           Transaction_Status = 3
	Transaction_AUTHORIZED         Transaction_Status = 4
	Transaction_CONFIRMED          Transaction_Status = 5
	Transaction_RECONCILED         Transaction_Status = 6
	Transaction_SETTLED            Transaction_Status = 7
)

var Transaction_Status_name = map[Transaction_Status]string{
	0: "STATUS_UNSPECIFIED",
	1: "NEW",
	2: "PENDING",
	3: "DECLINED",
	4: "AUTHORIZED",
	5: "CONFIRMED",
	6: "RECONCILED",
	7: "SETTLED",
}

func (x Transaction_Status) String() string {
	return Transaction_Status_name[x]
}

//type Customer struct {
//	Id            int32   `json:"id,omitempty"`
//	MerchantId    string  `json:"merchant_id,omitempty"`
//	AccountId     *string `json:"account_id,omitempty"`
//	Phone         *string `json:"phone,omitempty"`
//	Email         *string `json:"email,omitempty"`
//	Cookie        *string `json:"cookie,omitempty"`
//	Fingerprint   *string `json:"fingerprint,omitempty"`
//	Ip            *string `json:"ip,omitempty"`
//	IpCountryCode *string `json:"ip_country_code,omitempty"`
//	ParamsJson    *string `json:"params_json,omitempty"`
//}

// IsPayment
func (txn *Transaction) IsPayment() bool {
	if txn == nil {
		return false
	}
	return txn.TxnTypeId == Transaction_PAYMENT
}

// IsPayout
func (txn *Transaction) IsPayout() bool {
	if txn == nil {
		return false
	}
	return txn.TxnTypeId == Transaction_PAYOUT
}

func (txn *Transaction) IsCallback() bool {
	if txn == nil {
		return false
	}
	return txn.TxnTypeId == Transaction_CALLBACK
}

// Transaction status

// IsNew
func (txn *Transaction) IsNew() bool {
	if txn == nil {
		return false
	}
	return txn.TxnStatusId == Transaction_NEW.String()
}

// IsPending
func (txn *Transaction) IsPending() bool {
	if txn == nil {
		return false
	}
	return txn.TxnStatusId == Transaction_PENDING.String()
}

// IsDeclined
func (txn *Transaction) IsDeclined() bool {
	if txn == nil {
		return false
	}
	return txn.TxnStatusId == Transaction_DECLINED.String()
}

// IsReconciled
func (txn *Transaction) IsReconciled() bool {
	if txn == nil {
		return false
	}
	return txn.TxnStatusId == Transaction_RECONCILED.String()
}

// SetPending
func (txn *Transaction) SetPending(updatedAt time.Time) {
	txn.setStatus(Transaction_PENDING, nil, &updatedAt)
}

// SetDeclined
func (txn *Transaction) SetDeclined(updatedAt time.Time) {
	txn.setStatus(Transaction_DECLINED, nil, &updatedAt)
}

// SetReconciled
func (txn *Transaction) SetReconciled(reconciledAt time.Time) {
	txn.setStatus(Transaction_RECONCILED, nil, &reconciledAt)
}

// setStatus
func (txn *Transaction) setStatus(status Transaction_Status, dateField *time.Time, date *time.Time) {
	if txn == nil {
		return
	}

	if txn.TxnStatusId == status.String() {
		return
	}

	txn.TxnStatusId = status.String()
	if dateField != nil {
		*dateField = *date
	}
	txn.TxnUpdatedAt = *date
}
