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
	auth      = "auth/login"
	offer     = "offer/external"
	signature = "user/generate-signature-key"
)

func NewClient(ctx context.Context, baseAddress, email, password string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		email:       email,
		password:    password,
		client:      client,
	}
}

// MakeOffer
func (c *Client) MakeOffer(ctx context.Context, request *Request) (*Response, error) {
	apiKey, err := c.Auth()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, offer), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &Response{}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err // error EOF, because invalid url
	}

	return response, nil
}

func (c *Client) Auth() (string, error) {
	authCred := map[string]string{
		"email":    c.email,
		"password": c.password,
	}
	body, err := json.Marshal(authCred)
	if err != nil {
		return "", err
	}

	post, err := c.client.Post(helper.JoinUrl(c.baseAddress, auth), "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	defer post.Body.Close()

	response := map[string]string{}
	err = json.NewDecoder(post.Body).Decode(&response)
	if err != nil {
		return "", err // error EOF, because invalid url
	}

	return response["access_token"], err
}

func (c *Client) Sign(id, status, signInCallback string) (bool, error) {
	apiKey, err := c.Auth()
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, signature), nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	response := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, err // error EOF, because invalid url
	}

	sign := CreateSign(id, status, response["signature_key"])

	if signInCallback != sign {
		return false, nil
	}

	return true, nil
}
