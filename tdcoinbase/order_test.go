package tdcoinbase

import (
	"strings"
	"testing"
)

func TestCommonOrderRequestValidate_MandatoryFields(t *testing.T) {
	testCases := []struct {
		Order  commonOrderRequest
		ErrMsg string
	}{
		// Invalid order - empty type
		{Order: commonOrderRequest{
			Type: OrderType(""),
		}, ErrMsg: "tdcoinbase: order type '' invalid or unknown"},
		// Invalid order - invalid type
		{Order: commonOrderRequest{
			Type: OrderType("invalid"),
		}, ErrMsg: "tdcoinbase: order type 'invalid' invalid or unknown"},
		// Invalid order - empty side
		{Order: commonOrderRequest{
			Type: Limit,
			Side: OrderSide(""),
		}, ErrMsg: "tdcoinbase: order side '' invalid or unknown"},
		// Invalid order - invalid side
		{Order: commonOrderRequest{
			Type: Market,
			Side: OrderSide("invalid"),
		}, ErrMsg: "tdcoinbase: order side 'invalid' invalid or unknown"},
		// Invalid order - no product id
		{Order: commonOrderRequest{
			Type: Market,
			Side: Sell,
		}, ErrMsg: "tdcoinbase: order requires ProductID"},
		// Valid order - limit and buy
		{Order: commonOrderRequest{
			Type:      Limit,
			Side:      Buy,
			ProductID: "BTC-USD",
		}},
		// Valid order - market and sell
		{Order: commonOrderRequest{
			Type:      Market,
			Side:      Sell,
			ProductID: "BTC-USD",
		}},
	}

	for _, tc := range testCases {
		err := tc.Order.validate()
		if err != nil && tc.ErrMsg == "" {
			t.Errorf("Expected no errors but got: %s", err.Error())
		}

		if err == nil && tc.ErrMsg != "" {
			t.Errorf("Expected error: %s", tc.ErrMsg)
		}

		if err != nil && !strings.EqualFold(tc.ErrMsg, err.Error()) {
			t.Errorf("Expected error to be %q but got: %s", tc.ErrMsg, err.Error())
		}
	}
}

func TestCommonOrderRequestValidate_StopPriceIsCleared(t *testing.T) {
	o := commonOrderRequest{
		Type:      Limit,
		Side:      Buy,
		ProductID: "BTC-USD",
		Stop:      NoStop,
		StopPrice: "0.10000000",
	}

	err := o.validate()
	if err != nil {
		t.Errorf("Expected no errors but got: %s", err.Error())
	}

	if o.StopPrice != "" {
		t.Errorf("Expected StopPrice to be cleared because NoStop does not require it")
	}
}

func TestCommonOrderRequestValidate_StopPriceRequiredWhenStopIsSetted(t *testing.T) {
	o := commonOrderRequest{
		Type:      Limit,
		Side:      Buy,
		ProductID: "BTC-USD",
		Stop:      StopLoss,
	}

	err := o.validate()
	if err == nil {
		t.Error("Expected error when set Stop but not StopPrice")
	}
}

func TestLimitOrderRequestValidate_AlwaysOverrideType(t *testing.T) {
	o := LimitOrderRequest{
		commonOrderRequest: commonOrderRequest{
			Type:      Market,
			Side:      Buy,
			ProductID: "BTC-USD",
		},
		Price: "0.100",
		Size:  "0.01",
	}

	err := o.validate()
	if err != nil {
		t.Errorf("Expected no errors but got: %s", err.Error())
	}

	if o.Type != Limit {
		t.Errorf("Expected Type to be Limit but got: %q", o.Type)
	}
}

