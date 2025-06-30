package risk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

const defaultPromTimeout = 5 * time.Second

type Guard struct {
	logger      *zap.Logger
	cooldownSec int
	lastAlert   map[string]time.Time
	mu          sync.RWMutex

	client *http.Client

	promURL   string
	pnlMax    float64
	pnlMin    float64
	pnlMaxSet bool
}

func NewGuard(cooldownSec string) *Guard {
	sec, _ := strconv.Atoi(cooldownSec)
	promURL := os.Getenv("PROM_URL")

	timeout := defaultPromTimeout
	if v := os.Getenv("PROM_TIMEOUT_SEC"); v != "" {
		if t, err := strconv.Atoi(v); err == nil && t > 0 {
			timeout = time.Duration(t) * time.Second
		} else {
			fmt.Printf("Warning: Invalid PROM_TIMEOUT_SEC value '%s', using default %v. Error: %v\n", v, defaultPromTimeout, err)
		}
	}

	var (
		pnlMax    float64
		pnlMaxSet bool
	)
	if v, ok := os.LookupEnv("PNL_MAX"); ok {
		pnlMaxSet = true
		var err error
		pnlMax, err = strconv.ParseFloat(v, 64)
		if err != nil {
			fmt.Printf("Warning: Invalid PNL_MAX value '%s', using default 0.0. Error: %v\n", v, err)
			pnlMax = 0.0
		}
	}

	var pnlMin float64
	if v, ok := os.LookupEnv("PNL_MIN"); ok {
		var err error
		pnlMin, err = strconv.ParseFloat(v, 64)
		if err != nil {
			fmt.Printf("Warning: Invalid PNL_MIN value '%s', using default 0.0. Error: %v\n", v, err)
			pnlMin = 0.0
		}
	}

	return &Guard{
		logger:      zap.NewNop(),
		cooldownSec: sec,
		lastAlert:   make(map[string]time.Time),
		promURL:     promURL,
		client:      &http.Client{Timeout: timeout},
		pnlMax:      pnlMax,
		pnlMin:      pnlMin,
		pnlMaxSet:   pnlMaxSet,
	}
}

// SetLogger allows injecting a custom logger for debugging.
func (g *Guard) SetLogger(logger *zap.Logger) {
	if logger != nil {
		g.logger = logger
	}
}

func (g *Guard) Check(bot string) error {
	// Check cooldown
	if g.cooldownSec > 0 {
		g.mu.RLock()
		lastAlert, exists := g.lastAlert[bot]
		g.mu.RUnlock()

		if exists {
			cooldown := time.Duration(g.cooldownSec) * time.Second
			timeSinceLast := time.Since(lastAlert)
			if timeSinceLast < cooldown {
				g.logger.Warn("cooldown check failed",
					zap.String("bot", bot),
					zap.Duration("time_since_last", timeSinceLast),
					zap.Duration("cooldown", cooldown))
				return fmt.Errorf("cooldown period not elapsed for bot %s", bot)
			}
		}

		g.mu.Lock()
		g.lastAlert[bot] = time.Now()
		g.mu.Unlock()

		g.logger.Debug("cooldown check passed",
			zap.String("bot", bot),
			zap.Int("cooldown_sec", g.cooldownSec))
	}

	// Check PnL if Prometheus endpoint is available
	if err := g.checkPnL(bot); err != nil {
		return err
	}

	return nil
}

func (g *Guard) checkPnL(bot string) error {
	if g.promURL == "" {
		// Prometheus not configured
		g.logger.Debug("PnL check skipped - Prometheus not configured",
			zap.String("bot", bot))
		return nil
	}

	query := fmt.Sprintf("pnl{bot=\"%s\"}", bot)
	endpoint := fmt.Sprintf("%s/api/v1/query?query=%s", g.promURL, url.QueryEscape(query))

	g.logger.Debug("querying Prometheus for PnL",
		zap.String("bot", bot),
		zap.String("endpoint", endpoint))

	resp, err := g.client.Get(endpoint)
	if err != nil {
		g.logger.Error("failed to query Prometheus",
			zap.Error(err),
			zap.String("bot", bot),
			zap.String("endpoint", endpoint))
		return fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		g.logger.Error("Prometheus query failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("bot", bot),
			zap.String("endpoint", endpoint))
		return fmt.Errorf("Prometheus query failed with status code %d for endpoint %s", resp.StatusCode, endpoint)
	}

	var pr struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		g.logger.Error("failed to decode Prometheus response",
			zap.Error(err),
			zap.String("bot", bot))
		return fmt.Errorf("failed to decode Prometheus response: %w", err)
	}

	if len(pr.Data.Result) == 0 || len(pr.Data.Result[0].Value) < 2 {
		g.logger.Debug("no PnL data found",
			zap.String("bot", bot))
		return nil
	}

	valueStr, ok := pr.Data.Result[0].Value[1].(string)
	if !ok {
		g.logger.Error("unexpected PnL value type",
			zap.String("bot", bot),
			zap.Any("value", pr.Data.Result[0].Value[1]))
		return fmt.Errorf("unexpected PnL value type")
	}

	pnl, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		g.logger.Error("invalid PnL value",
			zap.Error(err),
			zap.String("bot", bot),
			zap.String("value", valueStr))
		return fmt.Errorf("invalid PnL value: %w", err)
	}

	g.logger.Debug("PnL check",
		zap.String("bot", bot),
		zap.Float64("pnl", pnl),
		zap.Float64("pnl_max", g.pnlMax),
		zap.Float64("pnl_min", g.pnlMin))

	if g.pnlMaxSet && pnl > g.pnlMax {
		g.logger.Warn("PnL exceeds maximum",
			zap.String("bot", bot),
			zap.Float64("pnl", pnl),
			zap.Float64("max", g.pnlMax))
		return fmt.Errorf("pnl %.2f exceeds max %.2f", pnl, g.pnlMax)
	}

	if g.pnlMin != 0 && pnl < g.pnlMin {
		g.logger.Warn("PnL below minimum",
			zap.String("bot", bot),
			zap.Float64("pnl", pnl),
			zap.Float64("min", g.pnlMin))
		return fmt.Errorf("pnl %.2f below min %.2f", pnl, g.pnlMin)
	}

	return nil
}
