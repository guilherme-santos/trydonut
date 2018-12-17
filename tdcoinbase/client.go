package tdcoinbase

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var DefaultTimeout = 100 * time.Millisecond

type Config struct {
	URL        string
	Key        string
	Secret     string
	Passphrase string
	Timeout    time.Duration
}

type Coinbase struct {
	httpClient *http.Client
	url        string
	key        string
	secret     []byte
	passphrase string
}

func NewClient(cfg Config) *Coinbase {
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}

	secret, err := base64.StdEncoding.DecodeString(cfg.Secret)
	if err != nil {
		panic(fmt.Sprintf("secret is not a valid base64 string: %s", err.Error()))
	}

	return &Coinbase{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		url:        cfg.URL,
		key:        cfg.Key,
		secret:     secret,
		passphrase: cfg.Passphrase,
	}
}

func (c *Coinbase) injectHeaders(req *http.Request) {
	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	req.Header.Set("CB-ACCESS-KEY", c.key)
	req.Header.Set("CB-ACCESS-PASSPHRASE", c.passphrase)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)

	signature := hmac.New(sha256.New, c.secret)
	signature.Write([]byte(timestamp + req.Method + req.URL.RequestURI()))

	if req.Body != nil {
		var buf bytes.Buffer
		tee := io.TeeReader(req.Body, &buf)

		req.Body = ioutil.NopCloser(&buf)
		io.Copy(signature, tee)
	}

	req.Header.Set("CB-ACCESS-SIGN", base64.StdEncoding.EncodeToString(signature.Sum(nil)))
}

func (c *Coinbase) Ticker(productID string) (Ticker, error) {
	var t Ticker

	req, err := http.NewRequest(http.MethodGet, c.url+fmt.Sprintf(productTicketEndpoint, productID), nil)
	if err != nil {
		return t, err
	}

	c.injectHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return t, err
	}

	err = handleResponse(resp, &t)
	return t, err
}

// PlaceOrder places a new order based in the OrderRequest which need to
// be LimitOrderRequest or MarketOrderRequest otherwise this function will panic.
func (c *Coinbase) PlaceOrder(orderReq OrderRequest) (Order, error) {
	var o Order

	reqBody, _ := json.Marshal(orderReq)

	req, err := http.NewRequest(http.MethodPost, c.url+orderEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return o, err
	}

	c.injectHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return o, err
	}

	err = handleResponse(resp, &o)
	return o, err
}

func handleResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		// some error happend
		err := Error{
			StatusCode: resp.StatusCode,
		}
		json.NewDecoder(resp.Body).Decode(&err)
		return &err
	}

	err := json.NewDecoder(resp.Body).Decode(v)
	return err
}
