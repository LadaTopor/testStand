package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	apiKey      string
	client      *http.Client
}

const (
	Reconciled = "executed"
	Decline    = "cancelled"
)

const (
	payout = "/api/v1/withdraw"
	status = "status"
)

func NewClient(ctx context.Context, baseAddress, apiKey string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		apiKey:      apiKey,
		client:      client,
	}
}

// MakePayout
func (c *Client) MakePayout(ctx context.Context, request *Request, secretKey string) (*Response, error) {
	sign := CreateSign(request.MerchId + request.CardData.CardNumber + request.Amount + secretKey)
	fmt.Println("--------SIGN1", sign)

	resp := &Response{}
	err := c.makeRequest(ctx, request, resp, sign, payout)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func CreateSign(input string) string {
	sum := helper.GenerateHash(sha256.New(), []byte(input))
	return hex.EncodeToString(sum)
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, sign, endpoint string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	fmt.Println("-------Endpoint", endpoint)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}

	log.Println("------API", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sign", sign)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	r, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(r))
	fmt.Println("----------------------------------------------------------------------------------")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	r, _ = httputil.DumpResponse(resp, true)
	fmt.Println(string(r))

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}

	return nil
}

func (c *Client) GetStatus(ctx context.Context, request *Request) (*Status, error) {
	resp := &Status{}
	err := c.makeRequest(ctx, request, resp, http.MethodGet, helper.JoinUrl(c.baseAddress, status))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
