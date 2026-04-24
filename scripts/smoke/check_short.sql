-- Operator short query (10 key columns) by order_id.
-- Replace order_id literal.
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
