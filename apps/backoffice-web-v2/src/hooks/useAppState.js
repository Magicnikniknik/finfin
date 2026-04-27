import { useState, useCallback, useEffect } from 'react'
import { DEMO_SCENARIOS, getScenarioById } from '../lib/scenarios'
import { currencyMeta, officeLabel, sourceLabel, cashierLabel } from '../lib/catalog'
import { asTrimmed, extractError, safeJson, mapStatusVariant, toIso } from '../lib/utils'

const LS_CONFIG = 'finfin.backoffice.config'
const LS_STATE  = 'finfin.backoffice.demo-state'

const DEFAULT_METRICS = { activeQuotes: 0, reservedOrders: 0, completedToday: 0, cancelledToday: 0, staleAlerts: 0, grossMargin: 0 }

const DEFAULT_QUOTE_FORM = { office_id: '', input_mode: 'GIVE', give_currency_id: '', get_currency_id: '', amount: '' }
const DEFAULT_RESERVE_FORM = { idempotency_key: 'idem-reserve-001', side: 'BUY', office_id: '', quote_id: '', give_amount: '', give_code: '', give_network: '', get_amount: '', get_code: '', get_network: '' }
const DEFAULT_COMPLETE_FORM = { idempotency_key: 'idem-complete-001', order_id: '', expected_version: 1, cashier_id: 'cashier_bkk_01' }
const DEFAULT_CANCEL_FORM   = { idempotency_key: 'idem-cancel-001', order_id: '', expected_version: 1, reason: 'manual_cancel' }

