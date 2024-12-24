package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"testStand/internal/acquirer/helper"
)

const (
	offer     = "offer/external"
	signature = "user/generate-signature-key"
	api       = "auth/login"
)

type Client struct {
	baseAddress string
	client      *http.Client
	login       Login
}

func NewClient(ctx context.Context, login Login, baseAddress string) *Client {
	return &Client{
		baseAddress: baseAddress,
		client:      http.DefaultClient,
		login: Login{
			Email:    login.Email,
			Password: login.Password,
		},
	}
}

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
	apiKey, err := c.GetToken()
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

func (c *Client) GetToken() (string, error) {
	jsonData, err := json.Marshal(c.login)
	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(jsonData)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, api), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiRes TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiRes); err != nil {
		return "", err
	}

	return apiRes.AccessToken, nil
}

func (c *Client) Signature() (string, error) {

	apiKey, err := c.GetToken()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, signature), nil)
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

	var signRes SignResponse
	if err := json.NewDecoder(resp.Body).Decode(&signRes); err != nil {
		return "", err
	}

	return signRes.SignatureKey, nil
}
