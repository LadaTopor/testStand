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

	"testStand/internal/acquirer/helper"

	"github.com/go-playground/form"
	"github.com/google/go-querystring/query"
)

// Endpoints
const (
	PaymentEndpoint = "/fim/est3dgate"
	StatusEndpoint  = "/fim/api"
)

type Client struct {
	baseAddress string
	currency    string
	client      *http.Client
}

func NewClient(ctx context.Context, baseAddress, currency string) *Client {

	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		currency:    currency,
		client:      client,
	}
}

func (c *Client) MakePayment(ctx context.Context, request *Request) (*PaymentResponse, error) {

	resp := &PaymentResponse{}

	body, err := query.Values(request)

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, PaymentEndpoint), bytes.NewBufferString(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := form.NewDecoder()

	defer response.Body.Close()

	byteRespBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(string(byteRespBody))
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(&resp, values)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) CheckStatus(ctx context.Context, request *StatusRequest) (*StatusResponse, error) {

	resp := &StatusResponse{}

	xmlBody, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, StatusEndpoint), bytes.NewReader(xmlBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/xml")

	response, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	err = xml.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) CreateSign(request *Request, storeKey string) (string, error) {

	req, err := query.Values(request)
	if err != nil {
		return "bababa bebebe", err
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
	hashString := base64.StdEncoding.EncodeToString(hash[:])

	return hashString, err
}
