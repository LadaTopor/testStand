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
)

type Client struct {
	baseAddress string
	storeKey    string
	client      *http.Client
}

const (
	payment = "/fim/api"
	auth    = "/fim/est3dgate"
)

func NewClient(ctx context.Context, storeKey, baseAddress string, timeout *int) *Client {
	client := http.DefaultClient
	return &Client{
		baseAddress: baseAddress,
		storeKey:    storeKey,
		client:      client,
	}
}

// MakePayment
func (c *Client) MakePayment(ctx context.Context, request *CC5Request, authData url.Values) (*CC5Response, error) {

	outResponse := &CC5Response{}

	//Первый запрос
	err := c.Auth(helper.JoinUrl(c.baseAddress, auth), http.MethodPost, request, authData)
	if err != nil {
		return nil, err
	}

	//второй запрос
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

	err = xml.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil, err
	}

	return outResponse, nil
}

func (c *Client) Auth(address, method string, r *CC5Request, authData url.Values) error {

	keys := make([]string, 0, len(authData))
	for key := range authData {
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
		hashValues[i] = authData.Get(key)
	}

	rawHash := strings.Join(hashValues, "|") + "|" + c.storeKey

	hash := CreateSign(rawHash)
	authData.Set("hash", hash)

	req, err := http.NewRequest(method, address, bytes.NewBufferString(authData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	params, err := url.ParseQuery(string(res))
	if err != nil {
		return err
	}

	r.Pan = params.Get("md")
	r.PayerTxnId = params.Get("xid")
	r.CAVV = params.Get("cavv")
	r.ECI = params.Get("eci")

	return nil
}

func CreateSign(input string) string {
	sum := helper.GenerateHash(sha512.New(), []byte(input))
	return base64.StdEncoding.EncodeToString(sum)
}
