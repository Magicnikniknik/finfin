# SQL seed checklist for manual gRPC smoke

This document contains two ready-to-run seed tracks:

1. **Fast path for `CompleteOrder` / `CancelOrder` smoke** (without full `ReserveOrder` flow).
2. **Full path for `ReserveOrder` smoke** (with `account_wiring` + `quote_snapshots`).

> Important: these are manual smoke seeds, not a replacement for full integration tests.

---

## A) Fast seed for `CompleteOrder` / `CancelOrder`

Creates:
- fixed tenant/office
- accounts and projection balance state
- 2 reserved orders + active holds:
  - `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa` (for complete)
  - `bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb` (for cancel)

```sql
BEGIN;

INSERT INTO core.accounts (id, tenant_id, office_id, currency_id, account_type) VALUES
  ('55555555-5555-5555-5555-555555555555','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','balance'),
  ('66666666-6666-6666-6666-666666666666','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','available_ledger'),
  ('77777777-7777-7777-7777-777777777777','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','reserved_ledger'),
  ('88888888-8888-8888-8888-888888888888','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','settlement')
ON CONFLICT DO NOTHING;

INSERT INTO core.account_balances (account_id, tenant_id, currency_id, available, reserved, updated_at)
VALUES ('55555555-5555-5555-5555-555555555555','11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333',800,200,now())
ON CONFLICT (account_id) DO UPDATE
SET available = EXCLUDED.available, reserved = EXCLUDED.reserved, updated_at = now();

INSERT INTO core.orders (
  id, tenant_id, office_id, client_ref, side, give_currency_id, get_currency_id,
  amount_give, amount_get, fixed_rate, quote_payload, status, reserved_at, expires_at,
  version, order_ref, created_at, updated_at
) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','client_complete_smoke','buy','33333333-3333-3333-3333-333333333333','44444444-4444-4444-4444-444444444444',100,3550,35.5,'{"source":"manual_seed","purpose":"complete_smoke"}'::jsonb,'reserved',now(),now()+interval '30 minutes',1,'ORD-COMPLETE-1',now(),now()),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','client_cancel_smoke','buy','33333333-3333-3333-3333-333333333333','44444444-4444-4444-4444-444444444444',100,3550,35.5,'{"source":"manual_seed","purpose":"cancel_smoke"}'::jsonb,'reserved',now(),now()+interval '30 minutes',1,'ORD-CANCEL-1',now(),now())
ON CONFLICT (id) DO NOTHING;

INSERT INTO core.order_holds (
  id, tenant_id, order_id, balance_account_id, available_ledger_account_id, reserved_ledger_account_id,
  settlement_ledger_account_id, currency_id, amount, status, expires_at, created_at
) VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc','11111111-1111-1111-1111-111111111111','aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','55555555-5555-5555-5555-555555555555','66666666-6666-6666-6666-666666666666','77777777-7777-7777-7777-777777777777','88888888-8888-8888-8888-888888888888','33333333-3333-3333-3333-333333333333',100,'active',now()+interval '30 minutes',now()),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd','11111111-1111-1111-1111-111111111111','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','55555555-5555-5555-5555-555555555555','66666666-6666-6666-6666-666666666666','77777777-7777-7777-7777-777777777777','88888888-8888-8888-8888-888888888888','33333333-3333-3333-3333-333333333333',100,'active',now()+interval '30 minutes',now())
ON CONFLICT (id) DO NOTHING;

COMMIT;
```

Quick check:

```sql
SELECT id, order_ref, status, version
FROM core.orders
WHERE id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb')
ORDER BY id;

SELECT order_id, status, amount
FROM core.order_holds
WHERE order_id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb')
ORDER BY order_id;

SELECT account_id, available, reserved
FROM core.account_balances
WHERE account_id = '55555555-5555-5555-5555-555555555555';
```

Expected:
- both orders `reserved`, `version=1`
- both holds `active`
- `available=800`, `reserved=200`

---

## B) Full seed for `ReserveOrder` smoke via gRPC

Creates:
- accounts
- account balance projection (`available=1000`, `reserved=0`)
- account wiring
- fresh quote snapshot (`quote-reserve-smoke-001`)

