package grpcserver

import (
	"errors"

	"finfin/internal/app"
	"finfin/internal/orders"
	"finfin/internal/pricing"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrReserveApplicationNotWired = errors.New("reserve application is not wired")

func MapDomainError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrReserveApplicationNotWired):
		return status.Error(codes.Unimplemented, err.Error())
	case errors.Is(err, app.ErrQuoteNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, app.ErrQuoteExpired):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, app.ErrAccountWiringNotFound):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrQuoteNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, orders.ErrQuoteExpired):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrQuoteAlreadyConsumed):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrAccountWiringNotFound):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrInvalidAmount):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, orders.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, orders.ErrVersionConflict):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, orders.ErrOrderNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrHoldNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrOrderAlreadyExpired):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrShiftNotOpen):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orders.ErrInsufficientAvailable):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, orders.ErrInsufficientReserved):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, orders.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, pricing.ErrInvalidQuoteInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, pricing.ErrBaseRateNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, pricing.ErrNoMarginRuleFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, pricing.ErrRateStale):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, pricing.ErrRateGuardrailTriggered):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
