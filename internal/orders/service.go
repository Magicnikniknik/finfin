package orders

import (
	"context"
	"encoding/json"
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

	now := s.now()
	quote, err := lockActiveQuoteForReserve(ctx, tx, cmd.TenantID, cmd.OfficeID, cmd.QuoteID, now)
	if err != nil {
		return ReserveOrderResult{}, err
	}
	holdCurrencyID, holdAmount, err := deriveHoldFromQuote(quote)
	if err != nil {
		return ReserveOrderResult{}, err
	}
	wiring, err := lockAccountWiringForReserve(ctx, tx, cmd.TenantID, cmd.OfficeID, holdCurrencyID)
	if err != nil {
		return ReserveOrderResult{}, err
	}

	reqHash := buildRequestHash(
		cmd.TenantID,
		cmd.ClientRef,
		cmd.IdempotencyKey,
		cmd.OfficeID,
		cmd.QuoteID,
		quote.Side,
		quote.GiveCurrencyID,
		quote.GetCurrencyID,
		quote.AmountGive,
		quote.AmountGet,
		quote.FixedRate,
		quote.ExpiresAt.UTC().Format(time.RFC3339),
	)

	rec, err := acquireIdempotency(ctx, tx, cmd.TenantID, "reserve_order", cmd.IdempotencyKey, reqHash)
	if err != nil {
		return ReserveOrderResult{}, err
	}
	if rec != nil {
		return decodeCachedResponse[ReserveOrderResult](rec)
	}

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
		holdAmount,
		wiring.BalanceAccountID,
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
		quote.Side,
		quote.GiveCurrencyID,
		quote.GetCurrencyID,
		quote.AmountGive,
		quote.AmountGet,
		quote.FixedRate,
		canonicalQuotePayload(quote, holdCurrencyID, holdAmount),
		now,
		quote.ExpiresAt.UTC(),
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
		wiring.BalanceAccountID,
		wiring.AvailableLedgerAccountID,
		wiring.ReservedLedgerAccountID,
		wiring.SettlementLedgerAccountID,
		holdCurrencyID,
		holdAmount,
		quote.ExpiresAt.UTC(),
		now,
	)
	if err != nil {
		return ReserveOrderResult{}, fmt.Errorf("insert order_hold: %w", err)
	}

	if err := s.poster.PostHoldCreate(ctx, tx, HoldCreatePosting{
		TenantID:                 cmd.TenantID,
		OrderID:                  orderID,
		HoldID:                   holdID,
		AvailableLedgerAccountID: wiring.AvailableLedgerAccountID,
		ReservedLedgerAccountID:  wiring.ReservedLedgerAccountID,
		CurrencyID:               holdCurrencyID,
		Amount:                   holdAmount,
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
			"expires_at_ts": quote.ExpiresAt.UTC().Unix(),
		},
	}); err != nil {
		return ReserveOrderResult{}, fmt.Errorf("insert outbox order_reserved: %w", err)
	}

	tag, err = tx.Exec(ctx, `
	UPDATE core.quotes
	SET
		status = 'consumed',
	consumed_at = $4
WHERE id = $1
  AND tenant_id = $2::uuid
  AND office_id = $3::uuid
  AND status = 'active'
`,
		cmd.QuoteID,
		cmd.TenantID,
		cmd.OfficeID,
		now,
	)
	consumed_at = $4
WHERE id = $1
  AND tenant_id = $2::uuid
  AND office_id = $3::uuid
  AND status = 'active'
`,
		cmd.QuoteID,
		cmd.TenantID,
		cmd.OfficeID,
		now,
	)
	if err != nil {
		return ReserveOrderResult{}, fmt.Errorf("consume quote: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ReserveOrderResult{}, ErrQuoteAlreadyConsumed
	}

	result := ReserveOrderResult{
		OrderID:   orderID,
		OrderRef:  orderRef,
		Status:    "reserved",
		ExpiresAt: quote.ExpiresAt.UTC(),
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

func lockActiveQuoteForReserve(ctx context.Context, tx pgx.Tx, tenantID, officeID, quoteID string, now time.Time) (lockedQuote, error) {
	const q = `
SELECT
	id,
	tenant_id::text,
	office_id::text,
	side,
	give_currency_id::text,
	get_currency_id::text,
	amount_give::text,
	amount_get::text,
	fixed_rate::text,
	expires_at,
	status
FROM core.quotes
WHERE id = $1
  AND tenant_id = $2::uuid
  AND office_id = $3::uuid
