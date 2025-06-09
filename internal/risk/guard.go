package risk

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Guard struct {
	cooldownSec int
	lastAlert   map[string]time.Time
	mu          sync.RWMutex
}

func NewGuard(cooldownSec string) *Guard {
	sec, _ := strconv.Atoi(cooldownSec)
	return &Guard{
		cooldownSec: sec,
		lastAlert:   make(map[string]time.Time),
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
	// TODO: Implement Prometheus PnL check
	// This would require a Prometheus client to query the metrics
	// For now, we'll skip this check
	return nil
}
