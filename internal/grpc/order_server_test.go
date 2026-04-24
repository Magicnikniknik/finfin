package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	orderv1 "finfin/gen/exchange/order/v1"
	"finfin/internal/app"
	"finfin/internal/orders"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type testOrderApp struct {
	reserveResult  orders.ReserveOrderResult
	reserveErr     error
	completeResult orders.CompleteOrderResult
	completeErr    error
	cancelResult   orders.CancelOrderResult
	cancelErr      error
}

func (a testOrderApp) ReserveOrder(_ context.Context, _ ReserveOrderCommand) (orders.ReserveOrderResult, error) {
	return a.reserveResult, a.reserveErr
}

func (a testOrderApp) CompleteOrder(_ context.Context, _ orders.CompleteOrderCommand) (orders.CompleteOrderResult, error) {
	return a.completeResult, a.completeErr
}

func (a testOrderApp) CancelOrder(_ context.Context, _ orders.CancelOrderCommand) (orders.CancelOrderResult, error) {
	return a.cancelResult, a.cancelErr
}

func TestReserveOrder_MissingTenantMetadata_ReturnsUnauthenticated(t *testing.T) {
	srv := NewOrderServer(testOrderApp{})
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-client-ref", "client-1"))

	_, err := srv.ReserveOrder(ctx, validReserveReq())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Fatalf("expected code %s, got %s", codes.Unauthenticated, got)
	}
}

func TestReserveOrder_MissingClientRefMetadata_ReturnsUnauthenticated(t *testing.T) {
	srv := NewOrderServer(testOrderApp{})
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-tenant-id", "tenant-1"))

	_, err := srv.ReserveOrder(ctx, validReserveReq())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Fatalf("expected code %s, got %s", codes.Unauthenticated, got)
	}
}

func TestReserveOrder_AppAndDomainErrors_AreMapped(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{name: "quote not found", err: app.ErrQuoteNotFound, want: codes.NotFound},
		{name: "quote expired", err: app.ErrQuoteExpired, want: codes.FailedPrecondition},
		{name: "quote mismatch", err: app.ErrQuoteMismatch, want: codes.InvalidArgument},
		{name: "account wiring missing", err: app.ErrAccountWiringNotFound, want: codes.FailedPrecondition},
		{name: "insufficient available", err: orders.ErrInsufficientAvailable, want: codes.ResourceExhausted},
	}

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-tenant-id", "tenant-1", "x-client-ref", "client-1"),
	)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewOrderServer(testOrderApp{reserveErr: tc.err})
			_, err := srv.ReserveOrder(ctx, validReserveReq())
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if got := status.Code(err); got != tc.want {
				t.Fatalf("expected code %s, got %s", tc.want, got)
			}
		})
	}
}

func TestCompleteOrder_DomainErrors_AreMapped(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{name: "version conflict", err: orders.ErrVersionConflict, want: codes.Aborted},
		{name: "insufficient reserved", err: orders.ErrInsufficientReserved, want: codes.ResourceExhausted},
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-tenant-id", "tenant-1"))
	req := &orderv1.CompleteOrderRequest{
		OrderId:         "order-1",
		ExpectedVersion: 2,
		IdempotencyKey:  "ik-1",
		CashierId:       "cashier-1",
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewOrderServer(testOrderApp{completeErr: tc.err})
			_, err := srv.CompleteOrder(ctx, req)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if got := status.Code(err); got != tc.want {
				t.Fatalf("expected code %s, got %s", tc.want, got)
			}
		})
	}
}

func TestMapDomainError_WiredSentinel(t *testing.T) {
	err := MapDomainError(errors.Join(ErrReserveApplicationNotWired, errors.New("extra")))
	if got := status.Code(err); got != codes.Unimplemented {
		t.Fatalf("expected code %s, got %s", codes.Unimplemented, got)
	}
}

func validReserveReq() *orderv1.ReserveOrderRequest {
	return &orderv1.ReserveOrderRequest{
		IdempotencyKey: "ik-1",
		OfficeId:       "office-1",
		QuoteId:        "quote-1",
		Side:           orderv1.OrderSide_BUY,
		Give: &orderv1.Money{
			Amount: "10.5",
			Currency: &orderv1.Currency{
				Code:    "USD",
				Network: "FIAT",
			},
		},
		Get: &orderv1.Money{
			Amount: "0.1",
			Currency: &orderv1.Currency{
				Code:    "BTC",
				Network: "BTC",
			},
		},
		ExpiresAtTs: time.Now().Add(5 * time.Minute).Unix(),
	}
}