FOR UPDATE
`
	var out lockedQuote
	if err := tx.QueryRow(ctx, q, quoteID, tenantID, officeID).Scan(
		&out.QuoteID,
		&out.TenantID,
		&out.OfficeID,
		&out.Side,
		&out.GiveCurrencyID,
		&out.GetCurrencyID,
		&out.AmountGive,
		&out.AmountGet,
		&out.FixedRate,
		&out.ExpiresAt,
		&out.Status,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return lockedQuote{}, ErrQuoteNotFound
		}
		return lockedQuote{}, fmt.Errorf("lock quote for reserve: %w", err)
	}

	if out.Status != "active" {
		if out.Status == "expired" {
			return lockedQuote{}, ErrQuoteExpired
		}
		return lockedQuote{}, ErrQuoteAlreadyConsumed
	}
	if !out.ExpiresAt.UTC().After(now.UTC()) {
		return lockedQuote{}, ErrQuoteExpired
	}

	return out, nil
}

func deriveHoldFromQuote(q lockedQuote) (string, string, error) {
	switch q.Side {
	case "buy":
		return q.GetCurrencyID, q.AmountGet, nil
	case "sell":
		return q.GiveCurrencyID, q.AmountGive, nil
	default:
		return "", "", fmt.Errorf("unsupported quote side %q", q.Side)
	}
}

func lockAccountWiringForReserve(ctx context.Context, tx pgx.Tx, tenantID, officeID, currencyID string) (resolvedWiring, error) {
	const q = `
SELECT
	balance_account_id::text,
	available_ledger_account_id::text,
	reserved_ledger_account_id::text,
	settlement_ledger_account_id::text
FROM core.account_wiring
WHERE tenant_id = $1::uuid
  AND office_id = $2::uuid
  AND currency_id = $3::uuid
LIMIT 1
`
	var out resolvedWiring
	if err := tx.QueryRow(ctx, q, tenantID, officeID, currencyID).Scan(
		&out.BalanceAccountID,
		&out.AvailableLedgerAccountID,
		&out.ReservedLedgerAccountID,
		&out.SettlementLedgerAccountID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resolvedWiring{}, ErrAccountWiringNotFound
		}
		return resolvedWiring{}, fmt.Errorf("resolve account wiring: %w", err)
	}
	return out, nil
}

func canonicalQuotePayload(q lockedQuote, holdCurrencyID, holdAmount string) json.RawMessage {
	payload, _ := json.Marshal(map[string]any{
		"quote_id":   q.QuoteID,
		"office_id":  q.OfficeID,
		"side":       q.Side,
		"expires_at": q.ExpiresAt.UTC().Format(time.RFC3339),
		"give":       map[string]any{"amount": q.AmountGive, "currency_id": q.GiveCurrencyID},
		"get":        map[string]any{"amount": q.AmountGet, "currency_id": q.GetCurrencyID},
		"fixed_rate": q.FixedRate,
		"hold":       map[string]any{"currency_id": holdCurrencyID, "amount": holdAmount},
	})
	return payload
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
		if ord.Status == "completed" || ord.Status == "cancelled" || ord.Status == "expired" {
			return CompleteOrderResult{}, ErrVersionConflict
		}
		return CompleteOrderResult{}, ErrOrderNotActive
	}
	if ord.HoldStatus != "active" {
		return CompleteOrderResult{}, ErrHoldNotActive
	}
	if ord.SettlementLedgerAccountID == nil || *ord.SettlementLedgerAccountID == "" {
		return CompleteOrderResult{}, fmt.Errorf("missing settlement ledger account for order %s", ord.OrderID)
	}
	if err := ensureOpenShiftForCashier(ctx, tx, cmd.TenantID, ord.OfficeID, cmd.CashierID); err != nil {
		return CompleteOrderResult{}, err
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

func ensureOpenShiftForCashier(ctx context.Context, tx pgx.Tx, tenantID, officeID, cashierID string) error {
	const q = `
SELECT id
FROM core.cash_shifts
WHERE tenant_id = $1::uuid
  AND office_id = $2::uuid
  AND cashier_id = $3
  AND status = 'open'
LIMIT 1
`
	var shiftID string
	if err := tx.QueryRow(ctx, q, tenantID, officeID, cashierID).Scan(&shiftID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrShiftNotOpen
		}
		return fmt.Errorf("check open shift: %w", err)
	}
	return nil
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
