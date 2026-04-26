package orders

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReserveOrder_EndToEnd(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
	clientRef := "client_001"

	giveCurrencyID := uuid.NewString()
	getCurrencyID := uuid.NewString()
	holdCurrencyID := giveCurrencyID

	balanceAccountID := uuid.NewString()
	availableLedgerAccountID := uuid.NewString()
	reservedLedgerAccountID := uuid.NewString()

	seedAccount(t, ctx, pool, balanceAccountID, tenantID, officeID, holdCurrencyID, "balance")
	seedAccount(t, ctx, pool, availableLedgerAccountID, tenantID, officeID, holdCurrencyID, "available_ledger")
	seedAccount(t, ctx, pool, reservedLedgerAccountID, tenantID, officeID, holdCurrencyID, "reserved_ledger")
	seedAccountWiring(t, ctx, pool, tenantID, officeID, holdCurrencyID, balanceAccountID, availableLedgerAccountID, reservedLedgerAccountID, nil)

	seedBalance(t, ctx, pool, balanceAccountID, tenantID, holdCurrencyID, "1000.000000000000000000", "0")
	seedCanonicalQuote(t, ctx, pool, "quote_001", tenantID, officeID, "sell", giveCurrencyID, getCurrencyID, "100.000000000000000000", "3550.000000000000000000", "35.500000000000000000", time.Now().UTC().Add(30*time.Minute).Truncate(time.Second), "active")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})

	cmd := ReserveOrderCommand{
		TenantID:       tenantID,
		ClientRef:      clientRef,
		IdempotencyKey: uuid.NewString(),
		OfficeID:       officeID,
		QuoteID:        "quote_001",
	}

	result, err := svc.ReserveOrder(ctx, cmd)
	if err != nil {
		t.Fatalf("ReserveOrder returned error: %v", err)
	}

	if result.OrderID == "" {
		t.Fatal("expected non-empty OrderID")
	}
	if result.OrderRef == "" {
		t.Fatal("expected non-empty OrderRef")
	}
	if result.Status != "reserved" {
		t.Fatalf("expected status reserved, got %q", result.Status)
	}
	if result.Version != 1 {
		t.Fatalf("expected version 1, got %d", result.Version)
	}

	var (
		orderStatus     string
		orderVersion    int64
		orderClientRef  string
		orderSide       string
		orderAmountGive string
		orderAmountGet  string
		orderFixedRate  string
	)
	err = pool.QueryRow(ctx, `
SELECT status, version, client_ref, side, amount_give::text, amount_get::text, fixed_rate::text
FROM core.orders
WHERE id = $1::uuid
`, result.OrderID).Scan(
		&orderStatus,
		&orderVersion,
		&orderClientRef,
		&orderSide,
		&orderAmountGive,
		&orderAmountGet,
		&orderFixedRate,
	)
	if err != nil {
		t.Fatalf("load order: %v", err)
	}

	if orderStatus != "reserved" {
		t.Fatalf("expected DB order status reserved, got %q", orderStatus)
	}
	if orderVersion != 1 {
		t.Fatalf("expected DB order version 1, got %d", orderVersion)
	}
	if orderClientRef != clientRef {
		t.Fatalf("expected client_ref %q, got %q", clientRef, orderClientRef)
	}
	if orderSide != "sell" {
		t.Fatalf("expected side buy, got %q", orderSide)
	}
	if orderAmountGive != "100.000000000000000000" {
		t.Fatalf("unexpected amount_give: %s", orderAmountGive)
	}
	if orderAmountGet != "3550.000000000000000000" {
		t.Fatalf("unexpected amount_get: %s", orderAmountGet)
	}
	if orderFixedRate != "35.500000000000000000" {
		t.Fatalf("unexpected fixed_rate: %s", orderFixedRate)
	}

	var (
		holdStatus     string
		holdAmount     string
		holdBalanceAcc string
		holdAvailAcc   string
		holdResAcc     string
	)
	err = pool.QueryRow(ctx, `
SELECT
	status,
	amount::text,
	balance_account_id::text,
	available_ledger_account_id::text,
	reserved_ledger_account_id::text
FROM core.order_holds
WHERE order_id = $1::uuid
`, result.OrderID).Scan(
		&holdStatus,
		&holdAmount,
		&holdBalanceAcc,
		&holdAvailAcc,
		&holdResAcc,
	)
	if err != nil {
		t.Fatalf("load hold: %v", err)
	}

	if holdStatus != "active" {
		t.Fatalf("expected active hold, got %q", holdStatus)
	}
	if holdAmount != "100.000000000000000000" {
		t.Fatalf("unexpected hold amount: %s", holdAmount)
	}
	if holdBalanceAcc != balanceAccountID {
		t.Fatalf("unexpected balance_account_id: %s", holdBalanceAcc)
	}
	if holdAvailAcc != availableLedgerAccountID {
		t.Fatalf("unexpected available_ledger_account_id: %s", holdAvailAcc)
	}
	if holdResAcc != reservedLedgerAccountID {
		t.Fatalf("unexpected reserved_ledger_account_id: %s", holdResAcc)
	}

	var available, reserved string
	err = pool.QueryRow(ctx, `
SELECT available::text, reserved::text
FROM core.account_balances
WHERE account_id = $1::uuid
`, balanceAccountID).Scan(&available, &reserved)
	if err != nil {
		t.Fatalf("load account_balances: %v", err)
	}

	if available != "900.000000000000000000" {
		t.Fatalf("expected available=900..., got %s", available)
	}
	if reserved != "100.000000000000000000" {
		t.Fatalf("expected reserved=100..., got %s", reserved)
	}

	var journalCount int
	err = pool.QueryRow(ctx, `
SELECT count(*)
FROM core.ledger_journals
WHERE order_id = $1::uuid
  AND kind = 'hold_create'
`, result.OrderID).Scan(&journalCount)
	if err != nil {
		t.Fatalf("count ledger_journals: %v", err)
	}
	if journalCount != 1 {
		t.Fatalf("expected 1 ledger_journal, got %d", journalCount)
	}

	type ledgerEntry struct {
		AccountID string
		Direction string
		Amount    string
	}

	rows, err := pool.Query(ctx, `
SELECT account_id::text, direction, amount::text
FROM core.ledger_entries
WHERE journal_id IN (
	SELECT id FROM core.ledger_journals
	WHERE order_id = $1::uuid
	  AND kind = 'hold_create'
)
ORDER BY id ASC
`, result.OrderID)
	if err != nil {
		t.Fatalf("query ledger_entries: %v", err)
	}
	defer rows.Close()

	var entries []ledgerEntry
	for rows.Next() {
		var e ledgerEntry
		if err := rows.Scan(&e.AccountID, &e.Direction, &e.Amount); err != nil {
			t.Fatalf("scan ledger entry: %v", err)
		}
		entries = append(entries, e)
	}
	if rows.Err() != nil {
		t.Fatalf("iterate ledger_entries: %v", rows.Err())
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 ledger entries, got %d", len(entries))
	}

	if entries[0].AccountID != availableLedgerAccountID || entries[0].Direction != "debit" {
		t.Fatalf("unexpected first ledger entry: %+v", entries[0])
	}
	if entries[1].AccountID != reservedLedgerAccountID || entries[1].Direction != "credit" {
		t.Fatalf("unexpected second ledger entry: %+v", entries[1])
	}
	if entries[0].Amount != "100.000000000000000000" || entries[1].Amount != "100.000000000000000000" {
		t.Fatalf("unexpected ledger amounts: %+v", entries)
	}

	var outboxCount int
	err = pool.QueryRow(ctx, `
SELECT count(*)
FROM core.outbox_events
WHERE aggregate_id = $1::uuid
  AND event_type = 'order_reserved'
`, result.OrderID).Scan(&outboxCount)
	if err != nil {
		t.Fatalf("count outbox events: %v", err)
	}
	if outboxCount != 1 {
		t.Fatalf("expected 1 outbox event, got %d", outboxCount)
	}

	var idemStatus string
	err = pool.QueryRow(ctx, `
SELECT status
FROM core.idempotency_keys
WHERE tenant_id = $1::uuid
  AND scope = 'reserve_order'
  AND idem_key = $2
`, tenantID, cmd.IdempotencyKey).Scan(&idemStatus)
	if err != nil {
		t.Fatalf("load idempotency key: %v", err)
	}
	if idemStatus != "completed" {
		t.Fatalf("expected idempotency status completed, got %q", idemStatus)
	}

	cachedResult, err := svc.ReserveOrder(ctx, cmd)
	if err != nil {
		t.Fatalf("second ReserveOrder returned error: %v", err)
	}

	if cachedResult.OrderID != result.OrderID {
		t.Fatalf("expected cached OrderID %q, got %q", result.OrderID, cachedResult.OrderID)
	}
	if cachedResult.OrderRef != result.OrderRef {
		t.Fatalf("expected cached OrderRef %q, got %q", result.OrderRef, cachedResult.OrderRef)
	}
	if cachedResult.Version != result.Version {
		t.Fatalf("expected cached Version %d, got %d", result.Version, cachedResult.Version)
	}

	var ordersCount int
	err = pool.QueryRow(ctx, `
SELECT count(*)
FROM core.orders
WHERE tenant_id = $1::uuid
`, tenantID).Scan(&ordersCount)
	if err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if ordersCount != 1 {
		t.Fatalf("expected exactly 1 order after idempotent replay, got %d", ordersCount)
	}
}

func openTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL_TEST")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("DATABASE_URL_TEST or DATABASE_URL is required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pgx pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func resetAndMigrateCoreSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS core CASCADE`); err != nil {
		t.Fatalf("drop core schema: %v", err)
	}

	for _, file := range []string{"0001_core_schema.sql", "0002_quote_snapshots.sql", "0003_account_wiring.sql", "0004_pricing_engine.sql", "0005_cash_shifts.sql"} {
		migrationPath := mustFindMigrationPath(t, file)
		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration file %s: %v", file, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("apply migration %s: %v", file, err)
		}
	}
}

func mustFindMigrationPath(t *testing.T, filename string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	base := filepath.Dir(currentFile)
	path := filepath.Clean(filepath.Join(base, "..", "..", "migrations", filename))

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("migration file not found at %s: %v", path, err)
	}

	return path
}

func seedCanonicalQuote(t *testing.T, ctx context.Context, pool *pgxpool.Pool, quoteID, tenantID, officeID, side, giveCurrencyID, getCurrencyID, amountGive, amountGet, fixedRate string, expiresAt time.Time, status string) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO core.quotes (
	id,
	tenant_id,
	office_id,
	client_ref,
	side,
	input_mode,
	requested_amount,
	give_currency_id,
	get_currency_id,
	amount_give,
	amount_get,
	fixed_rate,
	base_rate_snapshot,
	margin_bps_applied,
	fixed_fee_applied,
	rounding_precision,
	rounding_mode,
	status,
	expires_at,
	created_at
)
VALUES (
	$1,
	$2::uuid,
	$3::uuid,
	'seed',
	$4,
	'give',
	$5::numeric,
	$6::uuid,
	$7::uuid,
	$5::numeric,
	$8::numeric,
	$9::numeric,
	$9::numeric,
	0,
	0,
	2,
	'half_up',
	$10,
	$11,
	now()
)
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    expires_at = EXCLUDED.expires_at,
    amount_give = EXCLUDED.amount_give,
    amount_get = EXCLUDED.amount_get,
    fixed_rate = EXCLUDED.fixed_rate,
    consumed_at = NULL,
    expired_at = NULL
`, quoteID, tenantID, officeID, side, amountGive, giveCurrencyID, getCurrencyID, amountGet, fixedRate, status, expiresAt.UTC())
	if err != nil {
		t.Fatalf("seed canonical quote: %v", err)
	}
}

