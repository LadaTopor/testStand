package service

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"testStand/internal/models"
	"testStand/internal/repos"
	"testStand/internal/txnhandler"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
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

	txn := &models.Transaction{
		TxnId:          int64(uuid.New().ID()),
		ParentTxn:      nil,
		TxnTypeId:      models.Transaction_PAYOUT,
		PayMethodId:    req.PaymentData.Type,
		Customer:       &req.Customer,
		ChnName:        &req.GtwName,
		GtwName:        &req.GtwName,
		GtwTxnId:       nil,
		TxnAmountSrc:   req.Amount.Value,
		TxnCurrencySrc: req.Amount.Currency,
		TxnAmount:      0,
		TxnCurrency:    "",
		TxnInfo:        nil,
		TxnStatusId:    models.Transaction_NEW,
		TxnUpdatedAt:   time.Time{},
	}

	ctx := context.Background()
	s.process(ctx, txn)

	resp := &Response{
		TxnId:     txn.TxnId,
		TxnStatus: txn.TxnStatusId.String(),
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
			result.Credentials = desc
		}
		resp.Result = result
	}

	return c.JSON(http.StatusOK, resp)
}

// process
func (s *Service) process(ctx context.Context, txn *models.Transaction) {
	logger, _ := zap.NewDevelopment()

	// Choose acquirer by gateway
	acq, err := s.selectAcquirer(ctx, txn)
	if err != nil {
		logger.Error("Error creating acquirer for the gateway", zap.Error(err))
		if err == repos.ErrGtwNotFound || err == repos.ErrChnNotFound {
			logger.Fatal("Route not found")
		} else {
			txn.SetDeclined(time.Now())
			logger.Fatal(err.Error())
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
