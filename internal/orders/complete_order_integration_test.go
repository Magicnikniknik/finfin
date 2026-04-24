package orders

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCompleteOrder_VersionConflict(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
	clientRef := "client_occ_test"

	giveCurrencyID := uuid.NewString()
	getCurrencyID := uuid.NewString()
	holdCurrencyID := giveCurrencyID

	balanceAccountID := uuid.NewString()
	availableLedgerAccountID := uuid.NewString()
	reservedLedgerAccountID := uuid.NewString()
	settlementLedgerAccountID := uuid.NewString()

	seedAccount(t, ctx, pool, balanceAccountID, tenantID, officeID, holdCurrencyID, "balance")
	seedAccount(t, ctx, pool, availableLedgerAccountID, tenantID, officeID, holdCurrencyID, "available_ledger")
	seedAccount(t, ctx, pool, reservedLedgerAccountID, tenantID, officeID, holdCurrencyID, "reserved_ledger")
	seedAccount(t, ctx, pool, settlementLedgerAccountID, tenantID, officeID, holdCurrencyID, "settlement")

	seedBalance(t, ctx, pool, balanceAccountID, tenantID, holdCurrencyID, "1000.000000000000000000", "0")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})

	reserveCmd := ReserveOrderCommand{
		TenantID:                  tenantID,
		ClientRef:                 clientRef,
		IdempotencyKey:            uuid.NewString(),
		OfficeID:                  officeID,
		QuoteID:                   "quote_occ_001",
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
		SettlementLedgerAccountID: &settlementLedgerAccountID,
		QuotePayload:              json.RawMessage(`{"source":"occ-test"}`),
		ExpiresAt:                 time.Now().UTC().Add(10 * time.Minute).Truncate(time.Second),
	}

	resReserve, err := svc.ReserveOrder(ctx, reserveCmd)
	if err != nil {
		t.Fatalf("ReserveOrder returned error: %v", err)
	}
	if resReserve.Version != 1 {
		t.Fatalf("expected initial version 1, got %d", resReserve.Version)
	}

	completeCmdA := CompleteOrderCommand{
		TenantID:        tenantID,
		OrderID:         resReserve.OrderID,
		ExpectedVersion: 1,
		IdempotencyKey:  uuid.NewString(),
		CashierID:       "cashier_1",
	}

	resCompleteA, err := svc.CompleteOrder(ctx, completeCmdA)
	if err != nil {
		t.Fatalf("CompleteOrder A returned error: %v", err)
	}
	if resCompleteA.Status != "completed" {
		t.Fatalf("expected completed status, got %q", resCompleteA.Status)
	}
	if resCompleteA.Version != 2 {
		t.Fatalf("expected version 2 after first complete, got %d", resCompleteA.Version)
	}

	completeCmdB := CompleteOrderCommand{
		TenantID:        tenantID,
		OrderID:         resReserve.OrderID,
		ExpectedVersion: 1,
		IdempotencyKey:  uuid.NewString(),
		CashierID:       "cashier_2",
	}

	_, err = svc.CompleteOrder(ctx, completeCmdB)
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("expected ErrVersionConflict, got %v", err)
	}

	var finalVersion int64
	var finalStatus string
	err = pool.QueryRow(ctx, `
SELECT version, status
FROM core.orders
WHERE id = $1::uuid
`, resReserve.OrderID).Scan(&finalVersion, &finalStatus)
	if err != nil {
		t.Fatalf("load final order state: %v", err)
	}

	if finalVersion != 2 {
		t.Fatalf("expected final version 2, got %d", finalVersion)
	}
	if finalStatus != "completed" {
		t.Fatalf("expected final status completed, got %q", finalStatus)
	}

	var holdStatus string
	err = pool.QueryRow(ctx, `
SELECT status
FROM core.order_holds
WHERE order_id = $1::uuid
`, resReserve.OrderID).Scan(&holdStatus)
	if err != nil {
		t.Fatalf("load final hold status: %v", err)
	}
	if holdStatus != "consumed" {
		t.Fatalf("expected hold status consumed, got %q", holdStatus)
	}

	var available, reserved string
	err = pool.QueryRow(ctx, `
SELECT available::text, reserved::text
FROM core.account_balances
WHERE account_id = $1::uuid
`, balanceAccountID).Scan(&available, &reserved)
	if err != nil {
		t.Fatalf("load final balances: %v", err)
	}

	if available != "900.000000000000000000" {
		t.Fatalf("expected available to stay 900..., got %s", available)
	}
	if reserved != "0.000000000000000000" {
		t.Fatalf("expected reserved to become 0..., got %s", reserved)
	}

	var completedEventCount int
	err = pool.QueryRow(ctx, `
SELECT count(*)
FROM core.outbox_events
WHERE aggregate_id = $1::uuid
  AND event_type = 'order_completed'
`, resReserve.OrderID).Scan(&completedEventCount)
	if err != nil {
		t.Fatalf("count completed outbox events: %v", err)
	}

	if completedEventCount != 1 {
		t.Fatalf("expected exactly 1 order_completed outbox event, got %d", completedEventCount)
	}
}
