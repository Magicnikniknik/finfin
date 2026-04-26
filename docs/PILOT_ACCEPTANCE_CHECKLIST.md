# Pilot Acceptance Checklist

## Boot
- [ ] `make pilot-up` succeeds
- [ ] `make pilot-bootstrap` succeeds
- [ ] `GET /healthz` returns ok
- [ ] `GET /readyz` returns ready

## Auth
- [ ] owner login works
- [ ] cashier login works
- [ ] refresh token works
- [ ] disabled user is rejected

## Pricing
- [ ] quote calculation works for happy path
- [ ] stale rate scenario is rejected correctly
- [ ] expired quote scenario is visible

## Orders
- [ ] reserve works
- [ ] complete works
- [ ] cancel works
- [ ] OCC/version conflict is enforced

## RBAC / office scope
- [ ] cashier can act in own office
- [ ] cashier is blocked in foreign office
- [ ] manager can cancel
- [ ] owner can perform all actions

## Shifts
- [ ] cashier can open shift
- [ ] cashier can close shift
- [ ] complete is blocked without active shift
- [ ] cancel is blocked without active shift

## Audit
- [ ] login writes audit row
- [ ] quote writes audit row
- [ ] reserve writes audit row
- [ ] complete/cancel write audit row

## Persistence
- [ ] restart services
- [ ] data still present after restart
