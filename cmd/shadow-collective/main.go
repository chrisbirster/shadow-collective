package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/chrisbirster/shadow-collective/internal/config"
	"github.com/chrisbirster/shadow-collective/internal/gateway"
)

func main() {
	configPath := flag.String("config", envOr("GATEWAY_CONFIG", "/app/config/services.json"), "path to gateway config")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("configuration error", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("shadow collective gateway starting", "config", *configPath)
	if err := gateway.New(cfg, log).Run(ctx); err != nil {
		log.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
