package api

import (
	"bytes"         // встроенный пакет
	"context"       // встроенный пакет
	"crypto/sha256" // встроенный пакет
	"encoding/hex"  // встроенный пакет
	"encoding/json" // встроенный пакет
	"net/http"      // встроенный пакет

	"testStand/internal/acquirer/helper" // наш импорт
)

type Client struct {
	baseAddress string
	apiKey      string
	client      *http.Client
}

const (
	payout = "/api/v1/withdraw"
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
