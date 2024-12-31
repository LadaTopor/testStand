package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httputil"
	"net/url"

	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	client      *http.Client
}

const (
	payment = "/fim/api"
	auth    = "/fim/est3dgate"
)

func NewClient(ctx context.Context, baseAddress string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *CC5Request) (*CC5Response, error) {

	resp := &CC5Response{}

	c.Auth(helper.JoinUrl(c.baseAddress, auth), http.MethodPost, request)

	err := c.makeRequest(ctx, request, resp, http.MethodPost, helper.JoinUrl(c.baseAddress, payment))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// makeRequest
func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, method, address string) error {

	body, err := xml.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, address, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil // error EOF, because invalid url
	}

	return nil
}

func (c *Client) Auth(address, method string, request *CC5Request) error {
	data := url.Values{}
	data.Set("clientid", request.ClientId)
	data.Set("storetype", "3d_pay")
	data.Set("trantype", "Auth")
	data.Set("oid", request.OrderId)
	data.Set("Amount", string(request.Total))
	data.Set("currency", "944")
	data.Set("pan", request.CardNumber)
	data.Set("Ecom_Payment_Card_ExpDate_Year", request.CardYear)
	data.Set("Ecom_Payment_Card_ExpDate_Month", request.CardMonth)
	data.Set("cv2", "123")
	data.Set("encoding", "utf-8")
	data.Set("hashAlgorithm", request.HashAlgorithm)

	req, err := http.NewRequest(method, address, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	res, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	params, err := url.ParseQuery(string(res))
	if err != nil {
		return err
	}
	request.CardNumber = params["md"][0]
	request.PayerTxnId = params["xid"][0]
	request.CAVV = params["cavv"][0]
	request.ECI = params["eci"][0]

	return nil
}
