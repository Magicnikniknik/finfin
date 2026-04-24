package app

import (
	"context"
	"encoding/json"
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
	ErrQuoteMismatch         = errors.New("reserve request does not match resolved quote")
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
	if a.quotes == nil {
		return orders.ReserveOrderResult{}, fmt.Errorf("quote resolver is nil")
	}
	if a.accounts == nil {
		return orders.ReserveOrderResult{}, fmt.Errorf("account resolver is nil")
	}

	if err := validateReserveCommand(cmd); err != nil {
		return orders.ReserveOrderResult{}, err
	}

	resolvedQuote, err := a.quotes.ResolveQuoteForReserve(ctx, QuoteResolveRequest{TenantID: strings.TrimSpace(cmd.TenantID), OfficeID: strings.TrimSpace(cmd.OfficeID), QuoteID: strings.TrimSpace(cmd.QuoteID)})
	if err != nil {
		return orders.ReserveOrderResult{}, err
	}
	if strings.TrimSpace(resolvedQuote.QuoteID) == "" {
		return orders.ReserveOrderResult{}, ErrQuoteNotFound
	}
	if !resolvedQuote.ExpiresAt.After(a.now()) {
		return orders.ReserveOrderResult{}, ErrQuoteExpired
	}
	if err := ensureReserveMatchesQuote(cmd, resolvedQuote); err != nil {
		return orders.ReserveOrderResult{}, err
	}

	resolvedAccounts, err := a.accounts.ResolveAccountsForReserve(ctx, AccountResolveRequest{TenantID: strings.TrimSpace(cmd.TenantID), OfficeID: strings.TrimSpace(cmd.OfficeID), HoldCurrencyID: strings.TrimSpace(resolvedQuote.HoldCurrencyID)})
	if err != nil {
		return orders.ReserveOrderResult{}, err
	}
	if err := validateResolvedAccounts(resolvedAccounts); err != nil {
		return orders.ReserveOrderResult{}, err
	}

	domainCmd := orders.ReserveOrderCommand{
		TenantID:                  strings.TrimSpace(cmd.TenantID),
		ClientRef:                 strings.TrimSpace(cmd.ClientRef),
		IdempotencyKey:            strings.TrimSpace(cmd.IdempotencyKey),
		OfficeID:                  strings.TrimSpace(cmd.OfficeID),
		QuoteID:                   strings.TrimSpace(cmd.QuoteID),
		Side:                      strings.TrimSpace(resolvedQuote.Side),
		GiveCurrencyID:            strings.TrimSpace(resolvedQuote.GiveCurrencyID),
		GetCurrencyID:             strings.TrimSpace(resolvedQuote.GetCurrencyID),
		AmountGive:                strings.TrimSpace(resolvedQuote.AmountGive),
		AmountGet:                 strings.TrimSpace(resolvedQuote.AmountGet),
		FixedRate:                 strings.TrimSpace(resolvedQuote.FixedRate),
		HoldCurrencyID:            strings.TrimSpace(resolvedQuote.HoldCurrencyID),
		HoldAmount:                strings.TrimSpace(resolvedQuote.HoldAmount),
		BalanceAccountID:          strings.TrimSpace(resolvedAccounts.BalanceAccountID),
		AvailableLedgerAccountID:  strings.TrimSpace(resolvedAccounts.AvailableLedgerAccountID),
		ReservedLedgerAccountID:   strings.TrimSpace(resolvedAccounts.ReservedLedgerAccountID),
		SettlementLedgerAccountID: resolvedAccounts.SettlementLedgerAccountID,
		QuotePayload:              normalizedQuotePayload(resolvedQuote),
		ExpiresAt:                 resolvedQuote.ExpiresAt.UTC(),
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
	if strings.TrimSpace(cmd.TenantID) == "" || strings.TrimSpace(cmd.ClientRef) == "" || strings.TrimSpace(cmd.IdempotencyKey) == "" || strings.TrimSpace(cmd.OfficeID) == "" || strings.TrimSpace(cmd.QuoteID) == "" || strings.TrimSpace(cmd.Side) == "" {
		return fmt.Errorf("%w: required fields are missing", orders.ErrInvalidAmount)
	}
	if strings.TrimSpace(cmd.GiveAmount) == "" || strings.TrimSpace(cmd.GetAmount) == "" || strings.TrimSpace(cmd.GiveCurrencyCode) == "" || strings.TrimSpace(cmd.GetCurrencyCode) == "" {
		return fmt.Errorf("%w: amount/currency fields are required", orders.ErrInvalidAmount)
	}
	return nil
}

func validateResolvedAccounts(acc ResolvedAccounts) error {
	switch {
	case strings.TrimSpace(acc.BalanceAccountID) == "":
		return ErrAccountWiringNotFound
	case strings.TrimSpace(acc.AvailableLedgerAccountID) == "":
		return ErrAccountWiringNotFound
	case strings.TrimSpace(acc.ReservedLedgerAccountID) == "":
		return ErrAccountWiringNotFound
	default:
		return nil
	}
}

func ensureReserveMatchesQuote(cmd ReserveOrderInput, q ResolvedQuote) error {
	if normalize(cmd.OfficeID) != normalize(q.OfficeID) || normalize(cmd.Side) != normalize(q.Side) || normalize(cmd.GiveAmount) != normalize(q.AmountGive) || normalize(cmd.GetAmount) != normalize(q.AmountGet) || normalize(cmd.GiveCurrencyCode) != normalize(q.GiveCurrencyCode) || normalize(cmd.GiveCurrencyNetwork) != normalize(q.GiveCurrencyNetwork) || normalize(cmd.GetCurrencyCode) != normalize(q.GetCurrencyCode) || normalize(cmd.GetCurrencyNetwork) != normalize(q.GetCurrencyNetwork) {
		return ErrQuoteMismatch
	}
	return nil
}

func normalizedQuotePayload(q ResolvedQuote) json.RawMessage {
	if len(q.Payload) > 0 {
		return q.Payload
	}
	payload, _ := json.Marshal(map[string]any{
		"quote_id":   q.QuoteID,
		"office_id":  q.OfficeID,
		"side":       q.Side,
		"expires_at": q.ExpiresAt.UTC().Format(time.RFC3339),
		"give":       map[string]any{"amount": q.AmountGive, "currency_id": q.GiveCurrencyID, "code": q.GiveCurrencyCode, "network": q.GiveCurrencyNetwork},
		"get":        map[string]any{"amount": q.AmountGet, "currency_id": q.GetCurrencyID, "code": q.GetCurrencyCode, "network": q.GetCurrencyNetwork},
		"fixed_rate": q.FixedRate,
		"hold":       map[string]any{"currency_id": q.HoldCurrencyID, "amount": q.HoldAmount},
	})
	return payload
}

func normalize(s string) string { return strings.TrimSpace(strings.ToLower(s)) }
