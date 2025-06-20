package notify

import (
	"net/http"
	"net/http/httptest"
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
