# finfin

Pragmatic B2B white-label exchange platform blueprint with a working Go money-core start.

## Core modules implemented

- `internal/orders`: reserve/complete/cancel lifecycle with OCC + transactional SQL.
- `internal/orders/ttl_worker.go`: expiration worker for stale reserved orders (`FOR UPDATE SKIP LOCKED`).
- `internal/outbox`: outbox publisher worker.
- `cmd/ttl-worker`: CLI for TTL worker.
- `cmd/outbox-publisher`: CLI for outbox dispatch worker.

## Local run

1. Create PostgreSQL database and export `DATABASE_URL`.
2. Apply migration:
   ```bash
   psql "$DATABASE_URL" -f migrations/0001_core_schema.sql
   ```
3. Download dependencies:
   ```bash
   go mod tidy
   ```
4. Run tests:
   ```bash
   go test ./...
   ```
5. Run TTL worker:
   ```bash
   go run ./cmd/ttl-worker
   ```
6. Run outbox publisher worker:
   ```bash
   go run ./cmd/outbox-publisher
   ```

## Product phases

1. Core design (backend + DB).
2. Operations web panel.
3. White-label mobile apps.
4. Packaging & sales.

## Next milestones

- Add integration tests against PostgreSQL (reserve/complete/cancel/idempotency/ttl/outbox).
- Add real RabbitMQ/Kafka publisher implementation in `internal/outbox`.
- Add gRPC transport layer on top of `internal/orders`.
