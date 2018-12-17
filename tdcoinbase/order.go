package tdcoinbase

import (
	"errors"
	"fmt"
)

type (
	// OrderRequest could be LimitOrderRequest or MarketOrderRequest.
	OrderRequest interface{}

	OrderType string
	OrderSide string
	OrderStop string
)

var (
	Limit  OrderType = "limit"
	Market OrderType = "market"

	Buy  OrderSide = "buy"
	Sell OrderSide = "sell"

	NoStop    OrderStop = ""
	StopLoss  OrderStop = "loss"
	StopEntry OrderStop = "entry"
)

// commonOrderRequest will be used internaly by LimitOrderRequest and MarketOrderRequest.
type commonOrderRequest struct {
	ClientOID string    `json:"client_oid,omitempty"`
	Type      OrderType `json:"type"`
	Side      OrderSide `json:"side"`
	ProductID string    `json:"product_id"`
	Stp       string    `json:"stp,omitempty"`
	Stop      OrderStop `json:"stop,omitempty"`
	StopPrice string    `json:"stop_price,omitempty"`
}

func (o *commonOrderRequest) validate() error {
	switch o.Type {
	case Limit, Market:
	default:
		return fmt.Errorf("tdcoinbase: order type '%s' invalid or unknown", o.Type)
	}

	switch o.Side {
	case Buy, Sell:
	default:
		return fmt.Errorf("tdcoinbase: order side '%s' invalid or unknown", o.Side)
	}

	if o.ProductID == "" {
		return errors.New("tdcoinbase: order requires ProductID")
	}

	if o.Stop == NoStop {
		// case this order doesn't have stop, let's clear StopPrice
		o.StopPrice = ""
	} else {
		// case this order has stop, let's make sure we have a StopPrice
		if o.StopPrice == "" {
			return fmt.Errorf("tdcoinbase: order stop %s need to have a StopPrice", o.Stop)
		}
	}

	return nil
}

type LimitOrderRequest struct {
	commonOrderRequest `json:",inline"`
	Price              string `json:"price"`
	Size               string `json:"size"`
	TimeInForce        string `json:"time_in_force"`
	CancelAfter        string `json:"cancel_after,omitempty"`
	PostOnly           bool   `json:"post_only,omitempty"`
}

func (o *LimitOrderRequest) validate() error {
	o.Type = Limit

	err := o.commonOrderRequest.validate()
	if err != nil {
		return err
	}

	if o.Price == "" {
		return errors.New("tdcoinbase: LimitOrder requires Price")
	}
	if o.Size == "" {
		return errors.New("tdcoinbase: LimitOrder requires Size")
	}

	switch o.TimeInForce {
	case "":
		o.TimeInForce = "GTC"
	case "GTC", "GTT":
	case "IOC", "FOK":
		if o.PostOnly {
			return fmt.Errorf("tdcoinbase: PostOnly flag cannot be used with TimeInForce %s", o.TimeInForce)
		}
	default:
		return fmt.Errorf("tdcoinbase: TimeInForce '%s' invalid or unknown", o.TimeInForce)
	}

	switch o.CancelAfter {
	case "", "min", "hour", "day":
	default:
		return fmt.Errorf("tdcoinbase: CancelAfter '%s' invalid or unknown", o.CancelAfter)
	}

	return nil
}

type MarketOrderRequest struct {
	commonOrderRequest `json:",inline"`
	Size               string `json:"size"`
	Funds              string `json:"funds"`
}

func (o *MarketOrderRequest) validate() error {
	o.Type = Market

	err := o.commonOrderRequest.validate()
	if err != nil {
		return err
	}

	if o.Size == "" && o.Funds == "" {
		return errors.New("tdcoinbase: MarketOrder requires Size or Funds")
	}

	return nil
}
