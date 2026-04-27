BEGIN;

INSERT INTO core.base_rates (
  tenant_id,
  base_currency_id,
  quote_currency_id,
  bid,
  ask,
  source_name,
  updated_at
) VALUES (
  '11111111-1111-1111-1111-111111111111',
  '33333333-3333-3333-3333-333333333333',
  '44444444-4444-4444-4444-444444444444',
  35.10,
  35.30,
  'manual',
  now()
)
ON CONFLICT (tenant_id, base_currency_id, quote_currency_id) DO UPDATE
SET bid = EXCLUDED.bid,
    ask = EXCLUDED.ask,
    source_name = EXCLUDED.source_name,
    updated_at = EXCLUDED.updated_at;

INSERT INTO core.margin_rules (
  tenant_id, office_id, base_currency_id, quote_currency_id, side,
  volume_basis, min_volume, margin_bps, fixed_fee, priority,
  rounding_precision, rounding_mode, created_at, updated_at
) VALUES
(
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  '33333333-3333-3333-3333-333333333333',
  '44444444-4444-4444-4444-444444444444',
  'sell',
  'give',
  0,
  150,
  0,
  200,
  2,
  'half_up',
  now(),
  now()
),
(
  '11111111-1111-1111-1111-111111111111',
  NULL,
  '33333333-3333-3333-3333-333333333333',
  '44444444-4444-4444-4444-444444444444',
  'sell',
  'give',
  0,
  200,
  0,
  100,
  2,
  'half_up',
  now(),
  now()
);

COMMIT;
