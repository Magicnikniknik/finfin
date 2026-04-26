BEGIN;

-- Production-like Sandbox / Realistic Money Simulation seed pack.
-- Tenant: 11111111-1111-1111-1111-111111111111
-- Offices:
--   22222222-2222-2222-2222-222222222222 = Bangkok Main
--   33333333-3333-3333-3333-333333333333 = Pattaya Branch
--   44444444-4444-4444-4444-444444444444 = Airport Desk
-- Currencies:
--   90000000-0000-0000-0000-000000000001 = USDT TRC20
--   90000000-0000-0000-0000-000000000002 = THB cash
--   90000000-0000-0000-0000-000000000003 = USD cash
--   90000000-0000-0000-0000-000000000004 = BTC mainnet

-- ==============================
-- 1) Accounts + balances
-- ==============================
INSERT INTO core.accounts (id, tenant_id, office_id, currency_id, account_type) VALUES
  -- Bangkok / USDT
  ('91000000-0000-0000-0000-000000000011','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000001','balance'),
  ('91000000-0000-0000-0000-000000000012','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000001','available_ledger'),
  ('91000000-0000-0000-0000-000000000013','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000001','reserved_ledger'),
  ('91000000-0000-0000-0000-000000000014','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000001','settlement'),

  -- Bangkok / USD
  ('91000000-0000-0000-0000-000000000021','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000003','balance'),
  ('91000000-0000-0000-0000-000000000022','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000003','available_ledger'),
  ('91000000-0000-0000-0000-000000000023','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000003','reserved_ledger'),
  ('91000000-0000-0000-0000-000000000024','11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000003','settlement'),

  -- Pattaya / USDT (intentionally low balance for insufficient-liquidity scenario)
  ('91000000-0000-0000-0000-000000000031','11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333','90000000-0000-0000-0000-000000000001','balance'),
  ('91000000-0000-0000-0000-000000000032','11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333','90000000-0000-0000-0000-000000000001','available_ledger'),
  ('91000000-0000-0000-0000-000000000033','11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333','90000000-0000-0000-0000-000000000001','reserved_ledger'),
  ('91000000-0000-0000-0000-000000000034','11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333','90000000-0000-0000-0000-000000000001','settlement')
ON CONFLICT DO NOTHING;

INSERT INTO core.account_balances (account_id, tenant_id, currency_id, available, reserved, updated_at) VALUES
  ('91000000-0000-0000-0000-000000000011','11111111-1111-1111-1111-111111111111','90000000-0000-0000-0000-000000000001',15000,300,now()),
  ('91000000-0000-0000-0000-000000000021','11111111-1111-1111-1111-111111111111','90000000-0000-0000-0000-000000000003',5000,0,now()),
  ('91000000-0000-0000-0000-000000000031','11111111-1111-1111-1111-111111111111','90000000-0000-0000-0000-000000000001',20,0,now())
ON CONFLICT (account_id) DO UPDATE
SET available = EXCLUDED.available,
    reserved = EXCLUDED.reserved,
    updated_at = now();

INSERT INTO core.account_wiring (
  tenant_id, office_id, currency_id, balance_account_id,
  available_ledger_account_id, reserved_ledger_account_id, settlement_ledger_account_id,
  created_at, updated_at
) VALUES
  ('11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000001','91000000-0000-0000-0000-000000000011','91000000-0000-0000-0000-000000000012','91000000-0000-0000-0000-000000000013','91000000-0000-0000-0000-000000000014',now(),now()),
  ('11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222','90000000-0000-0000-0000-000000000003','91000000-0000-0000-0000-000000000021','91000000-0000-0000-0000-000000000022','91000000-0000-0000-0000-000000000023','91000000-0000-0000-0000-000000000024',now(),now()),
  ('11111111-1111-1111-1111-111111111111','33333333-3333-3333-3333-333333333333','90000000-0000-0000-0000-000000000001','91000000-0000-0000-0000-000000000031','91000000-0000-0000-0000-000000000032','91000000-0000-0000-0000-000000000033','91000000-0000-0000-0000-000000000034',now(),now())
