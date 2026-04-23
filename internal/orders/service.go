package orders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) ReserveOrder(ctx context.Context, cmd ReserveOrderCommand) (ReserveOrderResult, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return ReserveOrderResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '50ms'`); err != nil {
		return ReserveOrderResult{}, fmt.Errorf("set lock_timeout: %w", err)
	}

	reqHash := buildRequestHash(
		cmd.TenantID,
		cmd.ClientRef,
		cmd.IdempotencyKey,
		cmd.OfficeID,
		cmd.QuoteID,
		cmd.HoldAmount,
		cmd.HoldCurrencyID,
		cmd.ExpiresAt.UTC().Format(time.RFC3339),
	)

	rec, err := acquireIdempotency(ctx, tx, cmd.TenantID, "reserve_order", cmd.IdempotencyKey, reqHash)
	if err != nil {
		return ReserveOrderResult{}, err
	}
	if rec != nil {
		return decodeCachedResponse[ReserveOrderResult](rec)
	}

	now := s.now()
	orderID := uuid.NewString()
	holdID := uuid.NewString()
	orderRef := buildPublicOrderRef()

	tag, err := tx.Exec(ctx, `
UPDATE core.account_balances
SET
	available = available - ($1::numeric),
	reserved  = reserved  + ($1::numeric),
	updated_at = $3
WHERE account_id = $2::uuid
  AND tenant_id = $4::uuid
  AND available >= ($1::numeric)
`,
		cmd.HoldAmount,
		cmd.BalanceAccountID,
		now,
		cmd.TenantID,
	)
	if err != nil {
		return ReserveOrderResult{}, fmt.Errorf("reserve projection balance: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ReserveOrderResult{}, ErrInsufficientAvailable
	}

	_, err = tx.Exec(ctx, `
INSERT INTO core.orders (
	id,
	tenant_id,
	office_id,
	client_ref,
	side,
	give_currency_id,
	get_currency_id,
	amount_give,
	amount_get,
	fixed_rate,
	quote_payload,
	status,
	reserved_at,
	expires_at,
	version,
	order_ref,
	created_at,
	updated_at
)
VALUES (
	$1::uuid,
	$2::uuid,
	$3::uuid,
	$4,
	$5,
	$6::uuid,
	$7::uuid,
	$8::numeric,
	$9::numeric,
	$10::numeric,
	$11::jsonb,
	'reserved',
	$12,
	$13,
	1,
	$14,
	$12,
	$12
)
`,
		orderID,
		cmd.TenantID,
		cmd.OfficeID,
		cmd.ClientRef,
		cmd.Side,
		cmd.GiveCurrencyID,
		cmd.GetCurrencyID,
		cmd.AmountGive,
		cmd.AmountGet,
		cmd.FixedRate,
		cmd.QuotePayload,
		now,
		cmd.ExpiresAt.UTC(),
		orderRef,
	)
	if err != nil {
		return ReserveOrderResult{}, fmt.Errorf("insert order: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO core.order_holds (
	id,
	tenant_id,
	order_id,
	balance_account_id,
	available_ledger_account_id,
	reserved_ledger_account_id,
	settlement_ledger_account_id,
	currency_id,
	amount,
	status,
	expires_at,
	created_at
)
VALUES (
	$1::uuid,
	$2::uuid,
	$3::uuid,
	$4::uuid,
	$5::uuid,
	$6::uuid,
	$7::uuid,
	$8::uuid,
	$9::numeric,
	'active',
	$10,
	$11
)
`,
		holdID,
		cmd.TenantID,
		orderID,
		cmd.BalanceAccountID,
		cmd.AvailableLedgerAccountID,
		cmd.ReservedLedgerAccountID,
		cmd.SettlementLedgerAccountID,
		cmd.HoldCurrencyID,
		cmd.HoldAmount,
		cmd.ExpiresAt.UTC(),
		now,
	)
	if err != nil {
		return ReserveOrderResult{}, fmt.Errorf("insert order_hold: %w", err)
	}

	if err := s.poster.PostHoldCreate(ctx, tx, HoldCreatePosting{
		TenantID:                 cmd.TenantID,
		OrderID:                  orderID,
		HoldID:                   holdID,
		AvailableLedgerAccountID: cmd.AvailableLedgerAccountID,
		ReservedLedgerAccountID:  cmd.ReservedLedgerAccountID,
		CurrencyID:               cmd.HoldCurrencyID,
		Amount:                   cmd.HoldAmount,
		Reason:                   "reserve_order",
	}); err != nil {
		return ReserveOrderResult{}, fmt.Errorf("post hold_create journal: %w", err)
	}

	if err := insertOutboxEvent(ctx, tx, OutboxEvent{
		TenantID:      cmd.TenantID,
		AggregateType: "order",
		AggregateID:   orderID,
		EventType:     "order_reserved",
		Payload: map[string]any{
			"order_id":      orderID,
			"order_ref":     orderRef,
			"office_id":     cmd.OfficeID,
			"quote_id":      cmd.QuoteID,
			"expires_at_ts": cmd.ExpiresAt.UTC().Unix(),
		},
	}); err != nil {
		return ReserveOrderResult{}, fmt.Errorf("insert outbox order_reserved: %w", err)
	}

	result := ReserveOrderResult{
		OrderID:   orderID,
		OrderRef:  orderRef,
		Status:    "reserved",
		ExpiresAt: cmd.ExpiresAt.UTC(),
		Version:   1,
	}

	if err := finalizeIdempotency(ctx, tx, cmd.TenantID, "reserve_order", cmd.IdempotencyKey, "order", orderID, result); err != nil {
		return ReserveOrderResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ReserveOrderResult{}, fmt.Errorf("commit reserve tx: %w", err)
	}

	return result, nil
}

