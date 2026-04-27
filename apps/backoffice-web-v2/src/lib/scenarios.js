export const DEMO_SCENARIOS = [
  {
    id: 'buy-100-usdt',
    title: 'Buy 100 USDT',
    description: 'Retail happy path: quote → reserve → complete.',
    expected: 'Quote active, reserve succeeds, order completed.',
    quote: { office_id: '22222222-2222-2222-2222-222222222222', give_currency_id: '90000000-0000-0000-0000-000000000002', get_currency_id: '90000000-0000-0000-0000-000000000001', input_mode: 'GET', amount: '100.00' },
    reserve: { side: 'BUY', give_amount: '3530.00', get_amount: '100.00', give_currency_id: '90000000-0000-0000-0000-000000000002', get_currency_id: '90000000-0000-0000-0000-000000000001' },
    actionAfterReserve: 'complete',
  },
  {
    id: 'get-10000-thb',
    title: 'Get 10,000 THB',
    description: 'Reverse calculation flow: quote → reserve → cancel.',
    expected: 'Quote active, reserve succeeds, order cancelled.',
    quote: { office_id: '22222222-2222-2222-2222-222222222222', give_currency_id: '90000000-0000-0000-0000-000000000001', get_currency_id: '90000000-0000-0000-0000-000000000002', input_mode: 'GET', amount: '10000.00' },
    reserve: { side: 'SELL', give_amount: '285.10', get_amount: '10000.00', give_currency_id: '90000000-0000-0000-0000-000000000001', get_currency_id: '90000000-0000-0000-0000-000000000002' },
    actionAfterReserve: 'cancel',
  },
  {
    id: 'vip-tier',
    title: 'VIP Tier',
    description: 'Large notional to demonstrate better margin tier.',
    expected: 'Quote shows better client rate vs low-volume scenario.',
    quote: { office_id: '22222222-2222-2222-2222-222222222222', give_currency_id: '90000000-0000-0000-0000-000000000001', get_currency_id: '90000000-0000-0000-0000-000000000002', input_mode: 'GIVE', amount: '7500.00' },
    reserve: { side: 'SELL', give_amount: '7500.00', get_amount: '261000.00', give_currency_id: '90000000-0000-0000-0000-000000000001', get_currency_id: '90000000-0000-0000-0000-000000000002' },
    actionAfterReserve: 'none',
  },
  {
    id: 'stale-rate',
    title: 'Stale Rate',
    description: 'Use BTC/THB pair where rate may be stale and rejected.',
    expected: 'Quote rejected with stale-rate reason.',
    quote: { office_id: '22222222-2222-2222-2222-222222222222', give_currency_id: '90000000-0000-0000-0000-000000000004', get_currency_id: '90000000-0000-0000-0000-000000000002', input_mode: 'GIVE', amount: '0.10' },
    reserve: null,
    actionAfterReserve: 'none',
  },
  {
    id: 'expired-quote',
    title: 'Expired Quote',
    description: 'Use pre-seeded expired quote and try reserve.',
    expected: 'Reserve rejected with quote expired reason.',
    quote: { office_id: '22222222-2222-2222-2222-222222222222', give_currency_id: '90000000-0000-0000-0000-000000000001', get_currency_id: '90000000-0000-0000-0000-000000000002', input_mode: 'GIVE', amount: '100.00' },
    reserve: { quote_id: 'q_demo_expired_usdt_thb', side: 'SELL', give_amount: '100.00', get_amount: '3505.00', give_currency_id: '90000000-0000-0000-0000-000000000001', get_currency_id: '90000000-0000-0000-0000-000000000002' },
    actionAfterReserve: 'none',
  },
  {
    id: 'insufficient-liquidity',
    title: 'Insufficient Liquidity',
    description: 'Use low-liquidity office to force reserve rejection.',
    expected: 'Reserve rejected without corrupting quote/order state.',
    quote: { office_id: '33333333-3333-3333-3333-333333333333', give_currency_id: '90000000-0000-0000-0000-000000000002', get_currency_id: '90000000-0000-0000-0000-000000000001', input_mode: 'GET', amount: '250.00' },
    reserve: { side: 'BUY', give_amount: '8825.00', get_amount: '250.00', give_currency_id: '90000000-0000-0000-0000-000000000002', get_currency_id: '90000000-0000-0000-0000-000000000001' },
    actionAfterReserve: 'none',
  },
]

export const getScenarioById = (id) =>
  DEMO_SCENARIOS.find((s) => s.id === id) ?? DEMO_SCENARIOS[0]
