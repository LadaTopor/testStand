package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"testStand/internal/acquirer"
	"testStand/internal/acquirer/aplex"
	"testStand/internal/acquirer/asupayme"
	"testStand/internal/acquirer/auris"
	"testStand/internal/acquirer/paylink"
	"testStand/internal/acquirer/sequoia"
	"testStand/internal/models"
	"testStand/internal/repos"

	json "github.com/json-iterator/go"
	"github.com/labstack/gommon/log"
)

var ErrUnsupportedAcquirer = errors.New("unsupported acquirer")

const (
	AURIS    = "auris"
	SEQUOIA  = "sequoia"
	PAYLINK  = "paylink"
	ASUPAYME = "asupayme"
	APLEX    = "aplex"
)

type Factory struct {
	dbClient   *repos.Repo
	currentEnv string
}

// NewFactory
func NewFactory(dbClient *repos.Repo) *Factory {
	return &Factory{
		dbClient: dbClient,
	}
}

// Create
func (f *Factory) Create(ctx context.Context, txn *models.Transaction) (any, error) {
	logger := log.New("dev")

	gateway, err := f.dbClient.GetGateway(*txn.GtwName)
	if err != nil {
		logger.Error(fmt.Sprint(1, err))
		if err == sql.ErrNoRows {
			return nil, repos.ErrGtwNotFound
		}
		return nil, err
	}
	logger.Info(fmt.Sprintf("Loaded gateway: %v", gateway))
	channel, err := f.dbClient.GetChannel(*txn.ChnName)
	if err != nil {
		logger.Error(fmt.Sprint(2, err))
		if err == sql.ErrNoRows {
			return nil, repos.ErrChnNotFound
		}
		return nil, err
	}
	callbackUrl := "" // TODO ЗАПОЛНИТЬ

	acq, err := f.create(ctx, txn, gateway, channel.Params, callbackUrl)
	return acq, err
}

// create
func (f *Factory) create(ctx context.Context, txn *models.Transaction, gateway *repos.Gateway, channelParams repos.Params, callbackUrl string) (acquirer.Acquirer, error) {
	logger := log.New("dev")

	var err error
	var acq acquirer.Acquirer

	callbackUrl, err = url.JoinPath(callbackUrl, gateway.Adapter)
	if err != nil {
		return nil, err
	}

	switch gateway.Adapter {
	//case FAKE_BANK:
	//	acq = &fake.Acquirer{}
	case AURIS:
		var chParams auris.ChannelParams
		var gtwParams auris.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = auris.NewAcquirer(ctx, f.dbClient, chParams, gtwParams, callbackUrl)
	case SEQUOIA:
		var chParams sequoia.ChannelParams
		var gtwParams sequoia.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = sequoia.NewAcquirer(ctx, f.dbClient, &chParams, &gtwParams, callbackUrl)
	case PAYLINK:
		var chParams paylink.ChannelParams
		var gtwParams paylink.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = paylink.NewAcquirer(ctx, f.dbClient, &chParams, &gtwParams, callbackUrl)
	case ASUPAYME:
		var chParams asupayme.ChannelParams
		var gtwParams asupayme.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = asupayme.NewAcquirer(ctx, f.dbClient, chParams, gtwParams, callbackUrl)
	case APLEX:
		var chParams aplex.ChannelParams
		var gtwParams aplex.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = aplex.NewAcquirer(ctx, f.dbClient, &chParams, &gtwParams, callbackUrl)
	default:
		return nil, ErrUnsupportedAcquirer
	}
	logger.Info(fmt.Sprintf("Loaded acquirer: %s", gateway.Adapter))

	return acq, nil
}

// unmarshalParams
func (f *Factory) unmarshalParams(gatewayParamsJson string, channelParamsJson []byte, gatewayParams any, channelParams any) error {
	if err := json.Unmarshal(channelParamsJson, channelParams); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(gatewayParamsJson), gatewayParams); err != nil {
		return err
	}
	return nil
}
