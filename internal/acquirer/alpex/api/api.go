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
	offer   = "v1/offer/external"
	signUrl = "http://147.45.152.131:8022/v1/user/generate-signature-key"
	apiUrl  = "http://147.45.152.131:8022/v1/auth/login"
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
	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, offer)
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

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}

	return nil
}

func GetApi() string {
	var jsonData = []byte(`{"email": "buyer@dev.alpex.app", "password": "dev"}`)

	type response struct {
		Key string `json:"access_token"`
	}

	res := &response{}

	body := bytes.NewBuffer(jsonData)

	resp, _ := http.Post(apiUrl, "application/json", body)

	defer resp.Body.Close()

	_ = json.NewDecoder(resp.Body).Decode(&res)

	return res.Key
}

func GetSign() string {
	type response struct {
		Sign string `json:"signature_key"`
	}

	res := &response{}

	req, _ := http.NewRequest(http.MethodPost, signUrl, nil)

	req.Header.Set("Authorization", "Bearer "+GetApi())

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()

	_ = json.NewDecoder(resp.Body).Decode(&res)

	return res.Sign
}
