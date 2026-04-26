package pricing

import (
	"fmt"
	"math/big"
	"time"
)

func BuildQuote(cmd CalculateQuoteCommand, rate BaseRate, rule MarginRule, quoteID string, ttlMinutes int) (QuoteRecord, error) {
	if quoteID == "" || ttlMinutes <= 0 {
		return QuoteRecord{}, ErrInvalidQuoteInput
	}
	if cmd.InputMode != InputModeGive && cmd.InputMode != InputModeGet {
		return QuoteRecord{}, ErrInvalidQuoteInput
	}
	now := cmd.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	requested, err := ParseDecimalString(cmd.Amount)
	if err != nil || requested.Sign() <= 0 {
		return QuoteRecord{}, ErrInvalidQuoteInput
	}

	baseRateRaw := rate.Bid
	marginBps := -rule.MarginBps
	if rule.Side == SideBuy {
		baseRateRaw = rate.Ask
		marginBps = rule.MarginBps
	}
	baseRate, err := ParseDecimalString(baseRateRaw)
	if err != nil || baseRate.Sign() <= 0 {
		return QuoteRecord{}, fmt.Errorf("%w: invalid base rate", ErrInvalidQuoteInput)
	}
	clientRate := ApplyBps(baseRate, marginBps)
	if clientRate.Sign() <= 0 {
		return QuoteRecord{}, ErrRateGuardrailTriggered
	}
	fee, err := ParseDecimalString(rule.FixedFee)
	if err != nil || fee.Sign() < 0 {
		return QuoteRecord{}, ErrInvalidQuoteInput
	}

	var give, get *big.Rat
	switch cmd.InputMode {
	case InputModeGive:
		give = requested
		if rule.Side == SideSell {
			gross := new(big.Rat).Mul(give, clientRate)
			get = ApplyFixedFee(gross, fee, true)
		} else {
			netGive := ApplyFixedFee(give, fee, true)
			get = new(big.Rat).Quo(netGive, clientRate)
		}
	case InputModeGet:
		get = requested
		if rule.Side == SideSell {
			grossNeeded := ApplyFixedFee(get, fee, false)
			give = new(big.Rat).Quo(grossNeeded, clientRate)
		} else {
			give = new(big.Rat).Mul(get, clientRate)
			give = ApplyFixedFee(give, fee, false)
		}
	}
	if give.Sign() <= 0 || get.Sign() <= 0 {
		return QuoteRecord{}, ErrInvalidQuoteInput
	}

	giveRounded, err := RoundAmount(ratToString(give, 18), rule.RoundingPrecision, rule.RoundingMode)
	if err != nil {
		return QuoteRecord{}, err
	}
	getRounded, err := RoundAmount(ratToString(get, 18), rule.RoundingPrecision, rule.RoundingMode)
	if err != nil {
		return QuoteRecord{}, err
	}
	fixedRateRounded, err := RoundAmount(ratToString(clientRate, 18), 8, RoundingHalfUp)
	if err != nil {
		return QuoteRecord{}, err
	}

	ruleID := rule.ID
	updatedAt := rate.UpdatedAt.UTC()
	return QuoteRecord{
		ID:                    quoteID,
		TenantID:              cmd.TenantID,
		OfficeID:              cmd.OfficeID,
		ClientRef:             cmd.ClientRef,
		Side:                  rule.Side,
		InputMode:             cmd.InputMode,
		RequestedAmount:       cmd.Amount,
		GiveCurrencyID:        cmd.GiveCurrencyID,
		GetCurrencyID:         cmd.GetCurrencyID,
		AmountGive:            giveRounded,
		AmountGet:             getRounded,
		FixedRate:             fixedRateRounded,
		AppliedRuleID:         &ruleID,
		BaseRateSnapshot:      baseRateRaw,
		MarginBpsApplied:      rule.MarginBps,
		FixedFeeApplied:       rule.FixedFee,
		SourceNameSnapshot:    rate.SourceName,
		RateUpdatedAtSnapshot: &updatedAt,
		RoundingPrecision:     rule.RoundingPrecision,
		RoundingMode:          rule.RoundingMode,
		ExpiresAt:             now.Add(time.Duration(ttlMinutes) * time.Minute),
		CreatedAt:             now,
	}, nil
}
