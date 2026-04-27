package orders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type TTLWorkerConfig struct {
	TickInterval time.Duration
	BatchSize    int
	LockTimeout  time.Duration
}

type TTLWorker struct {
	db     TxBeginner
	log    *slog.Logger
	poster JournalPoster
	cfg    TTLWorkerConfig
	now    func() time.Time
}

func NewTTLWorker(db TxBeginner, log *slog.Logger, poster JournalPoster, cfg TTLWorkerConfig) *TTLWorker {
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.LockTimeout <= 0 {
		cfg.LockTimeout = 50 * time.Millisecond
	}
	if poster == nil {
		poster = RealJournalPoster{}
	}
	if log == nil {
		log = slog.Default()
	}
	return &TTLWorker{
		db:     db,
		log:    log,
		poster: poster,
		cfg:    cfg,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (w *TTLWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.cfg.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			processed, err := w.processTick(ctx)
			if err != nil {
				w.log.Error("ttl worker tick failed", "error", err)
				continue
			}
			if processed > 0 {
				w.log.Info("ttl worker processed orders", "count", processed)
			}
		}
	}
}

func (w *TTLWorker) processTick(ctx context.Context) (int, error) {
	total := 0
	for i := 0; i < w.cfg.BatchSize; i++ {
		ok, err := w.processOne(ctx)
		if err != nil {
			return total, err
		}
		if !ok {
			break
		}
		total++
	}
	return total, nil
}

func (w *TTLWorker) processOne(ctx context.Context) (bool, error) {
	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SET LOCAL lock_timeout = $1", w.cfg.LockTimeout.String()); err != nil {
		return false, fmt.Errorf("set lock_timeout: %w", err)
	}

	ord, found, err := lockOneExpiredOrder(ctx, tx)
	if err != nil {
		return false, err
	}
	if !found {
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("commit empty ttl tx: %w", err)
		}
		return false, nil
	}

	now := w.now()

	tag, err := tx.Exec(ctx, `
UPDATE core.account_balances
SET
	available = available + ($1::numeric),
	reserved  = reserved  - ($1::numeric),
	updated_at = $3
WHERE account_id = $2::uuid
  AND tenant_id = $4::uuid
  AND reserved >= ($1::numeric)
`,
		ord.Amount,
		ord.BalanceAccountID,
		now,
		ord.TenantID,
	)
	if err != nil {
		return false, fmt.Errorf("ttl release balance: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return false, ErrInsufficientReserved
	}

	tag, err = tx.Exec(ctx, `
UPDATE core.order_holds
SET
	status = 'expired',
	released_at = $2
WHERE id = $1::uuid
  AND status = 'active'
`,
		ord.HoldID,
		now,
	)
	if err != nil {
		return false, fmt.Errorf("ttl expire hold: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return false, ErrHoldNotActive
	}

	tag, err = tx.Exec(ctx, `
UPDATE core.orders
SET
	status = 'expired',
	version = version + 1,
	updated_at = $2
WHERE id = $1::uuid
  AND status = 'reserved'
`,
		ord.OrderID,
		now,
	)
	if err != nil {
		return false, fmt.Errorf("ttl expire order: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return false, ErrOrderAlreadyExpired
	}

	if err := w.poster.PostHoldRelease(ctx, tx, HoldReleasePosting{
		TenantID:                 ord.TenantID,
		OrderID:                  ord.OrderID,
		HoldID:                   ord.HoldID,
		AvailableLedgerAccountID: ord.AvailableLedgerAccountID,
		ReservedLedgerAccountID:  ord.ReservedLedgerAccountID,
		CurrencyID:               ord.CurrencyID,
		Amount:                   ord.Amount,
		Reason:                   "ttl_expired",
	}); err != nil {
		return false, fmt.Errorf("ttl post hold_release journal: %w", err)
	}

	if err := insertOutboxEvent(ctx, tx, OutboxEvent{
		TenantID:      ord.TenantID,
		AggregateType: "order",
		AggregateID:   ord.OrderID,
		EventType:     "order_expired",
		Payload: map[string]any{
			"order_id":      ord.OrderID,
			"order_ref":     ord.OrderRef,
			"office_id":     ord.OfficeID,
			"expired_at_ts": now.Unix(),
		},
	}); err != nil {
		return false, fmt.Errorf("ttl insert outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("ttl commit: %w", err)
	}

	return true, nil
}

func lockOneExpiredOrder(ctx context.Context, tx pgx.Tx) (lockedOrder, bool, error) {
	const q = `
SELECT
	o.id::text,
	o.order_ref,
	o.tenant_id::text,
	o.office_id::text,
	o.status,
	o.version,
	h.id::text,
	h.status,
	h.balance_account_id::text,
	h.available_ledger_account_id::text,
	h.reserved_ledger_account_id::text,
	h.settlement_ledger_account_id::text,
	h.currency_id::text,
	h.amount::text,
	o.expires_at
FROM core.orders o
JOIN core.order_holds h
	ON h.order_id = o.id
	AND h.status = 'active'
WHERE o.status = 'reserved'
  AND o.expires_at <= now()
ORDER BY o.expires_at ASC, o.id ASC
LIMIT 1
FOR UPDATE OF o, h SKIP LOCKED
`

	var out lockedOrder
	err := tx.QueryRow(ctx, q).Scan(
		&out.OrderID,
		&out.OrderRef,
		&out.TenantID,
		&out.OfficeID,
		&out.Status,
		&out.Version,
		&out.HoldID,
		&out.HoldStatus,
		&out.BalanceAccountID,
		&out.AvailableLedgerAccountID,
		&out.ReservedLedgerAccountID,
		&out.SettlementLedgerAccountID,
		&out.CurrencyID,
		&out.Amount,
		&out.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return lockedOrder{}, false, nil
		}
		return lockedOrder{}, false, fmt.Errorf("lock expired order: %w", err)
	}

	return out, true, nil
}