ON CONFLICT (tenant_id, office_id, currency_id) DO UPDATE
SET balance_account_id = EXCLUDED.balance_account_id,
    available_ledger_account_id = EXCLUDED.available_ledger_account_id,
    reserved_ledger_account_id = EXCLUDED.reserved_ledger_account_id,
    settlement_ledger_account_id = EXCLUDED.settlement_ledger_account_id,
    updated_at = now();

-- ==============================
-- 2) Pricing base rates + rules
-- ==============================
INSERT INTO core.base_rates (tenant_id, base_currency_id, quote_currency_id, bid, ask, source_name, updated_at) VALUES
  ('11111111-1111-1111-1111-111111111111','90000000-0000-0000-0000-000000000001','90000000-0000-0000-0000-000000000002',35.10,35.30,'manual_sandbox',now()), -- USDT/THB
  ('11111111-1111-1111-1111-111111111111','90000000-0000-0000-0000-000000000003','90000000-0000-0000-0000-000000000002',36.15,36.45,'manual_sandbox',now()), -- USD/THB
  ('11111111-1111-1111-1111-111111111111','90000000-0000-0000-0000-000000000004','90000000-0000-0000-0000-000000000002',2250000,2275000,'manual_sandbox',now() - interval '10 seconds') -- BTC/THB
ON CONFLICT (tenant_id, base_currency_id, quote_currency_id) DO UPDATE
SET bid = EXCLUDED.bid,
    ask = EXCLUDED.ask,
    source_name = EXCLUDED.source_name,
    updated_at = EXCLUDED.updated_at;

DELETE FROM core.margin_rules
WHERE tenant_id = '11111111-1111-1111-1111-111111111111'
  AND base_currency_id IN ('90000000-0000-0000-0000-000000000001','90000000-0000-0000-0000-000000000003');

INSERT INTO core.margin_rules (
  tenant_id, office_id, base_currency_id, quote_currency_id, side,
  volume_basis, min_volume, max_volume, margin_bps, fixed_fee, priority,
  rounding_precision, rounding_mode, created_at, updated_at
) VALUES
  -- Global default (USDT/THB sell)
  ('11111111-1111-1111-1111-111111111111',NULL,'90000000-0000-0000-0000-000000000001','90000000-0000-0000-0000-000000000002','sell','give',0,NULL,180,0,100,2,'half_up',now(),now()),
  -- VIP large-volume tier (better margin)
  ('11111111-1111-1111-1111-111111111111',NULL,'90000000-0000-0000-0000-000000000001','90000000-0000-0000-0000-000000000002','sell','give',5000,NULL,95,0,250,2,'half_up',now(),now()),
  -- Airport desk surcharge
  ('11111111-1111-1111-1111-111111111111','44444444-4444-4444-4444-444444444444','90000000-0000-0000-0000-000000000001','90000000-0000-0000-0000-000000000002','sell','give',0,NULL,260,20,400,2,'half_up',now(),now()),
  -- USD/THB default buy
  ('11111111-1111-1111-1111-111111111111',NULL,'90000000-0000-0000-0000-000000000003','90000000-0000-0000-0000-000000000002','buy','get',0,NULL,150,0,120,2,'half_up',now(),now());

