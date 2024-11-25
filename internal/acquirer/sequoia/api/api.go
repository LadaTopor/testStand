package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	apiKey      string
	client      *http.Client
}

const (
	payment = "api/pay"
	status  = "api/order"
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
func (c *Client) MakePayment(ctx context.Context, request *PaymentRequest) (*Response, error) {

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, http.MethodPost, helper.JoinUrl(c.baseAddress, payment))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckStatus
func (c *Client) CheckStatus(ctx context.Context, request *PaymentRequest) (*Status, error) {
	resp := &Status{}
	err := c.makeRequest(ctx, request, resp, http.MethodGet, helper.JoinUrl(c.baseAddress, status))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, method, address string) error {

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("HERE??????????///")
		return err
	}

	req, err := http.NewRequest(method, address, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

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