func seedAccount(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	accountID string,
	tenantID string,
	officeID string,
	currencyID string,
	accountType string,
) {
	t.Helper()

	_, err := pool.Exec(ctx, `
INSERT INTO core.accounts (
	id,
	tenant_id,
	office_id,
	currency_id,
	account_type
)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5)
`,
		accountID,
		tenantID,
		officeID,
		currencyID,
		accountType,
	)
	if err != nil {
		t.Fatalf("seed account %s: %v", accountType, err)
	}
}

func seedAccountWiring(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	tenantID string,
	officeID string,
	currencyID string,
	balanceAccountID string,
	availableLedgerAccountID string,
	reservedLedgerAccountID string,
	settlementLedgerAccountID *string,
) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO core.account_wiring (
	tenant_id,
	office_id,
	currency_id,
	balance_account_id,
	available_ledger_account_id,
	reserved_ledger_account_id,
	settlement_ledger_account_id,
	created_at,
	updated_at
)
VALUES (
	$1::uuid,
	$2::uuid,
	$3::uuid,
	$4::uuid,
	$5::uuid,
	$6::uuid,
	$7::uuid,
	now(),
	now()
)
ON CONFLICT (tenant_id, office_id, currency_id) DO UPDATE
SET
	balance_account_id = EXCLUDED.balance_account_id,
	available_ledger_account_id = EXCLUDED.available_ledger_account_id,
	reserved_ledger_account_id = EXCLUDED.reserved_ledger_account_id,
	settlement_ledger_account_id = EXCLUDED.settlement_ledger_account_id,
	updated_at = now()