-- ==============================
-- 3) Demo quotes (active/expired/consumed)
-- ==============================
INSERT INTO core.quotes (
  id, tenant_id, office_id, client_ref, side, input_mode, requested_amount,
  give_currency_id, get_currency_id, amount_give, amount_get, fixed_rate,
  applied_rule_id, base_rate_snapshot, margin_bps_applied, fixed_fee_applied,
  source_name_snapshot, rate_updated_at_snapshot, rounding_precision, rounding_mode,
  status, expires_at, consumed_at, expired_at, created_at
) VALUES
  (
    'q_demo_buy_100_usdt_bkk',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'client_demo_buy_100_usdt',
    'buy',
    'get',
    100,
    '90000000-0000-0000-0000-000000000002',
    '90000000-0000-0000-0000-000000000001',
    3530,
    100,
    35.30,
    NULL,
    35.20,
    120,
    0,
    'manual_sandbox',
    now(),
    2,
    'half_up',
    'active',
    now() + interval '20 minutes',
    NULL,
    NULL,
    now()
  ),
  (
    'q_demo_get_10000_thb_bkk',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'client_demo_get_10000_thb',
    'sell',
    'get',
    10000,
    '90000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000002',
    285.10,
    10000,
    35.08,
    NULL,
    35.10,
    80,
    0,
    'manual_sandbox',
    now(),
    2,
    'half_up',
    'active',
    now() + interval '20 minutes',
    NULL,
    NULL,
    now()
  ),
  (
    'q_demo_expired_usdt_thb',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'client_demo_expired',
    'sell',
    'give',
    100,
    '90000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000002',
    100,
    3505,
    35.05,
    NULL,
    35.10,
    140,
    0,
    'manual_sandbox',
    now() - interval '1 hour',
    2,
    'half_up',
    'expired',
    now() - interval '15 minutes',
    NULL,
    now() - interval '10 minutes',
    now() - interval '20 minutes'
  ),
  (
    'q_demo_consumed_usd_thb',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'client_demo_consumed',
    'sell',
    'give',
    500,
    '90000000-0000-0000-0000-000000000003',
    '90000000-0000-0000-0000-000000000002',
    500,
    18150,
    36.30,
    NULL,
    36.30,
    0,
    0,
    'manual_sandbox',
    now() - interval '30 minutes',
    2,
    'half_up',
    'consumed',
    now() + interval '10 minutes',
    now() - interval '5 minutes',
    NULL,
    now() - interval '25 minutes'
  )
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    expires_at = EXCLUDED.expires_at,
    consumed_at = EXCLUDED.consumed_at,
    expired_at = EXCLUDED.expired_at,
    amount_give = EXCLUDED.amount_give,
    amount_get = EXCLUDED.amount_get,
    fixed_rate = EXCLUDED.fixed_rate,
    margin_bps_applied = EXCLUDED.margin_bps_applied,
    source_name_snapshot = EXCLUDED.source_name_snapshot,
    rate_updated_at_snapshot = EXCLUDED.rate_updated_at_snapshot;

-- ==============================
-- 4) Demo reserved order + hold (for complete/cancel panel)
-- ==============================
INSERT INTO core.orders (
  id, tenant_id, office_id, client_ref, side, give_currency_id, get_currency_id,
  amount_give, amount_get, fixed_rate, quote_payload, status, reserved_at, expires_at,
  version, order_ref, created_at, updated_at
) VALUES
  (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'client_order_panel_demo',
    'buy',
    '90000000-0000-0000-0000-000000000002',
    '90000000-0000-0000-0000-000000000001',
    3530,
    100,
    35.30,
    '{"quote_id":"q_demo_buy_100_usdt_bkk","scenario":"buy_100_usdt"}'::jsonb,
    'reserved',
    now() - interval '2 minutes',
    now() + interval '28 minutes',
    1,
    'ORD-DEMO-RESERVED-1',
    now() - interval '2 minutes',
    now()
  )
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    version = EXCLUDED.version,
    amount_give = EXCLUDED.amount_give,
    amount_get = EXCLUDED.amount_get,
    fixed_rate = EXCLUDED.fixed_rate,
    reserved_at = EXCLUDED.reserved_at,
    expires_at = EXCLUDED.expires_at,
    updated_at = now();

INSERT INTO core.order_holds (
  id, tenant_id, order_id, balance_account_id, available_ledger_account_id,
  reserved_ledger_account_id, settlement_ledger_account_id, currency_id, amount,
  status, expires_at, created_at
) VALUES
  (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    '11111111-1111-1111-1111-111111111111',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    '91000000-0000-0000-0000-000000000011',
    '91000000-0000-0000-0000-000000000012',
    '91000000-0000-0000-0000-000000000013',
    '91000000-0000-0000-0000-000000000014',
    '90000000-0000-0000-0000-000000000001',
    100,
    'active',
    now() + interval '28 minutes',
    now() - interval '2 minutes'
  )
ON CONFLICT (id) DO UPDATE
SET amount = EXCLUDED.amount,
    status = EXCLUDED.status,
    expires_at = EXCLUDED.expires_at;

COMMIT;
