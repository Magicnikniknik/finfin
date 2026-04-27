package grpcserver

import (
	"context"
	"strings"

	orderv1 "finfin/gen/exchange/order/v1"
	"finfin/internal/orders"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ReserveOrderCommand struct {
	TenantID  string
	ClientRef string

	IdempotencyKey string
	OfficeID       string
	QuoteID        string
	Side           string

	GiveAmount          string
	GiveCurrencyCode    string
	GiveCurrencyNetwork string
	GetAmount           string
	GetCurrencyCode     string
	GetCurrencyNetwork  string
}

type OrderApplication interface {
	ReserveOrder(ctx context.Context, cmd ReserveOrderCommand) (orders.ReserveOrderResult, error)
	CompleteOrder(ctx context.Context, cmd orders.CompleteOrderCommand) (orders.CompleteOrderResult, error)
	CancelOrder(ctx context.Context, cmd orders.CancelOrderCommand) (orders.CancelOrderResult, error)
}

type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	app OrderApplication
}

func NewOrderServer(app OrderApplication) *OrderServer {
	return &OrderServer{app: app}
}

func (s *OrderServer) ReserveOrder(ctx context.Context, req *orderv1.ReserveOrderRequest) (*orderv1.ReserveOrderResponse, error) {
	if s.app == nil {
		return nil, status.Error(codes.Unimplemented, "order application is not configured")
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

	give, err := requireMoney(req.GetGive(), "give")
	if err != nil {
		return nil, err
	}
	get, err := requireMoney(req.GetGet(), "get")
	if err != nil {
		return nil, err
	}

	side, err := mapProtoSide(req.GetSide())
	if err != nil {
		return nil, err
	}

	res, err := s.app.ReserveOrder(ctx, ReserveOrderCommand{
		TenantID:            tenantID,
		ClientRef:           clientRef,
		IdempotencyKey:      strings.TrimSpace(req.GetIdempotencyKey()),
		OfficeID:            strings.TrimSpace(req.GetOfficeId()),
		QuoteID:             strings.TrimSpace(req.GetQuoteId()),
		Side:                side,
		GiveAmount:          give.amount,
		GiveCurrencyCode:    give.code,
		GiveCurrencyNetwork: give.network,
		GetAmount:           get.amount,
		GetCurrencyCode:     get.code,
		GetCurrencyNetwork:  get.network,
	})
	if err != nil {
		return nil, MapDomainError(err)
	}

	return &orderv1.ReserveOrderResponse{
		OrderId:     res.OrderID,
		Status:      mapOrderStatus(res.Status),
		ExpiresAtTs: res.ExpiresAt.Unix(),
		Version:     res.Version,
	}, nil
}

func (s *OrderServer) CompleteOrder(ctx context.Context, req *orderv1.CompleteOrderRequest) (*orderv1.CompleteOrderResponse, error) {
	if s.app == nil {
		return nil, status.Error(codes.Unimplemented, "order application is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	tenantID, err := requireMetadata(ctx, "x-tenant-id")
	if err != nil {
		return nil, err
	}

	res, err := s.app.CompleteOrder(ctx, orders.CompleteOrderCommand{
		TenantID:        tenantID,
		OrderID:         strings.TrimSpace(req.GetOrderId()),
		ExpectedVersion: req.GetExpectedVersion(),
		IdempotencyKey:  strings.TrimSpace(req.GetIdempotencyKey()),
		CashierID:       strings.TrimSpace(req.GetCashierId()),
	})
	if err != nil {
		return nil, MapDomainError(err)
	}

	return &orderv1.CompleteOrderResponse{
		OrderId:       res.OrderID,
		Status:        mapOrderStatus(res.Status),
		CompletedAtTs: res.CompletedAt.Unix(),
		Version:       res.Version,
	}, nil
}

func (s *OrderServer) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	if s.app == nil {
		return nil, status.Error(codes.Unimplemented, "order application is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	tenantID, err := requireMetadata(ctx, "x-tenant-id")
	if err != nil {
		return nil, err
	}

	res, err := s.app.CancelOrder(ctx, orders.CancelOrderCommand{
		TenantID:        tenantID,
		OrderID:         strings.TrimSpace(req.GetOrderId()),
		ExpectedVersion: req.GetExpectedVersion(),
		IdempotencyKey:  strings.TrimSpace(req.GetIdempotencyKey()),
		Reason:          strings.TrimSpace(req.GetReason()),
	})
	if err != nil {
		return nil, MapDomainError(err)
	}

	return &orderv1.CancelOrderResponse{
		OrderId: res.OrderID,
		Status:  mapOrderStatus(res.Status),
		Version: res.Version,
	}, nil
}

type moneyView struct {
	amount  string
	code    string
	network string
}

func requireMoney(m *orderv1.Money, field string) (moneyView, error) {
	if m == nil {
		return moneyView{}, status.Errorf(codes.InvalidArgument, "%s is required", field)
	}
	c := m.GetCurrency()
	if c == nil {
		return moneyView{}, status.Errorf(codes.InvalidArgument, "%s.currency is required", field)
	}

	amount := strings.TrimSpace(m.GetAmount())
	code := strings.TrimSpace(c.GetCode())
	network := strings.TrimSpace(c.GetNetwork())

	if amount == "" {
		return moneyView{}, status.Errorf(codes.InvalidArgument, "%s.amount is required", field)
	}
	if code == "" {
		return moneyView{}, status.Errorf(codes.InvalidArgument, "%s.currency.code is required", field)
	}
	if network == "" {
		return moneyView{}, status.Errorf(codes.InvalidArgument, "%s.currency.network is required", field)
	}

	return moneyView{amount: amount, code: code, network: network}, nil
}

func requireMetadata(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md.Get(key)
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata key %s", key)
	}
	return strings.TrimSpace(values[0]), nil
}

func mapProtoSide(side orderv1.OrderSide) (string, error) {
	switch side {
	case orderv1.OrderSide_BUY:
		return "buy", nil
	case orderv1.OrderSide_SELL:
		return "sell", nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid order side")
	}
}

func mapOrderStatus(statusText string) orderv1.OrderStatus {
	switch statusText {
	case "reserved":
		return orderv1.OrderStatus_RESERVED
	case "completed":
		return orderv1.OrderStatus_COMPLETED
	case "expired":
		return orderv1.OrderStatus_EXPIRED
	case "cancelled":
		return orderv1.OrderStatus_CANCELLED
	case "new":
		return orderv1.OrderStatus_NEW
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
