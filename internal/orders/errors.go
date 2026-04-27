package orders

import "errors"

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderNotActive        = errors.New("order is not active")
	ErrOrderAlreadyExpired   = errors.New("order already expired")
	ErrVersionConflict       = errors.New("order version conflict")
	ErrHoldNotActive         = errors.New("order hold is not active")
	ErrInsufficientAvailable = errors.New("insufficient available balance")
	ErrInsufficientReserved  = errors.New("insufficient reserved balance")
	ErrIdempotencyConflict   = errors.New("idempotency key is already in progress or mismatched")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrQuoteNotFound         = errors.New("quote not found")
	ErrQuoteExpired          = errors.New("quote expired")
	ErrQuoteAlreadyConsumed  = errors.New("quote already consumed")
	ErrAccountWiringNotFound = errors.New("account wiring not found")
	ErrShiftNotOpen          = errors.New("cash shift is not open")
)
