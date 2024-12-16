package api

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
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
	client      *http.Client
}

// NewClient
func NewClient(ctx context.Context, baseAddress, apiKey string, timeout *int) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		apiKey:      apiKey,
		client:      client,
	}
}

// MakeWithdraw
func (c *Client) MakeWithdraw(ctx context.Context, request Request) (*Response, error) {

	sign, err := createSign(request, c.apiKey, strconv.FormatInt(request.Amount, 10))
	if err != nil {
		return nil, err
	}
	request.Sign = sign

	resp := &Response{}
	err = c.makeRequest(ctx, request, resp, PayoutEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckStatus
func (c *Client) CheckStatus(ctx context.Context, request StatusRequest) (*StatusResponse, error) {

	sign, err := createSign(request, c.apiKey, "")
	if err != nil {
		return nil, err
	}
	request.Sign = sign

	resp := &StatusResponse{}
	err = c.makeRequest(ctx, request, resp, StatusEndpoint)
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
