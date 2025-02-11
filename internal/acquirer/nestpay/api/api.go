package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"

	"testStand/internal/acquirer/helper"
)

// Endpoints
const (
	PaymentEndpoint = "fim/api"
	AuthEndpoint    = "fim/est3dgate"
)

type Client struct {
	baseAddress string
	storeKey    string
	currency    string
	client      *http.Client
}

// NewClient
func NewClient(ctx context.Context, baseAddress, storeKey string, currency string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		storeKey:    storeKey,
		currency:    currency,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *CC5Request) (*CC5Response, error) {
	body, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, PaymentEndpoint), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &CC5Response{}
	err = xml.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// makeAuth
func (c *Client) MakeAuth(ctx context.Context, authParams url.Values) (url.Values, error) {
	bodyReader := bytes.NewReader([]byte(authParams.Encode()))

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, AuthEndpoint), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	text, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response, err := url.ParseQuery(string(text))
	if err != nil {
		return nil, err
	}

	return response, nil
}