```sql
BEGIN;

INSERT INTO core.accounts (id, tenant_id, office_id, currency_id, account_type) VALUES
  ('55555555-5555-5555-5555-555555555555','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','balance'),
  ('66666666-6666-6666-6666-666666666666','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','available_ledger'),
  ('77777777-7777-7777-7777-777777777777','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','reserved_ledger'),
  ('88888888-8888-8888-8888-888888888888','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','33333333-3333-3333-3333-333333333333','settlement')
ON CONFLICT DO NOTHING;

INSERT INTO core.account_balances (account_id, tenant_id, currency_id, available, reserved, updated_at)
VALUES ('55555555-5555-5555-5555-555555555555','11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333',1000,0,now())
ON CONFLICT (account_id) DO UPDATE
SET available = EXCLUDED.available, reserved = EXCLUDED.reserved, updated_at = now();

INSERT INTO core.account_wiring (
  tenant_id, office_id, currency_id, balance_account_id,
  available_ledger_account_id, reserved_ledger_account_id, settlement_ledger_account_id,
  created_at, updated_at
) VALUES (
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  '33333333-3333-3333-3333-333333333333',
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666666',
  '77777777-7777-7777-7777-777777777777',
  '88888888-8888-8888-8888-888888888888',
  now(), now()
)
ON CONFLICT DO NOTHING;

INSERT INTO core.quote_snapshots (
  id, tenant_id, office_id, side, expires_at,
  give_currency_id, give_currency_code, give_currency_network, amount_give,
  get_currency_id, get_currency_code, get_currency_network, amount_get,
  fixed_rate, hold_currency_id, hold_amount, payload, created_at
) VALUES (
  'quote-reserve-smoke-001',
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  'buy',
  now() + interval '30 minutes',
  '33333333-3333-3333-3333-333333333333','USDT','TRC20',100,
  '44444444-4444-4444-4444-444444444444','THB','cash',3550,
  35.5,
  '33333333-3333-3333-3333-333333333333',
  100,
  '{"source":"manual_seed","purpose":"grpc_reserve_smoke","label":"USDT_TRC20_to_THB_cash"}'::jsonb,
  now()
)
ON CONFLICT (id) DO UPDATE
SET
  expires_at = EXCLUDED.expires_at,
  amount_give = EXCLUDED.amount_give,
  amount_get = EXCLUDED.amount_get,
  fixed_rate = EXCLUDED.fixed_rate,
  hold_amount = EXCLUDED.hold_amount,
  payload = EXCLUDED.payload;

COMMIT;
```

Quick check:

```sql
SELECT id, side, amount_give, amount_get, fixed_rate, expires_at
FROM core.quote_snapshots
WHERE id = 'quote-reserve-smoke-001';

SELECT tenant_id, office_id, currency_id, balance_account_id
FROM core.account_wiring
WHERE tenant_id = '11111111-1111-1111-1111-111111111111'
  AND office_id = '22222222-2222-2222-2222-222222222222';

SELECT account_id, available, reserved
FROM core.account_balances
WHERE account_id = '55555555-5555-5555-5555-555555555555';
```

Expected:
- quote exists and not expired
- wiring exists
- balance projection = `available=1000`, `reserved=0`

---

## C) grpcurl examples

Reserve:

```bash
grpcurl -plaintext \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_reserve_smoke_001' \
  -d '{
    "idempotency_key": "99999999-9999-9999-9999-999999999999",
    "office_id": "22222222-2222-2222-2222-222222222222",
    "quote_id": "quote-reserve-smoke-001",
    "side": "BUY",
    "give": {"amount":"100.00","currency":{"code":"USDT","network":"TRC20"}},
    "get": {"amount":"3550.00","currency":{"code":"THB","network":"cash"}}
  }' \
  localhost:9090 \
  exchange.order.v1.OrderService/ReserveOrder
```

Complete:

```bash
grpcurl -plaintext \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_smoke_001' \
  -d '{
    "idempotency_key": "99999999-9999-9999-9999-999999999991",
    "order_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "expected_version": 1,
    "cashier_id": "cashier_01"
  }' \
  localhost:9090 \
  exchange.order.v1.OrderService/CompleteOrder
```

Cancel:

