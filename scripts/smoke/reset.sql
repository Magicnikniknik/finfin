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
  AND hold_currency_id = '33333333-3333-3333-3333-333333333333';

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
