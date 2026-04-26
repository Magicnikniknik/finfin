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

DELETE FROM core.quotes
WHERE id IN (
  'q_demo_buy_100_usdt_bkk',
  'q_demo_get_10000_thb_bkk',
  'q_demo_expired_usdt_thb',
  'q_demo_consumed_usd_thb'
);

DELETE FROM core.quote_snapshots
WHERE id = 'quote-reserve-smoke-001';

DELETE FROM core.account_wiring
WHERE tenant_id = '11111111-1111-1111-1111-111111111111'
  AND (
    (office_id = '22222222-2222-2222-2222-222222222222' AND currency_id IN (
      '33333333-3333-3333-3333-333333333333',
      '90000000-0000-0000-0000-000000000001',
      '90000000-0000-0000-0000-000000000003'
    )) OR
    (office_id = '33333333-3333-3333-3333-333333333333' AND currency_id = '90000000-0000-0000-0000-000000000001')
  );

DELETE FROM core.account_balances
WHERE account_id IN (
  '55555555-5555-5555-5555-555555555555',
  '91000000-0000-0000-0000-000000000011',
  '91000000-0000-0000-0000-000000000021',
  '91000000-0000-0000-0000-000000000031'
);

DELETE FROM core.accounts
WHERE id IN (
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666666',
  '77777777-7777-7777-7777-777777777777',
  '88888888-8888-8888-8888-888888888888',
  '91000000-0000-0000-0000-000000000011',
  '91000000-0000-0000-0000-000000000012',
  '91000000-0000-0000-0000-000000000013',
  '91000000-0000-0000-0000-000000000014',
  '91000000-0000-0000-0000-000000000021',
  '91000000-0000-0000-0000-000000000022',
  '91000000-0000-0000-0000-000000000023',
  '91000000-0000-0000-0000-000000000024',
  '91000000-0000-0000-0000-000000000031',
  '91000000-0000-0000-0000-000000000032',
  '91000000-0000-0000-0000-000000000033',
  '91000000-0000-0000-0000-000000000034'
);

DELETE FROM core.margin_rules
WHERE tenant_id = '11111111-1111-1111-1111-111111111111'
  AND base_currency_id IN (
    '90000000-0000-0000-0000-000000000001',
    '90000000-0000-0000-0000-000000000003'
  );

DELETE FROM core.base_rates
WHERE tenant_id = '11111111-1111-1111-1111-111111111111'
  AND (
    (base_currency_id = '90000000-0000-0000-0000-000000000001' AND quote_currency_id = '90000000-0000-0000-0000-000000000002') OR
    (base_currency_id = '90000000-0000-0000-0000-000000000003' AND quote_currency_id = '90000000-0000-0000-0000-000000000002') OR
    (base_currency_id = '90000000-0000-0000-0000-000000000004' AND quote_currency_id = '90000000-0000-0000-0000-000000000002')
  );

COMMIT;