func TestLimitOrderRequestValidate_DefaultTimeInForce(t *testing.T) {
	o := LimitOrderRequest{
		commonOrderRequest: commonOrderRequest{
			Type:      Market,
			Side:      Buy,
			ProductID: "BTC-USD",
		},
		Price: "0.100",
		Size:  "0.01",
	}

	err := o.validate()
	if err != nil {
		t.Errorf("Expected no errors but got: %s", err.Error())
	}

	if o.TimeInForce != "GTC" {
		t.Errorf("Expected TimeInForce to be \"GTC\" but got: %q", o.TimeInForce)
	}
}

func TestLimitOrderRequestValidate_MandatoryFields(t *testing.T) {
	testCases := []struct {
		Order  LimitOrderRequest
		ErrMsg string
	}{
		// Invalid order - missing price
		{
			Order: LimitOrderRequest{
				commonOrderRequest: commonOrderRequest{
					Side:      Sell,
					ProductID: "BTC-USD",
				},
			},
			ErrMsg: "tdcoinbase: LimitOrder requires Price",
		},
		// Invalid order - missing size
		{
			Order: LimitOrderRequest{
				commonOrderRequest: commonOrderRequest{
					Side:      Sell,
					ProductID: "BTC-USD",
				},
				Price: "0.100",
			},
			ErrMsg: "tdcoinbase: LimitOrder requires Size",
		},
		// Valid order
		{
			Order: LimitOrderRequest{
				commonOrderRequest: commonOrderRequest{
					Side:      Sell,
					ProductID: "BTC-USD",
				},
				Price: "0.100",
				Size:  "0.01",
			},
		},
	}

	for _, tc := range testCases {
		err := tc.Order.validate()
		if err != nil && tc.ErrMsg == "" {
			t.Errorf("Expected no errors but got: %s", err.Error())
		}

		if err == nil && tc.ErrMsg != "" {
			t.Errorf("Expected error: %s", tc.ErrMsg)
		}

		if err != nil && !strings.EqualFold(tc.ErrMsg, err.Error()) {
			t.Errorf("Expected error to be %q but got: %s", tc.ErrMsg, err.Error())
		}
	}
}

func TestMarketOrderRequestValidate_AlwaysOverrideType(t *testing.T) {
	o := MarketOrderRequest{
		commonOrderRequest: commonOrderRequest{
			Type:      Limit,
			Side:      Sell,
			ProductID: "BTC-USD",
		},
		Size: "0.01",
	}

	err := o.validate()
	if err != nil {
		t.Errorf("Expected no errors but got: %s", err.Error())
	}

	if o.Type != Market {
		t.Errorf("Expected Type to be Market but got: %q", o.Type)
	}
}

func TestMarketOrderRequestValidate_MandatoryFields(t *testing.T) {
	testCases := []struct {
		Order  MarketOrderRequest
		ErrMsg string
	}{
		// Invalid order - size and funds
		{
			Order: MarketOrderRequest{
				commonOrderRequest: commonOrderRequest{
					Side:      Buy,
					ProductID: "BTC-USD",
				},
			},
			ErrMsg: "tdcoinbase: MarketOrder requires Size or Funds",
		},
		// Valid order - with size
		{
			Order: MarketOrderRequest{
				commonOrderRequest: commonOrderRequest{
					Side:      Buy,
					ProductID: "BTC-USD",
				},
				Size: "0.01",
			},
		},
		// Valid order - with funds
		{
			Order: MarketOrderRequest{
				commonOrderRequest: commonOrderRequest{
					Side:      Buy,
					ProductID: "BTC-USD",
				},
				Funds: "150.00",
			},
		},
	}

	for _, tc := range testCases {
		err := tc.Order.validate()
		if err != nil && tc.ErrMsg == "" {
			t.Errorf("Expected no errors but got: %s", err.Error())
		}

		if err == nil && tc.ErrMsg != "" {
			t.Errorf("Expected error: %s", tc.ErrMsg)
		}

		if err != nil && !strings.EqualFold(tc.ErrMsg, err.Error()) {
			t.Errorf("Expected error to be %q but got: %s", tc.ErrMsg, err.Error())
		}
	}
}
