# Post-copy 5-minute checklist (Nest BFF)

After copying from `templates/nest-bff`, run this quick verification flow.

## 1) Files are present

```text
src/common/grpc-http-error.mapper.ts
src/common/filters/global-exception.filter.ts
src/common/decorators/current-tenant.decorator.ts
src/common/decorators/current-client-ref.decorator.ts

src/orders/interfaces/order-grpc.interface.ts
src/orders/dto/reserve-order.dto.ts
src/orders/dto/complete-order.dto.ts
src/orders/dto/cancel-order.dto.ts
src/orders/orders.service.ts
src/orders/orders.controller.ts
src/orders/orders.module.ts

src/app.module.ts
src/main.ts

test/orders.e2e-spec.ts
```

## 2) `OrdersModule` matches proto

Check in `src/orders/orders.module.ts`:

- `ORDER_GRPC_URL`
- `ORDER_GRPC_PACKAGE=exchange.order.v1`
- `ORDER_GRPC_PROTO_PATH` points to `proto/exchange/order/v1/order_service.proto`

## 3) gRPC service name matches

In `src/orders/orders.service.ts`:

```ts
this.grpcClient.getService<OrderGrpcService>('OrderService')
```

## 4) Bootstrap is wired

In `src/main.ts` ensure:

- `ValidationPipe`
- `GlobalExceptionFilter`

## 5) Controller uses decorators (no manual header duplication)

In `src/orders/orders.controller.ts` ensure:

- `@CurrentTenant()`
- `@CurrentClientRef()`

## 6) App starts

Minimum env:

```env
PORT=3000
ORDER_GRPC_URL=127.0.0.1:9090
ORDER_GRPC_PACKAGE=exchange.order.v1
ORDER_GRPC_PROTO_PATH=proto/exchange/order/v1/order_service.proto
```

Expected:
- Nest starts without DI/import errors.

## 7) E2E skeleton passes on mock service first

At minimum:

- missing tenant -> `401`
- missing client_ref -> `401`
- validation failed -> `400`
- mocked reserve/complete/cancel happy path

## 8) First manual HTTP smoke

Reserve request should reach controller/service and return expected JSON shape.

## 9) Common breakpoints

- wrong gRPC `package`
- wrong `protoPath`
- `OrdersModule` not imported into `AppModule`
- decorator/filter import path mismatches
- DTO payload mismatch
- global filter not registered

## 10) Ready for real backend

Transfer is considered successful if:

1. Nest starts
2. e2e skeleton passes (mock path)
3. `/orders/reserve|complete|cancel` respond
4. `x-tenant-id` and `x-client-ref` propagate to service
