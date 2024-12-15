package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	client      *http.Client
}

const (
	offerEndpoint = "v1/offer/external"
	signEndpoint  = "v1/user/generate-signature-key"
	apiEndpoint   = "v1/auth/login"
)

func NewClient(ctx context.Context, baseAddress string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
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
	apiKey, _ := c.GetApi()

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
	jsonData, err := json.Marshal(Login)
	if err != nil {
		log.Fatal(err)
	}

	var res map[string]string

	body := bytes.NewBuffer(jsonData)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, apiEndpoint), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, _ := c.client.Do(req)

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
