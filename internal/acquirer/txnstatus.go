package acquirer

import (
	"errors"
	"time"
)

// SetApproved
func (s *TransactionStatus) SetApproved() {
	if s != nil {
		s.Status = APPROVED
	}
}

// SetPending
func (s *TransactionStatus) SetPending() {
	if s != nil {
		s.Status = PENDING
	}
}

// SetRejected
func (s *TransactionStatus) SetRejected() {
	if s != nil {
		s.Status = REJECTED
	}
}

// SetDateFromString
func (s *TransactionStatus) SetDateFromString(layout string, timeValue string) error {
	if s == nil {
		return nil
	}

	t, err := time.Parse(layout, timeValue)
	if err != nil {
		return err
	}

	s.Date = &t
	return nil
}

// SetConvertedAmount
func (s *TransactionStatus) SetConvertedAmount(amount int64, currency string) error {
	if s == nil {
		return nil
	}

	if len(currency) == 0 {
		return errors.New("SetConvertedAmount: currency is empty string")
	}

	s.ConvertedAmount = &TransactionAmount{Amount: amount, Currency: currency}
	return nil
}

// SetGtwTxnId
func (s *TransactionStatus) SetGtwTxnId(gtwTxnId *string) {
	if s == nil {
		return
	}

	if gtwTxnId == nil {
		return
	}

	if len(*gtwTxnId) == 0 {
		return
	}

	s.GtwTxnId = gtwTxnId
}

// AddInfoItem
func (s *TransactionStatus) AddInfoItem(key string, value string) {
	if s == nil {
		return
	}

	if s.Info == nil {
		s.Info = map[string]string{}
	}

	s.Info[key] = value
}

// AddOutputItem
func (s *TransactionStatus) AddOutputItem(key string, value string) {
	if s == nil {
		return
	}

	if s.Outputs == nil {
		s.Outputs = map[string]string{}
	}

	s.Outputs[key] = value
}

// SetTxnError
func (s *TransactionStatus) SetTxnError(code int, message string) {
	if s == nil {
		return
	}

	s.TxnError.Code = code
	s.TxnError.Description = message
}
