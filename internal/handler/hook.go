package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/njdaniel/alertbridge/internal/adapter"
	"github.com/njdaniel/alertbridge/internal/auth"
	"github.com/njdaniel/alertbridge/internal/notify"
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
	logger        *zap.Logger
	alpacaClient  *adapter.AlpacaClient
	riskGuard     *risk.Guard
	tvSecret      []byte
	notifier      *notify.SlackNotifier
	notifySuccess bool
	notifyFailure bool
	fullLogging   bool // when true, log remote address and full request body
}

func NewHookHandler(
	logger *zap.Logger,
	alpacaClient *adapter.AlpacaClient,
	riskGuard *risk.Guard,
	tvSecret []byte,
	notifier *notify.SlackNotifier,
	success bool,
	failure bool,
	fullLogging bool,
) *HookHandler {
	return &HookHandler{
		logger:        logger,
		alpacaClient:  alpacaClient,
		riskGuard:     riskGuard,
		tvSecret:      tvSecret,
		notifier:      notifier,
		notifySuccess: success,
		notifyFailure: failure,
		fullLogging:   fullLogging,
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
		fields := []zap.Field{zap.Error(err)}
		if h.fullLogging {
			fields = append(fields, zap.String("remote_addr", r.RemoteAddr))
		}
		h.logger.Error("failed to read request body", fields...)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log incoming request
	reqFields := []zap.Field{zap.String("user_agent", r.UserAgent())}
	if h.fullLogging {
		reqFields = append(reqFields,
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("body", string(bodyBytes)))
	} else {
		reqFields = append(reqFields, zap.Int("body_len", len(bodyBytes)))
	}
	h.logger.Info("received webhook request", reqFields...)

	// Verify request signature when secret is provided
	sig := r.Header.Get("X-TV-Signature")
	if len(h.tvSecret) > 0 {
		if sig == "" {
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}
		if err := auth.VerifyHMAC(h.tvSecret, bodyBytes, sig); err != nil {
			fields := []zap.Field{zap.Error(err), zap.String("signature", sig[:8]+"...")}
			if h.fullLogging {
				fields = append(fields, zap.String("remote_addr", r.RemoteAddr))
			}
			h.logger.Error("invalid signature", fields...)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse request body
	var alert AlertRequest
	if err := json.Unmarshal(bodyBytes, &alert); err != nil {
		fields := []zap.Field{zap.Error(err)}
		if h.fullLogging {
			fields = append(fields, zap.String("body", string(bodyBytes)))
		}
		h.logger.Error("failed to decode request", fields...)
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
		if h.notifier != nil && h.notifyFailure {
			if notifyErr := h.notifier.SendMessage("Risk check failed for bot " + alert.Bot + ": " + err.Error()); notifyErr != nil {
				h.logger.Error("failed to send notification",
					zap.Error(notifyErr),
					zap.String("bot", alert.Bot))
			}
		}
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
		if h.notifier != nil && h.notifyFailure {
			h.notifier.SendMessage("Order creation failed for bot " + alert.Bot + ": " + err.Error())
		}
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
	if h.notifier != nil && h.notifySuccess {
		h.notifier.SendMessage("Order created: " + alert.Bot + " " + alert.Side + " " + alert.Symbol + " qty " + alert.Qty)
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
