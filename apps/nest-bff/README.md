# apps/nest-bff

Real Nest BFF app instantiated from `templates/nest-bff` inside this monorepo.

## Transfer log

From repo root (`/workspace/finfin`):

```bash
bash templates/nest-bff/apply.sh --dry-run apps/nest-bff
bash templates/nest-bff/apply.sh --force apps/nest-bff
```

## Post-copy checklist

Checklist source:

```bash
cat templates/nest-bff/POST_COPY_CHECKLIST.md
```

Verified in this app:

- `src/orders/orders.module.ts` has `ORDER_GRPC_URL`, `ORDER_GRPC_PACKAGE`, `ORDER_GRPC_PROTO_PATH`.
- `src/orders/orders.service.ts` uses `getService<OrderGrpcService>('OrderService')`.
- `src/main.ts` wires `ValidationPipe` and `GlobalExceptionFilter`.
- Mock e2e file exercises `reserve|complete|cancel`, missing headers, and validation failure.

## Run locally

```bash
cd apps/nest-bff
npm install
npm run test:e2e -- test/orders.e2e-spec.ts
```

For live gRPC path, run core server in another terminal:

```bash
cd /workspace/finfin
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/finfin_test?sslmode=disable'
go run ./cmd/grpc-server
```

Then start Nest BFF:

```bash
cd /workspace/finfin/apps/nest-bff
export PORT=3000
export ORDER_GRPC_URL=127.0.0.1:9090
export ORDER_GRPC_PACKAGE=exchange.order.v1
export ORDER_GRPC_PROTO_PATH=/workspace/finfin/apps/nest-bff/proto/exchange/order/v1/order_service.proto
npm run start:dev
```

## First HTTP smoke

```bash
curl -X POST http://localhost:3000/orders/reserve \
  -H 'content-type: application/json' \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_smoke_001' \
  -d '{
    "idempotency_key": "idem-111",
    "office_id": "office-001",
    "quote_id": "quote-001",
    "side": "BUY",
    "give": {"amount":"100.00","currency":{"code":"USDT","network":"TRC20"}},
    "get": {"amount":"3550.00","currency":{"code":"THB","network":"cash"}}
  }'
```

See also: `apps/nest-bff/FIRST_LIVE_SMOKE_CHECKLIST.md`.
