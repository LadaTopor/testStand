package api

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"testStand/internal/acquirer/helper"
)

const (
	payment = "/fim/api"
	auth    = "/fim/est3dgate"
)

type Client struct {
	baseAddress string
	storeKey    string
	client      *http.Client
}

func NewClient(ctx context.Context, storeKey, baseAddress string) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		storeKey:    storeKey,
		client:      client,
	}
}

func (c *Client) MakePayment(ctx context.Context, request *CC5Request, data url.Values) (*PaymentResponse, error) {

	err := c.Auth(helper.JoinUrl(c.baseAddress, auth), http.MethodPost, request, data)
	if err != nil {
		return nil, err
	}

	body, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, payment), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &PaymentResponse{}
	err = xml.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) Auth(address, method string, request *CC5Request, data url.Values) error {
	hash := c.GenerateHash(data)
	data.Set("hash", hash)
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

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	param, err := url.ParseQuery(string(res))
	if err != nil {
		return err
	}

	request.Pan = param.Get("md")
	request.PayerTxnId = param.Get("xid")
	request.CAVV = param.Get("cavv")
	request.ECI = param.Get("eci")

	if request.Pan == "" || request.PayerTxnId == "" || request.CAVV == "" || request.ECI == "" {
		return errors.New("missing required 3D-Secure parameters")
	}

	return nil
}
func (c *Client) GenerateHash(data url.Values) string {
	keys := make([]string, 0, len(data))
	for key := range data {
		if key != "encoding" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})

	hashValues := make([]string, len(keys))
	for i, key := range keys {
		hashValues[i] = data.Get(key)
	}

	rawHash := strings.Join(hashValues, "|") + "|" + c.storeKey

	hash := base64.StdEncoding.EncodeToString(helper.GenerateHash(sha512.New(), []byte(rawHash)))

	return hash
}
