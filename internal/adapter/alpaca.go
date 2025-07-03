package adapter

import (
	"fmt"
	"strings"
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

// isCrypto determines if a symbol represents a crypto pair
func isCrypto(symbol string) bool {
	return strings.HasSuffix(symbol, "USD") || strings.HasSuffix(symbol, "USDT") || strings.HasSuffix(symbol, "USDC")
}

func (c *AlpacaClient) CreateOrder(bot, symbol, side, qty string) (*alpaca.Order, error) {
	// Convert side to Alpaca side
	alpacaSide := alpaca.Side(side)

	// Parse quantity
	qtyDec, err := decimal.NewFromString(qty)
	if err != nil {
		return nil, fmt.Errorf("invalid qty: %w", err)
	}

	// Determine time in force based on asset type
	timeInForce := alpaca.Day
	if isCrypto(symbol) {
		timeInForce = alpaca.GTC
	}

	// Create order request
	orderRequest := alpaca.PlaceOrderRequest{
		Symbol:        symbol,
		Qty:           &qtyDec,
		Side:          alpacaSide,
		Type:          alpaca.Market,
		TimeInForce:   timeInForce,
		ClientOrderID: fmt.Sprintf("%s-%d", bot, time.Now().UnixNano()),
	}

	// Log outgoing request for debugging
	c.logger.Info("placing order",
		zap.String("baseURL", c.baseURL),
		zap.String("symbol", symbol),
		zap.String("side", side),
		zap.String("qty", qty),
		zap.String("timeInForce", string(timeInForce)),
		zap.Any("request", orderRequest))

	// Place order
	order, err := c.client.PlaceOrder(orderRequest)
	if err != nil {
		if apiErr, ok := err.(*alpaca.APIError); ok {
			c.logger.Error("alpaca API error",
				zap.String("symbol", symbol),
				zap.String("side", side),
				zap.String("qty", qty),
				zap.String("timeInForce", string(timeInForce)),
				zap.Int("status", apiErr.StatusCode),
				zap.Int("code", apiErr.Code),
				zap.String("message", apiErr.Message),
				zap.String("body", apiErr.Body))
		} else {
			c.logger.Error("failed to place order",
				zap.String("symbol", symbol),
				zap.String("side", side),
				zap.String("qty", qty),
				zap.String("timeInForce", string(timeInForce)),
				zap.Error(err))
		}
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	c.logger.Info("order placed successfully",
		zap.String("symbol", symbol),
		zap.String("side", side),
		zap.String("qty", qty),
		zap.String("timeInForce", string(timeInForce)),
		zap.String("orderID", order.ID))
	return order, nil
}
