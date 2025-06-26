package notify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendMessageWebhook(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL, "", "")
	if err := n.SendMessage("hello"); err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if !called {
		t.Fatalf("expected webhook to be called")
	}
}

func TestSendMessageWebhookError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL, "", "")
	err := n.SendMessage("hello")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendMessageToken(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewSlackNotifier("", "token", "chan")
	// override API url
	n.client = ts.Client()
	originalURL := slackAPIURL
	slackAPIURL = ts.URL
	defer func() { slackAPIURL = originalURL }()

	if err := n.SendMessage("hi"); err != nil {
		t.Fatalf("SendMessage token failed: %v", err)
	}
	if !called {
		t.Fatalf("expected api to be called")
	}
}

func TestSendMessageTokenError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	n := NewSlackNotifier("", "token", "chan")
	n.client = ts.Client()
	originalURL := slackAPIURL
	slackAPIURL = ts.URL
	defer func() { slackAPIURL = originalURL }()

	err := n.SendMessage("hi")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendMessageNoConfig(t *testing.T) {
	n := NewSlackNotifier("", "", "")
	if err := n.SendMessage("hi"); err == nil {
		t.Fatalf("expected error")
	}
}
