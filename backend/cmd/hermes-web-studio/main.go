package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.New(cfg).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		slog.Info("Hermes Web Studio API listening", "addr", addr, "gateway", cfg.GatewayBaseURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("API server stopped", "error", err)
			os.Exit(1)
		}
	}()
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("API server shutdown failed", "error", err)
		os.Exit(1)
	}
}