`,
		tenantID,
		officeID,
		currencyID,
		balanceAccountID,
		availableLedgerAccountID,
		reservedLedgerAccountID,
		settlementLedgerAccountID,
	)
	if err != nil {
		t.Fatalf("seed account wiring: %v", err)
	}
}

func seedBalance(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	accountID string,
	tenantID string,
	currencyID string,
	available string,
	reserved string,
) {
	t.Helper()

	_, err := pool.Exec(ctx, `
INSERT INTO core.account_balances (
	account_id,
	tenant_id,
	currency_id,
	available,
	reserved
)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4::numeric, $5::numeric)
`,
		accountID,
		tenantID,
		currencyID,
		available,
		reserved,
	)
	if err != nil {
		t.Fatalf("seed balance: %v", err)
	}
}

func seedOpenShift(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	tenantID string,
	officeID string,
	cashierID string,
) {
	t.Helper()
	_, err := pool.Exec(ctx, `
INSERT INTO core.cash_shifts (
	tenant_id,
	office_id,
	cashier_id,
	status,
	opened_at,
	created_at,
	updated_at
)
VALUES (
	$1::uuid,
	$2::uuid,
	$3,
	'open',
	now(),
	now(),
	now()
)
ON CONFLICT DO NOTHING
`, tenantID, officeID, cashierID)
	if err != nil {
		t.Fatalf("seed open shift: %v", err)
	}
}

func TestCompleteOrder_WrongExpectedVersionConflict(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
	clientRef := "client_002"

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
	seedAccountWiring(t, ctx, pool, tenantID, officeID, holdCurrencyID, balanceAccountID, availableLedgerAccountID, reservedLedgerAccountID, nil)
	seedAccount(t, ctx, pool, settlementLedgerAccountID, tenantID, officeID, holdCurrencyID, "settlement")
	seedAccountWiring(t, ctx, pool, tenantID, officeID, holdCurrencyID, balanceAccountID, availableLedgerAccountID, reservedLedgerAccountID, &settlementLedgerAccountID)

	seedBalance(t, ctx, pool, balanceAccountID, tenantID, holdCurrencyID, "500.000000000000000000", "0")
	seedCanonicalQuote(t, ctx, pool, "quote_002", tenantID, officeID, "sell", giveCurrencyID, getCurrencyID, "50.000000000000000000", "1800.000000000000000000", "36.000000000000000000", time.Now().UTC().Add(30*time.Minute).Truncate(time.Second), "active")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})

	reserveResult, err := svc.ReserveOrder(ctx, ReserveOrderCommand{
		TenantID:       tenantID,
		ClientRef:      clientRef,
		IdempotencyKey: uuid.NewString(),
		OfficeID:       officeID,
		QuoteID:        "quote_002",
	})
	if err != nil {
		t.Fatalf("ReserveOrder returned error: %v", err)
	}
	seedOpenShift(t, ctx, pool, tenantID, officeID, "cashier_001")

	_, err = svc.CompleteOrder(ctx, CompleteOrderCommand{
		TenantID:        tenantID,
		OrderID:         reserveResult.OrderID,
		ExpectedVersion: 2,
		IdempotencyKey:  uuid.NewString(),
		CashierID:       "cashier_001",
	})
	if err == nil {
		t.Fatal("expected ErrVersionConflict, got nil")
	}
	if err != ErrVersionConflict {
		t.Fatalf("expected ErrVersionConflict, got %v", err)
	}
}

func TestReserveOrder_FailsOnConsumedQuote(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
	giveCurrencyID := uuid.NewString()
	getCurrencyID := uuid.NewString()
	holdCurrencyID := giveCurrencyID

	balanceAccountID := uuid.NewString()
	availableLedgerAccountID := uuid.NewString()
	reservedLedgerAccountID := uuid.NewString()

	seedAccount(t, ctx, pool, balanceAccountID, tenantID, officeID, holdCurrencyID, "balance")
	seedAccount(t, ctx, pool, availableLedgerAccountID, tenantID, officeID, holdCurrencyID, "available_ledger")
	seedAccount(t, ctx, pool, reservedLedgerAccountID, tenantID, officeID, holdCurrencyID, "reserved_ledger")
	seedAccountWiring(t, ctx, pool, tenantID, officeID, holdCurrencyID, balanceAccountID, availableLedgerAccountID, reservedLedgerAccountID, nil)
	seedBalance(t, ctx, pool, balanceAccountID, tenantID, holdCurrencyID, "1000.000000000000000000", "0")
	seedCanonicalQuote(t, ctx, pool, "quote_consumed_1", tenantID, officeID, "sell", giveCurrencyID, getCurrencyID, "100.000000000000000000", "3550.000000000000000000", "35.500000000000000000", time.Now().UTC().Add(30*time.Minute), "consumed")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})
	_, err := svc.ReserveOrder(ctx, ReserveOrderCommand{
		TenantID:       tenantID,
		ClientRef:      "client_consumed",
		IdempotencyKey: uuid.NewString(),
		OfficeID:       officeID,
		QuoteID:        "quote_consumed_1",
	})
	if err == nil {
		t.Fatal("expected quote consumed error")
	}
	if !errors.Is(err, ErrQuoteAlreadyConsumed) {
		t.Fatalf("expected ErrQuoteAlreadyConsumed, got %v", err)
	}
}

func TestReserveOrder_RollbackKeepsQuoteActiveOnReserveFailure(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
	giveCurrencyID := uuid.NewString()
	getCurrencyID := uuid.NewString()
	holdCurrencyID := giveCurrencyID

	balanceAccountID := uuid.NewString()
	availableLedgerAccountID := uuid.NewString()
	reservedLedgerAccountID := uuid.NewString()

	seedAccount(t, ctx, pool, balanceAccountID, tenantID, officeID, holdCurrencyID, "balance")
	seedAccount(t, ctx, pool, availableLedgerAccountID, tenantID, officeID, holdCurrencyID, "available_ledger")
	seedAccount(t, ctx, pool, reservedLedgerAccountID, tenantID, officeID, holdCurrencyID, "reserved_ledger")
	seedAccountWiring(t, ctx, pool, tenantID, officeID, holdCurrencyID, balanceAccountID, availableLedgerAccountID, reservedLedgerAccountID, nil)
	seedBalance(t, ctx, pool, balanceAccountID, tenantID, holdCurrencyID, "10.000000000000000000", "0")
	seedCanonicalQuote(t, ctx, pool, "quote_active_rollback", tenantID, officeID, "sell", giveCurrencyID, getCurrencyID, "100.000000000000000000", "3550.000000000000000000", "35.500000000000000000", time.Now().UTC().Add(30*time.Minute), "active")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})
	_, err := svc.ReserveOrder(ctx, ReserveOrderCommand{
		TenantID:       tenantID,
		ClientRef:      "client_rollback",
		IdempotencyKey: uuid.NewString(),
		OfficeID:       officeID,
		QuoteID:        "quote_active_rollback",
	})
	if err == nil {
		t.Fatal("expected reserve to fail on insufficient funds")
	}

	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM core.quotes WHERE id = 'quote_active_rollback'`).Scan(&status); err != nil {
		t.Fatalf("load quote status: %v", err)
	}
	if status != "active" {
		t.Fatalf("expected quote to stay active on rollback, got %q", status)
	}
}

