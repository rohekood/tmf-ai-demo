package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tmf/services/pocv/internal/app"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("Starting POCV Service...")

	cfg := app.DefaultConfig()
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.DatabaseURL = v
	}
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		cfg.RabbitMQURL = v
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		logger.Info("Shutting down POCV Service...")
		cancel()
	}()

	if err := app.Run(ctx, cfg); err != nil {
		logger.Error("POCV Service failed", "error", err)
		os.Exit(1)
	}
	logger.Info("POCV Service Stopped.")
}
