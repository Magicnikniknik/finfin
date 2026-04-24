# NestJS BFF integration plan (OrderService gRPC)

This is a pragmatic next-step blueprint to expose `OrderService` over HTTP for mobile/web clients.

## 1) Goal

Build an HTTP façade in NestJS:

- `POST /orders/reserve`
- `POST /orders/complete`
- `POST /orders/cancel`

Backed by gRPC calls to `exchange.order.v1.OrderService`.

---

## 2) Suggested implementation order

1. Configure NestJS gRPC client (`@nestjs/microservices` + `Transport.GRPC`).
2. Add typed DTOs for HTTP requests/responses.
3. Implement a BFF service that maps HTTP DTO -> gRPC request + metadata.
4. Implement a gRPC-to-HTTP error mapper (status code + machine-readable error code).
5. Add HTTP controller endpoints.
6. Add minimal e2e tests for happy-path + key error mappings.

---

## 3) gRPC client wiring (NestJS)

```ts
// orders.grpc.module.ts
import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { join } from 'path';

@Module({
  imports: [
    ClientsModule.register([
      {
        name: 'ORDER_GRPC',
        transport: Transport.GRPC,
        options: {
          url: process.env.ORDER_GRPC_ADDR ?? 'localhost:9090',
          package: 'exchange.order.v1',
          protoPath: join(process.cwd(), 'proto/exchange/order/v1/order_service.proto'),
        },
      },
    ]),
  ],
  exports: [ClientsModule],
})
export class OrdersGrpcModule {}
```

---

## 4) Error mapping policy (recommended)

Map gRPC status to HTTP status:

- `INVALID_ARGUMENT` -> `400`
- `UNAUTHENTICATED` -> `401`
- `NOT_FOUND` -> `404`
- `FAILED_PRECONDITION` -> `409`
- `ABORTED` -> `409`
- `RESOURCE_EXHAUSTED` -> `409`
- `ALREADY_EXISTS` -> `409`
- fallback -> `500`

Return body shape:

```json
{
  "code": "ORDER_QUOTE_NOT_FOUND",
  "message": "quote not found"
}
```

Keep `code` stable for mobile clients; do not leak raw internal errors directly.

---

## 5) HTTP endpoint contracts (minimal)

### POST /orders/reserve

Headers:
- `x-tenant-id` (required)
- `x-client-ref` (required)

Body:

```json
{
  "idempotencyKey": "uuid-or-stable-key",
  "officeId": "uuid",
  "quoteId": "quote-reserve-smoke-001",
  "side": "BUY",
  "give": { "amount": "100.00", "currency": { "code": "USDT", "network": "TRC20" } },
  "get": { "amount": "3550.00", "currency": { "code": "THB", "network": "cash" } }
}
```

### POST /orders/complete

Headers:
- `x-tenant-id` (required)

Body:

```json
{
  "idempotencyKey": "uuid-or-stable-key",
  "orderId": "uuid",
  "expectedVersion": 1,
  "cashierId": "cashier_01"
}
```

### POST /orders/cancel

Headers:
- `x-tenant-id` (required)

Body:

```json
{
  "idempotencyKey": "uuid-or-stable-key",
  "orderId": "uuid",
  "expectedVersion": 1,
  "reason": "client_no_show"
}
```

---

## 6) First acceptance checklist

1. Reserve happy path via HTTP returns `orderId`, `status`, `version`, `expiresAtTs`.
2. Complete happy path returns `COMPLETED`, version increment.
3. Cancel happy path returns `CANCELLED`, version increment.
4. Missing `x-tenant-id` -> HTTP 401.
5. `ErrQuoteNotFound` -> HTTP 404.
6. OCC conflict (`ABORTED`) -> HTTP 409.

---

## 7) Runtime env

- `ORDER_GRPC_ADDR` (example: `localhost:9090`)
- optional request timeout for each gRPC call (recommended 2-5s)
- correlation/request id propagation from HTTP to gRPC metadata (recommended)
