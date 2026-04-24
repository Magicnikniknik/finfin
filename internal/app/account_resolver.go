package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
)

type AccountQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type SQLAccountResolver struct {
	db  AccountQueryer
	log *slog.Logger
}

func NewSQLAccountResolver(db AccountQueryer, log *slog.Logger) *SQLAccountResolver {
	if log == nil {
		log = slog.Default()
	}
	return &SQLAccountResolver{db: db, log: log}
}

func (r *SQLAccountResolver) ResolveAccountsForReserve(ctx context.Context, req AccountResolveRequest) (ResolvedAccounts, error) {
	tenantID := strings.TrimSpace(req.TenantID)
	officeID := strings.TrimSpace(req.OfficeID)
	holdCurrencyID := strings.TrimSpace(req.HoldCurrencyID)
	if tenantID == "" || officeID == "" || holdCurrencyID == "" {
		return ResolvedAccounts{}, ErrAccountWiringNotFound
	}

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

	var out ResolvedAccounts
	var settlementID *string
	if err := r.db.QueryRow(ctx, q, tenantID, officeID, holdCurrencyID).Scan(
		&out.BalanceAccountID,
		&out.AvailableLedgerAccountID,
		&out.ReservedLedgerAccountID,
		&settlementID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ResolvedAccounts{}, ErrAccountWiringNotFound
		}
		return ResolvedAccounts{}, fmt.Errorf("resolve account wiring from db: %w", err)
	}

	out.BalanceAccountID = strings.TrimSpace(out.BalanceAccountID)
	out.AvailableLedgerAccountID = strings.TrimSpace(out.AvailableLedgerAccountID)
	out.ReservedLedgerAccountID = strings.TrimSpace(out.ReservedLedgerAccountID)
	if settlementID != nil {
		val := strings.TrimSpace(*settlementID)
		if val != "" {
			out.SettlementLedgerAccountID = &val
		}
	}

	if strings.TrimSpace(out.BalanceAccountID) == "" ||
		strings.TrimSpace(out.AvailableLedgerAccountID) == "" ||
		strings.TrimSpace(out.ReservedLedgerAccountID) == "" {
		return ResolvedAccounts{}, ErrAccountWiringNotFound
	}

	r.log.Info("account wiring resolved for reserve",
		"tenant_id", tenantID,
		"office_id", officeID,
		"hold_currency_id", holdCurrencyID,
	)
	return out, nil
}
