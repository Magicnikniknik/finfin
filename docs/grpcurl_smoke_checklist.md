# grpcurl smoke checklist (CompleteOrder / CancelOrder)

This checklist is for manual smoke verification against `cmd/grpc-server` on `:9090`.

## 0) Start server

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/finfin_test?sslmode=disable'
export GRPC_ADDR=':9090'
go run ./cmd/grpc-server
```

## 1) Verify service discovery / reflection

```bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 describe exchange.order.v1.OrderService
```

Expected service:
- `exchange.order.v1.OrderService`

## 2) Metadata headers

Recommended to always pass both:

```bash
-H 'x-tenant-id: <TENANT_ID>'
-H 'x-client-ref: <CLIENT_REF>'
```

Notes:
- `CompleteOrder` requires `x-tenant-id`
- `CancelOrder` requires `x-tenant-id`

## 3) DB preconditions

For `CompleteOrder`/`CancelOrder`, ensure there is an existing `reserved` order and you know:
- `order_id`
- `expected_version` (usually `1` after reserve)

## 4) Smoke: CompleteOrder

```bash
grpcurl -plaintext \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_smoke_001' \
  -d '{
    "idempotency_key": "22222222-2222-2222-2222-222222222222",
    "order_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "expected_version": 1,
    "cashier_id": "cashier_01"
  }' \
  localhost:9090 \
  exchange.order.v1.OrderService/CompleteOrder
```

Expected response shape:

```json
{
  "orderId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "status": "COMPLETED",
  "completedAtTs": "1735100000",
  "version": "2"
}
```

Post-check SQL:

```sql
SELECT id, status, version, completed_at
FROM core.orders
WHERE id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';

SELECT order_id, status, consumed_at
FROM core.order_holds
WHERE order_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';

SELECT account_id, available, reserved
FROM core.account_balances;

SELECT kind, order_id, created_at
FROM core.ledger_journals
WHERE order_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
ORDER BY created_at DESC;

SELECT event_type, status, attempts
FROM core.outbox_events
WHERE aggregate_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
ORDER BY created_at DESC;
```

## 5) Smoke: CancelOrder

```bash
grpcurl -plaintext \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_smoke_001' \
  -d '{
    "idempotency_key": "33333333-3333-3333-3333-333333333333",
    "order_id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
    "expected_version": 1,
    "reason": "client_no_show"
  }' \
  localhost:9090 \
  exchange.order.v1.OrderService/CancelOrder
```

Expected response shape:

```json
{
  "orderId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
  "status": "CANCELLED",
  "version": "2"
}
```

Post-check SQL:

```sql
SELECT id, status, version, cancelled_at, cancel_reason
FROM core.orders
WHERE id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';

SELECT order_id, status, released_at
FROM core.order_holds
WHERE order_id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';

SELECT account_id, available, reserved
FROM core.account_balances;

SELECT kind, order_id, created_at
FROM core.ledger_journals
WHERE order_id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'
ORDER BY created_at DESC;

SELECT event_type, status, attempts
FROM core.outbox_events
WHERE aggregate_id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'
ORDER BY created_at DESC;
```

## 6) OCC manual check

Call `CompleteOrder` twice with same `order_id` and `expected_version=1`, but different idempotency keys.

Expected second call:
- gRPC status code `Aborted`

## 7) Transport auth/metadata check

Without `x-tenant-id`, call `CompleteOrder`:
- expected gRPC status code `Unauthenticated`

## 8) Metadata env helpers

```bash
export GRPC_META_TENANT="-H x-tenant-id:11111111-1111-1111-1111-111111111111"
export GRPC_META_CLIENT="-H x-client-ref:client_smoke_001"
```

Then:

```bash
grpcurl -plaintext \
  $GRPC_META_TENANT \
  $GRPC_META_CLIENT \
  -d '{
    "idempotency_key": "77777777-7777-7777-7777-777777777777",
    "order_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
    "expected_version": 1,
    "reason": "manual_cancel"
  }' \
  localhost:9090 \
  exchange.order.v1.OrderService/CancelOrder
```

## 9) If reflection is unavailable

```bash
grpcurl -plaintext \
  -import-path ./proto \
  -proto exchange/order/v1/order_service.proto \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -d '{
    "idempotency_key": "88888888-8888-8888-8888-888888888888",
    "order_id": "ffffffff-ffff-ffff-ffff-ffffffffffff",
    "expected_version": 1,
    "cashier_id": "cashier_01"
  }' \
  localhost:9090 \
  exchange.order.v1.OrderService/CompleteOrder
```

## Pre-run mini checklist

- server started
- reflection enabled
- `x-tenant-id` included
- reserved order exists in DB
- `expected_version` matches row version
- each independent call uses a new `idempotency_key`
