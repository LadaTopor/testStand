package repos

import (
	"database/sql"
	"encoding/json"
	"testStand/internal/models"

	_ "github.com/lib/pq"
)

type Repo struct {
	pgClient *sql.DB
}

// NewRepo
func NewRepo(pgClient *sql.DB) *Repo {
	return &Repo{pgClient: pgClient}
}

// GetGateway
func (db *Repo) GetGateway(gatewayId string) (*Gateway, error) {

	row := db.pgClient.QueryRow(
		`SELECT
		gtw_adapter_id,
		gtw_params_jsonb
		FROM gateway
		WHERE gtw_adapter_id = $1 AND gtw_is_active=true`,
		gatewayId)
	gateway := &Gateway{}
	err := row.Scan(&gateway.Adapter, &gateway.ParamsJson)
	if err != nil {
		return nil, err
	}

	return gateway, nil
}

// GetChannel
func (db *Repo) GetChannel(channelId string) (*Channel, error) {
	channel := new(Channel)
	var channelParams []byte
	row := db.pgClient.QueryRow(
		`SELECT
    	chn_id,
    	chn_name,
    	chn_is_active,
    	gtw_id,
    	chn_params_jsonb
		FROM channel
		WHERE chn_name = $1 AND chn_is_active=true`, channelId)
	err := row.Scan(&channel.Id, &channel.Name, &channel.IsActive, &channel.GtwId, &channelParams)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(channelParams, &channel.Params)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

// CreateTransaction
func (db *Repo) CreateTransaction(txn *models.Transaction) error {
	_, err := db.pgClient.Exec(`
		INSERT INTO transaction (
			txn_id,
		    txn_type_id,
		    pay_method_id,
		    chn_name,
		    gtw_name,
		    gtw_txn_id,
			txn_amount_src,
		    txn_currency_src,
		    txn_amount,
		    txn_currency,
		    txn_status_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`, txn.TxnId,
		txn.TxnTypeId,
		txn.PayMethodId,
		txn.ChnName,
		txn.GtwName,
		txn.GtwTxnId,
		txn.TxnAmountSrc,
		txn.TxnCurrencySrc,
		txn.TxnAmount,
		txn.TxnCurrency,
		txn.TxnStatusId)
	if err != nil {
		return err
	}

	return nil
}

// GetTransaction
func (db *Repo) GetTransaction(txnId int64) (*models.Transaction, error) {
	txn := &models.Transaction{}
	err := db.pgClient.QueryRow(`SELECT 
    txn_id,
    txn_type_id,
    pay_method_id,
    chn_name,
    gtw_name,
    gtw_txn_id,
    txn_amount_src,
    txn_currency_src,
    txn_amount,
    txn_currency,
    txn_status_id
	FROM 
		transaction
	WHERE 
		txn_id = $1`, txnId).Scan(
		txn.TxnId,
		txn.TxnTypeId,
		txn.PayMethodId,
		txn.ChnName,
		txn.GtwName,
		txn.GtwTxnId,
		txn.TxnAmountSrc,
		txn.TxnCurrencySrc,
		txn.TxnAmount,
		txn.TxnCurrency,
		txn.TxnStatusId)
	if err != nil {
		return nil, err
	}

	return txn, nil
}

// GetTransaction
func (db *Repo) UpdateTransactionStatus(txn *models.Transaction) error {
	_, err := db.pgClient.Exec(`
	UPDATE transaction
	SET txn_status_id = $1, txn_updated_at = CURRENT_TIMESTAMP
	WHERE txn_id = $2`, txn.TxnId, txn.TxnStatusId)
	if err != nil {
		return err
	}

	return nil
}