```bash
grpcurl -plaintext \
  -H 'x-tenant-id: 11111111-1111-1111-1111-111111111111' \
  -H 'x-client-ref: client_smoke_001' \
  -d '{
    "idempotency_key": "99999999-9999-9999-9999-999999999992",
    "order_id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
    "expected_version": 1,
    "reason": "client_no_show"
  }' \
  localhost:9090 \
  exchange.order.v1.OrderService/CancelOrder
```

---

## D) Cleanup (seed rollback)

```sql
BEGIN;

DELETE FROM core.outbox_events
WHERE aggregate_id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

DELETE FROM core.ledger_entries
WHERE journal_id IN (
  SELECT id FROM core.ledger_journals
  WHERE order_id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb')
);

DELETE FROM core.ledger_journals
WHERE order_id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

DELETE FROM core.order_holds
WHERE order_id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

DELETE FROM core.orders
WHERE id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

DELETE FROM core.quote_snapshots
WHERE id = 'quote-reserve-smoke-001';

DELETE FROM core.account_wiring
WHERE tenant_id = '11111111-1111-1111-1111-111111111111'
  AND office_id = '22222222-2222-2222-2222-222222222222'
  AND currency_id = '33333333-3333-3333-3333-333333333333';

DELETE FROM core.account_balances
WHERE account_id = '55555555-5555-5555-5555-555555555555';

DELETE FROM core.accounts
WHERE id IN (
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666666',
  '77777777-7777-7777-7777-777777777777',
  '88888888-8888-8888-8888-888888888888'
);

COMMIT;
```

---

## E) Unified post-action SQL snapshot (full)

Use this after `ReserveOrder` / `CompleteOrder` / `CancelOrder` to inspect in one row:
- order
- hold
- balances
- latest ledger journal
- latest outbox event

### Variant by `order_id`

```sql
WITH input AS (
  SELECT 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS order_id
),
ord AS (
  SELECT
    o.id,
    o.order_ref,
    o.tenant_id,
    o.office_id,
    o.client_ref,
    o.side,
    o.status,
    o.version,
    o.amount_give::text AS amount_give,
    o.amount_get::text AS amount_get,
    o.fixed_rate::text AS fixed_rate,
    o.reserved_at,
    o.expires_at,
    o.completed_at,
    o.cancelled_at,
    o.cancel_reason,
    o.created_at,
    o.updated_at
  FROM core.orders o
  JOIN input i ON i.order_id = o.id
),
hold AS (
  SELECT
    h.id AS hold_id,
    h.order_id,
    h.status AS hold_status,
    h.currency_id,
    h.amount::text AS hold_amount,
    h.balance_account_id,
    h.available_ledger_account_id,
    h.reserved_ledger_account_id,
    h.settlement_ledger_account_id,
    h.expires_at AS hold_expires_at,
    h.consumed_at,
    h.released_at,
    h.created_at AS hold_created_at
  FROM core.order_holds h
  JOIN input i ON i.order_id = h.order_id
),
bal AS (
  SELECT
    b.account_id,
    b.available::text AS available,
    b.reserved::text AS reserved,
    b.updated_at AS balances_updated_at
  FROM core.account_balances b
  JOIN hold h ON h.balance_account_id = b.account_id
),
last_journal AS (
  SELECT
    lj.id AS journal_id,
    lj.kind AS journal_kind,
    lj.note AS journal_note,
    lj.created_by AS journal_created_by,
    lj.created_at AS journal_created_at
  FROM core.ledger_journals lj
  JOIN input i ON i.order_id = lj.order_id
  ORDER BY lj.created_at DESC, lj.id DESC
  LIMIT 1
),
last_outbox AS (
  SELECT
    oe.id AS outbox_id,
    oe.event_type,
    oe.status AS outbox_status,
    oe.attempts,
    oe.available_at,
    oe.published_at,
    oe.last_error,
    oe.created_at AS outbox_created_at,
    oe.updated_at AS outbox_updated_at
  FROM core.outbox_events oe
  JOIN input i ON i.order_id = oe.aggregate_id
  ORDER BY oe.created_at DESC, oe.id DESC
  LIMIT 1
)
SELECT
  ord.id AS order_id,
  ord.order_ref,
  ord.tenant_id,
  ord.office_id,
  ord.client_ref,
  ord.side,
  ord.status AS order_status,
  ord.version,
  ord.amount_give,
  ord.amount_get,
  ord.fixed_rate,
  ord.reserved_at,
  ord.expires_at,
  ord.completed_at,
  ord.cancelled_at,
  ord.cancel_reason,
  ord.created_at AS order_created_at,
  ord.updated_at AS order_updated_at,

  hold.hold_id,
  hold.hold_status,
  hold.currency_id AS hold_currency_id,
  hold.hold_amount,
  hold.balance_account_id,
  hold.available_ledger_account_id,
  hold.reserved_ledger_account_id,
  hold.settlement_ledger_account_id,
  hold.hold_expires_at,
  hold.consumed_at,
  hold.released_at,
  hold.hold_created_at,

  bal.account_id AS balance_account_id_check,
  bal.available,
  bal.reserved,
  bal.balances_updated_at,

  last_journal.journal_id,
  last_journal.journal_kind,
  last_journal.journal_note,
  last_journal.journal_created_by,
  last_journal.journal_created_at,

  last_outbox.outbox_id,
  last_outbox.event_type AS last_event_type,
  last_outbox.outbox_status,
  last_outbox.attempts,
  last_outbox.available_at,
  last_outbox.published_at,
  last_outbox.last_error,
  last_outbox.outbox_created_at,
  last_outbox.outbox_updated_at
FROM ord
LEFT JOIN hold ON hold.order_id = ord.id
LEFT JOIN bal ON true
LEFT JOIN last_journal ON true
LEFT JOIN last_outbox ON true;
```

