package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testStand/internal/acquirer/helper"
)

// Endpoints
const (
	PaymentEndpoint = "payment"
	PayoutEndpoint  = "withdraw"
	StatusEndpoint  = "status"
)

type Client struct {
	baseAddress string
	apiKey      string
	secretKey   string
	client      *http.Client
}

// NewClient
func NewClient(ctx context.Context, baseAddress, apiKey string, secretKey string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		apiKey:      apiKey,
		secretKey:   secretKey,
		client:      client,
	}
}

// MakeDeposit
func (c *Client) MakeDeposit(ctx context.Context, request Request) (*Response, error) {
	resp := &Response{}

	return resp, nil
}

// MakeWithdraw
func (c *Client) MakeWithdraw(ctx context.Context, request Request) (*Response, error) {
	sign := createSign(request.MerchantId + request.CardData.CardNumber + request.Amount + c.secretKey)
	request.Sign = sign

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, request.ApiKey, PayoutEndpoint)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp)

	return resp, nil
}

// CheckStatus
func (c *Client) CheckStatus(ctx context.Context, request StatusRequest) (*Response, error) {
	resp := &Response{}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, apiKey string, endpoint string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Println(string(body))

	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), reader)
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"Host":           {"185.230.143.210"},
		"Authorization":  {"Bearer " + apiKey},
		"Content-Type":   {"application/json"},
		"Content-Length": {fmt.Sprintf("%d", reader.Size())},
	}

	fmt.Println(req.Header)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return errors.New("declined state due to network or internal error")
	}

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}
	return nil
}
