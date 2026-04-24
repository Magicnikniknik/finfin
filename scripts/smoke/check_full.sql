-- Unified full snapshot by order_id.
-- Replace order_id literal.
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
