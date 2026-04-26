package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"finfin/internal/orders"
)

var (
	ErrQuoteNotFound         = errors.New("quote not found")
	ErrQuoteExpired          = errors.New("quote expired")
	ErrAccountWiringNotFound = errors.New("account wiring not found")
)

type ReserveOrderInput struct {
	TenantID  string
	ClientRef string

	IdempotencyKey string
	OfficeID       string
	QuoteID        string
	Side           string

	GiveAmount          string
	GiveCurrencyCode    string
	GiveCurrencyNetwork string
	GetAmount           string
	GetCurrencyCode     string
	GetCurrencyNetwork  string
}

type QuoteResolveRequest struct {
	TenantID string
	OfficeID string
	QuoteID  string
}

type ResolvedQuote struct {
	QuoteID   string
	OfficeID  string
	Side      string
	ExpiresAt time.Time

	GiveCurrencyID      string
	GiveCurrencyCode    string
	GiveCurrencyNetwork string
	AmountGive          string

	GetCurrencyID      string
	GetCurrencyCode    string
	GetCurrencyNetwork string
	AmountGet          string

	FixedRate string

	HoldCurrencyID string
	HoldAmount     string
	Payload        json.RawMessage
}

type QuoteResolver interface {
	ResolveQuoteForReserve(ctx context.Context, req QuoteResolveRequest) (ResolvedQuote, error)
}

type AccountResolveRequest struct {
	TenantID       string
	OfficeID       string
	HoldCurrencyID string
}

type ResolvedAccounts struct {
	BalanceAccountID          string
	AvailableLedgerAccountID  string
	ReservedLedgerAccountID   string
	SettlementLedgerAccountID *string
}

type AccountResolver interface {
	ResolveAccountsForReserve(ctx context.Context, req AccountResolveRequest) (ResolvedAccounts, error)
}

type OrderApp struct {
	orders   *orders.Service
	quotes   QuoteResolver
	accounts AccountResolver
	log      *slog.Logger
	now      func() time.Time
}

func NewOrderApp(orderSvc *orders.Service, quoteResolver QuoteResolver, accountResolver AccountResolver, log *slog.Logger) *OrderApp {
	if log == nil {
		log = slog.Default()
	}
	return &OrderApp{orders: orderSvc, quotes: quoteResolver, accounts: accountResolver, log: log, now: func() time.Time { return time.Now().UTC() }}
}

func (a *OrderApp) ReserveOrder(ctx context.Context, cmd ReserveOrderInput) (orders.ReserveOrderResult, error) {
	if a.orders == nil {
		return orders.ReserveOrderResult{}, fmt.Errorf("orders service is nil")
	}

	if err := validateReserveCommand(cmd); err != nil {
		return orders.ReserveOrderResult{}, err
	}

	domainCmd := orders.ReserveOrderCommand{
		TenantID:       strings.TrimSpace(cmd.TenantID),
		ClientRef:      strings.TrimSpace(cmd.ClientRef),
		IdempotencyKey: strings.TrimSpace(cmd.IdempotencyKey),
		OfficeID:       strings.TrimSpace(cmd.OfficeID),
		QuoteID:        strings.TrimSpace(cmd.QuoteID),
	}

	return a.orders.ReserveOrder(ctx, domainCmd)
}

func (a *OrderApp) CompleteOrder(ctx context.Context, cmd orders.CompleteOrderCommand) (orders.CompleteOrderResult, error) {
	if a.orders == nil {
		return orders.CompleteOrderResult{}, fmt.Errorf("orders service is nil")
	}
	return a.orders.CompleteOrder(ctx, cmd)
}

func (a *OrderApp) CancelOrder(ctx context.Context, cmd orders.CancelOrderCommand) (orders.CancelOrderResult, error) {
	if a.orders == nil {
		return orders.CancelOrderResult{}, fmt.Errorf("orders service is nil")
	}
	return a.orders.CancelOrder(ctx, cmd)
}

func validateReserveCommand(cmd ReserveOrderInput) error {
	if strings.TrimSpace(cmd.TenantID) == "" || strings.TrimSpace(cmd.ClientRef) == "" || strings.TrimSpace(cmd.IdempotencyKey) == "" || strings.TrimSpace(cmd.OfficeID) == "" || strings.TrimSpace(cmd.QuoteID) == "" {
		return fmt.Errorf("%w: required fields are missing", orders.ErrInvalidAmount)
	}
	return nil
}
