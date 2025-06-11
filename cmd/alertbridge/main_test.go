package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/njdaniel/alertbridge/internal/adapter"
	"github.com/njdaniel/alertbridge/internal/handler"
	"github.com/njdaniel/alertbridge/internal/risk"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newTestAlpacaClient(t *testing.T) *adapter.AlpacaClient {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"1"}`))
	}))
	t.Cleanup(ts.Close)
	return adapter.NewAlpacaClient("key", "secret", ts.URL)
}

func TestMetricsEndpoint(t *testing.T) {
	t.Setenv("PROM_URL", "")
	t.Setenv("PNL_MAX", "")
	t.Setenv("PNL_MIN", "")

	alpacaClient := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := handler.NewHookHandler(zap.NewNop(), alpacaClient, g)

	mux := http.NewServeMux()
	mux.Handle("/hook", h)
	mux.Handle("/metrics", promhttp.Handler())

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
