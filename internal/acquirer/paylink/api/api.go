package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	apiKey      string
	client      *http.Client
}

const (
	payment = "payment"
	payout  = "payout"
	status  = "status"
)

func NewClient(ctx context.Context, baseAddress, apiKey string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		apiKey:      apiKey,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *Request) (*Response, error) {
	sign := createSign(request.MerchId, c.apiKey)

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, sign, payment)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakePayout
func (c *Client) MakePayout(ctx context.Context, request *Request, apiKey string) (*Response, error) {
	sign := createSign(payout+request.UserRef, apiKey)

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, sign, payout)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetStatus
func (c *Client) GetStatus(ctx context.Context, request *StatusRequest, apiKey string) (*Response, error) {
	sign := createSign(request.Id, apiKey)

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, sign, status)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, sign, endpoint string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sign", sign)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}
	return nil
}
