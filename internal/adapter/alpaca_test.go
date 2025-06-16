package adapter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestCreateOrderEquity(t *testing.T) {
	var requestPath string
	var requestBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		body, _ := ioutil.ReadAll(r.Body)
		requestBody = body
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"abc"}`))
	}))
	defer ts.Close()

	c := NewAlpacaClient("k", "s", ts.URL)
	c.SetLogger(zap.NewNop())

	order, err := c.CreateOrder("bot", "AAPL", "buy", "1")
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.ID != "abc" {
		t.Fatalf("expected order id abc, got %s", order.ID)
	}
	if requestPath != "/v2/orders" {
		t.Fatalf("expected path /v2/orders, got %s", requestPath)
	}
	var req map[string]interface{}
	if err := json.Unmarshal(requestBody, &req); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if req["symbol"] != "AAPL" {
		t.Fatalf("expected symbol AAPL, got %v", req["symbol"])
	}
	if req["time_in_force"] != "day" {
		t.Fatalf("expected time_in_force day, got %v", req["time_in_force"])
	}
}

func TestCreateOrderCrypto(t *testing.T) {
	var requestBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		requestBody = body
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"xyz"}`))
	}))
	defer ts.Close()

	c := NewAlpacaClient("k", "s", ts.URL)
	c.SetLogger(zap.NewNop())

	order, err := c.CreateOrder("bot", "ETHUSD", "sell", "0.5")
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.ID != "xyz" {
		t.Fatalf("expected order id xyz, got %s", order.ID)
	}
	var req map[string]interface{}
	if err := json.Unmarshal(requestBody, &req); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if req["time_in_force"] != "gtc" {
		t.Fatalf("expected time_in_force gtc, got %v", req["time_in_force"])
	}
}

func TestCreateOrderInvalidQty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	c := NewAlpacaClient("k", "s", ts.URL)
	c.SetLogger(zap.NewNop())

	if _, err := c.CreateOrder("bot", "AAPL", "buy", "bad"); err == nil {
		t.Fatalf("expected error for invalid qty")
	}
}
