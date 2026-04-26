-- Pricing MVP / Spread Engine
--
-- Notes:
-- 1) This migration introduces core.quotes as the canonical source of truth
--    for new quote generation and future ReserveOrder integration.
-- 2) Legacy core.quote_snapshots is intentionally left intact for transition.
--    New pricing flow should write/read core.quotes only.
-- 3) fixed_fee in margin_rules is denominated in quote_currency_id.
-- 4) bid/ask semantics:
--      bid = price at which the system buys base_currency
--      ask = price at which the system sells base_currency

CREATE SCHEMA IF NOT EXISTS core;

-- ==========================================
-- 1) BASE RATES (market data)
-- ==========================================

CREATE TABLE IF NOT EXISTS core.base_rates (
  tenant_id uuid NOT NULL,
  base_currency_id uuid NOT NULL,
  quote_currency_id uuid NOT NULL,

  bid numeric(38,18) NOT NULL,
  ask numeric(38,18) NOT NULL,

  source_name text NOT NULL DEFAULT 'manual',
  updated_at timestamptz NOT NULL DEFAULT now(),

  PRIMARY KEY (tenant_id, base_currency_id, quote_currency_id),

  CONSTRAINT chk_base_rates_bid_positive
    CHECK (bid > 0),

  CONSTRAINT chk_base_rates_ask_positive
    CHECK (ask > 0),

  CONSTRAINT chk_base_rates_spread
    CHECK (ask >= bid)
);

CREATE INDEX IF NOT EXISTS ix_base_rates_tenant_updated
  ON core.base_rates (tenant_id, updated_at DESC);

COMMENT ON TABLE core.base_rates IS
  'Canonical market data table for pricing. bid/ask are stored by (tenant_id, base_currency_id, quote_currency_id).';

COMMENT ON COLUMN core.base_rates.bid IS
  'Price at which the system buys base_currency from the client/market.';

COMMENT ON COLUMN core.base_rates.ask IS
  'Price at which the system sells base_currency to the client/market.';

-- ==========================================
-- 2) MARGIN RULES (spread config)
-- ==========================================

CREATE TABLE IF NOT EXISTS core.margin_rules (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  tenant_id uuid NOT NULL,
  office_id uuid, -- NULL = global fallback rule

  base_currency_id uuid NOT NULL,
  quote_currency_id uuid NOT NULL,

  side text NOT NULL, -- buy | sell

  -- Which amount is used to select the tier:
  -- give         = compare against amount_give
  -- get          = compare against amount_get
  -- quote_notional = compare against quote-currency notional
  volume_basis text NOT NULL DEFAULT 'give',

  min_volume numeric(38,18) NOT NULL DEFAULT 0,
  max_volume numeric(38,18),

  -- 100 bps = 1.00%
  margin_bps integer NOT NULL DEFAULT 0,

  -- Denominated in quote_currency_id
  fixed_fee numeric(38,18) NOT NULL DEFAULT 0,

  -- Deterministic conflict resolution: higher priority wins first
  priority integer NOT NULL DEFAULT 100,

  -- Rounding policy captured at rule level
  rounding_precision integer NOT NULL DEFAULT 2,
  rounding_mode text NOT NULL DEFAULT 'half_up', -- half_up | floor | ceil

  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT chk_margin_rules_side
    CHECK (side IN ('buy', 'sell')),

  CONSTRAINT chk_margin_rules_volume_basis
    CHECK (volume_basis IN ('give', 'get', 'quote_notional')),

  CONSTRAINT chk_margin_rules_min_volume_non_negative
    CHECK (min_volume >= 0),

  CONSTRAINT chk_margin_rules_max_volume_gt_min
    CHECK (max_volume IS NULL OR max_volume > min_volume),

  CONSTRAINT chk_margin_rules_fixed_fee_non_negative
    CHECK (fixed_fee >= 0),

  CONSTRAINT chk_margin_rules_rounding_precision
    CHECK (rounding_precision BETWEEN 0 AND 18),

  CONSTRAINT chk_margin_rules_rounding_mode
    CHECK (rounding_mode IN ('half_up', 'floor', 'ceil'))
);

-- Office-specific rules
CREATE INDEX IF NOT EXISTS ix_margin_rules_lookup_office
  ON core.margin_rules (
    tenant_id,
    office_id,
    base_currency_id,
    quote_currency_id,
    side,
    volume_basis,
    priority DESC,
    min_volume DESC
  )
  WHERE office_id IS NOT NULL;

-- Global fallback rules
CREATE INDEX IF NOT EXISTS ix_margin_rules_lookup_global
  ON core.margin_rules (
    tenant_id,
    base_currency_id,
    quote_currency_id,
    side,
    volume_basis,
    priority DESC,
    min_volume DESC
  )
  WHERE office_id IS NULL;

