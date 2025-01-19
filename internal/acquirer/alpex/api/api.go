package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"testStand/internal/acquirer/helper"
)

// Endpoints
const (
	PaymentEndpoint = "/v1/offer/external"
	PayoutEndpoint  = "/v1/offer/external"
	SignEndpoint    = "/v1/user/generate-signature-key"
)

type Client struct {
	baseAddress string
	login       string
	ID          string
	pass        string
	apiKey      string
	client      *http.Client
}

// NewClient
func NewClient(ctx context.Context, baseAddress, apiKey string, login string, id string, pass string) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		login:       login,
		ID:          id,
		pass:        pass,
		apiKey:      apiKey,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *Request) (*PaymentResponse, error) {
	signKey, err := c.makeSignatureKey(ctx)
	if err != nil {
		return nil, err
	}
	sign := createSign(request, signKey.SignatureKey)
	if err != nil {
		return nil, err
	}

	request.Sign = sign

	resp := &PaymentResponse{}
	err = c.makeRequest(ctx, request, resp, PaymentEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakePayout
func (c *Client) MakePayout(ctx context.Context, request *Request) (*Response, error) {
	signKey, err := c.makeSignatureKey(ctx)
	if err != nil {
		return nil, err
	}
	sign := createSign(request, signKey.SignatureKey)
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

	//if resp.StatusCode >= http.StatusInternalServerError {
	//	return errors.New("declined state due to network or internal error")
	//}

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}
	return nil
}

func (c *Client) makeSignatureKey(ctx context.Context) (*UserSign, error) {
	user := User{
		email:    c.login,
		password: c.pass,
	}
	StrKey := &UserSign{}
	err := c.makeRequest(ctx, user, StrKey, SignEndpoint)
	if err != nil {
		return nil, err
	}

	return StrKey, nil
}
