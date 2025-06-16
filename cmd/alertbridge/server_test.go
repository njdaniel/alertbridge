package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"
)

// Test that the main server starts and shuts down when it receives an interrupt signal.
func TestServerStartupShutdown(t *testing.T) {
	alpacaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"1"}`))
	}))
	defer alpacaSrv.Close()

	// find free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	os.Setenv("ALP_KEY", "k")
	os.Setenv("ALP_SECRET", "s")
	os.Setenv("ALP_BASE", alpacaSrv.URL)
	os.Setenv("PORT", fmt.Sprintf("%d", port))
	os.Setenv("COOLDOWN_SEC", "0")
	os.Setenv("PROM_URL", "")
	os.Setenv("PNL_MAX", "")
	os.Setenv("PNL_MIN", "")
	os.Setenv("TV_SECRET", "")

	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)
	var resp *http.Response
	for i := 0; i < 20; i++ {
		resp, err = http.Get(url)
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("server not responding: %v", err)
	}

	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGINT)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shutdown")
	}
}
