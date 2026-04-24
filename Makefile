.PHONY: db-up db-down migrate proto test test-integration run-ttl run-outbox smoke-seed-fast smoke-seed-reserve smoke-reset smoke-check-short smoke-check-full smoke-reserve smoke-complete smoke-cancel

db-up:
	docker compose up -d postgres

db-down:
	docker compose down -v

migrate:
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f migrations/0001_core_schema.sql
	psql "$$DATABASE_URL" -f migrations/0002_quote_snapshots.sql

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
