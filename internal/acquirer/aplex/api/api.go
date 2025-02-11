package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"testStand/internal/acquirer/helper"
)

type Client struct {
	client      *http.Client
	login       string
	password    string
	baseAddress string
}

const (
	loginEndpoint    = "auth/login"
	signEndpoint     = "user/generate-signature-key"
	externalEndpoint = "offer/external"
)

func NewClient(ctx context.Context, baseAddress, login, password string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		client:      client,
		login:       login,
		password:    password,
		baseAddress: baseAddress,
	}
}

// CreateOffer
func (c *Client) CreateOffer(ctx context.Context, request *Request) (*Response, error) {
	accessToken, err := c.Login()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, externalEndpoint), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) Sign(callback Callback) (bool, error) {
	accessToken, err := c.Login()
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, signEndpoint), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	response := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, err
	}

	sign := createSign(&callback, response["signature_key"])
	return callback.Signature != sign, nil
}

func (c *Client) Login() (string, error) {
	cred := map[string]string{
		"email":    c.login,
		"password": c.password,
	}
	body, err := json.Marshal(cred)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(body)
	post, err := c.client.Post(helper.JoinUrl(c.baseAddress, loginEndpoint), "application/json", reader)
	if err != nil {
		return "", err
	}
	defer post.Body.Close()

	response := map[string]string{}
	err = json.NewDecoder(post.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	if len(response["access_token"]) == 0 {
		return "", err
	}
	return response["access_token"], nil
}
