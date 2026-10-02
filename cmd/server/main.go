// Package main is the entry point for the github-widget server.
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

	"github.com/kirasync2748/github-widget/internal/cache"
	"github.com/kirasync2748/github-widget/internal/config"
	"github.com/kirasync2748/github-widget/internal/github"
	"github.com/kirasync2748/github-widget/internal/handler"
	"github.com/kirasync2748/github-widget/internal/middleware"
	"github.com/kirasync2748/github-widget/internal/service"
	"github.com/kirasync2748/github-widget/internal/version"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	// Build dependencies.
	ghOpts := []github.Option{
		github.WithBaseURL(cfg.GitHubAPIURL),
	}
	if cfg.GitHubToken != "" {
		ghOpts = append(ghOpts, github.WithToken(cfg.GitHubToken))
	}
	ghClient := github.NewClient(ghOpts...)

	cacheInstance := cache.New(cfg.CacheTTL, cfg.CacheMaxEntries)
	repoService := service.NewRepositoryService(ghClient, cacheInstance)

	widgetHandler := handler.NewWidgetHandler(repoService, logger)
	healthHandler := handler.NewHealthHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/", healthHandler.Root)
	mux.HandleFunc("/healthz", healthHandler.Healthz)
	mux.HandleFunc("/readyz", healthHandler.Readyz)
	mux.Handle("/api/widget/", widgetHandler)

	// Apply middleware.
	var handler http.Handler = mux
	handler = middleware.Logging(logger)(handler)
	handler = middleware.Recovery(logger)(handler)

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	logger.Info("server starting", "addr", cfg.Addr(), "version", version.Version)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}
