package api

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"

	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	apiKey      string
	client      *http.Client
}

const (
	Payment   = "est3dgate"
	Finish3ds = "api"
)

func NewClient(ctx context.Context, baseAddress string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, formData url.Values) (url.Values, error) {
	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, Payment), strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsedData, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil, err
	}

	return parsedData, nil
}

// makeRequest
func (c *Client) MakeFinish3ds(ctx context.Context, request *Request) (*Response, error) {
	xmlData, err := xml.MarshalIndent(request, "", "  ")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, Finish3ds), strings.NewReader(string(xmlData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response *Response
	err = xml.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, err
	}

	return response, err
}
