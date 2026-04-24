BEGIN;

-- Full seed for ReserveOrder smoke.
-- NOTE: if account_wiring column is currency_id in your schema, replace hold_currency_id below.

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
  tenant_id, office_id, hold_currency_id, balance_account_id,
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
