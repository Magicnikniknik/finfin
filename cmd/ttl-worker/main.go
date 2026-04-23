package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"finfin/internal/orders"
)

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

	worker := orders.NewTTLWorker(
		pool,
		logger,
		orders.RealJournalPoster{},
		orders.TTLWorkerConfig{
			TickInterval: time.Second,
			BatchSize:    100,
			LockTimeout:  50 * time.Millisecond,
		},
	)

	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("ttl worker stopped with error", "error", err)
		os.Exit(1)
	}
}
