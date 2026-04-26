package grpcserver

import (
	"context"
	"strings"

	pricingv1 "finfin/gen/exchange/pricing/v1"
	"finfin/internal/pricing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PricingApplication interface {
	CalculateQuote(ctx context.Context, cmd pricing.CalculateQuoteCommand) (pricing.QuoteResult, error)
}

type PricingServer struct {
	pricingv1.UnimplementedPricingServiceServer
	app PricingApplication
}

func NewPricingServer(app PricingApplication) *PricingServer {
	return &PricingServer{app: app}
}

func (s *PricingServer) CalculateQuote(ctx context.Context, req *pricingv1.CalculateQuoteRequest) (*pricingv1.CalculateQuoteResponse, error) {
	if s.app == nil {
		return nil, status.Error(codes.Unimplemented, "pricing application is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	tenantID, err := requireMetadata(ctx, "x-tenant-id")
	if err != nil {
		return nil, err
	}
	clientRef, err := requireMetadata(ctx, "x-client-ref")
	if err != nil {
		return nil, err
	}

	inputMode, err := mapProtoInputMode(req.GetInputMode())
	if err != nil {
		return nil, err
	}

	res, err := s.app.CalculateQuote(ctx, pricing.CalculateQuoteCommand{
		TenantID:       tenantID,
		OfficeID:       strings.TrimSpace(req.GetOfficeId()),
		ClientRef:      clientRef,
		GiveCurrencyID: strings.TrimSpace(req.GetGiveCurrencyId()),
		GetCurrencyID:  strings.TrimSpace(req.GetGetCurrencyId()),
		InputMode:      inputMode,
		Amount:         strings.TrimSpace(req.GetAmount()),
	})
	if err != nil {
		return nil, MapDomainError(err)
	}

	return &pricingv1.CalculateQuoteResponse{
		QuoteId: res.QuoteID,
		Side:    mapQuoteSide(res.Side),
		Give: &pricingv1.QuoteMoney{
			Amount:     res.AmountGive,
			CurrencyId: res.GiveCurrencyID,
		},
		Get: &pricingv1.QuoteMoney{
			Amount:     res.AmountGet,
			CurrencyId: res.GetCurrencyID,
		},
		FixedRate:   res.FixedRate,
		BaseRate:    res.BaseRate,
		FeeAmount:   res.FeeAmount,
		ExpiresAtTs: res.ExpiresAt.Unix(),
		SourceName:  res.SourceName,
	}, nil
}

func mapProtoInputMode(mode pricingv1.QuoteInputMode) (pricing.QuoteInputMode, error) {
	switch mode {
	case pricingv1.QuoteInputMode_GIVE:
		return pricing.InputModeGive, nil
	case pricingv1.QuoteInputMode_GET:
		return pricing.InputModeGet, nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid quote input mode")
	}
}

func mapQuoteSide(side pricing.QuoteSide) pricingv1.QuoteSide {
	switch side {
	case pricing.SideBuy:
		return pricingv1.QuoteSide_BUY
	case pricing.SideSell:
		return pricingv1.QuoteSide_SELL
	default:
		return pricingv1.QuoteSide_QUOTE_SIDE_UNSPECIFIED
	}
}
