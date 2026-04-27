const OFFICES = {
  '22222222-2222-2222-2222-222222222222': 'Bangkok Main',
  '33333333-3333-3333-3333-333333333333': 'Bangkok Low-Liq',
}

const CURRENCIES = {
  '90000000-0000-0000-0000-000000000001': { code: 'USDT', network: 'TRC20', label: 'USDT (TRC20)' },
  '90000000-0000-0000-0000-000000000002': { code: 'THB',  network: 'fiat',  label: 'THB' },
  '90000000-0000-0000-0000-000000000003': { code: 'USD',  network: 'fiat',  label: 'USD' },
  '90000000-0000-0000-0000-000000000004': { code: 'BTC',  network: 'BTC',   label: 'BTC' },
}

const CASHIERS = {
  cashier_bkk_01: 'Cashier Bangkok #1',
  cashier_bkk_02: 'Cashier Bangkok #2',
}

const SOURCES = {
  manual:   'Manual entry',
  feed:     'Market feed',
  oracle:   'Oracle',
  snapshot: 'Snapshot',
}

export const officeLabel   = (id)   => OFFICES[id]   ?? id ?? '—'
export const currencyMeta  = (id)   => CURRENCIES[id] ?? { code: id ?? '?', network: '?', label: id ?? '—' }
export const cashierLabel  = (id)   => CASHIERS[id]   ?? id ?? '—'
export const sourceLabel   = (name) => SOURCES[name]  ?? name ?? '—'
