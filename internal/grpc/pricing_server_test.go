package grpcserver

import (
	"context"
	"testing"
	"time"

	pricingv1 "finfin/gen/exchange/pricing/v1"
	"finfin/internal/pricing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type testPricingApp struct {
	result pricing.QuoteResult
	err    error
}

func (a testPricingApp) CalculateQuote(_ context.Context, _ pricing.CalculateQuoteCommand) (pricing.QuoteResult, error) {
	return a.result, a.err
}

func TestCalculateQuote_HappyPath(t *testing.T) {
	srv := NewPricingServer(testPricingApp{result: pricing.QuoteResult{
		QuoteID:        "q-1",
		Side:           pricing.SideSell,
		GiveCurrencyID: "cur-give",
		GetCurrencyID:  "cur-get",
		AmountGive:     "100.00",
		AmountGet:      "3550.00",
		FixedRate:      "35.5000",
		BaseRate:       "35.4500",
		FeeAmount:      "0.10",
		ExpiresAt:      time.Unix(1735000000, 0).UTC(),
		SourceName:     "provider-1",
	}})
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-tenant-id", "tenant-1", "x-client-ref", "client-1"),
	)

	res, err := srv.CalculateQuote(ctx, &pricingv1.CalculateQuoteRequest{
		OfficeId:       "office-1",
		GiveCurrencyId: "cur-give",
		GetCurrencyId:  "cur-get",
		InputMode:      pricingv1.QuoteInputMode_GIVE,
		Amount:         "100.00",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.GetQuoteId() != "q-1" {
		t.Fatalf("expected quote id q-1, got %q", res.GetQuoteId())
	}
	if res.GetSide() != pricingv1.QuoteSide_SELL {
		t.Fatalf("expected side SELL, got %v", res.GetSide())
	}
	if res.GetGive().GetCurrencyId() != "cur-give" {
		t.Fatalf("expected give currency cur-give, got %q", res.GetGive().GetCurrencyId())
	}
}

func TestCalculateQuote_MissingMetadata_ReturnsUnauthenticated(t *testing.T) {
	srv := NewPricingServer(testPricingApp{})

	_, err := srv.CalculateQuote(context.Background(), &pricingv1.CalculateQuoteRequest{
		OfficeId:       "office-1",
		GiveCurrencyId: "cur-give",
		GetCurrencyId:  "cur-get",
		InputMode:      pricingv1.QuoteInputMode_GIVE,
		Amount:         "100.00",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Fatalf("expected code %s, got %s", codes.Unauthenticated, got)
	}
}

func TestCalculateQuote_DomainErrors_AreMapped(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{name: "rate stale", err: pricing.ErrRateStale, want: codes.FailedPrecondition},
		{name: "no margin rule", err: pricing.ErrNoMarginRuleFound, want: codes.NotFound},
		{name: "base rate not found", err: pricing.ErrBaseRateNotFound, want: codes.NotFound},
	}

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-tenant-id", "tenant-1", "x-client-ref", "client-1"),
	)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewPricingServer(testPricingApp{err: tc.err})
			_, err := srv.CalculateQuote(ctx, &pricingv1.CalculateQuoteRequest{
				OfficeId:       "office-1",
				GiveCurrencyId: "cur-give",
				GetCurrencyId:  "cur-get",
				InputMode:      pricingv1.QuoteInputMode_GIVE,
				Amount:         "100.00",
			})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if got := status.Code(err); got != tc.want {
				t.Fatalf("expected code %s, got %s", tc.want, got)
			}
		})
	}
}
