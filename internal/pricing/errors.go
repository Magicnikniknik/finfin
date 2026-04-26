package pricing

import "errors"

var (
	ErrInvalidQuoteInput      = errors.New("invalid quote input")
	ErrBaseRateNotFound       = errors.New("base rate not found")
	ErrNoMarginRuleFound      = errors.New("no margin rule found")
	ErrRateStale              = errors.New("rate stale")
	ErrRateGuardrailTriggered = errors.New("rate guardrail triggered")
	ErrQuoteExpired           = errors.New("quote expired")
	ErrQuoteAlreadyConsumed   = errors.New("quote already consumed")
)
