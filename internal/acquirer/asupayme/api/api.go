package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testStand/internal/acquirer/helper"
)

// Endpoints
const (
	PayoutEndpoint = "/api/v1/withdraw"
)

type Client struct {
	baseAddress string
	apiKey      string
	secretKey   string
	client      *http.Client
}

// NewClient
func NewClient(ctx context.Context, baseAddress, apiKey string, secretKey string) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		apiKey:      apiKey,
		client:      client,
		secretKey:   secretKey,
	}
}

// MakeWithdraw
func (c *Client) MakeWithdraw(ctx context.Context, request Request) (*Response, error) {

	sign := createSign(request.Merchant + request.CardData.CardNumber + request.Amount + c.secretKey)
	request.Sign = sign

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, PayoutEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, endpoint string) error {

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