func TestReserveOrder_FailsOnExpiredQuote(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
	giveCurrencyID := uuid.NewString()
	getCurrencyID := uuid.NewString()
	holdCurrencyID := giveCurrencyID

	balanceAccountID := uuid.NewString()
	availableLedgerAccountID := uuid.NewString()
	reservedLedgerAccountID := uuid.NewString()

	seedAccount(t, ctx, pool, balanceAccountID, tenantID, officeID, holdCurrencyID, "balance")
	seedAccount(t, ctx, pool, availableLedgerAccountID, tenantID, officeID, holdCurrencyID, "available_ledger")
	seedAccount(t, ctx, pool, reservedLedgerAccountID, tenantID, officeID, holdCurrencyID, "reserved_ledger")
	seedAccountWiring(t, ctx, pool, tenantID, officeID, holdCurrencyID, balanceAccountID, availableLedgerAccountID, reservedLedgerAccountID, nil)
	seedBalance(t, ctx, pool, balanceAccountID, tenantID, holdCurrencyID, "1000.000000000000000000", "0")
	seedCanonicalQuote(t, ctx, pool, "quote_expired_1", tenantID, officeID, "sell", giveCurrencyID, getCurrencyID, "100.000000000000000000", "3550.000000000000000000", "35.500000000000000000", time.Now().UTC().Add(-5*time.Second), "active")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})
	_, err := svc.ReserveOrder(ctx, ReserveOrderCommand{
		TenantID:       tenantID,
		ClientRef:      "client_expired",
		IdempotencyKey: uuid.NewString(),
		OfficeID:       officeID,
		QuoteID:        "quote_expired_1",
	})
	if err == nil {
		t.Fatal("expected expired quote error")
	}
	if !errors.Is(err, ErrQuoteExpired) {
		t.Fatalf("expected ErrQuoteExpired, got %v", err)
	}
}

func TestReserveOrder_FailsWhenAccountWiringMissing(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	officeID := uuid.NewString()
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
	seedCanonicalQuote(t, ctx, pool, "quote_no_wiring_1", tenantID, officeID, "sell", giveCurrencyID, getCurrencyID, "100.000000000000000000", "3550.000000000000000000", "35.500000000000000000", time.Now().UTC().Add(30*time.Minute), "active")

	svc := NewService(pool, slog.Default(), RealJournalPoster{})
	_, err := svc.ReserveOrder(ctx, ReserveOrderCommand{
		TenantID:       tenantID,
		ClientRef:      "client_no_wiring",
		IdempotencyKey: uuid.NewString(),
		OfficeID:       officeID,
		QuoteID:        "quote_no_wiring_1",
	})
	if err == nil {
		t.Fatal("expected account wiring error")
	}
	if !errors.Is(err, ErrAccountWiringNotFound) {
		t.Fatalf("expected ErrAccountWiringNotFound, got %v", err)
	}

	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM core.quotes WHERE id = 'quote_no_wiring_1'`).Scan(&status); err != nil {
		t.Fatalf("load quote status: %v", err)
	}
	if status != "active" {
		t.Fatalf("expected quote to stay active when wiring is missing, got %q", status)
	}
}
