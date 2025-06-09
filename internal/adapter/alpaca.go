package adapter

import (
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AlpacaClient struct {
	client *alpaca.Client
	logger *zap.Logger
}

func NewAlpacaClient(key, secret, baseURL string) *AlpacaClient {
	client := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    key,
		APISecret: secret,
		BaseURL:   baseURL,
	})

	return &AlpacaClient{
		client: client,
		logger: zap.NewNop(),
	}
}

func (c *AlpacaClient) CreateOrder(bot, symbol, side, qty string) (*alpaca.Order, error) {
	// Convert side to Alpaca side
	alpacaSide := alpaca.Side(side)

	// Parse quantity
	qtyDec, err := decimal.NewFromString(qty)
	if err != nil {
		return nil, fmt.Errorf("invalid qty: %w", err)
	}

	// Create order request
	orderRequest := alpaca.PlaceOrderRequest{
		Symbol:        symbol,
		Qty:           &qtyDec,
		Side:          alpacaSide,
		Type:          alpaca.Market,
		TimeInForce:   alpaca.Day,
		ClientOrderID: fmt.Sprintf("%s-%d", bot, time.Now().UnixNano()),
	}

	// Place order
	order, err := c.client.PlaceOrder(orderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	return order, nil
}
