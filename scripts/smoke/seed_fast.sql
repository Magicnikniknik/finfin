BEGIN;

-- Fast seed for CompleteOrder/CancelOrder smoke.
-- Creates two reserved orders + active holds.

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
