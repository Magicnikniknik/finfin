package orders

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTTLWorker_ExpiresReservedOrder(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
	clientRef := "client_ttl_test"

	giveCurrencyID := uuid.NewString()
	getCurrencyID := uuid.NewString()
	holdCurrencyID := giveCurrencyID

	balanceAccountID := uuid.NewString()
	availableLedgerAccountID := uuid.NewString()
	reservedLedgerAccountID := uuid.NewString()

	seedAccount(t, ctx, pool, balanceAccountID, tenantID, officeID, holdCurrencyID, "balance")
	seedAccount(t, ctx, pool, availableLedgerAccountID, tenantID, officeID, holdCurrencyID, "available_ledger")
	seedAccount(t, ctx, pool, reservedLedgerAccountID, tenantID, officeID, holdCurrencyID, "reserved_ledger")

	seedBalance(t, ctx, pool, balanceAccountID, tenantID, holdCurrencyID, "1000.000000000000000000", "0")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})

	reserveCmd := ReserveOrderCommand{
		TenantID:                  tenantID,
		ClientRef:                 clientRef,
		IdempotencyKey:            uuid.NewString(),
		OfficeID:                  officeID,
		QuoteID:                   "quote_ttl_001",
		Side:                      "buy",
		GiveCurrencyID:            giveCurrencyID,
		GetCurrencyID:             getCurrencyID,
		AmountGive:                "100.000000000000000000",
		AmountGet:                 "3550.000000000000000000",
		FixedRate:                 "35.500000000000000000",
		HoldCurrencyID:            holdCurrencyID,
		HoldAmount:                "100.000000000000000000",
		BalanceAccountID:          balanceAccountID,
		AvailableLedgerAccountID:  availableLedgerAccountID,
		ReservedLedgerAccountID:   reservedLedgerAccountID,
		SettlementLedgerAccountID: nil,
		QuotePayload:              json.RawMessage(`{"source":"ttl-test"}`),
		ExpiresAt:                 time.Now().UTC().Add(-5 * time.Second),
	}

	reserveRes, err := svc.ReserveOrder(ctx, reserveCmd)
	if err != nil {
		t.Fatalf("ReserveOrder returned error: %v", err)
	}

	worker := NewTTLWorker(
		pool,
		slog.Default(),
		RealJournalPoster{},
		TTLWorkerConfig{
			TickInterval: time.Second,
			BatchSize:    10,
			LockTimeout:  50 * time.Millisecond,
		},
	)

	ok, err := worker.processOne(ctx)
	if err != nil {
		t.Fatalf("TTL worker processOne returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected TTL worker to process one expired order")
	}

	var orderStatus string
	var orderVersion int64
	err = pool.QueryRow(ctx, `
SELECT status, version
FROM core.orders
WHERE id = $1::uuid
`, reserveRes.OrderID).Scan(&orderStatus, &orderVersion)
	if err != nil {
		t.Fatalf("load order after ttl: %v", err)
	}

	if orderStatus != "expired" {
		t.Fatalf("expected order status expired, got %q", orderStatus)
	}
	if orderVersion != 2 {
		t.Fatalf("expected order version 2 after ttl expiration, got %d", orderVersion)
	}

	var holdStatus string
	err = pool.QueryRow(ctx, `
SELECT status
FROM core.order_holds
WHERE order_id = $1::uuid
`, reserveRes.OrderID).Scan(&holdStatus)
	if err != nil {
		t.Fatalf("load hold after ttl: %v", err)
	}

	if holdStatus != "expired" {
		t.Fatalf("expected hold status expired, got %q", holdStatus)
	}

	var available, reserved string
	err = pool.QueryRow(ctx, `
SELECT available::text, reserved::text
FROM core.account_balances
WHERE account_id = $1::uuid
`, balanceAccountID).Scan(&available, &reserved)
	if err != nil {
		t.Fatalf("load balances after ttl: %v", err)
	}

	if available != "1000.000000000000000000" {
		t.Fatalf("expected available restored to 1000..., got %s", available)
	}
	if reserved != "0.000000000000000000" {
		t.Fatalf("expected reserved restored to 0..., got %s", reserved)
	}

	var releaseJournalCount int
	err = pool.QueryRow(ctx, `
SELECT count(*)
FROM core.ledger_journals
WHERE order_id = $1::uuid
  AND kind = 'hold_release'
`, reserveRes.OrderID).Scan(&releaseJournalCount)
	if err != nil {
		t.Fatalf("count hold_release journals: %v", err)
	}

	if releaseJournalCount != 1 {
		t.Fatalf("expected 1 hold_release journal, got %d", releaseJournalCount)
	}

	var expiredOutboxCount int
	err = pool.QueryRow(ctx, `
SELECT count(*)
FROM core.outbox_events
WHERE aggregate_id = $1::uuid
  AND event_type = 'order_expired'
`, reserveRes.OrderID).Scan(&expiredOutboxCount)
	if err != nil {
		t.Fatalf("count order_expired outbox events: %v", err)
	}

	if expiredOutboxCount != 1 {
		t.Fatalf("expected 1 order_expired outbox event, got %d", expiredOutboxCount)
	}
}
