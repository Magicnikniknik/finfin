package orders

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type TxBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Service struct {
	db     TxBeginner
	log    *slog.Logger
	poster JournalPoster
	now    func() time.Time
}

func NewService(db TxBeginner, log *slog.Logger, poster JournalPoster) *Service {
	if log == nil {
		log = slog.Default()
	}
	if poster == nil {
		poster = RealJournalPoster{}
	}
	return &Service{
		db:     db,
		log:    log,
		poster: poster,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

type ReserveOrderCommand struct {
	TenantID       string
	ClientRef      string
	IdempotencyKey string

	OfficeID string
	QuoteID  string
}

type ReserveOrderResult struct {
	OrderID   string    `json:"order_id"`
	OrderRef  string    `json:"order_ref"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	Version   int64     `json:"version"`
}

type CompleteOrderCommand struct {
	TenantID        string
	OrderID         string
	ExpectedVersion int64
	IdempotencyKey  string
	CashierID       string
}

type CompleteOrderResult struct {
	OrderID     string    `json:"order_id"`
	OrderRef    string    `json:"order_ref"`
	Status      string    `json:"status"`
	Version     int64     `json:"version"`
	CompletedAt time.Time `json:"completed_at"`
}

type CancelOrderCommand struct {
	TenantID        string
	OrderID         string
	ExpectedVersion int64
	IdempotencyKey  string
	Reason          string
}

type CancelOrderResult struct {
	OrderID  string `json:"order_id"`
	OrderRef string `json:"order_ref"`
	Status   string `json:"status"`
	Version  int64  `json:"version"`
}

type lockedOrder struct {
	OrderID                   string
	OrderRef                  string
	TenantID                  string
	OfficeID                  string
	Status                    string
	Version                   int64
	HoldID                    string
	HoldStatus                string
	BalanceAccountID          string
	AvailableLedgerAccountID  string
	ReservedLedgerAccountID   string
	SettlementLedgerAccountID *string
	CurrencyID                string
	Amount                    string
	ExpiresAt                 time.Time
}

type lockedQuote struct {
	QuoteID        string
	TenantID       string
	OfficeID       string
	Status         string
	Side           string
	GiveCurrencyID string
	GetCurrencyID  string
	AmountGive     string
	AmountGet      string
	FixedRate      string
	ExpiresAt      time.Time
}

type resolvedWiring struct {
	BalanceAccountID          string
	AvailableLedgerAccountID  string
	ReservedLedgerAccountID   string
	SettlementLedgerAccountID *string
}
