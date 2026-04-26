package pricing

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repoMock struct {
	rate         BaseRate
	rateErr      error
	rules        []MarginRule
	rulesErr     error
	insertErr    error
	insertCalled bool
}

func (m *repoMock) GetBaseRate(ctx context.Context, tenantID, baseCurrencyID, quoteCurrencyID string) (BaseRate, error) {
	if m.rateErr != nil {
		return BaseRate{}, m.rateErr
	}
	return m.rate, nil
}

func (m *repoMock) FindCandidateMarginRules(ctx context.Context, tenantID, officeID, baseCurrencyID, quoteCurrencyID string, side QuoteSide) ([]MarginRule, error) {
	if m.rulesErr != nil {
		return nil, m.rulesErr
	}
	return m.rules, nil
}

func (m *repoMock) InsertQuote(ctx context.Context, quote QuoteRecord) error {
	m.insertCalled = true
	return m.insertErr
}

type staticID string

func (s staticID) NewID() string { return string(s) }

func TestService_CalculateQuote_StaleRate(t *testing.T) {
	repo := &repoMock{rate: BaseRate{Bid: "35", Ask: "36", UpdatedAt: time.Now().UTC().Add(-2 * time.Minute)}}
	svc := NewService(repo, staticID("q_1"))
	_, err := svc.CalculateQuote(context.Background(), CalculateQuoteCommand{TenantID: "t", OfficeID: "o", GiveCurrencyID: "b", GetCurrencyID: "q", InputMode: InputModeGive, Amount: "10"})
	if !errors.Is(err, ErrRateStale) {
		t.Fatalf("err = %v, want %v", err, ErrRateStale)
	}
}

func TestService_CalculateQuote_OK(t *testing.T) {
	now := time.Now().UTC()
	office := "o"
	repo := &repoMock{
		rate:  BaseRate{BaseCurrencyID: "b", QuoteCurrencyID: "q", Bid: "35.1", Ask: "35.3", SourceName: "manual", UpdatedAt: now},
		rules: []MarginRule{{ID: "r1", OfficeID: &office, Side: SideSell, MinVolume: "0", MarginBps: 100, FixedFee: "0", Priority: 100, RoundingPrecision: 2, RoundingMode: RoundingHalfUp}},
	}
	svc := NewService(repo, staticID("q_1"))
	res, err := svc.CalculateQuote(context.Background(), CalculateQuoteCommand{TenantID: "t", OfficeID: "o", GiveCurrencyID: "b", GetCurrencyID: "q", InputMode: InputModeGive, Amount: "10", Now: now})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !repo.insertCalled {
		t.Fatalf("quote was not inserted")
	}
	if res.QuoteID != "q_1" {
		t.Fatalf("QuoteID = %s", res.QuoteID)
	}
}
