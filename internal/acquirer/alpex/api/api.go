package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	email       string
	password    string
	client      *http.Client
}

const (
	auth      = "/auth/login"
	offer     = "/offer/external"
	signtaure = "/user/generate-signature-key"
)

func NewClient(baseAddress, email, password string) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
		email:       email,
		password:    password,
	}
}

func (c *Client) makeAuth() (string, error) {
	authReq := &AuthRequest{
		Email:    c.email,
		Password: c.password,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return "", err
	}

	response, err := http.Post(helper.JoinUrl(c.baseAddress, auth), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	var respBody map[string]string
	err = json.NewDecoder(response.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}

	return respBody["access_token"], nil
}

func (c *Client) GetSignatureKey() (string, error) {
	apiKey, err := c.makeAuth()
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, signtaure), nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)

	response, err := c.client.Do(request)
	if err != nil {
		return "", err
	}

	r, _ := httputil.DumpResponse(response, true)
	fmt.Println(string(r))

	var respBody map[string]string
	err = json.NewDecoder(response.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}

	return respBody["signature_key"], nil
}

func (c *Client) MakeOffer(req *Request) (*Response, error) {
	apiKey, err := c.makeAuth()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, offer), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	r, _ := httputil.DumpResponse(response, true)
	fmt.Println(string(r))

	resp := &Response{}
	err = json.NewDecoder(response.Body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
