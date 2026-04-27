package pricing

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type sqlQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type SQLRepository struct {
	db sqlQueryer
}

func NewSQLRepository(db sqlQueryer) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) GetBaseRate(ctx context.Context, tenantID, baseCurrencyID, quoteCurrencyID string) (BaseRate, error) {
	const q = `
SELECT tenant_id::text, base_currency_id::text, quote_currency_id::text,
       bid::text, ask::text, source_name, updated_at
FROM core.base_rates
WHERE tenant_id = $1::uuid AND base_currency_id = $2::uuid AND quote_currency_id = $3::uuid
LIMIT 1`
	var out BaseRate
	err := r.db.QueryRow(ctx, q, tenantID, baseCurrencyID, quoteCurrencyID).Scan(
		&out.TenantID, &out.BaseCurrencyID, &out.QuoteCurrencyID,
		&out.Bid, &out.Ask, &out.SourceName, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BaseRate{}, ErrBaseRateNotFound
		}
		return BaseRate{}, fmt.Errorf("get base rate: %w", err)
	}
	return out, nil
}

func (r *SQLRepository) FindCandidateMarginRules(ctx context.Context, tenantID, officeID, baseCurrencyID, quoteCurrencyID string, side QuoteSide) ([]MarginRule, error) {
	const q = `
SELECT id::text, tenant_id::text, office_id::text, base_currency_id::text, quote_currency_id::text,
       side, volume_basis, min_volume::text, max_volume::text, margin_bps, fixed_fee::text,
       priority, rounding_precision, rounding_mode, created_at
FROM core.margin_rules
WHERE tenant_id = $1::uuid
  AND base_currency_id = $2::uuid
  AND quote_currency_id = $3::uuid
  AND side = $4
  AND (office_id = $5::uuid OR office_id IS NULL)`
	rows, err := r.db.Query(ctx, q, tenantID, baseCurrencyID, quoteCurrencyID, side, officeID)
	if err != nil {
		return nil, fmt.Errorf("find candidate rules: %w", err)
	}
	defer rows.Close()

	out := make([]MarginRule, 0)
	for rows.Next() {
		var m MarginRule
		var officeID *string
		var maxVolume *string
		if err := rows.Scan(
			&m.ID, &m.TenantID, &officeID, &m.BaseCurrencyID, &m.QuoteCurrencyID,
			&m.Side, &m.VolumeBasis, &m.MinVolume, &maxVolume, &m.MarginBps, &m.FixedFee,
			&m.Priority, &m.RoundingPrecision, &m.RoundingMode, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan candidate rule: %w", err)
		}
		m.OfficeID = officeID
		m.MaxVolume = maxVolume
		out = append(out, m)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate candidate rules: %w", rows.Err())
	}
	return out, nil
}

func (r *SQLRepository) InsertQuote(ctx context.Context, quote QuoteRecord) error {
	const q = `
INSERT INTO core.quotes (
  id, tenant_id, office_id, client_ref, side, input_mode, requested_amount,
  give_currency_id, get_currency_id, amount_give, amount_get, fixed_rate,
  applied_rule_id, base_rate_snapshot, margin_bps_applied, fixed_fee_applied,
  source_name_snapshot, rate_updated_at_snapshot, rounding_precision, rounding_mode,
  expires_at, created_at
) VALUES (
  $1, $2::uuid, $3::uuid, $4, $5, $6, $7::numeric,
  $8::uuid, $9::uuid, $10::numeric, $11::numeric, $12::numeric,
  $13::uuid, $14::numeric, $15, $16::numeric,
  $17, $18, $19, $20,
  $21, $22
)`
	_, err := r.db.Exec(ctx, q,
		quote.ID, quote.TenantID, quote.OfficeID, quote.ClientRef, quote.Side, quote.InputMode, quote.RequestedAmount,
		quote.GiveCurrencyID, quote.GetCurrencyID, quote.AmountGive, quote.AmountGet, quote.FixedRate,
		quote.AppliedRuleID, quote.BaseRateSnapshot, quote.MarginBpsApplied, quote.FixedFeeApplied,
		quote.SourceNameSnapshot, quote.RateUpdatedAtSnapshot, quote.RoundingPrecision, quote.RoundingMode,
		quote.ExpiresAt, quote.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert quote: %w", err)
	}
	return nil
}
