# Pricing-first execution plan

## Decision

Build the **Pricing / Quote Engine first** and move `cash_shifts` immediately after.

Reasoning:

1. `Reserve -> Complete` and `Reserve -> Cancel` already work as transactional deal core.
2. `fixed_rate` and `quote_snapshot` currently act as storage, not as full product source-of-truth.
3. The next product question is revenue control: where the rate comes from and how margin is configured.

## Phase order

### Phase 1 — Pricing / Quote Engine

Target path:

`base_rates -> margin_rules -> quote calculation -> quote_snapshot -> ReserveOrder`

Deliverables:

- `core.base_rates` table (provider market rates, freshness, provenance).
- `core.margin_rules` table (tenant/office/route rules, spread model, priority).
- `CalculateQuote` application service.
- Quote TTL in payload and persistence into `core.quote_snapshots`.
- Volatility guardrails (hard limits for abrupt rate moves).
- BFF endpoint contract: `POST /quotes/calculate`.

Definition of done for Phase 1:

- Same input returns deterministic quote under same `base_rate` version.
- Quote response contains `expires_at` and `ttl_seconds`.
- `ReserveOrder` accepts only non-expired quotes and uses snapshot data as immutable source.
- Guardrail violations are observable and reject quote generation with explicit error code.

### Phase 2 — Cash Shifts / Reconciliation

Deliverables:

- `core.cash_shifts` table.
- `core.shift_reconciliations` table.
- Constraint: cashier cannot reserve/complete without an opened shift.
- Blind close + discrepancy workflow.

### Phase 3 — RBAC + audit

Deliverables:

- Role scopes for cashier/supervisor/backoffice/admin.
- Audit trail for shift/quote/order mutations with actor and timestamp.

### Phase 4 — License enforcement

Deliverables:

- Feature flags and tenant license policy checks.

### Phase 5 — AML/KYC hooks

Deliverables:

- Hook points for screening and compliance workflows before completion.

## Implementation notes for Phase 1

### Data model (minimum)

#### `core.base_rates`

- `id` (uuid pk)
- `provider` (text)
- `base_currency_code` (text)
- `quote_currency_code` (text)
- `rate` (numeric)
- `observed_at` (timestamptz)
- `received_at` (timestamptz)
- `metadata` (jsonb)

Indexes:

- `(provider, base_currency_code, quote_currency_code, observed_at desc)`

#### `core.margin_rules`

- `id` (uuid pk)
- `tenant_id` (uuid)
- `office_id` (uuid null for global)
- `side` (`buy`/`sell`)
- `base_currency_code` (text)
- `quote_currency_code` (text)
- `margin_bps` (int)
- `priority` (int)
- `active_from`, `active_to` (timestamptz)
- `enabled` (bool)

Indexes:

- `(tenant_id, office_id, side, base_currency_code, quote_currency_code, enabled, priority)`

### Quote calculation formula (v1)

- `market_rate = latest(base_rates)`
- `applied_margin_bps = first_matching(margin_rules by priority)`
- `client_rate = market_rate * (1 + sign(side) * margin_bps / 10_000)`

Where `sign(side)` is configurable by product policy; document final business convention in API spec.

### Volatility guardrails (v1)

Reject quote if either condition is met:

1. `abs(market_rate - last_used_rate) / last_used_rate > max_jump_percent`
2. `base_rate_age_seconds > max_rate_age_seconds`

Suggested defaults:

- `max_jump_percent = 3%`
- `max_rate_age_seconds = 30`

### `POST /quotes/calculate` response shape (minimum)

```json
{
  "quote_id": "q_...",
  "tenant_id": "...",
  "office_id": "...",
  "side": "buy",
  "give": { "currency_code": "USD", "amount": "1000.00" },
  "get": { "currency_code": "USDT", "amount": "998.40" },
  "fixed_rate": "0.9984",
  "margin_bps": 35,
  "market_rate": "1.0019",
  "expires_at": "2026-04-24T12:00:30Z",
  "ttl_seconds": 30,
  "guardrails": {
    "max_jump_percent": "3.0",
    "max_rate_age_seconds": 30
  }
}
```

The exact payload should be stored in `core.quote_snapshots.payload` for immutable reserve-time verification.