### Variant by `order_ref`

```sql
WITH input AS (
  SELECT 'ORD-COMPLETE-1'::text AS order_ref
),
ord AS (
  SELECT
    o.id,
    o.order_ref,
    o.tenant_id,
    o.office_id,
    o.client_ref,
    o.side,
    o.status,
    o.version,
    o.amount_give::text AS amount_give,
    o.amount_get::text AS amount_get,
    o.fixed_rate::text AS fixed_rate,
    o.reserved_at,
    o.expires_at,
    o.completed_at,
    o.cancelled_at,
    o.cancel_reason,
    o.created_at,
    o.updated_at
  FROM core.orders o
  JOIN input i ON i.order_ref = o.order_ref
),
hold AS (
  SELECT
    h.id AS hold_id,
    h.order_id,
    h.status AS hold_status,
    h.currency_id,
    h.amount::text AS hold_amount,
    h.balance_account_id,
    h.available_ledger_account_id,
    h.reserved_ledger_account_id,
    h.settlement_ledger_account_id,
    h.expires_at AS hold_expires_at,
    h.consumed_at,
    h.released_at,
    h.created_at AS hold_created_at
  FROM core.order_holds h
  JOIN ord o ON o.id = h.order_id
),
bal AS (
  SELECT
    b.account_id,
    b.available::text AS available,
    b.reserved::text AS reserved,
    b.updated_at AS balances_updated_at
  FROM core.account_balances b
  JOIN hold h ON h.balance_account_id = b.account_id
),
last_journal AS (
  SELECT
    lj.id AS journal_id,
    lj.kind AS journal_kind,
    lj.note AS journal_note,
    lj.created_by AS journal_created_by,
    lj.created_at AS journal_created_at
  FROM core.ledger_journals lj
  JOIN ord o ON o.id = lj.order_id
  ORDER BY lj.created_at DESC, lj.id DESC
  LIMIT 1
),
last_outbox AS (
  SELECT
    oe.id AS outbox_id,
    oe.event_type,
    oe.status AS outbox_status,
    oe.attempts,
    oe.available_at,
    oe.published_at,
    oe.last_error,
    oe.created_at AS outbox_created_at,
    oe.updated_at AS outbox_updated_at
  FROM core.outbox_events oe
  JOIN ord o ON o.id = oe.aggregate_id
  ORDER BY oe.created_at DESC, oe.id DESC
  LIMIT 1
)
SELECT
  ord.id AS order_id,
  ord.order_ref,
  ord.status AS order_status,
  ord.version,
  ord.amount_give,
  ord.amount_get,
  ord.fixed_rate,
  ord.reserved_at,
  ord.expires_at,
  hold.hold_id,
  hold.hold_status,
  hold.hold_amount,
  bal.available,
  bal.reserved,
  last_journal.journal_kind,
  last_journal.journal_created_at,
  last_outbox.event_type AS last_event_type,
  last_outbox.outbox_status,
  last_outbox.attempts,
  last_outbox.published_at,
  last_outbox.last_error
FROM ord
LEFT JOIN hold ON hold.order_id = ord.id
LEFT JOIN bal ON true
LEFT JOIN last_journal ON true
LEFT JOIN last_outbox ON true;
```

