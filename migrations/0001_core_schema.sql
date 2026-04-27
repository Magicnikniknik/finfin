CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS core;

-- ==========================================
-- 1) ACCOUNTS
-- ==========================================

CREATE TABLE IF NOT EXISTS core.accounts (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  office_id uuid,
  currency_id uuid NOT NULL,
  account_type text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_accounts_account_type
    CHECK (account_type IN (
      'balance',
      'available_ledger',
      'reserved_ledger',
      'settlement',
      'external',
      'system'
    ))
);

-- Для офисных счетов
CREATE UNIQUE INDEX IF NOT EXISTS ux_accounts_tenant_office_currency_type
  ON core.accounts (tenant_id, office_id, currency_id, account_type)
  WHERE office_id IS NOT NULL;

-- Для глобальных/системных счетов без office_id
CREATE UNIQUE INDEX IF NOT EXISTS ux_accounts_tenant_currency_type_global
  ON core.accounts (tenant_id, currency_id, account_type)
  WHERE office_id IS NULL;

CREATE INDEX IF NOT EXISTS ix_accounts_tenant_currency
  ON core.accounts (tenant_id, currency_id);

-- ==========================================
-- 2) ACCOUNT BALANCES (projection)
-- ==========================================

CREATE TABLE IF NOT EXISTS core.account_balances (
  account_id uuid PRIMARY KEY
    REFERENCES core.accounts(id) ON DELETE CASCADE,
  tenant_id uuid NOT NULL,
  currency_id uuid NOT NULL,
  available numeric(38,18) NOT NULL DEFAULT 0,
  reserved numeric(38,18) NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_account_balances_available_non_negative CHECK (available >= 0),
  CONSTRAINT chk_account_balances_reserved_non_negative CHECK (reserved >= 0)
);

CREATE INDEX IF NOT EXISTS ix_account_balances_tenant_currency
  ON core.account_balances (tenant_id, currency_id);

-- ==========================================
-- 3) ORDERS
-- ==========================================

CREATE TABLE IF NOT EXISTS core.orders (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  office_id uuid NOT NULL,
  client_ref text NOT NULL,
  side text NOT NULL,
  give_currency_id uuid NOT NULL,
  get_currency_id uuid NOT NULL,
  amount_give numeric(38,18) NOT NULL,
  amount_get numeric(38,18) NOT NULL,
  fixed_rate numeric(38,18) NOT NULL,
  quote_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL,
  reserved_at timestamptz,
  expires_at timestamptz NOT NULL,
  completed_at timestamptz,
  cancelled_at timestamptz,
  cancel_reason text,
  version bigint NOT NULL DEFAULT 1,
  order_ref text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT chk_orders_side
    CHECK (side IN ('buy', 'sell')),

  CONSTRAINT chk_orders_status
    CHECK (status IN ('new', 'reserved', 'completed', 'expired', 'cancelled')),

  CONSTRAINT chk_orders_amount_give_positive
    CHECK (amount_give > 0),

  CONSTRAINT chk_orders_amount_get_positive
    CHECK (amount_get > 0),

  CONSTRAINT chk_orders_fixed_rate_positive
    CHECK (fixed_rate > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_orders_tenant_order_ref
  ON core.orders (tenant_id, order_ref);

CREATE INDEX IF NOT EXISTS ix_orders_tenant_status_expires
  ON core.orders (tenant_id, status, expires_at);

CREATE INDEX IF NOT EXISTS ix_orders_ttl_worker
  ON core.orders (expires_at, id)
  WHERE status = 'reserved';

-- ==========================================
-- 4) ORDER HOLDS
-- ==========================================

CREATE TABLE IF NOT EXISTS core.order_holds (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  order_id uuid NOT NULL
    REFERENCES core.orders(id) ON DELETE CASCADE,

  balance_account_id uuid NOT NULL
    REFERENCES core.accounts(id),

  available_ledger_account_id uuid NOT NULL
    REFERENCES core.accounts(id),

  reserved_ledger_account_id uuid NOT NULL
    REFERENCES core.accounts(id),

  settlement_ledger_account_id uuid
    REFERENCES core.accounts(id),

  currency_id uuid NOT NULL,
  amount numeric(38,18) NOT NULL,
  status text NOT NULL,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  released_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT chk_order_holds_status
    CHECK (status IN ('active', 'consumed', 'released', 'expired')),

  CONSTRAINT chk_order_holds_amount_positive
    CHECK (amount > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_order_holds_order_id
  ON core.order_holds(order_id);

CREATE INDEX IF NOT EXISTS ix_order_holds_tenant_status_expires
  ON core.order_holds (tenant_id, status, expires_at);

CREATE INDEX IF NOT EXISTS ix_order_holds_active_by_order
  ON core.order_holds (order_id)
  WHERE status = 'active';

-- ==========================================
-- 5) LEDGER
-- ==========================================

CREATE TABLE IF NOT EXISTS core.ledger_journals (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL,
  kind text NOT NULL,
  order_id uuid,
  hold_id uuid,
  note text,
  created_by text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_ledger_journals_kind
    CHECK (kind IN ('hold_create', 'hold_release', 'trade_complete', 'manual_adjustment'))
);

CREATE INDEX IF NOT EXISTS ix_ledger_journals_tenant_created
  ON core.ledger_journals (tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS core.ledger_entries (
  id bigserial PRIMARY KEY,
  journal_id uuid NOT NULL
    REFERENCES core.ledger_journals(id) ON DELETE CASCADE,
  tenant_id uuid NOT NULL,
  account_id uuid NOT NULL
    REFERENCES core.accounts(id),
  currency_id uuid NOT NULL,
  direction text NOT NULL,
  amount numeric(38,18) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT chk_ledger_entries_direction
    CHECK (direction IN ('debit', 'credit')),

  CONSTRAINT chk_ledger_entries_amount_positive
    CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS ix_ledger_entries_journal_id
  ON core.ledger_entries(journal_id);

CREATE INDEX IF NOT EXISTS ix_ledger_entries_account_id_created
  ON core.ledger_entries(account_id, created_at DESC);

-- ==========================================
-- 6) IDEMPOTENCY
-- ==========================================

CREATE TABLE IF NOT EXISTS core.idempotency_keys (
  id bigserial PRIMARY KEY,
  tenant_id uuid NOT NULL,
  scope text NOT NULL,
  idem_key text NOT NULL,
  request_hash text NOT NULL,
  status text NOT NULL,
  response_code int,
  response_body jsonb,
  resource_type text,
  resource_id uuid,
  locked_until timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT chk_idempotency_status
    CHECK (status IN ('in_progress', 'completed', 'failed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_idempotency_tenant_scope_key
  ON core.idempotency_keys (tenant_id, scope, idem_key);

CREATE INDEX IF NOT EXISTS ix_idempotency_locked_until
  ON core.idempotency_keys (locked_until);

-- ==========================================
-- 7) OUTBOX
-- ==========================================

CREATE TABLE IF NOT EXISTS core.outbox_events (
  id bigserial PRIMARY KEY,
  tenant_id uuid NOT NULL,
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  headers jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'pending',
  attempts int NOT NULL DEFAULT 0,
  available_at timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT chk_outbox_status
    CHECK (status IN ('pending', 'processing', 'published', 'failed'))
);

CREATE INDEX IF NOT EXISTS ix_outbox_pending
  ON core.outbox_events (status, available_at, id)
  WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS ix_outbox_processing
  ON core.outbox_events (status, updated_at, id)
  WHERE status = 'processing';
