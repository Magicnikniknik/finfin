package pricing

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"
)

type IDGenerator interface {
	NewID() string
}

type defaultIDGenerator struct{}

func (defaultIDGenerator) NewID() string {
	return "q_" + time.Now().UTC().Format("20060102150405.000000000")
}

type Service struct {
	repo              Repository
	ids               IDGenerator
	now               func() time.Time
	quoteTTL          time.Duration
	maxRateAge        time.Duration
	maxAllowedSkewBps int
}

func NewService(repo Repository, ids IDGenerator) *Service {
	if ids == nil {
		ids = defaultIDGenerator{}
	}
	return &Service{
		repo:              repo,
		ids:               ids,
		now:               func() time.Time { return time.Now().UTC() },
		quoteTTL:          15 * time.Minute,
		maxRateAge:        30 * time.Second,
		maxAllowedSkewBps: 500,
	}
}

func (s *Service) CalculateQuote(ctx context.Context, cmd CalculateQuoteCommand) (QuoteResult, error) {
	if err := validateCommand(cmd); err != nil {
		return QuoteResult{}, err
	}
	now := cmd.Now.UTC()
	if now.IsZero() {
		now = s.now().UTC()
		cmd.Now = now
	}

	rate, side, err := s.loadRateAndSide(ctx, cmd)
	if err != nil {
		return QuoteResult{}, err
	}
	if now.Sub(rate.UpdatedAt.UTC()) > s.maxRateAge {
		return QuoteResult{}, ErrRateStale
	}

	rules, err := s.repo.FindCandidateMarginRules(ctx, cmd.TenantID, cmd.OfficeID, rate.BaseCurrencyID, rate.QuoteCurrencyID, side)
	if err != nil {
		return QuoteResult{}, err
	}
	rule, err := SelectBestMarginRule(rules, cmd, side, cmd.Amount)
	if err != nil {
		return QuoteResult{}, err
	}

	rec, err := BuildQuote(cmd, rate, rule, s.ids.NewID(), int(s.quoteTTL.Minutes()))
	if err != nil {
		return QuoteResult{}, err
	}

	ok, err := isSkewWithinGuardrail(rec.FixedRate, rec.BaseRateSnapshot, s.maxAllowedSkewBps)
	if err != nil {
		return QuoteResult{}, err
	}
	if !ok {
		return QuoteResult{}, ErrRateGuardrailTriggered
	}

	if err := s.repo.InsertQuote(ctx, rec); err != nil {
		return QuoteResult{}, err
	}

	return QuoteResult{
		QuoteID:        rec.ID,
		Side:           rec.Side,
		GiveCurrencyID: rec.GiveCurrencyID,
		GetCurrencyID:  rec.GetCurrencyID,
		AmountGive:     rec.AmountGive,
		AmountGet:      rec.AmountGet,
		FixedRate:      rec.FixedRate,
		BaseRate:       rec.BaseRateSnapshot,
		FeeAmount:      rec.FixedFeeApplied,
		ExpiresAt:      rec.ExpiresAt,
		SourceName:     rec.SourceNameSnapshot,
	}, nil
}

func validateCommand(cmd CalculateQuoteCommand) error {
	if strings.TrimSpace(cmd.TenantID) == "" || strings.TrimSpace(cmd.OfficeID) == "" ||
		strings.TrimSpace(cmd.GiveCurrencyID) == "" || strings.TrimSpace(cmd.GetCurrencyID) == "" ||
		strings.TrimSpace(cmd.Amount) == "" {
		return ErrInvalidQuoteInput
	}
	if cmd.InputMode != InputModeGive && cmd.InputMode != InputModeGet {
		return ErrInvalidQuoteInput
	}
	v, err := ParseDecimalString(cmd.Amount)
	if err != nil || v.Sign() <= 0 {
		return ErrInvalidQuoteInput
	}
	return nil
}

func (s *Service) loadRateAndSide(ctx context.Context, cmd CalculateQuoteCommand) (BaseRate, QuoteSide, error) {
	rate, err := s.repo.GetBaseRate(ctx, cmd.TenantID, cmd.GiveCurrencyID, cmd.GetCurrencyID)
	if err == nil {
		return rate, SideSell, nil
	}
	if !errors.Is(err, ErrBaseRateNotFound) {
		return BaseRate{}, "", err
	}

	rate2, err2 := s.repo.GetBaseRate(ctx, cmd.TenantID, cmd.GetCurrencyID, cmd.GiveCurrencyID)
	if err2 == nil {
		return rate2, SideBuy, nil
	}
	if !errors.Is(err2, ErrBaseRateNotFound) {
		return BaseRate{}, "", err2
	}
	return BaseRate{}, "", ErrBaseRateNotFound
}

func isSkewWithinGuardrail(fixedRate, baseRate string, maxBps int) (bool, error) {
	f, err := ParseDecimalString(fixedRate)
	if err != nil {
		return false, ErrInvalidQuoteInput
	}
	b, err := ParseDecimalString(baseRate)
	if err != nil {
		return false, ErrInvalidQuoteInput
	}
	diff := new(big.Rat).Sub(f, b)
	if diff.Sign() < 0 {
		diff.Neg(diff)
	}
	pct := new(big.Rat).Quo(diff, b)
	limit := new(big.Rat).SetFrac64(int64(maxBps), 10000)
	return pct.Cmp(limit) <= 0, nil
}
