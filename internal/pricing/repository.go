package pricing

import "context"

type Repository interface {
	GetBaseRate(ctx context.Context, tenantID, baseCurrencyID, quoteCurrencyID string) (BaseRate, error)
	FindCandidateMarginRules(ctx context.Context, tenantID, officeID, baseCurrencyID, quoteCurrencyID string, side QuoteSide) ([]MarginRule, error)
	InsertQuote(ctx context.Context, quote QuoteRecord) error
}
