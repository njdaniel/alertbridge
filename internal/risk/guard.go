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
)

type Guard struct {
	cooldownSec int
	lastAlert   map[string]time.Time
	mu          sync.RWMutex

	promURL   string
	pnlMax    float64
	pnlMin    float64
	pnlMaxSet bool
}

func NewGuard(cooldownSec string) *Guard {
	sec, _ := strconv.Atoi(cooldownSec)
	promURL := os.Getenv("PROM_URL")

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
		cooldownSec: sec,
		lastAlert:   make(map[string]time.Time),
		promURL:     promURL,
		pnlMax:      pnlMax,
		pnlMin:      pnlMin,
		pnlMaxSet:   pnlMaxSet,
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
			if time.Since(lastAlert) < cooldown {
				return fmt.Errorf("cooldown period not elapsed for bot %s", bot)
			}
		}

		g.mu.Lock()
		g.lastAlert[bot] = time.Now()
		g.mu.Unlock()
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
		return nil
	}

	query := fmt.Sprintf("pnl{bot=\"%s\"}", bot)
	endpoint := fmt.Sprintf("%s/api/v1/query?query=%s", g.promURL, url.QueryEscape(query))

	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
		return fmt.Errorf("failed to decode Prometheus response: %w", err)
	}

	if len(pr.Data.Result) == 0 || len(pr.Data.Result[0].Value) < 2 {
		return nil
	}

	valueStr, ok := pr.Data.Result[0].Value[1].(string)
	if !ok {
		return fmt.Errorf("unexpected PnL value type")
	}

	pnl, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid PnL value: %w", err)
	}

	if g.pnlMaxSet && pnl > g.pnlMax {
		return fmt.Errorf("pnl %.2f exceeds max %.2f", pnl, g.pnlMax)
	}

	if g.pnlMin != 0 && pnl < g.pnlMin {
		return fmt.Errorf("pnl %.2f below min %.2f", pnl, g.pnlMin)
	}

	return nil
}
