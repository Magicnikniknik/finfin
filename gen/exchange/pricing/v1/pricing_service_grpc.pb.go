package pricingv1

import (
	"context"

	"google.golang.org/grpc"
)

type PricingServiceClient interface {
	CalculateQuote(ctx context.Context, in *CalculateQuoteRequest, opts ...grpc.CallOption) (*CalculateQuoteResponse, error)
}

type pricingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPricingServiceClient(cc grpc.ClientConnInterface) PricingServiceClient {
	return &pricingServiceClient{cc}
}

func (c *pricingServiceClient) CalculateQuote(ctx context.Context, in *CalculateQuoteRequest, opts ...grpc.CallOption) (*CalculateQuoteResponse, error) {
	out := new(CalculateQuoteResponse)
	err := c.cc.Invoke(ctx, "/exchange.pricing.v1.PricingService/CalculateQuote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type UnimplementedPricingServiceServer struct{}

func RegisterPricingServiceServer(s grpc.ServiceRegistrar, srv PricingServiceServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "exchange.pricing.v1.PricingService",
		HandlerType: (*PricingServiceServer)(nil),
		Methods:     []grpc.MethodDesc{},
		Streams:     []grpc.StreamDesc{},
		Metadata:    "pricing_service.proto",
	}, srv)
}
