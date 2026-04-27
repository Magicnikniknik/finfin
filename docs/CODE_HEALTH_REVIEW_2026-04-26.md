# Code Health Review — 2026-04-26

## Scope
- Quick operational review of pricing/order calculation paths and backoffice-web-v2 UI composition.
- Execution of local checks available in this environment.

## Findings

### 1) Pricing calculation logic (code review)
- `BuildQuote` validates required inputs (`quoteID`, positive `ttlMinutes`, valid `InputMode`, positive amount).
- Side-dependent base rate choice is consistent:
  - BUY uses `ask` + positive margin bps.
  - SELL uses `bid` - margin bps.
- Fee application in GIVE/GET modes is symmetric:
  - SELL/GIVE: `get = give * rate - fee`
  - SELL/GET: `give = (get + fee) / rate`
  - BUY/GIVE: `get = (give - fee) / rate`
  - BUY/GET: `give = get * rate + fee`
- Guardrails exist for non-positive client rate and non-positive resulting amounts.
- Final quote fields include snapshots of applied rule/base rate, and expiry uses UTC `now + ttl`.

### 2) Design quality (UI structure review)
- App layout in `backoffice-web-v2` is modular and dashboard-oriented: KPI/config/demo/scenario controls on top, operational split-pane workspace, and walkthrough footer.
- Styling choices in `index.css` are coherent for dark operator UI (near-black base, subtle contrast, shared glass utility classes).
- Overall design is strong for demo/operations usage: clear hierarchy and grouped workflows.
- Minor UX risk: high density in right column (`OrderOperations`, `OrderSummary`, `StateTimeline`, `AuditPanel`, optional `DebugPanel`) can feel crowded on smaller screens.

### 3) Runtime verification status
- Full runtime verification is **blocked in this environment** due dependency download restrictions (Go modules from `proxy.golang.org` return `403 Forbidden`).
- Frontend lint run is also blocked because npm dependencies are not installed locally (`@eslint/js` missing).

## Commands run
- `go test ./internal/pricing/...`
- `go test ./internal/orders/...`
- `npm --prefix apps/backoffice-web-v2 run -s lint`

## Conclusion
- By code inspection, calculation formulas and validation flow look internally consistent.
- However, because tests could not execute in this environment, calculations are **not fully runtime-confirmed here**.

## Why smoke/tests failed in this container
- This container does not have required runtime tools installed (`docker`, `psql`, `grpcurl`), while smoke scripts explicitly require them.
- `go test` cannot complete because module downloads for `pgx` transitive dependencies are blocked from `proxy.golang.org` (`403 Forbidden`).
- So failures are infrastructure/environment-related, not evidence that quote/reserve/complete/cancel business logic regressed.

## How to get back to "works well" locally/CI
1. Install prerequisites: Docker Engine, `psql`, `grpcurl`.
2. Export `DATABASE_URL` (and optionally `DATABASE_URL_TEST`) to a running Postgres.
3. Validate with `make smoke-preflight`.
4. Run `make smoke-bootstrap` and then `make smoke-full-cycle`.
5. If Go proxy remains blocked, use an internal mirror or allowlist required module sources.
