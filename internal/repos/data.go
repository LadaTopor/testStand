package repos

import (
	"context"
	"database/sql"
	"encoding/json"

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
func (db *Repo) GetGateway(ctx context.Context, gatewayId string) (*Gateway, error) {

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
func (db *Repo) GetChannel(ctx context.Context, channelId string) (*Channel, error) {
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