func (s *Service) CompleteOrder(ctx context.Context, cmd CompleteOrderCommand) (CompleteOrderResult, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return CompleteOrderResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '50ms'`); err != nil {
		return CompleteOrderResult{}, fmt.Errorf("set lock_timeout: %w", err)
	}

	reqHash := buildRequestHash(
		cmd.TenantID,
		cmd.OrderID,
		cmd.IdempotencyKey,
		cmd.CashierID,
		fmt.Sprintf("%d", cmd.ExpectedVersion),
	)

	rec, err := acquireIdempotency(ctx, tx, cmd.TenantID, "complete_order", cmd.IdempotencyKey, reqHash)
	if err != nil {
		return CompleteOrderResult{}, err
	}
	if rec != nil {
		return decodeCachedResponse[CompleteOrderResult](rec)
	}

	ord, err := lockOrderForMutation(ctx, tx, cmd.TenantID, cmd.OrderID)
	if err != nil {
		return CompleteOrderResult{}, err
	}
	if ord.Status != "reserved" {
		return CompleteOrderResult{}, ErrOrderNotActive
	}
	if ord.HoldStatus != "active" {
		return CompleteOrderResult{}, ErrHoldNotActive
	}
	if ord.SettlementLedgerAccountID == nil || *ord.SettlementLedgerAccountID == "" {
		return CompleteOrderResult{}, fmt.Errorf("missing settlement ledger account for order %s", ord.OrderID)
	}

	now := s.now()

	var newVersion int64
	err = tx.QueryRow(ctx, `
UPDATE core.orders
SET
	status = 'completed',
	completed_at = $3,
	version = version + 1,
	updated_at = $3
WHERE id = $1::uuid
  AND tenant_id = $2::uuid
  AND status = 'reserved'
  AND version = $4
RETURNING version
`,
		cmd.OrderID,
		cmd.TenantID,
		now,
		cmd.ExpectedVersion,
	).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CompleteOrderResult{}, ErrVersionConflict
		}
		return CompleteOrderResult{}, fmt.Errorf("cas update complete order: %w", err)
	}

	tag, err := tx.Exec(ctx, `
UPDATE core.order_holds
SET
	status = 'consumed',
	consumed_at = $2
WHERE id = $1::uuid
  AND status = 'active'
`,
		ord.HoldID,
		now,
	)
	if err != nil {
		return CompleteOrderResult{}, fmt.Errorf("consume hold: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return CompleteOrderResult{}, ErrHoldNotActive
	}

	tag, err = tx.Exec(ctx, `
UPDATE core.account_balances
SET
	reserved = reserved - ($1::numeric),
	updated_at = $3
WHERE account_id = $2::uuid
  AND tenant_id = $4::uuid
  AND reserved >= ($1::numeric)
`,
		ord.Amount,
		ord.BalanceAccountID,
		now,
		cmd.TenantID,
	)
	if err != nil {
		return CompleteOrderResult{}, fmt.Errorf("decrease reserved balance: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return CompleteOrderResult{}, ErrInsufficientReserved
	}

	if err := s.poster.PostTradeComplete(ctx, tx, TradeCompletePosting{
		TenantID:                  cmd.TenantID,
		OrderID:                   ord.OrderID,
		HoldID:                    ord.HoldID,
		ReservedLedgerAccountID:   ord.ReservedLedgerAccountID,
		SettlementLedgerAccountID: *ord.SettlementLedgerAccountID,
		CurrencyID:                ord.CurrencyID,
		Amount:                    ord.Amount,
		CashierID:                 cmd.CashierID,
		Reason:                    "manual_complete",
	}); err != nil {
		return CompleteOrderResult{}, fmt.Errorf("post trade_complete journal: %w", err)
	}

	if err := insertOutboxEvent(ctx, tx, OutboxEvent{
		TenantID:      cmd.TenantID,
		AggregateType: "order",
		AggregateID:   ord.OrderID,
		EventType:     "order_completed",
		Payload: map[string]any{
			"order_id":        ord.OrderID,
			"order_ref":       ord.OrderRef,
			"cashier_id":      cmd.CashierID,
			"completed_at_ts": now.Unix(),
		},
	}); err != nil {
		return CompleteOrderResult{}, fmt.Errorf("insert outbox order_completed: %w", err)
	}

	result := CompleteOrderResult{
		OrderID:     ord.OrderID,
		OrderRef:    ord.OrderRef,
		Status:      "completed",
		Version:     newVersion,
		CompletedAt: now,
	}

	if err := finalizeIdempotency(ctx, tx, cmd.TenantID, "complete_order", cmd.IdempotencyKey, "order", ord.OrderID, result); err != nil {
		return CompleteOrderResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CompleteOrderResult{}, fmt.Errorf("commit complete tx: %w", err)
	}

	return result, nil
}

func (s *Service) CancelOrder(ctx context.Context, cmd CancelOrderCommand) (CancelOrderResult, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return CancelOrderResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '50ms'`); err != nil {
		return CancelOrderResult{}, fmt.Errorf("set lock_timeout: %w", err)
	}

	reqHash := buildRequestHash(
		cmd.TenantID,
		cmd.OrderID,
		cmd.IdempotencyKey,
		cmd.Reason,
		fmt.Sprintf("%d", cmd.ExpectedVersion),
	)

	rec, err := acquireIdempotency(ctx, tx, cmd.TenantID, "cancel_order", cmd.IdempotencyKey, reqHash)
	if err != nil {
		return CancelOrderResult{}, err
	}
	if rec != nil {
		return decodeCachedResponse[CancelOrderResult](rec)
	}

	ord, err := lockOrderForMutation(ctx, tx, cmd.TenantID, cmd.OrderID)
	if err != nil {
		return CancelOrderResult{}, err
	}
	if ord.Status != "reserved" {
		return CancelOrderResult{}, ErrOrderNotActive
	}
	if ord.HoldStatus != "active" {
		return CancelOrderResult{}, ErrHoldNotActive
	}

	now := s.now()

	var newVersion int64
	err = tx.QueryRow(ctx, `
UPDATE core.orders
SET
	status = 'cancelled',
	cancelled_at = $3,
	cancel_reason = $4,
	version = version + 1,
	updated_at = $3
WHERE id = $1::uuid
  AND tenant_id = $2::uuid
  AND status = 'reserved'
  AND version = $5
RETURNING version
`,
		cmd.OrderID,
		cmd.TenantID,
		now,
		cmd.Reason,
		cmd.ExpectedVersion,
	).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CancelOrderResult{}, ErrVersionConflict
		}
		return CancelOrderResult{}, fmt.Errorf("cas update cancel order: %w", err)
	}

	tag, err := tx.Exec(ctx, `
UPDATE core.order_holds
SET
	status = 'released',
	released_at = $2
WHERE id = $1::uuid
  AND status = 'active'
`,
		ord.HoldID,
		now,
	)
	if err != nil {
		return CancelOrderResult{}, fmt.Errorf("release hold: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return CancelOrderResult{}, ErrHoldNotActive
	}

	tag, err = tx.Exec(ctx, `
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
		cmd.TenantID,
	)
	if err != nil {
		return CancelOrderResult{}, fmt.Errorf("release reserved balance: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return CancelOrderResult{}, ErrInsufficientReserved
	}

	if err := s.poster.PostHoldRelease(ctx, tx, HoldReleasePosting{
		TenantID:                 cmd.TenantID,
		OrderID:                  ord.OrderID,
		HoldID:                   ord.HoldID,
		AvailableLedgerAccountID: ord.AvailableLedgerAccountID,
		ReservedLedgerAccountID:  ord.ReservedLedgerAccountID,
		CurrencyID:               ord.CurrencyID,
		Amount:                   ord.Amount,
		Reason:                   cmd.Reason,
	}); err != nil {
		return CancelOrderResult{}, fmt.Errorf("post hold_release journal: %w", err)
	}

	if err := insertOutboxEvent(ctx, tx, OutboxEvent{
		TenantID:      cmd.TenantID,
		AggregateType: "order",
		AggregateID:   ord.OrderID,
		EventType:     "order_cancelled",
		Payload: map[string]any{
			"order_id":        ord.OrderID,
			"order_ref":       ord.OrderRef,
			"cancel_reason":   cmd.Reason,
			"cancelled_at_ts": now.Unix(),
		},
	}); err != nil {
		return CancelOrderResult{}, fmt.Errorf("insert outbox order_cancelled: %w", err)
	}

	result := CancelOrderResult{
		OrderID:  ord.OrderID,
		OrderRef: ord.OrderRef,
		Status:   "cancelled",
		Version:  newVersion,
	}

	if err := finalizeIdempotency(ctx, tx, cmd.TenantID, "cancel_order", cmd.IdempotencyKey, "order", ord.OrderID, result); err != nil {
		return CancelOrderResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CancelOrderResult{}, fmt.Errorf("commit cancel tx: %w", err)
	}

	return result, nil
}

func lockOrderForMutation(ctx context.Context, tx pgx.Tx, tenantID, orderID string) (lockedOrder, error) {
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
WHERE o.id = $1::uuid
  AND o.tenant_id = $2::uuid
FOR UPDATE OF o, h
`
	var out lockedOrder
	err := tx.QueryRow(ctx, q, orderID, tenantID).Scan(
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
			return lockedOrder{}, ErrOrderNotFound
		}
		return lockedOrder{}, fmt.Errorf("lock order for mutation: %w", err)
	}
	return out, nil
}

func buildPublicOrderRef() string {
	return fmt.Sprintf("ORD-%s", uuid.NewString()[:8])
}
