package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"testStand/internal/acquirer/helper"
)

// Endpoints
const (
	PaymentEndpoint = "/v1/offer/external"
	PayoutEndpoint  = "/v1/offer/external"
	SignEndpoint    = "/v1/user/generate-signature-key"
	BearerEndpoint  = "/v1/auth/login"
)

type Client struct {
	baseAddress string
	login       string
	pass        string
	client      *http.Client
}

// NewClient
func NewClient(ctx context.Context, baseAddress, login string, pass string) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		login:       login,
		pass:        pass,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *Request) (*Response, error) {

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, PaymentEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakePayout
func (c *Client) MakePayout(ctx context.Context, request *Request) (*Response, error) {

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, PayoutEndpoint)
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

	Token := c.MakeBearerToken(ctx)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Token)

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

func (c *Client) MakeBearerToken(ctx context.Context) string {
	body, err := json.Marshal(User{
		Email:    c.login,
		Password: c.pass,
	})

	if err != nil {
		return "Invalid BearerBody"
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, BearerEndpoint), bytes.NewReader(body))
	if err != nil {
		return ""
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "_"
	}

	Token := &UserSignToken{}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&Token)
	if err != nil {
		return "nil"
	}

	return Token.AccessToken
}

func (c *Client) MakeSignatureKey(ctx context.Context) (*UserSignToken, error) {
	user := User{
		Email:    c.login,
		Password: c.pass,
	}
	StrKey := &UserSignToken{}
	err := c.makeRequest(ctx, user, StrKey, SignEndpoint)
	if err != nil {
		return nil, err
	}

	return StrKey, nil
}

func (c *Client) CreateSign(id, status, key string) string {
	offerString := fmt.Sprintf("id=%s\nstatus=%s", id, status)
	hmac := hmac.New(sha256.New, []byte(key))
	hmac.Write([]byte(offerString))
	sign := hex.EncodeToString(hmac.Sum(nil))

	return sign
}
