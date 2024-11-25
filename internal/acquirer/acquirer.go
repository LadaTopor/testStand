package acquirer

import (
	"context"
	"testStand/internal/models"
	"time"
)

type Status int32

const (
	UNSPECIFIED = iota
	APPROVED
	REJECTED
	PENDING
)

type TransactionAmount struct {
	Amount   int64
	Currency string
}

func NewTxnError(code int, desc string) *models.TxnError {
	return &models.TxnError{Code: code, Description: desc}
}

type TransactionStatus struct {
	Status           Status
	Date             *time.Time
	ConvertedAmount  *TransactionAmount
	GtwAmountSettled *TransactionAmount
	TxnError         *models.TxnError
	GtwTxnId         *string
	Outputs          map[string]string
	Info             map[string]string
}

// Acquirer interface for communicating with acquirer.
//
// Objects implementing this interface should never change transaction object directly,
// instead they should return information regarding operation status and appropriate fields.
type Acquirer interface {
	Payment(ctx context.Context, txn *models.Transaction) (*TransactionStatus, error)
	Payout(ctx context.Context, txn *models.Transaction) (*TransactionStatus, error)
	HandleCallback(ctx context.Context, txn *models.Transaction) (*TransactionStatus, error)
	FinalizePending(ctx context.Context, txn *models.Transaction) (*TransactionStatus, error)
}
