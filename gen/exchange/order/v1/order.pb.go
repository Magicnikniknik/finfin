package orderv1

import (
	"context"

	"google.golang.org/grpc"
)

type OrderSide int32

const (
	OrderSide_ORDER_SIDE_UNSPECIFIED OrderSide = 0
	OrderSide_BUY                    OrderSide = 1
	OrderSide_SELL                   OrderSide = 2
)

type OrderStatus int32

const (
	OrderStatus_ORDER_STATUS_UNSPECIFIED OrderStatus = 0
	OrderStatus_NEW                      OrderStatus = 1
	OrderStatus_RESERVED                 OrderStatus = 2
	OrderStatus_COMPLETED                OrderStatus = 3
	OrderStatus_EXPIRED                  OrderStatus = 4
	OrderStatus_CANCELLED                OrderStatus = 5
)

type Currency struct {
	Code    string
	Network string
}

func (c *Currency) GetCode() string {
	if c == nil {
		return ""
	}
	return c.Code
}
func (c *Currency) GetNetwork() string {
	if c == nil {
		return ""
	}
	return c.Network
}

type Money struct {
	Amount   string
	Currency *Currency
}

func (m *Money) GetAmount() string {
	if m == nil {
		return ""
	}
	return m.Amount
}
func (m *Money) GetCurrency() *Currency {
	if m == nil {
		return nil
	}
	return m.Currency
}

type ReserveOrderRequest struct {
	IdempotencyKey string
	OfficeId       string
	QuoteId        string
	Side           OrderSide
	Give           *Money
	Get            *Money
}

func (r *ReserveOrderRequest) GetIdempotencyKey() string {
	if r == nil {
		return ""
	}
	return r.IdempotencyKey
}
func (r *ReserveOrderRequest) GetOfficeId() string {
	if r == nil {
		return ""
	}
	return r.OfficeId
}
func (r *ReserveOrderRequest) GetQuoteId() string {
	if r == nil {
		return ""
	}
	return r.QuoteId
}
func (r *ReserveOrderRequest) GetSide() OrderSide {
	if r == nil {
		return OrderSide_ORDER_SIDE_UNSPECIFIED
	}
	return r.Side
}
func (r *ReserveOrderRequest) GetGive() *Money {
	if r == nil {
		return nil
	}
	return r.Give
}
func (r *ReserveOrderRequest) GetGet() *Money {
	if r == nil {
		return nil
	}
	return r.Get
}

type ReserveOrderResponse struct {
	OrderId     string
	Status      OrderStatus
	ExpiresAtTs int64
	Version     int64
}

type CompleteOrderRequest struct {
	OrderId         string
	ExpectedVersion int64
	IdempotencyKey  string
	CashierId       string
}

func (r *CompleteOrderRequest) GetOrderId() string {
	if r == nil {
		return ""
	}
	return r.OrderId
}
func (r *CompleteOrderRequest) GetExpectedVersion() int64 {
	if r == nil {
		return 0
	}
	return r.ExpectedVersion
}
func (r *CompleteOrderRequest) GetIdempotencyKey() string {
	if r == nil {
		return ""
	}
	return r.IdempotencyKey
}
func (r *CompleteOrderRequest) GetCashierId() string {
	if r == nil {
		return ""
	}
	return r.CashierId
}

type CompleteOrderResponse struct {
	OrderId       string
	Status        OrderStatus
	CompletedAtTs int64
	Version       int64
}

type CancelOrderRequest struct {
	OrderId         string
	ExpectedVersion int64
	IdempotencyKey  string
	Reason          string
}

func (r *CancelOrderRequest) GetOrderId() string {
	if r == nil {
		return ""
	}
	return r.OrderId
}
func (r *CancelOrderRequest) GetExpectedVersion() int64 {
	if r == nil {
		return 0
	}
	return r.ExpectedVersion
}
func (r *CancelOrderRequest) GetIdempotencyKey() string {
	if r == nil {
		return ""
	}
	return r.IdempotencyKey
}
func (r *CancelOrderRequest) GetReason() string {
	if r == nil {
		return ""
	}
	return r.Reason
}

type CancelOrderResponse struct {
	OrderId string
	Status  OrderStatus
	Version int64
}

type OrderServiceServer interface {
	ReserveOrder(context.Context, *ReserveOrderRequest) (*ReserveOrderResponse, error)
	CompleteOrder(context.Context, *CompleteOrderRequest) (*CompleteOrderResponse, error)
	CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderResponse, error)
}

type UnimplementedOrderServiceServer struct{}

func RegisterOrderServiceServer(s grpc.ServiceRegistrar, srv OrderServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "exchange.order.v1.OrderService",
		HandlerType: (*OrderServiceServer)(nil),
		Methods:     []grpc.MethodDesc{},
		Streams:     []grpc.StreamDesc{},
		Metadata:    "order.proto",
	}, srv)
}
