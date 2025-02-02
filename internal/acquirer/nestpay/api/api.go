package api

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/go-playground/form"
	"github.com/google/go-querystring/query"
	"testStand/internal/acquirer/helper"
)

// Endpoints
const (
	PaymentEndpoint = "/fim/est3dgate"
	StatusEndpoint  = "/fim/api"
)

type Client struct {
	baseAddress string
	name        string
	password    string
	storeKey    string
	client      *http.Client
}

func NewClient(ctx context.Context, baseAddress, name string, password string, key string) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		name:        name,
		password:    password,
		storeKey:    key,
		client:      client,
	}
}

func (c *Client) MakePayment(ctx context.Context, request *Request) (*PaymentResponse, error) {

	resp := &PaymentResponse{}
	err := c.makeRequest(ctx, request, resp, PaymentEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, endpoint string) error {

	body, err := query.Values(payload)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewBufferString(body.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	decoder := form.NewDecoder()

	defer resp.Body.Close()

	Body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	values, err := url.ParseQuery(string(Body))
	if err != nil {
		return nil
	}

	err = decoder.Decode(&outResponse, values)
	if err != nil {
		return nil
	}

	return nil
}

func (c *Client) CheckStatus(ctx context.Context, request *StatusRequest) (*StatusResponse, error) {

	resp := &StatusResponse{}
	err := c.makeStatusRequest(ctx, request, resp, StatusEndpoint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) makeStatusRequest(ctx context.Context, payload, outResponse any, endpoint string) error {

	xmlBody, err := xml.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(xmlBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	err = xml.Unmarshal(body, &outResponse)
	if err != nil {
		return nil
	}

	return nil
}

func (c *Client) CreateSign(request *Request, storeKey string) string {

	req, err := query.Values(request)
	if err != nil {
		return "bababa bebebe"
	}

	var sortedKeys []string
	for key := range req {
		lowerKey := strings.ToLower(key)
		if lowerKey != "encoding" && lowerKey != "hash" {
			sortedKeys = append(sortedKeys, key)
		}
	}

	sort.Strings(sortedKeys)

	var hashDataBuilder strings.Builder

	for _, key := range sortedKeys {
		value := req.Get(key)
		hashDataBuilder.WriteString(value)
		hashDataBuilder.WriteString("|")
	}

	hashDataBuilder.WriteString(storeKey)

	hashVal := hashDataBuilder.String()
	hash := sha512.Sum512([]byte(hashVal))
	OkSign := base64.StdEncoding.EncodeToString(hash[:])

	return OkSign
}
