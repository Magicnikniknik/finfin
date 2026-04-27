package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	orderv1 "finfin/gen/exchange/order/v1"
	pricingv1 "finfin/gen/exchange/pricing/v1"
	"finfin/internal/app"
	grpcserver "finfin/internal/grpc"
	"finfin/internal/orders"
	"finfin/internal/pricing"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type orderApplicationAdapter struct {
	app *app.OrderApp
}

func (a *orderApplicationAdapter) ReserveOrder(ctx context.Context, cmd grpcserver.ReserveOrderCommand) (orders.ReserveOrderResult, error) {
	return a.app.ReserveOrder(ctx, app.ReserveOrderInput{
		TenantID:            cmd.TenantID,
		ClientRef:           cmd.ClientRef,
		IdempotencyKey:      cmd.IdempotencyKey,
		OfficeID:            cmd.OfficeID,
		QuoteID:             cmd.QuoteID,
		Side:                cmd.Side,
		GiveAmount:          cmd.GiveAmount,
		GiveCurrencyCode:    cmd.GiveCurrencyCode,
		GiveCurrencyNetwork: cmd.GiveCurrencyNetwork,
		GetAmount:           cmd.GetAmount,
		GetCurrencyCode:     cmd.GetCurrencyCode,
		GetCurrencyNetwork:  cmd.GetCurrencyNetwork,
	})
}

func (a *orderApplicationAdapter) CompleteOrder(ctx context.Context, cmd orders.CompleteOrderCommand) (orders.CompleteOrderResult, error) {
	return a.app.CompleteOrder(ctx, cmd)
}

func (a *orderApplicationAdapter) CancelOrder(ctx context.Context, cmd orders.CancelOrderCommand) (orders.CancelOrderResult, error) {
	return a.app.CancelOrder(ctx, cmd)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":9090"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("failed to init db pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	domainSvc := orders.NewService(pool, logger, orders.RealJournalPoster{})
	quoteResolver := app.NewSQLQuoteResolver(pool, logger)
	accountResolver := app.NewSQLAccountResolver(pool, logger)
	orderApp := app.NewOrderApp(domainSvc, quoteResolver, accountResolver, logger)
	adapter := &orderApplicationAdapter{app: orderApp}
	pricingRepo := pricing.NewSQLRepository(pool)
	pricingSvc := pricing.NewService(pricingRepo, nil)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to listen", "addr", addr, "error", err)
		os.Exit(1)
	}

	grpcSrv := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcSrv, grpcserver.NewOrderServer(adapter))
	pricingv1.RegisterPricingServiceServer(grpcSrv, grpcserver.NewPricingServer(pricingSvc))
	reflection.Register(grpcSrv)

	go func() {
		logger.Info("grpc server started",
			"addr", addr,
			"quote_resolver", "sql",
			"account_resolver", "sql",
		)
		if serveErr := grpcSrv.Serve(lis); serveErr != nil {
			logger.Error("grpc server stopped with error", "error", serveErr)
			cancel()
		}
	}()

	<-ctx.Done()
	logger.Info("grpc server shutting down")
	grpcSrv.GracefulStop()
	logger.Info("grpc server stopped")
}
