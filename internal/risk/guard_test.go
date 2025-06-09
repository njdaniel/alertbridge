package risk

import (
	"os"
	"testing"
	"time"
)

func TestGuardCooldown(t *testing.T) {
	os.Unsetenv("PROM_URL")
	os.Unsetenv("PNL_MAX")
	os.Unsetenv("PNL_MIN")

	g := NewGuard("1")
	if err := g.Check("bot"); err != nil {
		t.Fatalf("first check failed: %v", err)
	}
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected cooldown error")
	}
	time.Sleep(1100 * time.Millisecond)
	if err := g.Check("bot"); err != nil {
		t.Fatalf("expected no error after cooldown: %v", err)
	}
}
