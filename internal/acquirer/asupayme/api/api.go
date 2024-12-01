package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

const (
	payment = "api/v1/withdraw"
	status  = "api/order"
)

type Client struct {
	baseAddress string
	apiKey      string
	MerchantID  string
	secretKey   string
	client      *http.Client
}

func NewClient(ctx context.Context, baseAddress, apiKey, merchantID, secretKey string) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		apiKey:      apiKey,
		MerchantID:  merchantID,
		secretKey:   secretKey,
		client:      client,
	}
}

func (c *Client) GenerateSignature(cardNumber, amount string) string {
	data := fmt.Sprintf("%s%s%s%s", c.MerchantID, cardNumber, amount, c.secretKey)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (c *Client) MakeWithdraw(ctx context.Context, request *WithdrawRequest) (*Response, error) {

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s", c.baseAddress, payment), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	r, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(r))
	fmt.Println("----------------------------------------------------------------------------------")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r, _ = httputil.DumpResponse(resp, true)
	fmt.Println(string(r))

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("failed with error: %s", response.Error)
	}

	return &response, nil
}
