package adapter

import (
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AlpacaClient struct {
	client  *alpaca.Client
	logger  *zap.Logger
	baseURL string
}

func NewAlpacaClient(key, secret, baseURL string) *AlpacaClient {
	client := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    key,
		APISecret: secret,
		BaseURL:   baseURL,
	})

	return &AlpacaClient{
		client:  client,
		logger:  zap.NewNop(),
		baseURL: baseURL,
	}
}

// SetLogger allows injecting a custom logger for debugging.
func (c *AlpacaClient) SetLogger(logger *zap.Logger) {
	if logger != nil {
		c.logger = logger
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

	// Log outgoing request for debugging
	c.logger.Debug("placing order",
		zap.String("url", fmt.Sprintf("%s/v2/orders", c.baseURL)),
		zap.Any("request", orderRequest))

	// Place order
	order, err := c.client.PlaceOrder(orderRequest)
	if err != nil {
		if apiErr, ok := err.(*alpaca.APIError); ok {
			c.logger.Error("alpaca error", zap.Any("request", orderRequest),
				zap.Int("status", apiErr.StatusCode),
				zap.Int("code", apiErr.Code),
				zap.String("message", apiErr.Message),
				zap.String("body", apiErr.Body))
		} else {
			c.logger.Error("failed to place order", zap.Error(err), zap.Any("request", orderRequest))
		}
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	c.logger.Debug("order placed", zap.Any("order", order))
	return order, nil
}
