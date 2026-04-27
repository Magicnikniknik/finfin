# Operational Smoke Checklist

This document is the reproducible smoke runbook for local and CI environments.

## 1) Prepare env file

```bash
cp scripts/smoke/env.smoke.example .env.smoke
set -a
source .env.smoke
set +a
```

## 2) Install required tools

Required binaries:
- docker
- psql
- grpcurl
- go
- curl

## 3) Startup order

```bash
# infra
make run-services

# wait for DB
make smoke-preflight-db

# migrations + seed
make migrate-up
make seed-smoke

# endpoint/tooling checks
make smoke-preflight

# full happy-path scenario
make smoke-full-cycle
```

## 4) Smoke coverage

`make smoke-full-cycle` executes:
1. Reserve order
2. Complete order
3. Cancel order

Use SQL checks for post-state validation:
- `make smoke-check-short`
- `make smoke-check-full`

## 5) CI-compatible gate

```bash
go test ./...
npm --prefix apps/nest-bff run -s build
npm --prefix apps/nest-bff run -s test
npm --prefix apps/nest-bff run -s test:e2e
npm --prefix apps/backoffice-web-v2 run -s lint
make smoke-preflight
make smoke-full-cycle-assertive
```

For local "first real green" run with infra, use one command:

```bash
make smoke-first-green
```

`smoke-first-green` fails fast if `DATABASE_URL`, `HTTP_BASE_URL`, or `GRPC_ADDR` is missing, then runs:
1. `make run-services`
2. `make smoke-preflight-db`
3. `make migrate-up`
4. `make seed-smoke`
5. `make smoke-preflight`
6. `make smoke-full-cycle-assertive`

GitHub Actions workflow:
- `.github/workflows/smoke-full-cycle.yml`

## Smoke full-cycle troubleshooting

### `docker: command not found`

Docker is not available on the runner/machine.

Fix:

```bash
docker --version
```

Use a runner with Docker support (or enable Docker service).

### `psql: command not found`

PostgreSQL client is missing.

Fix:

```bash
# macOS
brew install postgresql@16

# Ubuntu
sudo apt-get update
sudo apt-get install -y postgresql-client
```

### `grpcurl: command not found`

gRPC smoke dependency is missing.

Fix:

```bash
grpcurl --version
```

Install via Homebrew or release artifact (same way as CI workflow).

### `DATABASE_URL is required`

Smoke environment is not loaded.

Fix:

```bash
cp scripts/smoke/env.smoke.example .env.smoke
source .env.smoke
echo "$DATABASE_URL"
```

### `database is not reachable`

Postgres is not ready/running or URL points to wrong host/port.

Fix:

```bash
make smoke-preflight-db
psql "$DATABASE_URL" -c "select 1;"
```

### `HTTP endpoint is not reachable`

HTTP service is not running or `HTTP_BASE_URL` is incorrect.

Fix:

```bash
make run-services
curl -i "$HTTP_BASE_URL/healthz"
```

### `gRPC endpoint is not reachable`

gRPC service is not running or `GRPC_ADDR` is incorrect.

Fix:

```bash
grpcurl -plaintext "$GRPC_ADDR" list
```

### `smoke-full-cycle` fails after reserve/complete

Service is up but scenario does not match seeded data or current env values.

Fix:

```bash
make seed-smoke
make smoke-full-cycle
```

Also verify `TENANT_ID`, `CASHIER_ID`, and `CLIENT_REF` in `.env.smoke`.
