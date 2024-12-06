package txnhandler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"testStand/internal/acquirer"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/gommon/log"
)

var (
	errUnknownAcquirer = errors.New("unknown acquirer interface")
)

type handler struct {
	dbClient *repos.Repo
	acquirer acquirer.Acquirer
}

// NewHandler
func NewHandler(db *repos.Repo, acq any) (*handler, error) {
	h := handler{dbClient: db}
	if acqNew, ok := acq.(acquirer.Acquirer); ok {
		h.acquirer = acqNew
	} else {
		return nil, errUnknownAcquirer
	}

	return &h, nil
}

// HandleTxn
func (h *handler) HandleTxn(ctx context.Context, txn *models.Transaction) {

	if h.acquirer != nil {
		err := h.dbClient.CreateTransaction(txn)
		if err != nil {
			log.Error("error with saving transaction - ", err)
			return
		}
		h.handle(ctx, txn)

		err = h.dbClient.UpdateTransactionStatus(txn)
		if err != nil {
			log.Error("error with updating transaction - ", err)
			return
		}
	}
}

// handle
func (h *handler) handle(ctx context.Context, txn *models.Transaction) {
	logger := log.New("dev")

	var status *acquirer.TransactionStatus
	var err error

	if txn.IsPayment() {
		status, err = h.acquirer.Payment(ctx, txn)
	} else if txn.IsPayout() {
		status, err = h.acquirer.Payout(ctx, txn)
	} else if txn.IsCallback() {
		status, err = h.acquirer.HandleCallback(ctx, txn)
	} else {
		logger.Info("unknown transaction status")
	}

	if err != nil {
		logger.Error("error processing txn by acquirer interface: ", err)
		txnFillHandlingError(txn, err)
		return
	}

	updatedAt := time.Now()

	errorCode, errorMessage := "", ""
	if status.Info != nil {
		errorCode = status.Info["ps_error_code"]
		errorMessage = status.Info["ps_error_message"]
	}
	switch status.Status {
	case acquirer.APPROVED:
		txn.SetReconciled(updatedAt)

		logger.Info(fmt.Sprintf("approving txn with status: %s", txn.TxnStatusId))
	case acquirer.REJECTED:
		txn.SetDeclined(updatedAt)
		if status.TxnError == nil {
			errorCodeInt, _ := strconv.Atoi(errorCode)

			status.TxnError = acquirer.NewTxnError(errorCodeInt, errorMessage)
		}

		errInfo := fmt.Sprintf("declining txn with code: %d", status.TxnError.Code)
		if len(status.TxnError.Description) > 0 {
			errInfo += fmt.Sprintf(" and message %s", status.TxnError.Description)
		}
		logger.Info(errInfo)

	case acquirer.PENDING:
		h.SetSafePending(ctx, txn, &updatedAt)
	case acquirer.UNSPECIFIED:
		logger.Error("acquirer implementation returned unspecified status for transaction")
	}

	if status.GtwTxnId != nil {
		txn.GtwTxnId = status.GtwTxnId
	}

	if status.ConvertedAmount != nil {
		txn.TxnAmount = status.ConvertedAmount.Amount
		txn.TxnCurrency = status.ConvertedAmount.Currency
	}

	txn.Outputs = status.Outputs
}

// SetSafePending
func (h *handler) SetSafePending(ctx context.Context, txn *models.Transaction, updatedAt *time.Time) {
	logger := log.New("dev")

	if txn == nil {
		logger.Warn("cannot set pending for null transaction")
		return
	}

	if txn.IsReconciled() || txn.IsDeclined() {
		logger.Warn("can't mark txn as pending, status already final")

		if txn.IsDeclined() {
			if txn.Err != nil {
				logger.Error("error restoring last txn error")
			}
		}
	} else {
		txn.SetPending(*updatedAt)
		logger.Info("marking txn as pending")
	}
}

func txnFillHandlingError(txn *models.Transaction, err error) {
	if txn == nil || err == nil {
		return
	}

	txn.Err = acquirer.NewTxnError(5012, err.Error())
}
