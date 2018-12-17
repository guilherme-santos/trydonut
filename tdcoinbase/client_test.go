package tdcoinbase_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/guilherme-santos/trydonut/tdcoinbase"
)

func assert(t *testing.T, expected, received interface{}) {
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %q but got %q", expected, received)
	}
}

func TestClientTicker(t *testing.T) {
	cfg := tdcoinbase.Config{
		Key:        "my-key",
		Secret:     "bXktc2VjcmV0",
		Passphrase: "my-passphrase",
	}
	testStarted := time.Now().UTC().Unix()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert(t, "GET", req.Method)
		assert(t, "/products/BTC-USD/ticker", req.RequestURI)
		assert(t, cfg.Key, req.Header.Get("CB-ACCESS-KEY"))
		assert(t, cfg.Passphrase, req.Header.Get("CB-ACCESS-PASSPHRASE"))

		timestamp, err := strconv.ParseInt(req.Header.Get("CB-ACCESS-TIMESTAMP"), 10, 64)
		if err != nil {
			t.Errorf("Expected no errors but got: %s", err.Error())
		}

		if timestamp < testStarted {
			t.Error("Expected timestamp to be after test started but it's before")
		}

		receivedSign, _ := base64.StdEncoding.DecodeString(req.Header.Get("CB-ACCESS-SIGN"))
		secret, _ := base64.StdEncoding.DecodeString(cfg.Secret)

		signature := hmac.New(sha256.New, secret)
		signature.Write([]byte(req.Header.Get("CB-ACCESS-TIMESTAMP") + "GET" + "/products/BTC-USD/ticker"))

		if !hmac.Equal(receivedSign, signature.Sum(nil)) {
			t.Error("Signature is not valid")
		}

		fmt.Fprint(w, `{
			"trade_id": 4729088,
			"price": "333.99",
			"size": "0.193",
			"bid": "333.98",
			"ask": "333.99",
			"volume": "5957.11914015",
			"time": "2015-11-14T20:46:03.511254Z"
		  }`)
	}))
	defer ts.Close()

	cfg.URL = ts.URL
	c := tdcoinbase.NewClient(cfg)

	ticker, err := c.Ticker("BTC-USD")
	if err != nil {
		t.Errorf("Expected no errors but got: %s", err.Error())
	}

	assert(t, ticker.TradeID, 4729088)
	assert(t, ticker.Price, "333.99")
	assert(t, ticker.Size, "0.193")
	assert(t, ticker.Bid, "333.98")
	assert(t, ticker.Ask, "333.99")
	assert(t, ticker.Volume, "5957.11914015")
	assert(t, ticker.Time.Format(time.RFC3339Nano), "2015-11-14T20:46:03.511254Z")
}

func TestClientTicker_WithError(t *testing.T) {
	cfg := tdcoinbase.Config{
		Key:        "my-key",
		Secret:     "bXktc2VjcmV0",
		Passphrase: "my-passphrase",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{
			"message": "Invalid Price"
		}`)
	}))
	defer ts.Close()

	cfg.URL = ts.URL
	c := tdcoinbase.NewClient(cfg)

	_, err := c.Ticker("BTC-USD")
	if err == nil {
		t.Errorf("Expected an error")
	}

	assert(t, "tdcoinbase: Bad Request - Invalid Price", err.Error())
}

func TestClientPlaceOrder(t *testing.T) {
	cfg := tdcoinbase.Config{
		Key:        "my-key",
		Secret:     "bXktc2VjcmV0",
		Passphrase: "my-passphrase",
	}
	order := tdcoinbase.LimitOrderRequest{}
	testStarted := time.Now().UTC().Unix()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert(t, "POST", req.Method)
		assert(t, "/orders", req.RequestURI)
		assert(t, cfg.Key, req.Header.Get("CB-ACCESS-KEY"))
		assert(t, cfg.Passphrase, req.Header.Get("CB-ACCESS-PASSPHRASE"))

		timestamp, err := strconv.ParseInt(req.Header.Get("CB-ACCESS-TIMESTAMP"), 10, 64)
		if err != nil {
			t.Errorf("Expected no errors but got: %s", err.Error())
		}

		if timestamp < testStarted {
			t.Error("Expected timestamp to be after test started but it's before")
		}

		receivedSign, _ := base64.StdEncoding.DecodeString(req.Header.Get("CB-ACCESS-SIGN"))
		secret, _ := base64.StdEncoding.DecodeString(cfg.Secret)

		signature := hmac.New(sha256.New, secret)
		signature.Write([]byte(req.Header.Get("CB-ACCESS-TIMESTAMP") + "POST" + "/orders"))
		io.Copy(signature, req.Body)

		if !hmac.Equal(receivedSign, signature.Sum(nil)) {
			t.Error("Signature is not valid")
		}

		fmt.Fprint(w, `{
			"id": "d0c5340b-6d6c-49d9-b567-48c4bfca13d2",
			"price": "0.10000000",
			"size": "0.01000000",
			"product_id": "BTC-USD",
			"side": "buy",
			"stp": "dc",
			"type": "limit",
			"time_in_force": "GTC",
			"post_only": false,
			"created_at": "2016-12-08T20:02:28.53864Z",
			"fill_fees": "0.0000000000000000",
			"filled_size": "0.00000000",
			"executed_value": "0.0000000000000000",
			"status": "pending",
			"settled": false
		}`)
	}))
	defer ts.Close()

	cfg.URL = ts.URL
	c := tdcoinbase.NewClient(cfg)

	newOrder, err := c.PlaceOrder(order)
	if err != nil {
		t.Errorf("Expected no errors but got: %s", err.Error())
	}

	assert(t, newOrder.ID, "d0c5340b-6d6c-49d9-b567-48c4bfca13d2")
	assert(t, newOrder.Price, "0.10000000")
	assert(t, newOrder.Size, "0.01000000")
	assert(t, newOrder.ProductID, "BTC-USD")
	assert(t, newOrder.Side, tdcoinbase.Buy)
	assert(t, newOrder.Stp, "dc")
	assert(t, newOrder.Type, tdcoinbase.Limit)
	assert(t, newOrder.TimeInForce, "GTC")
	assert(t, newOrder.PostOnly, false)
	assert(t, newOrder.CreatedAt.Format(time.RFC3339Nano), "2016-12-08T20:02:28.53864Z")
	assert(t, newOrder.FillFees, "0.0000000000000000")
	assert(t, newOrder.FilledSize, "0.00000000")
	assert(t, newOrder.ExecutedValue, "0.0000000000000000")
	assert(t, newOrder.Status, "pending")
	assert(t, newOrder.Settled, false)
}
