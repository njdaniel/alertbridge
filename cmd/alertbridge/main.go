package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/njdaniel/alertbridge/internal/adapter"
	"github.com/njdaniel/alertbridge/internal/auth"
	"github.com/njdaniel/alertbridge/internal/handler"
	"github.com/njdaniel/alertbridge/internal/risk"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	port := getEnv("PORT", "3000")
	alpacaKey := getEnv("ALP_KEY", "")
	alpacaSecret := getEnv("ALP_SECRET", "")
	alpacaBase := getEnv("ALP_BASE", "https://paper-api.alpaca.markets")
	tvSecret := getEnv("TV_SECRET", "")
	cooldownSec := getEnv("COOLDOWN_SEC", "0")

	// Initialize components
	alpacaClient := adapter.NewAlpacaClient(alpacaKey, alpacaSecret, alpacaBase)
	hmacVerifier := auth.NewHMACVerifier(tvSecret)
	riskGuard := risk.NewGuard(cooldownSec)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize handlers
	hookHandler := handler.NewHookHandler(logger, alpacaClient, hmacVerifier, riskGuard)

	// Routes
	r.Post("/hook", hookHandler.Handle)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out... forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	logger.Info("starting server", zap.String("port", port))
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
