package service

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"

	"testStand/internal/models"
	"testStand/internal/repos"
	"testStand/internal/txnhandler"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Service struct {
	dbClient *repos.Repo
}

func NewService(
	pgClient *sql.DB,
) *Service {
	return &Service{
		dbClient: repos.NewRepo(pgClient),
	}
}

func (s *Service) CreatePayoutTransaction(c echo.Context) error {
	req := &Request{}
	err := c.Bind(req)
	if err != nil {
		return err
	}
	resp := s.createTransaction(req, models.Transaction_PAYOUT)

	return c.JSON(http.StatusOK, resp)
}

func (s *Service) CreatePaymentTransaction(c echo.Context) error {
	req := &Request{}
	err := c.Bind(req)
	if err != nil {
		return err
	}
	resp := s.createTransaction(req, models.Transaction_PAYMENT)

	return c.JSON(http.StatusOK, resp)
}

func (s *Service) createTransaction(req *Request, txnType models.Transaction_Type) *Response {
	txn := &models.Transaction{
		TxnId:          int64(uuid.New().ID()),
		ParentTxn:      nil,
		TxnTypeId:      txnType,
		PayMethodId:    req.PaymentData.Type,
		PaymentData:    req.PaymentData,
		Customer:       &req.Customer,
		ChnName:        &req.GtwName,
		GtwName:        &req.GtwName,
		GtwTxnId:       nil,
		TxnAmountSrc:   req.Amount.Value,
		TxnCurrencySrc: req.Amount.Currency,
		TxnAmount:      0,
		TxnCurrency:    "",
		TxnInfo:        nil,
		TxnStatusId:    models.Transaction_NEW.String(),
		TxnUpdatedAt:   time.Time{},
	}

	ctx := context.Background()
	s.process(ctx, txn)

	resp := &Response{
		TxnId:     txn.TxnId,
		TxnStatus: txn.TxnStatusId,
	}

	if txn.Outputs != nil {
		result := &Result{}
		if cred, ok := txn.Outputs["credentials"]; ok && len(cred) > 0 {
			result.Credentials = cred
		}
		if bank, ok := txn.Outputs["bank"]; ok && len(bank) > 0 {
			result.Bank = bank
		}
		if desc, ok := txn.Outputs["description"]; ok && len(desc) > 0 {
			result.Description = desc
		}
		resp.Result = result
	}

	return resp
}

// process
func (s *Service) process(ctx context.Context, txn *models.Transaction) {
	logger := log.New("dev")

	// Choose acquirer by gateway
	acq, err := s.selectAcquirer(ctx, txn)
	if err != nil {
		logger.Error("Error creating acquirer for the gateway - ", err)
		if err == repos.ErrGtwNotFound || err == repos.ErrChnNotFound {
			logger.Error("Route not found")
		} else {
			txn.SetDeclined(time.Now())
			logger.Error(err.Error())
		}
	}

	// Handler transaction
	handler, _ := txnhandler.NewHandler(s.dbClient, acq)
	handler.HandleTxn(ctx, txn)
}

// selectAcquirer
func (s *Service) selectAcquirer(ctx context.Context, txn *models.Transaction) (any, error) {
	factory := NewFactory(s.dbClient)
	return factory.Create(ctx, txn)
}
