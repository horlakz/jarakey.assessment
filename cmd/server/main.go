package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/horlakz/jarakey.assessment/internal/app"
	"github.com/horlakz/jarakey.assessment/internal/config"
)

func main() {
	cfg := config.Load()
	application, err := app.Build(cfg)
	if err != nil {
		slog.Error("failed to build application", "error", err)
		os.Exit(1)
	}

	go func() {
		if err := application.Fiber.Listen(cfg.ServerPort); err != nil {
			slog.Error("fiber server stopped", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}
