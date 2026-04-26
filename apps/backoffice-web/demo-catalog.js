export const DEMO_CATALOG = {
  offices: {
    '22222222-2222-2222-2222-222222222222': 'Bangkok Main',
    '33333333-3333-3333-3333-333333333333': 'Pattaya Branch',
    '44444444-4444-4444-4444-444444444444': 'Airport Desk',
  },
  currencies: {
    '90000000-0000-0000-0000-000000000001': { label: 'USDT (TRC20)', code: 'USDT', network: 'TRC20' },
    '90000000-0000-0000-0000-000000000002': { label: 'THB Cash', code: 'THB', network: 'cash' },
    '90000000-0000-0000-0000-000000000003': { label: 'USD Cash', code: 'USD', network: 'cash' },
    '90000000-0000-0000-0000-000000000004': { label: 'BTC Mainnet', code: 'BTC', network: 'mainnet' },
  },
  cashiers: {
    cashier_bkk_01: 'Cashier Bangkok 01',
    cashier_bkk_02: 'Cashier Bangkok 02',
    cashier_ptt_01: 'Cashier Pattaya 01',
  },
  sources: {
    manual_sandbox: 'Manual Sandbox Feed',
    manual: 'Manual Feed',
  },
};

export function officeLabel(id) {
  return DEMO_CATALOG.offices[id] || id || '—';
}

export function currencyMeta(id) {
  return DEMO_CATALOG.currencies[id] || { label: id || '—', code: 'UNK', network: 'n/a' };
}

export function sourceLabel(name) {
  return DEMO_CATALOG.sources[name] || name || 'n/a';
}

export function cashierLabel(id) {
  return DEMO_CATALOG.cashiers[id] || id || 'n/a';
}