Expected trends:
- after reserve: `order_status=reserved`, `hold_status=active`, `journal_kind=hold_create`, `last_event_type=order_reserved`
- after complete: `order_status=completed`, `hold_status=consumed`, `journal_kind=trade_complete`, `last_event_type=order_completed`
- after cancel: `order_status=cancelled`, `hold_status=released`, `journal_kind=hold_release`, `last_event_type=order_cancelled`

---

## F) Operator short query (10-key columns)

### Variant by `order_id`

```sql
WITH input AS (
  SELECT 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid AS order_id
),
o AS (
  SELECT id, status, version
  FROM core.orders
  WHERE id = (SELECT order_id FROM input)
),
h AS (
  SELECT order_id, status AS hold_status, balance_account_id
  FROM core.order_holds
  WHERE order_id = (SELECT order_id FROM input)
),
b AS (
  SELECT available::text AS available, reserved::text AS reserved
  FROM core.account_balances
  WHERE account_id = (SELECT balance_account_id FROM h)
),
j AS (
  SELECT kind AS last_journal_kind
  FROM core.ledger_journals
  WHERE order_id = (SELECT order_id FROM input)
  ORDER BY created_at DESC, id DESC
  LIMIT 1
),
e AS (
  SELECT event_type AS last_event_type, status AS outbox_status, attempts
  FROM core.outbox_events
  WHERE aggregate_id = (SELECT order_id FROM input)
  ORDER BY created_at DESC, id DESC
  LIMIT 1
)
SELECT
  o.id AS order_id,
  o.status AS order_status,
  o.version,
  h.hold_status,
  b.available,
  b.reserved,
  j.last_journal_kind,
  e.last_event_type,
  e.outbox_status,
  e.attempts
FROM o
LEFT JOIN h ON h.order_id = o.id
LEFT JOIN b ON true
LEFT JOIN j ON true
LEFT JOIN e ON true;
```

### Variant by `order_ref`

```sql
WITH input AS (
  SELECT 'ORD-COMPLETE-1'::text AS order_ref
),
o AS (
  SELECT id, order_ref, status, version
  FROM core.orders
  WHERE order_ref = (SELECT order_ref FROM input)
),
h AS (
  SELECT order_id, status AS hold_status, balance_account_id
  FROM core.order_holds
  WHERE order_id = (SELECT id FROM o)
),
b AS (
  SELECT available::text AS available, reserved::text AS reserved
  FROM core.account_balances
  WHERE account_id = (SELECT balance_account_id FROM h)
),
j AS (
  SELECT kind AS last_journal_kind
  FROM core.ledger_journals
  WHERE order_id = (SELECT id FROM o)
  ORDER BY created_at DESC, id DESC
  LIMIT 1
),
e AS (
  SELECT event_type AS last_event_type, status AS outbox_status, attempts
  FROM core.outbox_events
  WHERE aggregate_id = (SELECT id FROM o)
  ORDER BY created_at DESC, id DESC
  LIMIT 1
)
SELECT
  o.order_ref,
  o.status AS order_status,
  o.version,
  h.hold_status,
  b.available,
  b.reserved,
  j.last_journal_kind,
  e.last_event_type,
  e.outbox_status,
  e.attempts
FROM o
LEFT JOIN h ON h.order_id = o.id
LEFT JOIN b ON true
LEFT JOIN j ON true
LEFT JOIN e ON true;
```

Expected readings:
- reserve: `reserved`, `v1`, `active`, reserved up, `hold_create`, `order_reserved`
- complete: `completed`, `v2`, `consumed`, reserved down, `trade_complete`, `order_completed`
- cancel: `cancelled`, `v2`, `released`, available up, `hold_release`, `order_cancelled`

---

## Caveat

The fast seed path intentionally bypasses reserve-phase accounting side effects (`hold_create` journal and `order_reserved` outbox event).  
This is acceptable for manual `CompleteOrder` / `CancelOrder` smoke, but it is **not** equivalent to full reserve-flow validation.
