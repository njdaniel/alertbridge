package risk

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestGuardCooldown(t *testing.T) {
	t.Setenv("PROM_URL", "")
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("1", zap.NewNop())
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

func TestGuardCheckPnLPass(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"result":[{"value":[0,"10"]}]}}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "15")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGuardCheckPnLMaxFail(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"result":[{"value":[0,"10"]}]}}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "5")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected pnl max error")
	}
}

func TestGuardCheckPnLMinFail(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"result":[{"value":[0,"0"]}]}}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "1")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected pnl min error")
	}
}

func TestGuardCheckPnLQueryError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected query error")
	}
}

func TestGuardCheckPnLStatusFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected status error")
	}
}

func TestGuardCheckPnLDecodeFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("bad"))
	}))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestGuardCheckPnLValueTypeFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"result":[{"value":[0,10]}]}}`))
	}))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected value type error")
	}
}

func TestGuardCheckPnLParseFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"result":[{"value":[0,"bad"]}]}}`))
	}))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

	g := NewGuard("0", zap.NewNop())
	if err := g.Check("bot"); err == nil {
		t.Fatalf("expected parse error")
	}
}
