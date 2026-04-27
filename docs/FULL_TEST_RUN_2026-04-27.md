# Full test run snapshot — 2026-04-27

## Goal
Validate correctness and execution readiness across backend calculations, transport/services, smoke flow, and web layers.

## Commands executed

1) Backend Go suite
- `go test ./...`

Result: **PASS**.
- `internal/grpc`: PASS
- `internal/orders`: PASS
- `internal/outbox`: PASS
- `internal/pricing`: PASS

2) Nest BFF checks
- `npm --prefix apps/nest-bff ci` -> PASS
- `npm --prefix apps/nest-bff run -s build` -> PASS
- `npm --prefix apps/nest-bff run -s test` -> PASS (`--passWithNoTests`, no unit tests matched)
- `npm --prefix apps/nest-bff run -s test:e2e` -> PASS (6/6 suites, 29/29 tests)

3) Smoke checks
- `make smoke-preflight` -> FAIL (missing `docker`, `psql`, `grpcurl`, `DATABASE_URL` unset)
- `make smoke-full-cycle` -> FAIL (`grpcurl` missing)

4) Frontend checks
- `node --check apps/backoffice-web/app.js` -> PASS
- `node --check apps/backoffice-web/design.js` -> PASS
- `npm --prefix apps/backoffice-web-v2 run -s lint` -> PASS with warnings (0 errors)

## Interpretation
- Core money/order calculation paths validated by Go suite are green in this environment.
- Nest BFF build/e2e path is green; unit test command is green with `--passWithNoTests` until a dedicated unit test set is added.
- End-to-end smoke is still blocked by missing runtime tooling in this container.
- UI script syntax for backoffice-web is valid; backoffice-web-v2 lint is now runnable and reports warnings only.

## Immediate priority to reach fully green
1. Install missing smoke runtime tools: `docker`, `psql`, `grpcurl`.
2. Provide `DATABASE_URL` and run smoke flow against live stack.
3. Decide whether to keep `npm test` failing on no matches or switch to `--passWithNoTests` / explicit unit pattern.
4. Reduce current lint warning volume in `backoffice-web-v2` (non-blocking).
