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
	"strconv"
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

	response := &CC5Response{}

	err = xml.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) Auth(address, method string, r *CC5Request, authData url.Values) error {

	hash := c.CreateHash(authData)
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

	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	params, err := url.ParseQuery(string(res))
	if err != nil {
		return err
	}

	mdStatus, err := strconv.Atoi(params.Get("mdStatus"))
	if err != nil {
		return err
	}

	if mdStatus > 1 && mdStatus <= 4 {
		switch mdStatus {
		case 0:
			return errors.New("Authentication failed")
		case 6:
			return errors.New("No payments should be made")
		default:
			return errors.New("MPI fallback")
		}
	}

	r.Pan = params.Get("md")
	r.PayerTxnId = params.Get("xid")
	r.CAVV = params.Get("cavv")
	r.ECI = params.Get("eci")

	return nil
}

func (c *Client) CreateHash(authData url.Values) string {
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

	sum := helper.GenerateHash(sha512.New(), []byte(rawHash))
	hash := base64.StdEncoding.EncodeToString(sum)

	return hash
}
