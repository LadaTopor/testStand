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
	PayoutEndpoint = "withdraw"
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
func (c *Client) MakeDeposit(ctx context.Context, request PayoutRequest) (*Response, error) {
	resp := &Response{}

	return resp, nil
}

// MakeWithdraw
func (c *Client) MakeWithdraw(ctx context.Context, request PayoutRequest, apiKey string) (*Response, error) {
	sign := createSign(request, c.secretKey)
	request.Sign = sign

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, apiKey, PayoutEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, apiKey string, endpoint string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), reader)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

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
