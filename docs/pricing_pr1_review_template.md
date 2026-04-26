# PR review template — Pricing PR #1 (Schema + Calculator)

## Summary
Add Pricing MVP schema and calculator kernel.

## Included
- `migrations/0004_pricing_engine.sql`
- `internal/pricing/*`
- pricing tests
- `scripts/smoke/seed_pricing.sql`

## Not included
- Reserve integration
- gRPC pricing transport
- Nest BFF `/quotes/calculate`

## Goal
Keep PR #1 focused on one question:
Can the core pricing module calculate and persist quotes correctly?

## Reviewer focus
- schema
- rule selection
- pricing math
- rounding
- quote persistence

## Recommended review order
1. `migrations/0004_pricing_engine.sql`
2. `internal/pricing/types.go`
3. `internal/pricing/rule_selector.go`
4. `internal/pricing/rounding.go`
5. `internal/pricing/calculator.go`
6. `internal/pricing/service.go`
7. tests
