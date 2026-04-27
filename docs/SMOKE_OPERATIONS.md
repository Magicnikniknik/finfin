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
make smoke-full-cycle
```
