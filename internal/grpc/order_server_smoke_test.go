package grpcserver

import (
	"context"
	"testing"
	"time"

	orderv1 "finfin/gen/exchange/order/v1"
	"finfin/internal/orders"

	"google.golang.org/grpc/metadata"
)

type fakeSmokeOrderApp struct {
	reserveResult  orders.ReserveOrderResult
	reserveErr     error
	lastReserveCmd ReserveOrderCommand

	completeResult  orders.CompleteOrderResult
	completeErr     error
	lastCompleteCmd orders.CompleteOrderCommand

	cancelResult  orders.CancelOrderResult
	cancelErr     error
	lastCancelCmd orders.CancelOrderCommand
}

func (f *fakeSmokeOrderApp) ReserveOrder(_ context.Context, cmd ReserveOrderCommand) (orders.ReserveOrderResult, error) {
	f.lastReserveCmd = cmd
	return f.reserveResult, f.reserveErr
}

func (f *fakeSmokeOrderApp) CompleteOrder(_ context.Context, cmd orders.CompleteOrderCommand) (orders.CompleteOrderResult, error) {
	f.lastCompleteCmd = cmd
	return f.completeResult, f.completeErr
}

func (f *fakeSmokeOrderApp) CancelOrder(_ context.Context, cmd orders.CancelOrderCommand) (orders.CancelOrderResult, error) {
	f.lastCancelCmd = cmd
	return f.cancelResult, f.cancelErr
}

func TestOrderServer_ReserveOrder_HappyPath(t *testing.T) {
	expiresAt := time.Unix(1_735_000_000, 0).UTC()

	app := &fakeSmokeOrderApp{
		reserveResult: orders.ReserveOrderResult{
			OrderID:   "order-123",
			OrderRef:  "ORD-ABC12345",
			Status:    "reserved",
			ExpiresAt: expiresAt,
			Version:   1,
		},
	}

	srv := NewOrderServer(app)
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-tenant-id", "tenant-1", "x-client-ref", "client-42"),
	)

	req := &orderv1.ReserveOrderRequest{
		IdempotencyKey: "idem-reserve-1",
		OfficeId:       "office-1",
		QuoteId:        "quote-1",
		Side:           orderv1.OrderSide_BUY,
		Give: &orderv1.Money{
			Amount: "100.00",
			Currency: &orderv1.Currency{
				Code:    "USDT",
				Network: "TRC20",
			},
		},
		Get: &orderv1.Money{
			Amount: "3550.00",
			Currency: &orderv1.Currency{
				Code:    "THB",
				Network: "cash",
			},
		},
	}

	resp, err := srv.ReserveOrder(ctx, req)
	if err != nil {
		t.Fatalf("ReserveOrder returned error: %v", err)
	}

	if resp.OrderId != "order-123" {
		t.Fatalf("expected order id order-123, got %q", resp.OrderId)
	}
	if resp.Status != orderv1.OrderStatus_RESERVED {
		t.Fatalf("expected RESERVED status, got %v", resp.Status)
	}
	if resp.ExpiresAtTs != expiresAt.Unix() {
		t.Fatalf("expected expires_at_ts %d, got %d", expiresAt.Unix(), resp.ExpiresAtTs)
	}
	if resp.Version != 1 {
		t.Fatalf("expected version 1, got %d", resp.Version)
	}

	if app.lastReserveCmd.TenantID != "tenant-1" ||
		app.lastReserveCmd.ClientRef != "client-42" ||
		app.lastReserveCmd.IdempotencyKey != "idem-reserve-1" ||
		app.lastReserveCmd.OfficeID != "office-1" ||
		app.lastReserveCmd.QuoteID != "quote-1" ||
		app.lastReserveCmd.Side != "buy" ||
		app.lastReserveCmd.GiveAmount != "100.00" ||
		app.lastReserveCmd.GiveCurrencyCode != "USDT" ||
		app.lastReserveCmd.GiveCurrencyNetwork != "TRC20" ||
		app.lastReserveCmd.GetAmount != "3550.00" ||
		app.lastReserveCmd.GetCurrencyCode != "THB" ||
		app.lastReserveCmd.GetCurrencyNetwork != "cash" {
		t.Fatalf("unexpected reserve command mapping: %+v", app.lastReserveCmd)
	}
}

func TestOrderServer_CompleteOrder_HappyPath(t *testing.T) {
	completedAt := time.Unix(1_735_100_000, 0).UTC()
	app := &fakeSmokeOrderApp{
		completeResult: orders.CompleteOrderResult{
			OrderID:     "order-456",
			OrderRef:    "ORD-COMPLETE1",
			Status:      "completed",
			Version:     2,
			CompletedAt: completedAt,
		},
	}

	srv := NewOrderServer(app)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-tenant-id", "tenant-2"))
	req := &orderv1.CompleteOrderRequest{
		IdempotencyKey:  "idem-complete-1",
		OrderId:         "order-456",
		ExpectedVersion: 1,
		CashierId:       "cashier-7",
	}

	resp, err := srv.CompleteOrder(ctx, req)
	if err != nil {
		t.Fatalf("CompleteOrder returned error: %v", err)
	}

	if resp.OrderId != "order-456" ||
		resp.Status != orderv1.OrderStatus_COMPLETED ||
		resp.CompletedAtTs != completedAt.Unix() ||
		resp.Version != 2 {
		t.Fatalf("unexpected complete response: %+v", resp)
	}

	if app.lastCompleteCmd.TenantID != "tenant-2" ||
		app.lastCompleteCmd.OrderID != "order-456" ||
		app.lastCompleteCmd.ExpectedVersion != 1 ||
		app.lastCompleteCmd.IdempotencyKey != "idem-complete-1" ||
		app.lastCompleteCmd.CashierID != "cashier-7" {
		t.Fatalf("unexpected complete command mapping: %+v", app.lastCompleteCmd)
	}
}

func TestOrderServer_CancelOrder_HappyPath(t *testing.T) {
	app := &fakeSmokeOrderApp{
		cancelResult: orders.CancelOrderResult{
			OrderID:  "order-789",
			OrderRef: "ORD-CANCEL01",
			Status:   "cancelled",
			Version:  2,
		},
	}

	srv := NewOrderServer(app)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-tenant-id", "tenant-3"))
	req := &orderv1.CancelOrderRequest{
		IdempotencyKey:  "idem-cancel-1",
		OrderId:         "order-789",
		ExpectedVersion: 1,
		Reason:          "client_no_show",
	}

	resp, err := srv.CancelOrder(ctx, req)
	if err != nil {
		t.Fatalf("CancelOrder returned error: %v", err)
	}

	if resp.OrderId != "order-789" ||
		resp.Status != orderv1.OrderStatus_CANCELLED ||
		resp.Version != 2 {
		t.Fatalf("unexpected cancel response: %+v", resp)
	}

	if app.lastCancelCmd.TenantID != "tenant-3" ||
		app.lastCancelCmd.OrderID != "order-789" ||
		app.lastCancelCmd.ExpectedVersion != 1 ||
		app.lastCancelCmd.IdempotencyKey != "idem-cancel-1" ||
		app.lastCancelCmd.Reason != "client_no_show" {
		t.Fatalf("unexpected cancel command mapping: %+v", app.lastCancelCmd)
	}
}
