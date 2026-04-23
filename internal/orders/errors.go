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
)
