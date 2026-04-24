package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type QuoteQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type SQLQuoteResolver struct {
	db  QuoteQueryer
	log *slog.Logger
	now func() time.Time
}

func NewSQLQuoteResolver(db QuoteQueryer, log *slog.Logger) *SQLQuoteResolver {
	if log == nil {
		log = slog.Default()
	}

	return &SQLQuoteResolver{
		db:  db,
		log: log,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (r *SQLQuoteResolver) ResolveQuoteForReserve(ctx context.Context, req QuoteResolveRequest) (ResolvedQuote, error) {
	if strings.TrimSpace(req.TenantID) == "" {
		return ResolvedQuote{}, fmt.Errorf("%w: tenant_id is required", ordersInvalidInput())
	}
	if strings.TrimSpace(req.OfficeID) == "" {
		return ResolvedQuote{}, fmt.Errorf("%w: office_id is required", ordersInvalidInput())
	}
	if strings.TrimSpace(req.QuoteID) == "" {
		return ResolvedQuote{}, fmt.Errorf("%w: quote_id is required", ordersInvalidInput())
	}

	const q = `
SELECT
	id,
	tenant_id::text,
	office_id::text,
	side,
	expires_at,
	give_currency_id::text,
	give_currency_code,
	give_currency_network,
	amount_give::text,
	get_currency_id::text,
	get_currency_code,
	get_currency_network,
	amount_get::text,
	fixed_rate::text,
	hold_currency_id::text,
	hold_amount::text,
	payload
FROM core.quote_snapshots
WHERE id = $1
  AND tenant_id = $2::uuid
  AND office_id = $3::uuid
LIMIT 1
`

	var out ResolvedQuote
	var tenantID string
	var officeID string
	var payload []byte

	err := r.db.QueryRow(ctx, q,
		strings.TrimSpace(req.QuoteID),
		strings.TrimSpace(req.TenantID),
		strings.TrimSpace(req.OfficeID),
	).Scan(
		&out.QuoteID,
		&tenantID,
		&officeID,
		&out.Side,
		&out.ExpiresAt,
		&out.GiveCurrencyID,
		&out.GiveCurrencyCode,
		&out.GiveCurrencyNetwork,
		&out.AmountGive,
		&out.GetCurrencyID,
		&out.GetCurrencyCode,
		&out.GetCurrencyNetwork,
		&out.AmountGet,
		&out.FixedRate,
		&out.HoldCurrencyID,
		&out.HoldAmount,
		&payload,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ResolvedQuote{}, ErrQuoteNotFound
		}
		return ResolvedQuote{}, fmt.Errorf("resolve quote from db: %w", err)
	}

	out.QuoteID = strings.TrimSpace(out.QuoteID)
	out.OfficeID = strings.TrimSpace(officeID)
	out.Side = normalizeSide(out.Side)
	out.GiveCurrencyID = strings.TrimSpace(out.GiveCurrencyID)
	out.GiveCurrencyCode = strings.TrimSpace(out.GiveCurrencyCode)
	out.GiveCurrencyNetwork = strings.TrimSpace(out.GiveCurrencyNetwork)
	out.AmountGive = strings.TrimSpace(out.AmountGive)
	out.GetCurrencyID = strings.TrimSpace(out.GetCurrencyID)
	out.GetCurrencyCode = strings.TrimSpace(out.GetCurrencyCode)
	out.GetCurrencyNetwork = strings.TrimSpace(out.GetCurrencyNetwork)
	out.AmountGet = strings.TrimSpace(out.AmountGet)
	out.FixedRate = strings.TrimSpace(out.FixedRate)
	out.HoldCurrencyID = strings.TrimSpace(out.HoldCurrencyID)
	out.HoldAmount = strings.TrimSpace(out.HoldAmount)
	out.Payload = json.RawMessage(payload)

	now := r.now()
	if !out.ExpiresAt.UTC().After(now) {
		return ResolvedQuote{}, ErrQuoteExpired
	}

	r.log.Info("quote resolved for reserve",
		"tenant_id", tenantID,
		"office_id", officeID,
		"quote_id", out.QuoteID,
		"side", out.Side,
	)

	return out, nil
}

func ordersInvalidInput() error {
	return errors.New("invalid reserve request")
}
