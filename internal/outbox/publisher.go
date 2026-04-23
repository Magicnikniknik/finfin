package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type TxBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type MessagePublisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

type Event struct {
	ID       int64
	TenantID string
	Type     string
	Topic    string
	Key      string
	Payload  []byte
	Attempts int
}

type WorkerConfig struct {
	TickInterval time.Duration
	BatchSize    int
}

type Worker struct {
	db        TxBeginner
	publisher MessagePublisher
	log       *slog.Logger
	cfg       WorkerConfig
}

func NewWorker(db TxBeginner, publisher MessagePublisher, log *slog.Logger, cfg WorkerConfig) *Worker {
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = 500 * time.Millisecond
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if log == nil {
		log = slog.Default()
	}
	return &Worker{db: db, publisher: publisher, log: log, cfg: cfg}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.cfg.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.log.Error("outbox batch failed", "error", err)
			}
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) error {
	for i := 0; i < w.cfg.BatchSize; i++ {
		processed, err := w.processOne(ctx)
		if err != nil {
			return err
		}
		if !processed {
			return nil
		}
	}
	return nil
}

func (w *Worker) processOne(ctx context.Context) (bool, error) {
	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	evt, found, err := lockNextEvent(ctx, tx)
	if err != nil {
		return false, err
	}
	if !found {
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("commit empty tx: %w", err)
		}
		return false, nil
	}

	if err := w.publisher.Publish(ctx, evt.Topic, evt.Key, evt.Payload); err != nil {
		_, updErr := tx.Exec(ctx, `
UPDATE core.outbox_events
SET
  status = 'pending',
  attempts = attempts + 1,
  last_error = $2,
  available_at = now() + interval '2 seconds',
  updated_at = now()
WHERE id = $1
`, evt.ID, err.Error())
		if updErr != nil {
			return false, fmt.Errorf("mark publish failure: %w", updErr)
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return false, fmt.Errorf("commit publish failure: %w", commitErr)
		}
		return true, nil
	}

	_, err = tx.Exec(ctx, `
UPDATE core.outbox_events
SET
  status = 'published',
  published_at = now(),
  updated_at = now(),
  last_error = NULL
WHERE id = $1
`, evt.ID)
	if err != nil {
		return false, fmt.Errorf("mark published: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}
	return true, nil
}

func lockNextEvent(ctx context.Context, tx pgx.Tx) (Event, bool, error) {
	const q = `
SELECT
  id,
  tenant_id::text,
  event_type,
  aggregate_type,
  aggregate_id::text,
  payload::text,
  attempts
FROM core.outbox_events
WHERE status = 'pending'
  AND available_at <= now()
ORDER BY available_at ASC, id ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	var (
		evt         Event
		aggType     string
		aggID       string
		payloadText string
	)
	if err := tx.QueryRow(ctx, q).Scan(
		&evt.ID,
		&evt.TenantID,
		&evt.Type,
		&aggType,
		&aggID,
		&payloadText,
		&evt.Attempts,
	); err != nil {
		if err == pgx.ErrNoRows {
			return Event{}, false, nil
		}
		return Event{}, false, fmt.Errorf("lock outbox row: %w", err)
	}

	evt.Topic = evt.Type
	evt.Key = aggType + ":" + aggID
	evt.Payload = json.RawMessage(payloadText)

	_, err := tx.Exec(ctx, `
UPDATE core.outbox_events
SET
  status = 'processing',
  updated_at = now()
WHERE id = $1
`, evt.ID)
	if err != nil {
		return Event{}, false, fmt.Errorf("mark processing: %w", err)
	}

	return evt, true, nil
}
