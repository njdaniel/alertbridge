package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/njdaniel/alertbridge/internal/adapter"
	"github.com/njdaniel/alertbridge/internal/auth"
	"github.com/njdaniel/alertbridge/internal/risk"
	"github.com/njdaniel/alertbridge/pkg/metrics"
)

type AlertRequest struct {
	Bot    string `json:"bot"`
	Symbol string `json:"symbol"`
	Side   string `json:"side"`
	Qty    string `json:"qty"`
	TS     int64  `json:"ts,omitempty"`
}

type HookHandler struct {
	logger       *zap.Logger
	alpacaClient *adapter.AlpacaClient
	riskGuard    *risk.Guard
	tvSecret     []byte
}

func NewHookHandler(
	logger *zap.Logger,
	alpacaClient *adapter.AlpacaClient,
	riskGuard *risk.Guard,
	tvSecret []byte,
) *HookHandler {
	return &HookHandler{
		logger:       logger,
		alpacaClient: alpacaClient,
		riskGuard:    riskGuard,
		tvSecret:     tvSecret,
	}
}

// ServeHTTP implements http.Handler interface
func (h *HookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hook" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.Handle(w, r)
}

func (h *HookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body",
			zap.Error(err),
			zap.String("remote_addr", r.RemoteAddr))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log incoming request
	h.logger.Info("received webhook request",
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("body", string(bodyBytes)))

	// Verify request signature when secret is provided
	sig := r.Header.Get("X-TV-Signature")
	if len(h.tvSecret) > 0 {
		if sig == "" {
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}
		if err := auth.VerifyHMAC(h.tvSecret, bodyBytes, sig); err != nil {
			h.logger.Error("invalid signature",
				zap.Error(err),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("signature", sig))
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse request body
	var alert AlertRequest
	if err := json.Unmarshal(bodyBytes, &alert); err != nil {
		h.logger.Error("failed to decode request",
			zap.Error(err),
			zap.String("body", string(bodyBytes)))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if alert.Bot == "" || alert.Symbol == "" || alert.Side == "" || alert.Qty == "" {
		h.logger.Error("missing required fields",
			zap.String("bot", alert.Bot),
			zap.String("symbol", alert.Symbol),
			zap.String("side", alert.Side),
			zap.String("qty", alert.Qty))
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Validate side
	if alert.Side != "buy" && alert.Side != "sell" {
		h.logger.Error("invalid side",
			zap.String("side", alert.Side),
			zap.String("bot", alert.Bot))
		http.Error(w, "Invalid side", http.StatusBadRequest)
		return
	}

	// Check risk rules
	if err := h.riskGuard.Check(alert.Bot); err != nil {
		h.logger.Error("risk check failed",
			zap.Error(err),
			zap.String("bot", alert.Bot))
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Create order
	order, err := h.alpacaClient.CreateOrder(alert.Bot, alert.Symbol, alert.Side, alert.Qty)
	if err != nil {
		h.logger.Error("failed to create order",
			zap.Error(err),
			zap.String("bot", alert.Bot),
			zap.String("symbol", alert.Symbol),
			zap.String("side", alert.Side),
			zap.String("qty", alert.Qty))
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Increment metrics
	metrics.OrderTotal.WithLabelValues(alert.Bot, alert.Side).Inc()

	// Log success
	h.logger.Info("order created successfully",
		zap.String("bot", alert.Bot),
		zap.String("symbol", alert.Symbol),
		zap.String("side", alert.Side),
		zap.String("qty", alert.Qty),
		zap.String("order_id", order.ID))

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
