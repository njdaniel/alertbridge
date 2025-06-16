package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/njdaniel/alertbridge/internal/adapter"
	"github.com/njdaniel/alertbridge/internal/handler"
	"github.com/njdaniel/alertbridge/internal/risk"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Get environment variables
	alpacaKey := os.Getenv("ALP_KEY")
	alpacaSecret := os.Getenv("ALP_SECRET")
	alpacaBase := os.Getenv("ALP_BASE")
	if alpacaBase == "" {
		alpacaBase = "https://paper-api.alpaca.markets"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	cooldownSec := os.Getenv("COOLDOWN_SEC")
	tvSecret := os.Getenv("TV_SECRET")

	// Initialize Alpaca client
	alpacaClient := adapter.NewAlpacaClient(alpacaKey, alpacaSecret, alpacaBase)
	alpacaClient.SetLogger(logger)

	// Initialize risk guard
	riskGuard := risk.NewGuard(cooldownSec)

	// Initialize handler
	hookHandler := handler.NewHookHandler(logger, alpacaClient, riskGuard, []byte(tvSecret))

	// Create mux and register handlers
	mux := http.NewServeMux()
	mux.Handle("/hook", hookHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	logger.Info("Registered /metrics and /healthz endpoints")

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
}
