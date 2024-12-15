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
	client      *http.Client
	email       string
	password    string
}

const (
	offerEndpoint = "v1/offer/external"
	signEndpoint  = "v1/user/generate-signature-key"
	apiEndpoint   = "v1/auth/login"
)

var Login = map[string]string{
	"email":    "buyer@dev.alpex.app",
	"password": "dev",
}

func NewClient(ctx context.Context, email, password, baseAddress string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
		email:       email,
		password:    password,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *Request) (*Response, error) {
	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, offerEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, endpoint string) error {
	apiKey, err := c.GetApi()
	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

func (c *Client) GetApi() (string, error) {
	var Login = map[string]string{
		"email":    c.email,
		"password": c.password,
	}

	jsonData, err := json.Marshal(Login)
	if err != nil {
		return "", err
	}

	var res map[string]string

	body := bytes.NewBuffer(jsonData)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, apiEndpoint), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	return res["access_token"], nil
}

func (c *Client) GetSign() (string, error) {

	var res map[string]string
	apiKey, err := c.GetApi()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, signEndpoint), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	return res["signature_key"], nil
}
