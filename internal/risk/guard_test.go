package risk

import (
    "testing"
    "time"
)

func TestGuardCooldown(t *testing.T) {
	t.Setenv("PROM_URL", "")
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

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
