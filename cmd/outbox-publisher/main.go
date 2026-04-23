package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"finfin/internal/outbox"
)

type LogPublisher struct {
	log *slog.Logger
}

func (p LogPublisher) Publish(_ context.Context, topic string, key string, payload []byte) error {
	p.log.Info("outbox publish", "topic", topic, "key", key, "payload", string(payload))
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("failed to init db pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	worker := outbox.NewWorker(pool, LogPublisher{log: logger}, logger, outbox.WorkerConfig{
		TickInterval: time.Second,
		BatchSize:    50,
	})

	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("outbox worker stopped", "error", err)
		os.Exit(1)
	}
}
