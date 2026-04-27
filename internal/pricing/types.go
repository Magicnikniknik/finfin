package pricing

import "time"

type QuoteInputMode string
type QuoteSide string
type VolumeBasis string
type RoundingMode string

const (
	InputModeGive QuoteInputMode = "give"
	InputModeGet  QuoteInputMode = "get"

	SideBuy  QuoteSide = "buy"
	SideSell QuoteSide = "sell"

	VolumeBasisGive          VolumeBasis = "give"
	VolumeBasisGet           VolumeBasis = "get"
	VolumeBasisQuoteNotional VolumeBasis = "quote_notional"

	RoundingHalfUp RoundingMode = "half_up"
	RoundingFloor  RoundingMode = "floor"
	RoundingCeil   RoundingMode = "ceil"
)

type CalculateQuoteCommand struct {
	TenantID       string
	OfficeID       string
	ClientRef      string
	GiveCurrencyID string
	GetCurrencyID  string
	InputMode      QuoteInputMode
	Amount         string
	Now            time.Time
}

type QuoteResult struct {
	QuoteID        string
	Side           QuoteSide
	GiveCurrencyID string
	GetCurrencyID  string
	AmountGive     string
	AmountGet      string
	FixedRate      string
	BaseRate       string
	FeeAmount      string
	ExpiresAt      time.Time
	SourceName     string
}

type BaseRate struct {
	TenantID        string
	BaseCurrencyID  string
	QuoteCurrencyID string
	Bid             string
	Ask             string
	SourceName      string
	UpdatedAt       time.Time
}

type MarginRule struct {
	ID                string
	TenantID          string
	OfficeID          *string
	BaseCurrencyID    string
	QuoteCurrencyID   string
	Side              QuoteSide
	VolumeBasis       VolumeBasis
	MinVolume         string
	MaxVolume         *string
	MarginBps         int
	FixedFee          string
	Priority          int
	RoundingPrecision int
	RoundingMode      RoundingMode
	CreatedAt         time.Time
}

type QuoteRecord struct {
	ID                    string
	TenantID              string
	OfficeID              string
	ClientRef             string
	Side                  QuoteSide
	InputMode             QuoteInputMode
	RequestedAmount       string
	GiveCurrencyID        string
	GetCurrencyID         string
	AmountGive            string
	AmountGet             string
	FixedRate             string
	AppliedRuleID         *string
	BaseRateSnapshot      string
	MarginBpsApplied      int
	FixedFeeApplied       string
	SourceNameSnapshot    string
	RateUpdatedAtSnapshot *time.Time
	RoundingPrecision     int
	RoundingMode          RoundingMode
	ExpiresAt             time.Time
	CreatedAt             time.Time
}
