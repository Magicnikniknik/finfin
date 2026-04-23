package orders

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type HoldCreatePosting struct {
	TenantID                 string
	OrderID                  string
	HoldID                   string
	AvailableLedgerAccountID string
	ReservedLedgerAccountID  string
	CurrencyID               string
	Amount                   string
	Reason                   string
}

type HoldReleasePosting struct {
	TenantID                 string
	OrderID                  string
	HoldID                   string
	AvailableLedgerAccountID string
	ReservedLedgerAccountID  string
	CurrencyID               string
	Amount                   string
	Reason                   string
}

type TradeCompletePosting struct {
	TenantID                  string
	OrderID                   string
	HoldID                    string
	ReservedLedgerAccountID   string
	SettlementLedgerAccountID string
	CurrencyID                string
	Amount                    string
	CashierID                 string
	Reason                    string
}

type JournalPoster interface {
	PostHoldCreate(ctx context.Context, tx pgx.Tx, posting HoldCreatePosting) error
	PostHoldRelease(ctx context.Context, tx pgx.Tx, posting HoldReleasePosting) error
	PostTradeComplete(ctx context.Context, tx pgx.Tx, posting TradeCompletePosting) error
}

type RealJournalPoster struct{}

func (RealJournalPoster) PostHoldCreate(ctx context.Context, tx pgx.Tx, p HoldCreatePosting) error {
	var journalID string
	err := tx.QueryRow(ctx, `
INSERT INTO core.ledger_journals (
	tenant_id,
	kind,
	order_id,
	hold_id,
	note,
	created_at
)
VALUES (
	$1::uuid,
	'hold_create',
	$2::uuid,
	$3::uuid,
	$4,
	now()
)
RETURNING id::text
`,
		p.TenantID,
		p.OrderID,
		p.HoldID,
		p.Reason,
	).Scan(&journalID)
	if err != nil {
		return fmt.Errorf("insert hold_create journal: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO core.ledger_entries (
	journal_id,
	tenant_id,
	account_id,
	currency_id,
	direction,
	amount,
	created_at
)
VALUES
	($1::uuid, $2::uuid, $3::uuid, $4::uuid, 'debit',  $5::numeric, now()),
	($1::uuid, $2::uuid, $6::uuid, $4::uuid, 'credit', $5::numeric, now())
`,
		journalID,
		p.TenantID,
		p.AvailableLedgerAccountID,
		p.CurrencyID,
		p.Amount,
		p.ReservedLedgerAccountID,
	)
	if err != nil {
		return fmt.Errorf("insert hold_create entries: %w", err)
	}

	return nil
}

func (RealJournalPoster) PostHoldRelease(ctx context.Context, tx pgx.Tx, p HoldReleasePosting) error {
	var journalID string
	err := tx.QueryRow(ctx, `
INSERT INTO core.ledger_journals (
	tenant_id,
	kind,
	order_id,
	hold_id,
	note,
	created_at
)
VALUES (
	$1::uuid,
	'hold_release',
	$2::uuid,
	$3::uuid,
	$4,
	now()
)
RETURNING id::text
`,
		p.TenantID,
		p.OrderID,
		p.HoldID,
		p.Reason,
	).Scan(&journalID)
	if err != nil {
		return fmt.Errorf("insert hold_release journal: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO core.ledger_entries (
	journal_id,
	tenant_id,
	account_id,
	currency_id,
	direction,
	amount,
	created_at
)
VALUES
	($1::uuid, $2::uuid, $3::uuid, $4::uuid, 'debit',  $5::numeric, now()),
	($1::uuid, $2::uuid, $6::uuid, $4::uuid, 'credit', $5::numeric, now())
`,
		journalID,
		p.TenantID,
		p.ReservedLedgerAccountID,
		p.CurrencyID,
		p.Amount,
		p.AvailableLedgerAccountID,
	)
	if err != nil {
		return fmt.Errorf("insert hold_release entries: %w", err)
	}

	return nil
}

func (RealJournalPoster) PostTradeComplete(ctx context.Context, tx pgx.Tx, p TradeCompletePosting) error {
	var journalID string
	err := tx.QueryRow(ctx, `
INSERT INTO core.ledger_journals (
	tenant_id,
	kind,
	order_id,
	hold_id,
	note,
	created_by,
	created_at
)
VALUES (
	$1::uuid,
	'trade_complete',
	$2::uuid,
	$3::uuid,
	$4,
	$5,
	now()
)
RETURNING id::text
`,
		p.TenantID,
		p.OrderID,
		p.HoldID,
		p.Reason,
		p.CashierID,
	).Scan(&journalID)
	if err != nil {
		return fmt.Errorf("insert trade_complete journal: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO core.ledger_entries (
	journal_id,
	tenant_id,
	account_id,
	currency_id,
	direction,
	amount,
	created_at
)
VALUES
	($1::uuid, $2::uuid, $3::uuid, $4::uuid, 'debit',  $5::numeric, now()),
	($1::uuid, $2::uuid, $6::uuid, $4::uuid, 'credit', $5::numeric, now())
`,
		journalID,
		p.TenantID,
		p.ReservedLedgerAccountID,
		p.CurrencyID,
		p.Amount,
		p.SettlementLedgerAccountID,
	)
	if err != nil {
		return fmt.Errorf("insert trade_complete entries: %w", err)
	}

	return nil
}
