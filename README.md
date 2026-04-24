# finfin

Pragmatic B2B white-label exchange platform blueprint with a working Go money-core start.

## Core modules implemented

- `internal/orders`: reserve/complete/cancel lifecycle with OCC + transactional SQL.
- `internal/orders/ttl_worker.go`: expiration worker for stale reserved orders (`FOR UPDATE SKIP LOCKED`).
- `internal/outbox`: outbox publisher worker.
- `cmd/ttl-worker`: CLI for TTL worker.
- `cmd/outbox-publisher`: CLI for outbox dispatch worker.

## Local run

1. Start local PostgreSQL:
   ```bash
   docker compose up -d postgres
   ```
2. Export DB URLs:
   ```bash
   export DATABASE_URL=postgres://postgres:postgres@localhost:5432/finfin_test?sslmode=disable
   export DATABASE_URL_TEST=$DATABASE_URL
   ```
3. Apply migration:
   ```bash
   psql "$DATABASE_URL" -f migrations/0001_core_schema.sql
   psql "$DATABASE_URL" -f migrations/0002_quote_snapshots.sql
   ```
4. Download dependencies:
   ```bash
   go mod tidy
   ```
5. Regenerate protobuf stubs (after proto changes):
   ```bash
   make proto
   ```
6. Run tests:
   ```bash
   go test ./...
   ```
7. Run TTL worker:
   ```bash
   go run ./cmd/ttl-worker
   ```
8. Run outbox publisher worker:
   ```bash
   go run ./cmd/outbox-publisher
   ```
9. Run gRPC server:
   ```bash
   go run ./cmd/grpc-server
   ```
10. Run integration test only:
   ```bash
   make test-integration
   ```

## Protobuf workflow

- Source of truth: `proto/exchange/order/v1/order_service.proto`
- Generated Go stubs path: `gen/exchange/order/v1`
- Regenerate:
  ```bash
  make proto
  ```

## Manual gRPC smoke

- Use `docs/grpcurl_smoke_checklist.md` for step-by-step `grpcurl` verification of:
  - `CompleteOrder`
  - `CancelOrder`
  - metadata/auth checks
  - OCC conflict checks
  - post-call DB validation queries
- Use `docs/sql_seed_checklist.md` for DB seed scripts to prepare:
  - quick reserved orders for `CompleteOrder`/`CancelOrder` smoke
  - full wiring + quote snapshot for `ReserveOrder` smoke
  - unified post-action SQL snapshots (full and operator-short variants)
- Executable SQL files are in `scripts/smoke/`:
  - `seed_fast.sql`, `seed_reserve.sql`, `reset.sql`
  - `check_short.sql`, `check_full.sql`
- grpcurl wrappers are in `scripts/smoke/`:
  - `reserve.sh`, `complete.sh`, `cancel.sh`
  - shared defaults loader: `common.sh`
  - sample env file: `env.example` (copy to `.env`)
- Handy targets:
  - `make smoke-seed-fast`
  - `make smoke-seed-reserve`
  - `make smoke-check-short`
  - `make smoke-check-full`
  - `make smoke-reset`
  - `make smoke-reserve`
  - `make smoke-complete`
  - `make smoke-cancel`
- Optional setup:
  ```bash
  cp scripts/smoke/env.example scripts/smoke/.env
  ```
- Dependency checks:
  - grpc wrappers fail fast with a clear message if `grpcurl` is missing
  - SQL smoke targets fail fast with a clear message if `psql` is missing

## Product phases

1. Core design (backend + DB).
2. BFF / operations integration.
3. White-label mobile apps.
4. Packaging & sales.

## Next integration step

- See `docs/nest_bff_plan.md` for a pragmatic NestJS BFF integration plan:
  - gRPC client wiring
  - HTTP endpoints (`/orders/reserve|complete|cancel`)
  - gRPC -> HTTP error mapping policy
  - first acceptance checklist
- See `docs/nest_bff_hardening.md` for:
  - global mobile-safe exception envelope
  - metadata extraction helper pattern
  - minimal BFF e2e smoke matrix
- Starter template files for quick bootstrapping live in `templates/nest-bff/src/`:
  - `app.module.ts`
  - `common/grpc-http-error.mapper.ts`
  - `common/filters/global-exception.filter.ts`
  - `common/decorators/current-tenant.decorator.ts`
  - `common/decorators/current-client-ref.decorator.ts`
  - `orders/interfaces/order-grpc.interface.ts`
  - `orders/orders.module.ts`
  - `orders/orders.service.ts`
  - `orders/orders.controller.ts`
  - `main.ts`
- e2e skeleton: `templates/nest-bff/test/orders.e2e-spec.ts`
- template notes: `templates/nest-bff/README.md`
- copy helper: `templates/nest-bff/apply.sh`
- post-copy verifier: `templates/nest-bff/POST_COPY_CHECKLIST.md`

## Next milestones

- Add integration tests against PostgreSQL (reserve/complete/cancel/idempotency/ttl/outbox).
- Add real RabbitMQ/Kafka publisher implementation in `internal/outbox`.
- Add gRPC transport layer on top of `internal/orders`.
