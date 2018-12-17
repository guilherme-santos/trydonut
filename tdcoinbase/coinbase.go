package tdcoinbase

import (
	"fmt"
	"net/http"
	"time"
)

var (
	productTicketEndpoint = "/products/%s/ticker"
	orderEndpoint         = "/orders"
)

// Coinbase responses
type (
	Ticker struct {
		TradeID int       `json:"trade_id"`
		Price   string    `json:"price"`
		Size    string    `json:"size"`
		Bid     string    `json:"bid"`
		Ask     string    `json:"ask"`
		Volume  string    `json:"volume"`
		Time    time.Time `json:"time"`
	}

	Order struct {
		ID            string    `json:"id"`
		Price         string    `json:"price"`
		Size          string    `json:"size"`
		ProductID     string    `json:"product_id"`
		Side          OrderSide `json:"side"`
		Stp           string    `json:"stp"`
		Type          OrderType `json:"type"`
		TimeInForce   string    `json:"time_in_force"`
		PostOnly      bool      `json:"post_only"`
		CreatedAt     time.Time `json:"created_at"`
		FillFees      string    `json:"fill_fees"`
		FilledSize    string    `json:"filled_size"`
		ExecutedValue string    `json:"executed_value"`
		Status        string    `json:"status"`
		Settled       bool      `json:"settled"`
	}
)

// Error represents and status code error different of 2XX
type Error struct {
	StatusCode int
	Message    string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("tdcoinbase: %s - %s", http.StatusText(e.StatusCode), e.Message)
}