export function useAppState() {
  const [config, setConfig] = useState(() => {
    try {
      const raw = localStorage.getItem(LS_CONFIG)
      const c = raw ? JSON.parse(raw) : {}
      return { baseUrl: c.baseUrl || 'http://localhost:3000', tenantId: c.tenantId || '', clientRef: c.clientRef || '', cashierId: c.cashierId || 'cashier_bkk_01' }
    } catch { return { baseUrl: 'http://localhost:3000', tenantId: '', clientRef: '', cashierId: 'cashier_bkk_01' } }
  })

  const [appState, setAppState] = useState(() => {
    try {
      const raw = localStorage.getItem(LS_STATE)
      const s = raw ? JSON.parse(raw) : {}
      return {
        currentScenarioId: s.currentScenarioId || DEMO_SCENARIOS[0].id,
        currentQuote: s.currentQuote || null,
        currentOrder: s.currentOrder || null,
        lastAction: s.lastAction || 'idle',
        lastError: s.lastError || null,
        presentationMode: Boolean(s.presentationMode),
        metrics: s.metrics || { ...DEFAULT_METRICS },
      }
    } catch { return { currentScenarioId: DEMO_SCENARIOS[0].id, currentQuote: null, currentOrder: null, lastAction: 'idle', lastError: null, presentationMode: false, metrics: { ...DEFAULT_METRICS } } }
  })

  const [quoteForm, setQuoteForm]     = useState(DEFAULT_QUOTE_FORM)
  const [reserveForm, setReserveForm] = useState(DEFAULT_RESERVE_FORM)
  const [completeForm, setCompleteForm] = useState(DEFAULT_COMPLETE_FORM)
  const [cancelForm, setCancelForm]   = useState(DEFAULT_CANCEL_FORM)

  const [response, setResponse] = useState({ status: 'idle', text: 'idle', json: { message: 'Send a request to see JSON response' }, loading: false })

  // Persist config
  useEffect(() => {
    localStorage.setItem(LS_CONFIG, JSON.stringify(config))
  }, [config])

  // Persist state
  useEffect(() => {
    localStorage.setItem(LS_STATE, JSON.stringify(appState))
  }, [appState])

  // Apply scenario to forms on mount
  useEffect(() => {
    applyScenarioToForms(getScenarioById(appState.currentScenarioId))
  }, [])

  // ── helpers ────────────────────────────────────────────────────────────

  const applyScenarioToForms = useCallback((scenario) => {
    setQuoteForm({
      office_id: scenario.quote.office_id,
      input_mode: scenario.quote.input_mode,
      give_currency_id: scenario.quote.give_currency_id,
      get_currency_id: scenario.quote.get_currency_id,
      amount: scenario.quote.amount,
    })

    const giveMeta = currencyMeta(scenario.reserve?.give_currency_id ?? scenario.quote.give_currency_id)
    const getMeta  = currencyMeta(scenario.reserve?.get_currency_id  ?? scenario.quote.get_currency_id)

    setReserveForm((prev) => ({
      ...prev,
      office_id:   scenario.quote.office_id,
      quote_id:    scenario.reserve?.quote_id ?? 'quote-from-calculate',
      side:        scenario.reserve?.side ?? 'BUY',
      give_amount: scenario.reserve?.give_amount ?? '0',
      get_amount:  scenario.reserve?.get_amount  ?? '0',
      give_code:   giveMeta.code,
      give_network: giveMeta.network,
      get_code:    getMeta.code,
      get_network: getMeta.network,
    }))

    setCompleteForm((prev) => ({ ...prev, cashier_id: config.cashierId || 'cashier_bkk_01' }))
  }, [config.cashierId])

  const sendRequest = useCallback(async (path, payload) => {
    const baseUrl   = asTrimmed(config.baseUrl).replace(/\/$/, '')
    const tenantId  = asTrimmed(config.tenantId)
    const clientRef = asTrimmed(config.clientRef)

    if (!baseUrl || !tenantId || !clientRef) {
      const err = { error: { code: 'LOCAL_VALIDATION', message: 'Set base URL, tenant_id and client_ref first' } }
      setResponse({ status: 'warn', text: 'local validation', json: err, loading: false })
      return { ok: false, json: err }
    }

    setResponse((r) => ({ ...r, status: 'idle', text: `sending ${path}`, loading: true }))

    try {
      const resp = await fetch(`${baseUrl}${path}`, {
        method: 'POST',
        headers: { 'content-type': 'application/json', 'x-tenant-id': tenantId, 'x-client-ref': clientRef },
        body: JSON.stringify(payload),
      })
      const json = await safeJson(resp)
      const status = resp.ok ? 'ok' : resp.status >= 500 ? 'err' : 'warn'
      setResponse({ status, text: `${resp.status} ${resp.statusText}`, json: { request: payload, response: json }, loading: false })
      return { ok: resp.ok, json }
    } catch (error) {
      const json = { error: { code: 'NETWORK_ERROR', message: String(error) } }
      setResponse({ status: 'err', text: 'network error', json: { request: payload, response: json }, loading: false })
      return { ok: false, json }
    }
  }, [config])

  const normalizeQuote = (json, req) => {
    const give = json.give ?? {}
    const get  = json.get  ?? {}
    return {
      quote_id:       json.quote_id ?? json.quoteId ?? '',
      office_id:      req.office_id,
      status:         (json.status ?? 'active').toLowerCase(),
      base_rate:      json.base_rate ?? json.baseRate ?? '—',
      fixed_rate:     json.fixed_rate ?? json.fixedRate ?? '—',
      fee_amount:     json.fee_amount ?? json.feeAmount ?? '0',
      source_name:    json.source_name ?? json.sourceName ?? 'manual',
      expires_at_ts:  json.expires_at_ts ?? json.expiresAtTs ?? '',
      applied_rule:   json.applied_rule ?? json.appliedRule ?? 'n/a',
      give_amount:    give.amount ?? '—',
      get_amount:     get.amount  ?? '—',
      give_currency_id: give.currency_id ?? give.currencyId ?? req.give_currency_id,
      get_currency_id:  get.currency_id  ?? get.currencyId  ?? req.get_currency_id,
    }
  }

  const estimateMargin = (quote) => {
    if (!quote) return 0
    const base = Number(quote.base_rate ?? 0)
    const fixed = Number(quote.fixed_rate ?? 0)
    const fee = Number(quote.fee_amount ?? 0)
    const notional = Number(quote.give_amount ?? 0)
    if ([base, fixed, fee, notional].some(Number.isNaN)) return 0
    return Math.max(0, (fixed - base) * notional + fee)
  }

  // ── actions ────────────────────────────────────────────────────────────

  const runQuote = useCallback(async () => {
    const payload = {
      office_id:        asTrimmed(quoteForm.office_id),
      give_currency_id: asTrimmed(quoteForm.give_currency_id),
      get_currency_id:  asTrimmed(quoteForm.get_currency_id),
      input_mode:       asTrimmed(quoteForm.input_mode),
      amount:           asTrimmed(quoteForm.amount),
    }

    const { ok, json } = await sendRequest('/quotes/calculate', payload)

    setAppState((prev) => {
      const next = { ...prev, lastAction: 'quote' }
      if (ok) {
        const q = normalizeQuote(json, payload)
        next.currentQuote = q
        next.lastError = null
        if ((q.status ?? 'active') === 'active') {
          next.metrics = { ...prev.metrics, activeQuotes: prev.metrics.activeQuotes + 1 }
        }
        setReserveForm((rf) => ({ ...rf, quote_id: q.quote_id }))
      } else {
        next.lastError = extractError(json)
        const code = next.lastError.code.toLowerCase()
        const msg  = next.lastError.message.toLowerCase()
        if (code.includes('stale') || msg.includes('stale')) {
          next.metrics = { ...prev.metrics, staleAlerts: prev.metrics.staleAlerts + 1 }
        }
      }
      return next
    })
  }, [quoteForm, sendRequest])

  const runReserve = useCallback(async () => {
    const payload = {
      idempotency_key: asTrimmed(reserveForm.idempotency_key),
      office_id:       asTrimmed(reserveForm.office_id),
      quote_id:        asTrimmed(reserveForm.quote_id),
      side:            asTrimmed(reserveForm.side),
      give: { amount: asTrimmed(reserveForm.give_amount), currency: { code: asTrimmed(reserveForm.give_code), network: asTrimmed(reserveForm.give_network) } },
      get:  { amount: asTrimmed(reserveForm.get_amount),  currency: { code: asTrimmed(reserveForm.get_code),  network: asTrimmed(reserveForm.get_network)  } },
    }

    const { ok, json } = await sendRequest('/orders/reserve', payload)

    setAppState((prev) => {
      const next = { ...prev, lastAction: 'reserve' }
      if (ok) {
        next.currentOrder = {
          order_id:      json.order_id,
          order_ref:     json.order_ref ?? 'n/a',
          status:        (json.status ?? 'reserved').toLowerCase(),
          version:       json.version,
          expires_at_ts: json.expires_at_ts,
          held_amount:   reserveForm.get_amount,
        }
        next.currentQuote = prev.currentQuote ? { ...prev.currentQuote, status: 'consumed' } : null
        next.lastError = null
        next.metrics = {
          ...prev.metrics,
          reservedOrders: prev.metrics.reservedOrders + 1,
          activeQuotes:   Math.max(0, prev.metrics.activeQuotes - 1),
          grossMargin:    prev.metrics.grossMargin + estimateMargin(prev.currentQuote),
        }
        setCompleteForm((f) => ({ ...f, order_id: json.order_id, expected_version: json.version ?? 1 }))
        setCancelForm((f)   => ({ ...f, order_id: json.order_id, expected_version: json.version ?? 1 }))
      } else {
        next.lastError = extractError(json)
      }
      return next
    })
  }, [reserveForm, sendRequest])

  const runComplete = useCallback(async () => {
    const payload = {
      idempotency_key:  asTrimmed(completeForm.idempotency_key),
      order_id:         asTrimmed(completeForm.order_id),
      expected_version: Number(completeForm.expected_version),
      cashier_id:       asTrimmed(completeForm.cashier_id),
    }

    const { ok, json } = await sendRequest('/orders/complete', payload)

    setAppState((prev) => {
      const next = { ...prev, lastAction: 'complete' }
      if (ok) {
        next.currentOrder = { ...(prev.currentOrder ?? {}), order_id: json.order_id ?? prev.currentOrder?.order_id, status: (json.status ?? 'completed').toLowerCase(), version: json.version, completed_at_ts: json.completed_at_ts }
        next.lastError = null
        next.metrics = { ...prev.metrics, completedToday: prev.metrics.completedToday + 1, reservedOrders: Math.max(0, prev.metrics.reservedOrders - 1) }
      } else {
        next.lastError = extractError(json)
      }
      return next
    })
  }, [completeForm, sendRequest])

  const runCancel = useCallback(async () => {
    const payload = {
      idempotency_key:  asTrimmed(cancelForm.idempotency_key),
      order_id:         asTrimmed(cancelForm.order_id),
      expected_version: Number(cancelForm.expected_version),
      reason:           asTrimmed(cancelForm.reason),
    }

    const { ok, json } = await sendRequest('/orders/cancel', payload)

    setAppState((prev) => {
      const next = { ...prev, lastAction: 'cancel' }
      if (ok) {
        next.currentOrder = { ...(prev.currentOrder ?? {}), order_id: json.order_id ?? prev.currentOrder?.order_id, status: (json.status ?? 'cancelled').toLowerCase(), version: json.version }
        next.lastError = null
        next.metrics = { ...prev.metrics, cancelledToday: prev.metrics.cancelledToday + 1, reservedOrders: Math.max(0, prev.metrics.reservedOrders - 1) }
      } else {
        next.lastError = extractError(json)
      }
      return next
    })
  }, [cancelForm, sendRequest])

  const runFullFlow = useCallback(async () => {
    await runQuote()
    setAppState((prev) => {
      if (prev.lastError) return prev
      return prev
    })
    // We need to check lastError after runQuote; use a ref-based approach via promise chain
    const scenario = getScenarioById(appState.currentScenarioId)
    await runReserve()
    if (scenario.actionAfterReserve === 'complete') await runComplete()
    if (scenario.actionAfterReserve === 'cancel')   await runCancel()
  }, [appState.currentScenarioId, runQuote, runReserve, runComplete, runCancel])

  const selectScenario = useCallback((id) => {
    const scenario = getScenarioById(id)
    setAppState((prev) => ({ ...prev, currentScenarioId: id }))
    applyScenarioToForms(scenario)
  }, [applyScenarioToForms])

  const runHappyDemo = useCallback(async () => {
    selectScenario('buy-100-usdt')
    await runFullFlow()
  }, [selectScenario, runFullFlow])

  const runErrorDemo = useCallback(async () => {
    selectScenario('expired-quote')
    await runReserve()
  }, [selectScenario, runReserve])

  const loadSandboxDefaults = useCallback(() => {
    const first = DEMO_SCENARIOS[0]
    setAppState({ currentScenarioId: first.id, currentQuote: null, currentOrder: null, lastAction: 'sandbox-defaults', lastError: null, presentationMode: false, metrics: { ...DEFAULT_METRICS } })
    applyScenarioToForms(first)
    setResponse({ status: 'ok', text: 'Production-like sandbox defaults loaded', json: { message: 'Sandbox defaults loaded' }, loading: false })
  }, [applyScenarioToForms])

  const resetDemoState = useCallback(() => {
    const first = DEMO_SCENARIOS[0]
    setAppState({ currentScenarioId: first.id, currentQuote: null, currentOrder: null, lastAction: 'idle', lastError: null, presentationMode: false, metrics: { ...DEFAULT_METRICS } })
    setCompleteForm(DEFAULT_COMPLETE_FORM)
    setCancelForm(DEFAULT_CANCEL_FORM)
    applyScenarioToForms(first)
    localStorage.removeItem(LS_STATE)
    setResponse({ status: 'ok', text: 'Demo state reset', json: { message: 'Demo state reset' }, loading: false })
  }, [applyScenarioToForms])

  const togglePresentationMode = useCallback(() => {
    setAppState((prev) => ({ ...prev, presentationMode: !prev.presentationMode }))
  }, [])

  const copyQuoteId = useCallback(async () => {
    if (!appState.currentQuote?.quote_id) {
      setResponse((r) => ({ ...r, status: 'warn', text: 'No quote_id to copy' }))
      return
    }
    try {
      await navigator.clipboard.writeText(appState.currentQuote.quote_id)
      setResponse((r) => ({ ...r, status: 'ok', text: 'quote_id copied' }))
    } catch {
      setResponse((r) => ({ ...r, status: 'warn', text: 'clipboard unavailable' }))
    }
  }, [appState.currentQuote])

  const useQuoteInReserve = useCallback(() => {
    const q = appState.currentQuote
    if (!q) { setResponse((r) => ({ ...r, status: 'warn', text: 'No quote yet. Run Quote first.' })); return }
    const give = currencyMeta(q.give_currency_id)
    const get  = currencyMeta(q.get_currency_id)
    setReserveForm((f) => ({ ...f, office_id: q.office_id, quote_id: q.quote_id, give_amount: q.give_amount, get_amount: q.get_amount, give_code: give.code, give_network: give.network, get_code: get.code, get_network: get.network }))
    setResponse((r) => ({ ...r, status: 'ok', text: 'Quote applied to reserve form' }))
  }, [appState.currentQuote])

  // ── derived: timeline ───────────────────────────────────────────────────
  const timeline = (() => {
    const { currentQuote: q, currentOrder: o, lastAction, lastError } = appState

    const steps = [
      { key: 'quote',    label: q ? `Quote: ${q.status} (${q.quote_id?.slice(0, 8) ?? 'n/a'}…)` : 'Quote: not started',           variant: q ? mapStatusVariant(q.status) : 'neutral' },
      { key: 'reserve',  label: o ? `Reserve: ${o.status} (${o.order_id?.slice(0, 8) ?? 'n/a'}…)` : 'Order reserve: not started',  variant: o ? mapStatusVariant(o.status) : 'neutral' },
      { key: 'complete', label: o?.status === 'completed' ? `Complete: completed (v${o.version ?? 'n/a'})` : 'Order complete: not started', variant: o?.status === 'completed' ? 'ok' : 'neutral' },
      { key: 'cancel',   label: o?.status === 'cancelled' ? `Cancel: cancelled (v${o.version ?? 'n/a'})` : 'Order cancel: not started',    variant: o?.status === 'cancelled' ? 'warn' : 'neutral' },
    ]

    if (lastError) {
      const target = lastAction === 'quote' ? 0 : lastAction === 'reserve' ? 1 : lastAction === 'complete' ? 2 : 3
      if (steps[target]) { steps[target].label += ` • ${lastError.code}`; steps[target].variant = 'err' }
    }
    return steps
  })()

  // ── derived: audit rows ─────────────────────────────────────────────────
  const { currentQuote: q, currentOrder: o } = appState
  const auditRows = [
    { label: 'Scenario',     value: getScenarioById(appState.currentScenarioId).title },
    { label: 'Applied rule', value: q?.applied_rule ?? 'n/a' },
    { label: 'Margin / fee', value: `${q?.base_rate ?? '—'} → ${q?.fixed_rate ?? '—'} / fee ${q?.fee_amount ?? '—'}` },
    { label: 'Held asset',   value: q ? `${q.get_amount} ${currencyMeta(q.get_currency_id).label}` : '—' },
    { label: 'Quote source', value: sourceLabel(q?.source_name) },
    { label: 'Last action',  value: appState.lastAction },
    { label: 'Last error',   value: appState.lastError ? `${appState.lastError.code}: ${appState.lastError.message}` : 'none' },
    { label: 'Cashier',      value: cashierLabel(asTrimmed(config.cashierId)) },
    { label: 'HTTP status',  value: response.text },
  ]

  // ── quote summary data ──────────────────────────────────────────────────
  const quoteSummaryData = q ? {
    quote_id:    q.quote_id,
    office:      officeLabel(q.office_id),
    source_name: sourceLabel(q.source_name),
    base_rate:   q.base_rate,
    fixed_rate:  q.fixed_rate,
    fee:         q.fee_amount,
    rule:        q.applied_rule,
    expires_at:  toIso(q.expires_at_ts),
    status:      q.status ?? 'active',
    give:        `${q.give_amount} ${currencyMeta(q.give_currency_id).label}`,
    get:         `${q.get_amount}  ${currencyMeta(q.get_currency_id).label}`,
  } : null

  // ── order summary data ──────────────────────────────────────────────────
  const orderSummaryData = o ? {
    order_id:    o.order_id,
    order_ref:   o.order_ref ?? 'n/a',
    status:      o.status,
    version:     String(o.version ?? '—'),
    expires_at:  toIso(o.expires_at_ts ?? o.completed_at_ts),
    quote_status: q?.status ?? '—',
    held_currency: q ? `${q.get_amount} ${currencyMeta(q.get_currency_id).label}` : '—',
    held_amount: o.held_amount ?? '—',
  } : null

  return {
    config, setConfig,
    appState,
    quoteForm, setQuoteForm,
    reserveForm, setReserveForm,
    completeForm, setCompleteForm,
    cancelForm, setCancelForm,
    response,
    timeline,
    auditRows,
    quoteSummaryData,
    orderSummaryData,
    // actions
    selectScenario,
    runQuote, runReserve, runComplete, runCancel,
    runFullFlow, runHappyDemo, runErrorDemo,
    loadSandboxDefaults, resetDemoState,
    togglePresentationMode, copyQuoteId, useQuoteInReserve,
  }
}
