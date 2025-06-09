package handler

import (
	"bytes"
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
	hmacVerifier *auth.HMACVerifier
	riskGuard    *risk.Guard
}

func NewHookHandler(
	logger *zap.Logger,
	alpacaClient *adapter.AlpacaClient,
	hmacVerifier *auth.HMACVerifier,
	riskGuard *risk.Guard,
) *HookHandler {
	return &HookHandler{
		logger:       logger,
		alpacaClient: alpacaClient,
		hmacVerifier: hmacVerifier,
		riskGuard:    riskGuard,
	}
}

func (h *HookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify HMAC if secret is set
	if h.hmacVerifier.IsEnabled() {
		signature := r.Header.Get("X-TV-Signature")
		if !h.hmacVerifier.Verify(bytes.NewReader(bodyBytes), signature) {
			h.logger.Error("invalid signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse request body
	var alert AlertRequest
	if err := json.Unmarshal(bodyBytes, &alert); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if alert.Bot == "" || alert.Symbol == "" || alert.Side == "" || alert.Qty == "" {
		h.logger.Error("missing required fields")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Validate side
	if alert.Side != "buy" && alert.Side != "sell" {
		h.logger.Error("invalid side", zap.String("side", alert.Side))
		http.Error(w, "Invalid side", http.StatusBadRequest)
		return
	}

	// Check risk rules
	if err := h.riskGuard.Check(alert.Bot); err != nil {
		h.logger.Error("risk check failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Create order
	order, err := h.alpacaClient.CreateOrder(alert.Bot, alert.Symbol, alert.Side, alert.Qty)
	if err != nil {
		h.logger.Error("failed to create order", zap.Error(err))
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Increment metrics
	metrics.OrderTotal.WithLabelValues(alert.Bot, alert.Side).Inc()

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
