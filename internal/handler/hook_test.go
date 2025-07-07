package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/njdaniel/alertbridge/internal/adapter"
	"github.com/njdaniel/alertbridge/internal/auth"
	"github.com/njdaniel/alertbridge/internal/risk"
)

// sign calculates the TradingView HMAC signature used in tests.
func sign(secret string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func newTestAlpacaClient(t *testing.T) *adapter.AlpacaClient {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"1"}`))
	}))
	t.Cleanup(ts.Close)
	return adapter.NewAlpacaClient("key", "secret", ts.URL)
}

func TestHandleSuccess(t *testing.T) {
	client := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, nil, nil, true, true, true)

	body := []byte(`{"bot":"b","symbol":"AAPL","side":"buy","qty":"1"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleCooldown(t *testing.T) {
	client := newTestAlpacaClient(t)
	g := risk.NewGuard("1")
	h := NewHookHandler(zap.NewNop(), client, g, nil, nil, true, true, true)

	body := []byte(`{"bot":"b","symbol":"AAPL","side":"buy","qty":"1"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Handle(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected first call 200, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	h.Handle(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("expected cooldown 403, got %d", rr2.Code)
	}
}

func TestVerifyHMAC(t *testing.T) {
	body := []byte("test")
	secret := []byte("s")
	sig := sign(string(secret), body)
	if err := auth.VerifyHMAC(secret, body, sig); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
}

func TestVerifyHMACInvalid(t *testing.T) {
	body := []byte("test")
	secret := []byte("s")
	if err := auth.VerifyHMAC(secret, body, "bad"); err == nil {
		t.Fatalf("expected error for invalid signature")
	}
}

func TestVerifyHMACDisabled(t *testing.T) {
	if err := auth.VerifyHMAC(nil, []byte("test"), "anything"); err != nil {
		t.Fatalf("expected nil when secret empty, got %v", err)
	}
}

func TestHandleMissingSignature(t *testing.T) {
	client := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, []byte("s"), nil, true, true, true)

	body := []byte(`{"bot":"b","symbol":"AAPL","side":"buy","qty":"1"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandleInvalidSignature(t *testing.T) {
	client := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, []byte("s"), nil, true, true, true)

	body := []byte(`{"bot":"b","symbol":"AAPL","side":"buy","qty":"1"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	req.Header.Set("X-TV-Signature", "badbadbad")
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandleInvalidJSON(t *testing.T) {
	client := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, nil, nil, true, true, true)

	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader([]byte("{")))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleMissingFields(t *testing.T) {
	client := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, nil, nil, true, true, true)

	body := []byte(`{"bot":"b"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleInvalidSide(t *testing.T) {
	client := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, nil, nil, true, true, true)

	body := []byte(`{"bot":"b","symbol":"AAPL","side":"bad","qty":"1"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleRiskFail(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"result":[{"value":[0,"10"]}]}}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	t.Setenv("PROM_URL", ts.URL)
	t.Setenv("PNL_MAX", "5")
	t.Setenv("PNL_MIN", "")

	client := newTestAlpacaClient(t)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, nil, nil, true, true, true)

	body := []byte(`{"bot":"b","symbol":"AAPL","side":"buy","qty":"1"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestHandleOrderError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	t.Cleanup(ts.Close)

	client := adapter.NewAlpacaClient("key", "secret", ts.URL)
	g := risk.NewGuard("0")
	h := NewHookHandler(zap.NewNop(), client, g, nil, nil, true, true, true)

	body := []byte(`{"bot":"b","symbol":"AAPL","side":"buy","qty":"1"}`)
	req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
