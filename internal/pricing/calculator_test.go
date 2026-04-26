package pricing

import (
	"testing"
	"time"
)

func TestBuildQuote_GiveSell(t *testing.T) {
	cmd := CalculateQuoteCommand{TenantID: "t", OfficeID: "o", GiveCurrencyID: "b", GetCurrencyID: "q", InputMode: InputModeGive, Amount: "100", Now: time.Now().UTC()}
	rate := BaseRate{Bid: "35.10", Ask: "35.30", SourceName: "manual", UpdatedAt: time.Now().UTC()}
	rule := MarginRule{ID: "r1", Side: SideSell, MarginBps: 100, FixedFee: "10", RoundingPrecision: 2, RoundingMode: RoundingHalfUp}

	q, err := BuildQuote(cmd, rate, rule, "q1", 15)
	if err != nil {
		t.Fatalf("BuildQuote err = %v", err)
	}
	if q.AmountGive != "100" {
		t.Fatalf("AmountGive = %s, want 100", q.AmountGive)
	}
	if q.AmountGet == "0" {
		t.Fatalf("AmountGet should be positive")
	}
}

func TestBuildQuote_GetBuy(t *testing.T) {
	cmd := CalculateQuoteCommand{TenantID: "t", OfficeID: "o", GiveCurrencyID: "q", GetCurrencyID: "b", InputMode: InputModeGet, Amount: "10", Now: time.Now().UTC()}
	rate := BaseRate{Bid: "35.10", Ask: "35.30", SourceName: "manual", UpdatedAt: time.Now().UTC()}
	rule := MarginRule{ID: "r1", Side: SideBuy, MarginBps: 200, FixedFee: "1", RoundingPrecision: 2, RoundingMode: RoundingCeil}

	q, err := BuildQuote(cmd, rate, rule, "q2", 15)
	if err != nil {
		t.Fatalf("BuildQuote err = %v", err)
	}
	if q.AmountGet != "10" {
		t.Fatalf("AmountGet = %s, want 10", q.AmountGet)
	}
	if q.RoundingMode != RoundingCeil {
		t.Fatalf("RoundingMode = %s", q.RoundingMode)
	}
}
