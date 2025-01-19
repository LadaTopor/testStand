package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"golang.org/x/exp/slices"
)

type callbackMapperFunc func(*repos.Repo, string, []byte, string, http.Header) (int64, any, error)

type callbackMapper struct {
	Handler callbackMapperFunc
	Methods []string
}

var callbackMappersMap = map[string]*callbackMapper{
	"paylink": newCallbackMapper(paylinkCallbackMapper, http.MethodPost),
	"auris":   newCallbackMapper(aurisCallbackMapper, http.MethodPost),
	"sequoia": newCallbackMapper(sequoiaCallbackMapper, http.MethodPost),
	"alpex":   newCallbackMapper(alpexCallbackMapper, http.MethodPost),
}

func newCallbackMapper(handler callbackMapperFunc, methods ...string) *callbackMapper {
	return &callbackMapper{
		Handler: handler,
		Methods: methods,
	}
}

func (s *Service) CallbackHandler(c echo.Context) error {
	logger := log.New("dev")

	acq := c.Param("acquirer")
	if len(acq) == 0 {
		logger.Error("acquirer param is not present")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "acquirer param is not present"})
	}

	cbHandler, ok := callbackMappersMap[acq]
	if !ok {
		logger.Error(fmt.Sprintf("unknown acquirer: %s", acq))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("unknown acquirer: %s", acq)})
	}
	logger.Info(fmt.Sprintf("selected handler: %s", acq))

	if !slices.Contains(cbHandler.Methods, c.Request().Method) {
		logger.Error("unsupported http method")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported http method"})
	}

	query := c.QueryParams().Encode()

	// Check body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil || len(bodyBytes) == 0 {
		if query == "" {
			logger.Error("request body read error - ", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "request body read error"})
		}
	}

	headers := c.Request().Header

	txnId, payload, err := cbHandler.Handler(s.dbClient, acq, bodyBytes, query, headers)
	if err != nil {
		logger.Error("request body parsing error - ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "request body read error"})
	}

	// Skip callback handling if txnId is missing
	if txnId == 0 {
		logger.Info("Skipping callback handling")
		return c.String(200, "OK")
	}

	logger.Info(fmt.Sprintf("callback parsed. txn_id = %d", txnId))

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("payload marshaling error - ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "payload marshaling error"})
	}
	logger.Info("callback marshalled")

	txn, err := s.dbClient.GetTransaction(txnId)
	if err != nil {
		logger.Error("getting txn error")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "getting txn error"})
	}

	txn.TxnInfo["callback"] = string(payloadBytes)
	txn.TxnTypeId = models.Transaction_CALLBACK

	go s.process(context.Background(), txn)

	return c.String(http.StatusOK, "OK")
}

// aurisCallbackMapper
func aurisCallbackMapper(payRepo *repos.Repo, gtwAdapterId string, content []byte, query string, headers http.Header) (int64, any, error) {
	logger := log.New("dev")

	callback := Auris{}
	if err := json.Unmarshal(content, &callback); err != nil {
		logger.Error("callback body unmarshalling error - ", err)
		return 0, nil, err
	}

	txnId, err := strconv.ParseInt(callback.Label, 10, 64)
	if err != nil {
		logger.Error("error parsing txnId - ", err)
		return 0, nil, err
	}

	return txnId, callback, nil
}

func sequoiaCallbackMapper(payRepo *repos.Repo, gtwAdapterId string, content []byte, query string, headers http.Header) (int64, any, error) {
	logger := log.New("dev")

	callback := Sequoia{}
	if err := json.Unmarshal(content, &callback); err != nil {
		logger.Error("callback body unmarshalling error - ", err)
		return 0, nil, err
	}

	txnId, err := strconv.ParseInt(callback.OrderId, 10, 64)
	if err != nil {
		logger.Error("error parsing txnId - ", err)
		return 0, nil, err
	}

	return txnId, callback, nil
}

func paylinkCallbackMapper(payRepo *repos.Repo, gtwAdapterId string, content []byte, query string, headers http.Header) (int64, any, error) {
	logger := log.New("dev")

	callback := Paylink{}
	if err := json.Unmarshal(content, &callback); err != nil {
		logger.Error("callback body unmarshalling error - ", err)
		return 0, nil, err
	}

	txnId, err := strconv.ParseInt(callback.UserRef, 10, 64)
	if err != nil {
		logger.Error("error parsing txnId - ", err)
		return 0, nil, err
	}

	return txnId, callback, nil
}

func alpexCallbackMapper(payRepo *repos.Repo, gtwAdapterId string, content []byte, query string, headers http.Header) (int64, any, error) {
	logger := log.New("dev")

	callback := Alpex{}
	if err := json.Unmarshal(content, &callback); err != nil {
		logger.Error("callback body unmarshalling error - ", err)
		return 0, nil, err
	}
	fmt.Println(callback.Id, callback.Status)
	txnId, err := strconv.ParseInt(callback.ExternalId, 10, 64)
	if err != nil {
		logger.Error("error parsing txnId - ", err)
		return 0, nil, err
	}

	return txnId, callback, nil
}
