# First live smoke checklist (`apps/nest-bff`)

Goal: validate real chain end-to-end.

```text
curl -> apps/nest-bff -> gRPC -> money-core -> Postgres
```

## 0) Prerequisites

- Postgres is running (`docker compose up -d` from repo root).
- Migrations applied.
- `cmd/grpc-server` starts without DB errors.
- `apps/nest-bff` dependencies installed.

## 1) Start sequence (recommended)

Terminal A (core):

```bash
cd /workspace/finfin
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/finfin_test?sslmode=disable'
export GRPC_ADDR=':9090'
go run ./cmd/grpc-server
```

Terminal B (BFF):

```bash
cd /workspace/finfin/apps/nest-bff
export PORT=3000
export ORDER_GRPC_URL='127.0.0.1:9090'
export ORDER_GRPC_PACKAGE='exchange.order.v1'
export ORDER_GRPC_PROTO_PATH='/workspace/finfin/apps/nest-bff/proto/exchange/order/v1/order_service.proto'
npm run start:dev
```

## 2) First live requests

### 2.1 Complete path (fast seed)

```bash
cd /workspace/finfin
make smoke-seed-fast
curl -X POST http://localhost:3000/orders/complete \
  -H 'content-type: application/json' \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_smoke_001' \
  -d '{
    "idempotency_key": "22222222-2222-2222-2222-222222222222",
    "order_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "expected_version": 1,
    "cashier_id": "cashier_01"
  }'
```

Then:

```bash
make smoke-check-short
make smoke-check-full
```

### 2.2 Reserve path

```bash
cd /workspace/finfin
make smoke-seed-reserve
curl -X POST http://localhost:3000/orders/reserve \
  -H 'content-type: application/json' \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_reserve_smoke_001' \
  -d '{
    "idempotency_key": "99999999-9999-9999-9999-999999999999",
    "office_id": "22222222-2222-2222-2222-222222222222",
    "quote_id": "quote-reserve-smoke-001",
    "side": "BUY",
    "give": {
      "amount": "100.00",
      "currency": { "code": "USDT", "network": "TRC20" }
    },
    "get": {
      "amount": "3550.00",
      "currency": { "code": "THB", "network": "cash" }
    }
  }'
```

## 3) Quick diagnosis map ("if X -> check Y")

- **BFF startup fails with proto errors**
  - check `ORDER_GRPC_PROTO_PATH` file exists
  - check `ORDER_GRPC_PACKAGE=exchange.order.v1`
- **HTTP 500 with gRPC unavailable/refused**
  - core server not running on `:9090`
  - wrong `ORDER_GRPC_URL`
- **HTTP 401 MISSING_TENANT / MISSING_CLIENT_REF**
  - missing headers in curl or UI form
- **HTTP 400 VALIDATION_FAILED**
  - invalid DTO payload shape/types
- **HTTP 409 / FAILED_PRECONDITION-like envelope**
  - business preconditions in core (version mismatch, invalid transition)
- **smoke-check SQL mismatches**
  - wrong seed used for scenario
  - request idempotency key reused unexpectedly

## 4) Success criteria

Live contour is considered healthy when:

1. HTTP request reaches BFF and returns mobile-safe envelope.
2. BFF forwards into gRPC (`OrderService`) without transport errors.
3. core persists expected state transitions.
4. `smoke-check-short/full` snapshots match expected records.