COMMENT ON TABLE core.margin_rules IS
  'Pricing/spread rules. Deterministic selection should prefer office-specific rules over global rules, then higher priority, then narrower tier.';

COMMENT ON COLUMN core.margin_rules.fixed_fee IS
  'Fixed fee denominated in quote_currency_id.';

COMMENT ON COLUMN core.margin_rules.volume_basis IS
  'Specifies which amount is used to choose the pricing tier.';

-- ==========================================
-- 3) QUOTES (canonical quote store)
-- ==========================================

CREATE TABLE IF NOT EXISTS core.quotes (
  id text PRIMARY KEY, -- e.g. ULID / q_...

  tenant_id uuid NOT NULL,
  office_id uuid NOT NULL,
  client_ref text,

  -- buy | sell relative to base/crypto business semantics
  side text NOT NULL,

  -- give | get
  input_mode text NOT NULL,

  -- Original amount requested by caller before calculation
  requested_amount numeric(38,18) NOT NULL,

  -- Final calculated exchange legs
  give_currency_id uuid NOT NULL,
  get_currency_id uuid NOT NULL,
  amount_give numeric(38,18) NOT NULL,
  amount_get numeric(38,18) NOT NULL,

  -- Final client-facing rate
  fixed_rate numeric(38,18) NOT NULL,

  -- Audit snapshot ("why this quote was produced")
  applied_rule_id uuid REFERENCES core.margin_rules(id),
  base_rate_snapshot numeric(38,18) NOT NULL,
  margin_bps_applied integer NOT NULL DEFAULT 0,
  fixed_fee_applied numeric(38,18) NOT NULL DEFAULT 0,
  source_name_snapshot text,
  rate_updated_at_snapshot timestamptz,

  -- Rounding policy frozen into the quote
  rounding_precision integer NOT NULL DEFAULT 2,
  rounding_mode text NOT NULL DEFAULT 'half_up',

  status text NOT NULL DEFAULT 'active', -- active | consumed | expired
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  expired_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT chk_quotes_side
    CHECK (side IN ('buy', 'sell')),

  CONSTRAINT chk_quotes_input_mode
    CHECK (input_mode IN ('give', 'get')),

  CONSTRAINT chk_quotes_status
    CHECK (status IN ('active', 'consumed', 'expired')),

  CONSTRAINT chk_quotes_requested_amount_positive
    CHECK (requested_amount > 0),

  CONSTRAINT chk_quotes_amount_give_positive
    CHECK (amount_give > 0),

  CONSTRAINT chk_quotes_amount_get_positive
    CHECK (amount_get > 0),

  CONSTRAINT chk_quotes_fixed_rate_positive
    CHECK (fixed_rate > 0),

  CONSTRAINT chk_quotes_base_rate_snapshot_positive
    CHECK (base_rate_snapshot > 0),

  CONSTRAINT chk_quotes_fixed_fee_applied_non_negative
    CHECK (fixed_fee_applied >= 0),

  CONSTRAINT chk_quotes_rounding_precision
    CHECK (rounding_precision BETWEEN 0 AND 18),

  CONSTRAINT chk_quotes_rounding_mode
    CHECK (rounding_mode IN ('half_up', 'floor', 'ceil'))
);

-- Fast lookup for ReserveOrder / ConsumeQuote
CREATE INDEX IF NOT EXISTS ix_quotes_active_lookup
  ON core.quotes (tenant_id, office_id, expires_at)
  WHERE status = 'active';

-- Cleanup / TTL sweeps
CREATE INDEX IF NOT EXISTS ix_quotes_status_expires
  ON core.quotes (status, expires_at);

-- Optional operator / customer traceability
CREATE INDEX IF NOT EXISTS ix_quotes_client_created
  ON core.quotes (tenant_id, client_ref, created_at DESC);

COMMENT ON TABLE core.quotes IS
  'Canonical quote table for Pricing Engine. Intended to replace legacy core.quote_snapshots for new ReserveOrder flow.';

COMMENT ON COLUMN core.quotes.requested_amount IS
  'Original caller-provided amount before calculation, interpreted according to input_mode.';

COMMENT ON COLUMN core.quotes.base_rate_snapshot IS
  'Raw market/base rate captured at quote generation time for audit and PnL analysis.';

COMMENT ON COLUMN core.quotes.fixed_fee_applied IS
  'Fixed fee actually applied to this quote, denominated in quote currency context.';

COMMENT ON COLUMN core.quotes.rate_updated_at_snapshot IS
  'Timestamp of the market rate used to generate the quote, for freshness/audit checks.';
