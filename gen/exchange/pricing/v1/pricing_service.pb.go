package pricingv1

import "context"

type QuoteInputMode int32

const (
	QuoteInputMode_QUOTE_INPUT_MODE_UNSPECIFIED QuoteInputMode = 0
	QuoteInputMode_GIVE                         QuoteInputMode = 1
	QuoteInputMode_GET                          QuoteInputMode = 2
)

type QuoteSide int32

const (
	QuoteSide_QUOTE_SIDE_UNSPECIFIED QuoteSide = 0
	QuoteSide_BUY                    QuoteSide = 1
	QuoteSide_SELL                   QuoteSide = 2
)

type QuoteMoney struct {
	Amount     string
	CurrencyId string
}

func (m *QuoteMoney) GetAmount() string {
	if m == nil {
		return ""
	}
	return m.Amount
}

func (m *QuoteMoney) GetCurrencyId() string {
	if m == nil {
		return ""
	}
	return m.CurrencyId
}

type CalculateQuoteRequest struct {
	OfficeId       string
	GiveCurrencyId string
	GetCurrencyId  string
	InputMode      QuoteInputMode
	Amount         string
}

func (r *CalculateQuoteRequest) GetOfficeId() string {
	if r == nil {
		return ""
	}
	return r.OfficeId
}

func (r *CalculateQuoteRequest) GetGiveCurrencyId() string {
	if r == nil {
		return ""
	}
	return r.GiveCurrencyId
}

func (r *CalculateQuoteRequest) GetGetCurrencyId() string {
	if r == nil {
		return ""
	}
	return r.GetCurrencyId
}

func (r *CalculateQuoteRequest) GetInputMode() QuoteInputMode {
	if r == nil {
		return QuoteInputMode_QUOTE_INPUT_MODE_UNSPECIFIED
	}
	return r.InputMode
}

func (r *CalculateQuoteRequest) GetAmount() string {
	if r == nil {
		return ""
	}
	return r.Amount
}

type CalculateQuoteResponse struct {
	QuoteId     string
	Side        QuoteSide
	Give        *QuoteMoney
	Get         *QuoteMoney
	FixedRate   string
	BaseRate    string
	FeeAmount   string
	ExpiresAtTs int64
	SourceName  string
}

func (r *CalculateQuoteResponse) GetQuoteId() string {
	if r == nil {
		return ""
	}
	return r.QuoteId
}

func (r *CalculateQuoteResponse) GetSide() QuoteSide {
	if r == nil {
		return QuoteSide_QUOTE_SIDE_UNSPECIFIED
	}
	return r.Side
}

func (r *CalculateQuoteResponse) GetGive() *QuoteMoney {
	if r == nil {
		return nil
	}
	return r.Give
}

func (r *CalculateQuoteResponse) GetGet() *QuoteMoney {
	if r == nil {
		return nil
	}
	return r.Get
}

func (r *CalculateQuoteResponse) GetFixedRate() string {
	if r == nil {
		return ""
	}
	return r.FixedRate
}

func (r *CalculateQuoteResponse) GetBaseRate() string {
	if r == nil {
		return ""
	}
	return r.BaseRate
}

func (r *CalculateQuoteResponse) GetFeeAmount() string {
	if r == nil {
		return ""
	}
	return r.FeeAmount
}

func (r *CalculateQuoteResponse) GetExpiresAtTs() int64 {
	if r == nil {
		return 0
	}
	return r.ExpiresAtTs
}

func (r *CalculateQuoteResponse) GetSourceName() string {
	if r == nil {
		return ""
	}
	return r.SourceName
}

type PricingServiceServer interface {
	CalculateQuote(context.Context, *CalculateQuoteRequest) (*CalculateQuoteResponse, error)
}
