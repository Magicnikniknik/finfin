# Full regression run — 2026-04-27

Requested scope: run all key scenarios, smoke flow, and available automated tests.

## 1) Go backend tests

### Command
`go test ./...`

### Result
PASS.

### Notes
- Packages with tests passing: `internal/grpc`, `internal/orders`, `internal/outbox`, `internal/pricing`.
- Cmd/gen/app packages report "[no test files]" as expected.

## 2) Smoke flow (Reserve -> Complete -> Cancel)

### Commands
- `make smoke-preflight`
- `make smoke-bootstrap`
- `make smoke-full-cycle`

### Result
FAILED in this container before business-flow execution.

### Root causes
- `grpcurl` is missing (required by smoke wrappers via `scripts/smoke/common.sh`).
- `psql` is missing (required by bootstrap SQL step).
- `docker` is missing (required for local Postgres/pilot stack).
- `DATABASE_URL` is not set.

## 3) Nest BFF automated tests

### Commands
- `npm --prefix apps/nest-bff ci`
- `npm --prefix apps/nest-bff run -s build`
- `npm --prefix apps/nest-bff run -s test`
- `npm --prefix apps/nest-bff run -s test:e2e`

### Result
- `ci`: PASS.
- `build`: PASS.
- `test`: FAIL (no unit tests matched Jest default pattern).
- `test:e2e`: FAIL.

### e2e failure summary
- `auth.e2e-spec.ts`: endpoint expectations fail on `/auth/login` and `/auth/refresh` status assertions.
- `orders.e2e-spec.ts`: multiple assertions receive 401 instead of expected success/validation/forbidden statuses.
- `pricing.e2e-spec.ts`: multiple assertions receive 401 instead of expected statuses.
- `audit.e2e-spec.ts`, `rbac.e2e-spec.ts`, `shifts.e2e-spec.ts`: PASS.

## 4) Backoffice web v2 checks

### Commands
- `npm --prefix apps/backoffice-web-v2 ci`
- `npm --prefix apps/backoffice-web-v2 run -s lint`
- `npm --prefix apps/backoffice-web-v2 run -s build`

### Result
FAILED / BLOCKED.

### Root causes
- dependency installation did not complete reliably in this environment.
- `lint` cannot resolve `@eslint/js` from local `node_modules`.
- `build` fails because `vite` binary is unavailable (not installed in local `node_modules/.bin`).

## 5) "All buttons" manual UI walkthrough status

Not executable in this container because there is no working browser runtime/tooling attached to run interactive click-through checks against running services.

## 6) Overall verdict

- **Backend Go test suite: green.**
- **End-to-end smoke/UI acceptance: not green in this container due missing runtime tools/services and frontend dependency install issues.**
- **Nest BFF e2e: partially red (3/6 suites failed).**
