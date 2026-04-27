.PHONY: db-up db-down migrate migrate-up proto test test-integration run-ttl run-outbox smoke-preflight smoke-preflight-db smoke-bootstrap seed-smoke run-services smoke-full-cycle smoke-full-cycle-assertive smoke-first-green smoke-seed-pricing smoke-seed-demo smoke-seed-fast smoke-seed-reserve smoke-reset smoke-check-short smoke-check-full smoke-reserve smoke-complete smoke-cancel pilot-up pilot-down pilot-bootstrap pilot-smoke

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
	psql "$$DATABASE_URL" -f migrations/0005_cash_shifts.sql
	psql "$$DATABASE_URL" -f migrations/0006_auth_schema.sql
	psql "$$DATABASE_URL" -f migrations/0007_audit_logs.sql

migrate-up: migrate

proto:
	@command -v protoc >/dev/null 2>&1 || (echo "protoc is required" && exit 1)
	@command -v protoc-gen-go >/dev/null 2>&1 || (echo "protoc-gen-go is required: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" && exit 1)
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || (echo "protoc-gen-go-grpc is required: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" && exit 1)
	protoc \
		-I proto \
		--go_out=gen --go_opt=paths=source_relative \
		--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
		proto/exchange/order/v1/order_service.proto \
		proto/exchange/pricing/v1/pricing_service.proto

test:
	go test ./...

test-integration:
	@test -n "$$DATABASE_URL_TEST" || (echo "DATABASE_URL_TEST is required. Run: export DATABASE_URL_TEST=postgres://postgres:postgres@localhost:5432/finfin_test?sslmode=disable" && exit 1)
	go test ./internal/orders ./internal/outbox -run Test -count=1 -v

run-ttl:
	go run ./cmd/ttl-worker

run-outbox:
	go run ./cmd/outbox-publisher


smoke-preflight:
	./scripts/smoke/preflight.sh

smoke-preflight-db:
	./scripts/smoke/preflight_db.sh

smoke-bootstrap:
	./scripts/smoke/bootstrap.sh

seed-smoke:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/seed_pricing.sql
	psql "$$DATABASE_URL" -f scripts/smoke/seed_demo.sql
	psql "$$DATABASE_URL" -f scripts/smoke/seed_reserve.sql
	psql "$$DATABASE_URL" -f scripts/smoke/seed_fast.sql

run-services:
	docker compose -f docker-compose.pilot.yml --env-file .env.smoke up -d postgres grpc-server nest-bff

smoke-full-cycle:
	./scripts/smoke/full_cycle.sh

smoke-full-cycle-assertive:
	./scripts/smoke/full_cycle_assertive.sh

smoke-first-green:
	@set -e; \
	: "$${DATABASE_URL:?DATABASE_URL is required}"; \
	: "$${HTTP_BASE_URL:?HTTP_BASE_URL is required}"; \
	: "$${GRPC_ADDR:?GRPC_ADDR is required}"; \
	echo "== smoke-first-green: run-services -> preflight-db -> migrate -> seed -> preflight -> assertive"; \
	$(MAKE) run-services; \
	$(MAKE) smoke-preflight-db; \
	$(MAKE) migrate-up; \
	$(MAKE) seed-smoke; \
	$(MAKE) smoke-preflight; \
	$(MAKE) smoke-full-cycle-assertive

smoke-seed-pricing:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/seed_pricing.sql

smoke-seed-demo:
	@command -v psql >/dev/null 2>&1 || (echo "psql is required to run smoke SQL scripts" && exit 1)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f scripts/smoke/seed_demo.sql

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


pilot-up:
	docker compose -f docker-compose.pilot.yml --env-file .env up -d --build

pilot-down:
	docker compose -f docker-compose.pilot.yml --env-file .env down -v

pilot-bootstrap:
	bash scripts/bootstrap_pilot.sh

pilot-smoke:
	@echo "Run docs/PILOT_ACCEPTANCE_CHECKLIST.md"
