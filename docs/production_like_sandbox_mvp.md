# Production-like Sandbox MVP

This document defines a realistic demo environment (without real money) on top of the existing backend.

## Goal

Show a complete and understandable loop in 3-5 minutes:

1. `CalculateQuote` (Give / Get)
2. active quote with TTL
3. `ReserveOrder` from canonical `core.quotes`
4. `CompleteOrder` or `CancelOrder`
5. visible safe failures (`stale`, `expired`, `insufficient liquidity`)

## Seed pack (SQL)

Use:

```bash
make smoke-seed-demo
```

Seed file: `scripts/smoke/seed_demo.sql`.

### Entities (demo naming)

- Tenant: `11111111-1111-1111-1111-111111111111`
- Offices:
  - `22222222-2222-2222-2222-222222222222` = Bangkok Main
  - `33333333-3333-3333-3333-333333333333` = Pattaya Branch
  - `44444444-4444-4444-4444-444444444444` = Airport Desk
- Currencies:
  - `90000000-0000-0000-0000-000000000001` = USDT TRC20
  - `90000000-0000-0000-0000-000000000002` = THB cash
  - `90000000-0000-0000-0000-000000000003` = USD cash
  - `90000000-0000-0000-0000-000000000004` = BTC mainnet

## Demo scenarios

### 1) Buy 100 USDT (retail happy path)

- Quote id: `q_demo_buy_100_usdt_bkk`
- Status: `active`
- Flow: reserve -> complete

### 2) Get 10,000 THB (reverse calculation)

- Quote id: `q_demo_get_10000_thb_bkk`
- Status: `active`
- Flow: reserve -> cancel

### 3) VIP tier

- Rule set includes a high-volume better tier (`margin_bps` lower for volume >= 5000).
- Compare rates for small vs large notional on the same pair.

### 4) Stale rate

- Trigger by forcing `core.base_rates.updated_at` older than pricing threshold.
- Expected: `CalculateQuote` fails with stale-rate error.

### 5) Error resilience

- Expired quote id: `q_demo_expired_usdt_thb` (status `expired`).
- Consumed quote id: `q_demo_consumed_usd_thb` (status `consumed`).
- Insufficient liquidity path is simulated via low Pattaya USDT balance.

## Suggested run order

1. `make smoke-bootstrap`
2. `make smoke-seed-demo`
3. Run gRPC server + BFF
4. Execute scenarios from UI / grpcurl
5. Inspect state with:
   - `make smoke-check-short`
   - `make smoke-check-full`

## Naming

Use these terms externally:

- **Production-like Sandbox**
- **Realistic Money Simulation**
- **Operator Demo Environment**

Avoid saying “real money mode” for this stage.
