package pricing

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestServiceIntegration_CalculateQuote_InsertsCoreQuote(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL_TEST")
	if dsn == "" {
		t.Skip("DATABASE_URL_TEST is required")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	for _, migration := range []string{
		"migrations/0001_core_schema.sql",
		"migrations/0002_quote_snapshots.sql",
		"migrations/0003_account_wiring.sql",
		"migrations/0004_pricing_engine.sql",
		"migrations/0005_cash_shifts.sql",
	} {
		b, readErr := os.ReadFile(filepath.Join("..", "..", filepath.Clean(migration)))
		if readErr != nil {
			t.Fatalf("read migration %s: %v", migration, readErr)
		}
		if _, execErr := pool.Exec(ctx, string(b)); execErr != nil {
			t.Fatalf("exec migration %s: %v", migration, execErr)
		}
	}

	_, _ = pool.Exec(ctx, "DELETE FROM core.quotes")
	_, _ = pool.Exec(ctx, "DELETE FROM core.margin_rules")
	_, _ = pool.Exec(ctx, "DELETE FROM core.base_rates")

	_, err = pool.Exec(ctx, `INSERT INTO core.base_rates (tenant_id, base_currency_id, quote_currency_id, bid, ask, source_name, updated_at)
VALUES ('11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333','44444444-4444-4444-4444-444444444444',35.10,35.30,'manual',now())`)
	if err != nil {
		t.Fatalf("insert base rate: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO core.margin_rules (tenant_id, office_id, base_currency_id, quote_currency_id, side, volume_basis, min_volume, margin_bps, fixed_fee, priority, rounding_precision, rounding_mode)
VALUES ('11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','44444444-4444-4444-4444-444444444444','sell','give',0,150,0,100,2,'half_up')`)
	if err != nil {
		t.Fatalf("insert margin rule: %v", err)
	}

	svc := NewService(NewSQLRepository(pool), staticID("q_integration_1"))
	_, err = svc.CalculateQuote(ctx, CalculateQuoteCommand{
		TenantID:       "11111111-1111-1111-1111-111111111111",
		OfficeID:       "22222222-2222-2222-2222-222222222222",
		GiveCurrencyID: "33333333-3333-3333-3333-333333333333",
		GetCurrencyID:  "44444444-4444-4444-4444-444444444444",
		InputMode:      InputModeGive,
		Amount:         "100",
		Now:            time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CalculateQuote: %v", err)
	}

	var id string
	if err := pool.QueryRow(ctx, "SELECT id FROM core.quotes WHERE id = 'q_integration_1' LIMIT 1").Scan(&id); err != nil {
		t.Fatalf("quote not inserted: %v", err)
	}
}
