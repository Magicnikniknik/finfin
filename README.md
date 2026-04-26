FINFIN PR #3 — CONFLICT RESOLUTION PACK

Use this in GitHub “Resolve conflicts”.

========================================
1) Makefile — REPLACE WHOLE FILE WITH THIS
========================================
.PHONY: db-up db-down migrate proto test test-integration run-ttl run-outbox smoke-seed-fast smoke-seed-reserve smoke-reset smoke-check-short smoke-check-full smoke-reserve smoke-complete smoke-cancel

db-up:
	docker compose up -d postgres

db-down:
	docker compose down -v

migrate:
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f migrations/0001_core_schema.sql
	psql "$$DATABASE_URL" -f migrations/0002_quote_snapshots.sql
	psql "$$DATABASE_URL" -f migrations/0003_account_wiring.sql
	psql "$$DATABASE_URL" -f migrations/0004_pricing_engine.sql

proto:
	@command -v protoc >/dev/null 2>&1 || (echo "protoc is required" && exit 1)
	@command -v protoc-gen-go >/dev/null 2>&1 || (echo "protoc-gen-go is required: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" && exit 1)
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || (echo "protoc-gen-go-grpc is required: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" && exit 1)
	protoc \
		-I proto \
		--go_out=gen --go_opt=paths=source_relative \
		--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
		proto/exchange/order/v1/order_service.proto

test:
	go test ./...

test-integration:
	@test -n "$$DATABASE_URL_TEST" || (echo "DATABASE_URL_TEST is required. Run: export DATABASE_URL_TEST=postgres://postgres:postgres@localhost:5432/finfin_test?sslmode=disable" && exit 1)
	go test ./internal/orders ./internal/outbox -run Test -count=1 -v

run-ttl:
	go run ./cmd/ttl-worker

run-outbox:
	go run ./cmd/outbox-publisher

smoke-seed-fast:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/seed_fast.sql

smoke-seed-reserve:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/seed_reserve.sql

smoke-reset:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/reset.sql

smoke-check-short:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/check_short.sql

smoke-check-full:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/check_full.sql

smoke-reserve:
	./scripts/smoke/reserve.sh

smoke-complete:
	./scripts/smoke/complete.sh

smoke-cancel:
	./scripts/smoke/cancel.sh

========================================
2) README.md — REPLACE WHOLE FILE WITH THIS
========================================
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
   psql "$DATABASE_URL" -f migrations/0003_account_wiring.sql
   psql "$DATABASE_URL" -f migrations/0004_pricing_engine.sql
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

- Source of truth:
  - `proto/exchange/order/v1/order_service.proto`
  - `proto/exchange/pricing/v1/pricing_service.proto`
- Generated Go stubs path:
  - `gen/exchange/order/v1`
  - `gen/exchange/pricing/v1`
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
  - full wiring + canonical `core.quotes` seed for `ReserveOrder` smoke
  - unified post-action SQL snapshots (full and operator-short variants)
- Executable SQL files are in `scripts/smoke/`:
  - `seed_demo.sql`, `seed_fast.sql`, `seed_reserve.sql`, `reset.sql`
  - `check_short.sql`, `check_full.sql`
- grpcurl wrappers are in `scripts/smoke/`:
  - `reserve.sh`, `complete.sh`, `cancel.sh`
  - shared defaults loader: `common.sh`
  - sample env file: `env.example` (copy to `.env`)
- Handy targets:
  - `make smoke-seed-pricing`
  - `make smoke-seed-demo`
  - `make smoke-bootstrap`
  - `make smoke-full-cycle`
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
- For a production-like sandbox plan and scenario pack, see `docs/production_like_sandbox_mvp.md`.

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

========================================
3) go.mod — REPLACE WHOLE FILE WITH THIS
========================================
module finfin

go 1.22

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	google.golang.org/grpc v1.76.0
)

========================================
4) internal/orders/service.go — ONLY REPLACE THIS BLOCK
========================================
Find the block inside CompleteOrder that starts with:

	if ord.Status != "reserved" {

Replace that whole block with this:

	if ord.Status != "reserved" {
		if ord.Status == "completed" || ord.Status == "cancelled" || ord.Status == "expired" {
			return CompleteOrderResult{}, ErrVersionConflict
		}
		return CompleteOrderResult{}, ErrOrderNotActive
	}

========================================
5) WHAT TO CLICK IN GITHUB
========================================
For each of the 4 files:
- remove all lines with <<<<<<<, =======, >>>>>>>
- paste the final text from this file
- click “Mark as resolved”

Then:
- click “Commit merge”
- go back to the PR
- click “Merge pull request”
- click “Confirm merge”


## Docs

- [Pricing-first execution plan](docs/pricing_first_execution_plan.md)
- [Pricing PR#1 review template](docs/pricing_pr1_review_template.md)
- Pricing transport is available via gRPC `PricingService/CalculateQuote` and Nest BFF `POST /quotes/calculate`.

## Auth foundation (PR1)

- New migration: `migrations/0006_auth_schema.sql` (`auth.users`, `auth.refresh_tokens`).
- `make migrate` and `scripts/smoke/bootstrap.sh` now apply auth schema migration.
- Copy env defaults: `cp .env.example .env`.
- Seed owner user (uses `ADMIN_BOOTSTRAP_*` vars from `.env.example`):

```bash
export $(grep -v '^#' .env.example | xargs)
./scripts/bootstrap_admin.sh
```

- Login/refresh via Nest BFF:

```bash
curl -X POST http://localhost:3000/auth/login \
  -H 'content-type: application/json' \
  -d '{"tenant_id":"11111111-1111-1111-1111-111111111111","login":"owner_demo","password":"owner_demo_password"}'

curl -X POST http://localhost:3000/auth/refresh \
  -H 'content-type: application/json' \
  -d '{"refresh_token":"<token>"}'
```

## Ops layer (PR3): audit + shifts

- New migration: `migrations/0007_audit_logs.sql` for `audit.audit_logs`.
- New BFF shift endpoints:
  - `POST /shifts/open`
  - `POST /shifts/close`
  - `GET /shifts/current`
- Cashier shift gate is enforced in BFF for `POST /orders/complete` and `POST /orders/cancel`.
- Audit events are recorded for:
  - `login`
  - `quote_calculate`
  - `reserve`
  - `complete`
  - `cancel`
  - `open_shift`
  - `close_shift`

## Pilot install

```bash
cp .env.example .env
make pilot-up
make pilot-bootstrap
```

Then run the checklist in `docs/PILOT_ACCEPTANCE_CHECKLIST.md`.

> Follow-up note: shifts are currently tracked in-memory in BFF thin slice. Before real pilot rollout, shift state checks should be backed by DB source-of-truth.
