.PHONY: migrate test run-ttl run-outbox

migrate:
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	psql "$$DATABASE_URL" -f migrations/0001_core_schema.sql

test:
	go test ./...

run-ttl:
	go run ./cmd/ttl-worker

run-outbox:
	go run ./cmd/outbox-publisher
