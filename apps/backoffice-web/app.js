import { DEMO_SCENARIOS, getScenarioById } from './scenarios.js';
import { cashierLabel, currencyMeta, officeLabel, sourceLabel } from './demo-catalog.js';

const LS_CONFIG_KEY = 'finfin.backoffice.config';
const LS_STATE_KEY = 'finfin.backoffice.demo-state';

const responseNode = document.getElementById('responseJson');
const statusBadge = document.getElementById('statusBadge');
const quoteStatusBadge = document.getElementById('quoteStatusBadge');
const orderStatusBadge = document.getElementById('orderStatusBadge');
const scenarioStatus = document.getElementById('scenarioStatus');
const presentationBadge = document.getElementById('presentationBadge');
const debugPanel = document.getElementById('debugPanel');
const rawResponsePanel = document.getElementById('rawResponsePanel');
const modePresentationBtn = document.getElementById('modePresentation');
const modeDebugBtn = document.getElementById('modeDebug');
const orderModeCompleteBtn = document.getElementById('orderModeComplete');
const orderModeCancelBtn = document.getElementById('orderModeCancel');
const quoteInputGiveBtn = document.getElementById('quoteInputGive');
const quoteInputGetBtn = document.getElementById('quoteInputGet');
const demoHappyModeBtn = document.getElementById('demoHappyMode');
const demoErrorModeBtn = document.getElementById('demoErrorMode');

const kpis = {
  activeQuotes: document.getElementById('kpiActiveQuotes'),
  reservedOrders: document.getElementById('kpiReservedOrders'),
  completedToday: document.getElementById('kpiCompletedToday'),
  cancelledToday: document.getElementById('kpiCancelledToday'),
  grossMargin: document.getElementById('kpiGrossMargin'),
  staleAlerts: document.getElementById('kpiStaleAlerts'),
};

const fields = {
  baseUrl: document.getElementById('baseUrl'),
  tenantId: document.getElementById('tenantId'),
  clientRef: document.getElementById('clientRef'),
  cashierId: document.getElementById('cashierId'),
};

const quoteForm = document.getElementById('quoteForm');
const reserveForm = document.getElementById('reserveForm');
const completeForm = document.getElementById('completeForm');
const cancelForm = document.getElementById('cancelForm');

const quoteSummary = {
  quoteId: document.getElementById('qQuoteId'),
  office: document.getElementById('qOffice'),
  source: document.getElementById('qSource'),
  baseRate: document.getElementById('qBaseRate'),
  fixedRate: document.getElementById('qFixedRate'),
  fee: document.getElementById('qFee'),
  rule: document.getElementById('qRule'),
  expiresAt: document.getElementById('qExpiresAt'),
  status: document.getElementById('qStatus'),
  give: document.getElementById('qGive'),
  get: document.getElementById('qGet'),
};

const orderSummary = {
  orderId: document.getElementById('oOrderId'),
  orderRef: document.getElementById('oOrderRef'),
  status: document.getElementById('oStatus'),
  version: document.getElementById('oVersion'),
  expiresAt: document.getElementById('oExpiresAt'),
  quoteStatus: document.getElementById('oQuoteStatus'),
  heldCurrency: document.getElementById('oHeldCurrency'),
  heldAmount: document.getElementById('oHeldAmount'),
};

const scenarioView = {
  chips: document.getElementById('scenarioChips'),
  title: document.getElementById('scenarioTitle'),
  description: document.getElementById('scenarioDescription'),
  expected: document.getElementById('scenarioExpected'),
  nameInput: document.getElementById('scenarioName'),
};

const timelineRoot = document.getElementById('stateTimeline');
const auditRows = document.getElementById('auditRows');

const state = {
  currentScenarioId: DEMO_SCENARIOS[0].id,
  currentQuote: null,
  currentOrder: null,
  lastAction: 'idle',
  lastError: null,
  presentationMode: false,
  metrics: {
    activeQuotes: 0,
    reservedOrders: 0,
    completedToday: 0,
    cancelledToday: 0,
    staleAlerts: 0,
    grossMargin: 0,
  },
  orderActionMode: 'complete',
  quoteInputMode: 'GIVE',
  demoFlowMode: 'happy',
};

init();

function init() {
  loadConfig();
  loadState();
  renderScenarioChips();
  renderScenarioConsole();
  applyScenarioToForms(getScenarioById(state.currentScenarioId));
  renderTimeline();
  renderWhyAudit();
  renderKpiBar();
  renderPresentationMode();

  Object.values(fields).forEach((input) => input.addEventListener('change', saveConfig));

  quoteForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    await runQuote();
  });

  reserveForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    await runReserve();
  });

  completeForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    await runComplete();
  });

  cancelForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    await runCancel();
  });

  document.getElementById('copyQuoteId').addEventListener('click', copyQuoteId);
  document.getElementById('useQuoteInReserve').addEventListener('click', useQuoteInReserve);
  document.getElementById('loadScenario').addEventListener('click', () => {
    applyScenarioToForms(getScenarioById(state.currentScenarioId));
    renderScenarioConsole();
  });
  document.getElementById('runQuote').addEventListener('click', runQuote);
  document.getElementById('runReserve').addEventListener('click', runReserve);
  document.getElementById('runFullFlow').addEventListener('click', runFullFlow);
  document.getElementById('resetDemoState').addEventListener('click', resetDemoState);
  document.getElementById('loadSandboxDefaults').addEventListener('click', loadProductionLikeSandboxDefaults);
  document.getElementById('runSelectedDemo').addEventListener('click', () => {
    if (state.demoFlowMode === 'error') return runErrorDemo();
    return runHappyDemo();
  });
  modePresentationBtn.addEventListener('click', () => setPresentationMode(true));
  modeDebugBtn.addEventListener('click', () => setPresentationMode(false));
  orderModeCompleteBtn.addEventListener('click', () => setOrderActionMode('complete'));
  orderModeCancelBtn.addEventListener('click', () => setOrderActionMode('cancel'));
  quoteInputGiveBtn.addEventListener('click', () => setQuoteInputMode('GIVE'));
  quoteInputGetBtn.addEventListener('click', () => setQuoteInputMode('GET'));
  demoHappyModeBtn.addEventListener('click', () => setDemoFlowMode('happy'));
  demoErrorModeBtn.addEventListener('click', () => setDemoFlowMode('error'));

  setOrderActionMode(state.orderActionMode);
  setQuoteInputMode(quoteForm.elements.input_mode.value || state.quoteInputMode);
  setDemoFlowMode(state.demoFlowMode);
}

function renderScenarioChips() {
  scenarioView.chips.innerHTML = '';
  DEMO_SCENARIOS.forEach((scenario) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = `scenario-chip ${scenario.id === state.currentScenarioId ? 'active' : ''}`;
    button.innerHTML = `<span class="icon">${scenarioIcon(scenario.id)}</span><span>${scenario.title}</span>`;
    button.addEventListener('click', () => {
      state.currentScenarioId = scenario.id;
      applyScenarioToForms(scenario);
      renderScenarioChips();
      renderScenarioConsole();
      persistState();
    });
    scenarioView.chips.appendChild(button);
  });
}

function renderScenarioConsole() {
  const scenario = getScenarioById(state.currentScenarioId);
  scenarioView.title.textContent = scenario.title;
  scenarioView.description.textContent = scenario.description;
  scenarioView.expected.textContent = scenario.expected;
  scenarioView.nameInput.value = scenario.id;
  scenarioStatus.className = 'status neutral';
  scenarioStatus.textContent = `${scenario.title} loaded`;
}

function applyScenarioToForms(scenario) {
  quoteForm.elements.office_id.value = scenario.quote.office_id;
  quoteForm.elements.give_currency_id.value = scenario.quote.give_currency_id;
  quoteForm.elements.get_currency_id.value = scenario.quote.get_currency_id;
  setQuoteInputMode(scenario.quote.input_mode);
  quoteForm.elements.amount.value = scenario.quote.amount;

  reserveForm.elements.office_id.value = scenario.quote.office_id;
  reserveForm.elements.quote_id.value = scenario.reserve?.quote_id || 'quote-from-calculate';
  reserveForm.elements.side.value = scenario.reserve?.side || 'BUY';

  reserveForm.elements.give_amount.value = scenario.reserve?.give_amount || '0';
  reserveForm.elements.get_amount.value = scenario.reserve?.get_amount || '0';

  const giveMeta = currencyMeta(scenario.reserve?.give_currency_id || scenario.quote.give_currency_id);
  const getMeta = currencyMeta(scenario.reserve?.get_currency_id || scenario.quote.get_currency_id);

  reserveForm.elements.give_code.value = giveMeta.code;
  reserveForm.elements.give_network.value = giveMeta.network;
  reserveForm.elements.get_code.value = getMeta.code;
  reserveForm.elements.get_network.value = getMeta.network;

  fields.cashierId.value = fields.cashierId.value || 'cashier_bkk_01';
  completeForm.elements.cashier_id.value = fields.cashierId.value;
}

async function runQuote() {
  const payload = {
    office_id: asTrimmed(quoteForm.elements.office_id.value),
    give_currency_id: asTrimmed(quoteForm.elements.give_currency_id.value),
    get_currency_id: asTrimmed(quoteForm.elements.get_currency_id.value),
    input_mode: asTrimmed(quoteForm.elements.input_mode.value),
    amount: asTrimmed(quoteForm.elements.amount.value),
  };

  const { ok, json } = await sendRequest('/quotes/calculate', payload);
  state.lastAction = 'quote';

  if (ok) {
    state.currentQuote = normalizeQuote(json, payload);
    state.lastError = null;
    renderQuoteSummary(state.currentQuote);
    reserveForm.elements.quote_id.value = state.currentQuote.quote_id;
    renderQuotePill(state.currentQuote.status || 'active');
    if ((state.currentQuote.status || 'active') === 'active') {
      state.metrics.activeQuotes += 1;
    }
  } else {
    state.lastError = extractError(json);
    applyQuoteFailureToUI(state.lastError);
    if (state.lastError.code.toLowerCase().includes('stale') || state.lastError.message.toLowerCase().includes('stale')) {
      state.metrics.staleAlerts += 1;
    }
  }

  renderTimeline();
  renderWhyAudit();
  renderKpiBar();
  persistState();
}

async function runReserve() {
  const payload = {
    idempotency_key: asTrimmed(reserveForm.elements.idempotency_key.value),
    office_id: asTrimmed(reserveForm.elements.office_id.value),
    quote_id: asTrimmed(reserveForm.elements.quote_id.value),
    side: asTrimmed(reserveForm.elements.side.value),
    give: {
      amount: asTrimmed(reserveForm.elements.give_amount.value),
      currency: {
        code: asTrimmed(reserveForm.elements.give_code.value),
        network: asTrimmed(reserveForm.elements.give_network.value),
      },
    },
    get: {
      amount: asTrimmed(reserveForm.elements.get_amount.value),
      currency: {
        code: asTrimmed(reserveForm.elements.get_code.value),
        network: asTrimmed(reserveForm.elements.get_network.value),
      },
    },
  };

  const { ok, json } = await sendRequest('/orders/reserve', payload);
  state.lastAction = 'reserve';

  if (ok) {
    state.currentOrder = {
      order_id: json.order_id,
      order_ref: json.order_ref || 'n/a',
      status: (json.status || 'reserved').toLowerCase(),
      version: json.version,
      expires_at_ts: json.expires_at_ts,
      held_currency: quoteSummary.get.textContent,
      held_amount: reserveForm.elements.get_amount.value,
    };
    if (state.currentQuote) {
      state.currentQuote.status = 'consumed';
      quoteSummary.status.textContent = 'consumed';
      renderQuotePill('ready');
    }
    state.lastError = null;
    renderOrderSummary();
    state.metrics.reservedOrders += 1;
    state.metrics.activeQuotes = Math.max(0, state.metrics.activeQuotes - 1);
    state.metrics.grossMargin += estimateMarginFromQuote();
  } else {
    state.lastError = extractError(json);
    if (state.lastError.message.toLowerCase().includes('expired')) {
      renderQuotePill('expired');
      quoteSummary.status.textContent = 'expired';
    }
    if (state.lastError.message.toLowerCase().includes('consumed')) {
      renderQuotePill('failed');
      quoteSummary.status.textContent = 'consumed';
    }
  }

  renderTimeline();
  renderWhyAudit();
  renderKpiBar();
  persistState();
}

async function runComplete() {
  if (!completeForm.elements.order_id.value && state.currentOrder?.order_id) {
    completeForm.elements.order_id.value = state.currentOrder.order_id;
  }
  const payload = {
    idempotency_key: asTrimmed(completeForm.elements.idempotency_key.value),
    order_id: asTrimmed(completeForm.elements.order_id.value),
    expected_version: Number(completeForm.elements.expected_version.value),
    cashier_id: asTrimmed(completeForm.elements.cashier_id.value),
  };

  const { ok, json } = await sendRequest('/orders/complete', payload);
  state.lastAction = 'complete';

  if (ok) {
    state.currentOrder = {
      ...(state.currentOrder || {}),
      order_id: json.order_id || state.currentOrder?.order_id,
      status: (json.status || 'completed').toLowerCase(),
      version: json.version,
      completed_at_ts: json.completed_at_ts,
    };
    state.lastError = null;
    renderOrderSummary();
    state.metrics.completedToday += 1;
    state.metrics.reservedOrders = Math.max(0, state.metrics.reservedOrders - 1);
  } else {
    state.lastError = extractError(json);
  }

  renderTimeline();
  renderWhyAudit();
  renderKpiBar();
  persistState();
}

async function runCancel() {
  if (!cancelForm.elements.order_id.value && state.currentOrder?.order_id) {
    cancelForm.elements.order_id.value = state.currentOrder.order_id;
  }
  const payload = {
    idempotency_key: asTrimmed(cancelForm.elements.idempotency_key.value),
    order_id: asTrimmed(cancelForm.elements.order_id.value),
    expected_version: Number(cancelForm.elements.expected_version.value),
    reason: asTrimmed(cancelForm.elements.reason.value),
  };

  const { ok, json } = await sendRequest('/orders/cancel', payload);
  state.lastAction = 'cancel';

  if (ok) {
    state.currentOrder = {
      ...(state.currentOrder || {}),
      order_id: json.order_id || state.currentOrder?.order_id,
      status: (json.status || 'cancelled').toLowerCase(),
      version: json.version,
    };
    state.lastError = null;
    renderOrderSummary();
    state.metrics.cancelledToday += 1;
    state.metrics.reservedOrders = Math.max(0, state.metrics.reservedOrders - 1);
  } else {
    state.lastError = extractError(json);
  }

  renderTimeline();
  renderWhyAudit();
  renderKpiBar();
  persistState();
}

async function runFullFlow() {
  await runQuote();
  if (state.lastError) return;
  await runReserve();
  if (state.lastError) return;

  const scenario = getScenarioById(state.currentScenarioId);
  if (scenario.actionAfterReserve === 'complete') {
    completeForm.elements.order_id.value = state.currentOrder?.order_id || '';
    completeForm.elements.expected_version.value = String(state.currentOrder?.version || 1);
    await runComplete();
  }
  if (scenario.actionAfterReserve === 'cancel') {
    cancelForm.elements.order_id.value = state.currentOrder?.order_id || '';
    cancelForm.elements.expected_version.value = String(state.currentOrder?.version || 1);
    await runCancel();
  }
}

async function runHappyDemo() {
  state.currentScenarioId = 'buy-100-usdt';
  applyScenarioToForms(getScenarioById(state.currentScenarioId));
  renderScenarioChips();
  renderScenarioConsole();
  await runFullFlow();
}

async function runErrorDemo() {
  state.currentScenarioId = 'expired-quote';
  applyScenarioToForms(getScenarioById(state.currentScenarioId));
  renderScenarioChips();
  renderScenarioConsole();
  await runReserve();
}

function loadProductionLikeSandboxDefaults() {
  state.currentScenarioId = DEMO_SCENARIOS[0].id;
  state.currentQuote = null;
  state.currentOrder = null;
  state.lastAction = 'sandbox-defaults';
  state.lastError = null;
  state.metrics = {
    activeQuotes: 0,
    reservedOrders: 0,
    completedToday: 0,
    cancelledToday: 0,
    staleAlerts: 0,
    grossMargin: 0,
  };
  applyScenarioToForms(getScenarioById(state.currentScenarioId));
  renderScenarioChips();
  renderScenarioConsole();
  clearSummaries();
  renderTimeline();
  renderWhyAudit();
  renderKpiBar();
  setStatus('ok', 'Production-like sandbox defaults loaded');
  persistState();
}

function setPresentationMode(on) {
  state.presentationMode = Boolean(on);
  renderPresentationMode();
  persistState();
}

function setOrderActionMode(mode) {
  state.orderActionMode = mode === 'cancel' ? 'cancel' : 'complete';
  completeForm.classList.toggle('hidden', state.orderActionMode !== 'complete');
  cancelForm.classList.toggle('hidden', state.orderActionMode !== 'cancel');
  orderModeCompleteBtn.classList.toggle('active', state.orderActionMode === 'complete');
  orderModeCancelBtn.classList.toggle('active', state.orderActionMode === 'cancel');
  persistState();
}

function setQuoteInputMode(mode) {
  state.quoteInputMode = mode === 'GET' ? 'GET' : 'GIVE';
  quoteForm.elements.input_mode.value = state.quoteInputMode;
  quoteInputGiveBtn.classList.toggle('active', state.quoteInputMode === 'GIVE');
  quoteInputGetBtn.classList.toggle('active', state.quoteInputMode === 'GET');
  persistState();
}

function setDemoFlowMode(mode) {
  state.demoFlowMode = mode === 'error' ? 'error' : 'happy';
  demoHappyModeBtn.classList.toggle('active', state.demoFlowMode === 'happy');
  demoErrorModeBtn.classList.toggle('active', state.demoFlowMode === 'error');
  persistState();
}

function useQuoteInReserve() {
  if (!state.currentQuote) {
    renderLocalError('No quote yet. Run Quote first.');
    return;
  }
  reserveForm.elements.office_id.value = state.currentQuote.office_id;
  reserveForm.elements.quote_id.value = state.currentQuote.quote_id;
  reserveForm.elements.give_amount.value = state.currentQuote.give_amount;
  reserveForm.elements.get_amount.value = state.currentQuote.get_amount;

  const give = currencyMeta(state.currentQuote.give_currency_id);
  const get = currencyMeta(state.currentQuote.get_currency_id);
  reserveForm.elements.give_code.value = give.code;
  reserveForm.elements.give_network.value = give.network;
  reserveForm.elements.get_code.value = get.code;
  reserveForm.elements.get_network.value = get.network;
  setStatus('ok', 'Quote applied to reserve form');
}

async function copyQuoteId() {
  if (!state.currentQuote?.quote_id) {
    renderLocalError('No quote_id available to copy');
    return;
  }
  try {
    await navigator.clipboard.writeText(state.currentQuote.quote_id);
    setStatus('ok', 'quote_id copied');
  } catch {
    setStatus('warn', 'clipboard unavailable');
  }
}

function renderQuoteSummary(quote) {
  quoteSummary.quoteId.textContent = quote.quote_id || '—';
  quoteSummary.office.textContent = officeLabel(quote.office_id);
  quoteSummary.source.textContent = sourceLabel(quote.source_name);
  quoteSummary.baseRate.textContent = quote.base_rate || '—';
  quoteSummary.fixedRate.textContent = quote.fixed_rate || '—';
  quoteSummary.fee.textContent = quote.fee_amount || '—';
  quoteSummary.rule.textContent = quote.applied_rule || 'n/a';
  quoteSummary.expiresAt.textContent = toIso(quote.expires_at_ts);
  quoteSummary.status.textContent = quote.status || 'active';
  quoteSummary.give.textContent = `${quote.give_amount} ${currencyMeta(quote.give_currency_id).label}`;
  quoteSummary.get.textContent = `${quote.get_amount} ${currencyMeta(quote.get_currency_id).label}`;
  renderQuotePill(quote.status || 'active');
}

function renderOrderSummary() {
  const order = state.currentOrder;
  if (!order) return;

  orderSummary.orderId.textContent = order.order_id || '—';
  orderSummary.orderRef.textContent = order.order_ref || 'n/a';
  orderSummary.status.textContent = order.status || '—';
  orderSummary.version.textContent = String(order.version ?? '—');
  orderSummary.expiresAt.textContent = toIso(order.expires_at_ts || order.completed_at_ts || '');
  orderSummary.quoteStatus.textContent = quoteSummary.status.textContent;
  orderSummary.heldCurrency.textContent = quoteSummary.get.textContent;
  orderSummary.heldAmount.textContent = reserveForm.elements.get_amount.value || '—';

  renderOrderPill(order.status || 'pending');

  completeForm.elements.order_id.value = order.order_id || completeForm.elements.order_id.value;
  cancelForm.elements.order_id.value = order.order_id || cancelForm.elements.order_id.value;
}

function renderTimeline() {
  const steps = {
    quote: 'Quote: not started',
    reserve: 'Order reserve: not started',
    complete: 'Order complete: not started',
    cancel: 'Order cancel: not started',
  };
  const classes = {
    quote: 'neutral',
    reserve: 'neutral',
    complete: 'neutral',
    cancel: 'neutral',
  };

  if (state.currentQuote) {
    steps.quote = `Quote: ${state.currentQuote.status || 'active'} (${state.currentQuote.quote_id || 'n/a'})`;
    classes.quote = mapStatusClass(state.currentQuote.status || 'active');
  }
  if (state.currentOrder) {
    steps.reserve = `Order reserve: ${state.currentOrder.status || 'reserved'} (${state.currentOrder.order_id || 'n/a'})`;
    classes.reserve = mapStatusClass(state.currentOrder.status || 'reserved');
  }
  if (state.currentOrder?.status === 'completed') {
    steps.complete = `Order complete: completed (v${state.currentOrder.version ?? 'n/a'})`;
    classes.complete = 'ok';
  }
  if (state.currentOrder?.status === 'cancelled') {
    steps.cancel = `Order cancel: cancelled (v${state.currentOrder.version ?? 'n/a'})`;
    classes.cancel = 'warn';
  }
  if (state.lastError) {
    const target = state.lastAction === 'quote' ? 'quote' : state.lastAction;
    if (steps[target]) {
      steps[target] = `${steps[target]} • error: ${state.lastError.code}`;
      classes[target] = 'err';
    }
  }

  [...timelineRoot.querySelectorAll('li')].forEach((node) => {
    const step = node.getAttribute('data-step');
    node.innerHTML = `<span class="timeline-line-main">${steps[step]}</span><span class="timeline-meta">step: ${step}</span>`;
    node.className = classes[step] || 'neutral';
  });
}

function renderWhyAudit() {
  const scenario = getScenarioById(state.currentScenarioId);
  const rows = [
    { symbol: '◎', label: 'Scenario', value: scenario.title },
    { symbol: '⚙', label: 'Applied rule', value: state.currentQuote?.applied_rule || 'n/a' },
    { symbol: '％', label: 'Margin / fee', value: `${state.currentQuote?.base_rate || '—'} -> ${state.currentQuote?.fixed_rate || '—'} / fee ${state.currentQuote?.fee_amount || '—'}` },
    { symbol: '◈', label: 'Held asset', value: quoteSummary.get.textContent || '—' },
    { symbol: '⌁', label: 'Quote source', value: sourceLabel(state.currentQuote?.source_name) },
    { symbol: '→', label: 'Last action', value: state.lastAction },
    { symbol: '!', label: 'Last error', value: state.lastError ? `${state.lastError.code}: ${state.lastError.message}` : 'none', danger: true },
    { symbol: '👤', label: 'Cashier', value: cashierLabel(asTrimmed(fields.cashierId.value)) },
    { symbol: '#', label: 'HTTP status', value: statusBadge.textContent },
  ];

  auditRows.innerHTML = '';
  rows.forEach((rowData) => {
    const row = document.createElement('div');
    row.className = 'audit-row';
    row.innerHTML = `
      <span class="audit-line-main">
        <span class="insight-symbol ${rowData.danger ? 'danger' : ''}">${rowData.symbol}</span>
        ${rowData.label}
      </span>
      <strong class="audit-meta">${rowData.value || '—'}</strong>
    `;
    auditRows.appendChild(row);
  });
}

function renderKpiBar() {
  kpis.activeQuotes.textContent = String(state.metrics.activeQuotes);
  kpis.reservedOrders.textContent = String(state.metrics.reservedOrders);
  kpis.completedToday.textContent = String(state.metrics.completedToday);
  kpis.cancelledToday.textContent = String(state.metrics.cancelledToday);
  kpis.staleAlerts.textContent = String(state.metrics.staleAlerts);
  kpis.grossMargin.textContent = state.metrics.grossMargin ? state.metrics.grossMargin.toFixed(2) : '—';
}

function renderPresentationMode() {
  const on = state.presentationMode;
  document.body.classList.toggle('presentation', on);
  debugPanel.style.display = on ? 'none' : '';
  rawResponsePanel.open = !on && rawResponsePanel.open;
  modePresentationBtn.classList.toggle('active', on);
  modeDebugBtn.classList.toggle('active', !on);
  presentationBadge.className = `status ${on ? 'ok' : 'neutral'}`;
  presentationBadge.textContent = `presentation mode: ${on ? 'on' : 'off'}`;
}

function estimateMarginFromQuote() {
  if (!state.currentQuote) return 0;
  const base = Number(state.currentQuote.base_rate || 0);
  const fixed = Number(state.currentQuote.fixed_rate || 0);
  const fee = Number(state.currentQuote.fee_amount || 0);
  const notional = Number(state.currentQuote.give_amount || 0);
  if (Number.isNaN(base) || Number.isNaN(fixed) || Number.isNaN(fee) || Number.isNaN(notional)) return 0;
  return Math.max(0, (fixed - base) * notional + fee);
}

function applyQuoteFailureToUI(error) {
  const code = error.code.toLowerCase();
  const message = error.message.toLowerCase();

  if (code.includes('stale') || message.includes('stale')) {
    renderQuotePill('expired');
    quoteSummary.status.textContent = 'stale';
  } else if (code.includes('expired') || message.includes('expired')) {
    renderQuotePill('expired');
    quoteSummary.status.textContent = 'expired';
  } else {
    renderQuotePill('failed');
    quoteSummary.status.textContent = 'rejected';
  }
}

function normalizeQuote(json, request) {
  const give = json.give || {};
  const get = json.get || {};
  return {
    quote_id: json.quote_id || json.quoteId || '',
    office_id: request.office_id,
    status: (json.status || 'active').toLowerCase(),
    base_rate: json.base_rate || json.baseRate || '—',
    fixed_rate: json.fixed_rate || json.fixedRate || '—',
    fee_amount: json.fee_amount || json.feeAmount || '0',
    source_name: json.source_name || json.sourceName || 'manual',
    expires_at_ts: json.expires_at_ts || json.expiresAtTs || '',
    applied_rule: json.applied_rule || json.appliedRule || 'n/a',
    give_amount: give.amount || '—',
    get_amount: get.amount || '—',
    give_currency_id: give.currency_id || give.currencyId || request.give_currency_id,
    get_currency_id: get.currency_id || get.currencyId || request.get_currency_id,
  };
}

async function sendRequest(path, payload) {
  const ctx = validateConnection();
  if (!ctx) return { ok: false, json: { error: { code: 'LOCAL_VALIDATION', message: 'Missing connection config' } } };

  setStatus('idle', `sending ${path}`);

  try {
    const resp = await fetch(`${ctx.baseUrl}${path}`, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        'x-tenant-id': ctx.tenantId,
        'x-client-ref': ctx.clientRef,
      },
      body: JSON.stringify(payload),
    });
    const json = await safeJson(resp);
    const badge = resp.ok ? 'ok' : resp.status >= 500 ? 'err' : 'warn';
    setStatus(badge, `${resp.status} ${resp.statusText}`);
    responseNode.textContent = JSON.stringify({ request: payload, response: json }, null, 2);
    return { ok: resp.ok, json };
  } catch (error) {
    setStatus('err', 'network error');
    const json = { error: { code: 'NETWORK_ERROR', message: String(error) } };
    responseNode.textContent = JSON.stringify({ request: payload, response: json }, null, 2);
    return { ok: false, json };
  }
}

function validateConnection() {
  const baseUrl = asTrimmed(fields.baseUrl.value).replace(/\/$/, '');
  const tenantId = asTrimmed(fields.tenantId.value);
  const clientRef = asTrimmed(fields.clientRef.value);
  if (!baseUrl || !tenantId || !clientRef) {
    renderLocalError('Set base URL, tenant_id and client_ref first');
    return null;
  }
  return { baseUrl, tenantId, clientRef };
}

function extractError(json) {
  return {
    code: String(json?.error?.code || json?.code || 'UNKNOWN_ERROR'),
    message: String(json?.error?.message || json?.message || 'unknown error'),
  };
}

function resetDemoState() {
  state.currentQuote = null;
  state.currentOrder = null;
  state.lastAction = 'idle';
  state.lastError = null;
  state.currentScenarioId = DEMO_SCENARIOS[0].id;
  state.metrics = {
    activeQuotes: 0,
    reservedOrders: 0,
    completedToday: 0,
    cancelledToday: 0,
    staleAlerts: 0,
    grossMargin: 0,
  };
  localStorage.removeItem(LS_STATE_KEY);

  renderScenarioChips();
  renderScenarioConsole();
  applyScenarioToForms(getScenarioById(state.currentScenarioId));
  clearSummaries();
  renderTimeline();
  renderWhyAudit();
  renderKpiBar();
  renderPresentationMode();
  setStatus('ok', 'Demo state reset. Sandbox defaults loaded.');
}

function clearSummaries() {
  Object.values(quoteSummary).forEach((node) => {
    node.textContent = '—';
  });
  Object.values(orderSummary).forEach((node) => {
    node.textContent = '—';
  });
  renderQuotePill('draft');
  renderOrderPill('pending');
}

function setStatus(kind, text) {
  statusBadge.className = `status ${kind}`;
  statusBadge.textContent = text;
}

function setBadge(node, status) {
  const klass = mapStatusClass(status);
  node.className = `status ${klass}`;
  node.textContent = status;
}

function mapStatusClass(status) {
  const value = String(status || '').toLowerCase();
  if (value.includes('active') || value.includes('completed') || value.includes('ok')) return 'ok';
  if (value.includes('consumed') || value.includes('reserved') || value.includes('cancelled') || value.includes('warn') || value.includes('stale')) return 'warn';
  if (value.includes('expired') || value.includes('error') || value.includes('failed') || value.includes('err')) return 'err';
  return 'neutral';
}

function renderQuotePill(status) {
  const value = String(status || '').toLowerCase();
  if (!status || value === 'neutral') {
    quoteStatusBadge.className = 'status neutral';
    quoteStatusBadge.textContent = '● Draft';
    return;
  }
  if (value.includes('active') || value.includes('ready')) {
    quoteStatusBadge.className = 'status ok';
    quoteStatusBadge.textContent = '● Ready';
    return;
  }
  if (value.includes('expired') || value.includes('stale')) {
    quoteStatusBadge.className = 'status warn';
    quoteStatusBadge.textContent = '● Expired';
    return;
  }
  if (value.includes('failed') || value.includes('error') || value.includes('rejected')) {
    quoteStatusBadge.className = 'status err';
    quoteStatusBadge.textContent = '● Failed';
    return;
  }
  quoteStatusBadge.className = 'status neutral';
  quoteStatusBadge.textContent = '● Draft';
}

function renderOrderPill(status) {
  const value = String(status || '').toLowerCase();
  if (value.includes('completed')) {
    orderStatusBadge.className = 'status ok';
    orderStatusBadge.textContent = 'Order · Completed';
    return;
  }
  if (value.includes('reserved')) {
    orderStatusBadge.className = 'status warn';
    orderStatusBadge.textContent = 'Order · Reserved';
    return;
  }
  if (value.includes('cancelled')) {
    orderStatusBadge.className = 'status err';
    orderStatusBadge.textContent = 'Order · Cancelled';
    return;
  }
  if (value.includes('failed') || value.includes('error')) {
    orderStatusBadge.className = 'status err';
    orderStatusBadge.textContent = 'Order · Failed';
    return;
  }
  orderStatusBadge.className = 'status neutral';
  orderStatusBadge.textContent = 'Order · Pending';
}

function saveConfig() {
  localStorage.setItem(
    LS_CONFIG_KEY,
    JSON.stringify({
      baseUrl: fields.baseUrl.value,
      tenantId: fields.tenantId.value,
      clientRef: fields.clientRef.value,
      cashierId: fields.cashierId.value,
    }),
  );
}

function loadConfig() {
  const raw = localStorage.getItem(LS_CONFIG_KEY);
  if (!raw) return;
  try {
    const cfg = JSON.parse(raw);
    fields.baseUrl.value = cfg.baseUrl || fields.baseUrl.value;
    fields.tenantId.value = cfg.tenantId || fields.tenantId.value;
    fields.clientRef.value = cfg.clientRef || fields.clientRef.value;
    fields.cashierId.value = cfg.cashierId || fields.cashierId.value;
  } catch {
    // ignore invalid config
  }
}

function persistState() {
  localStorage.setItem(
    LS_STATE_KEY,
    JSON.stringify({
      currentScenarioId: state.currentScenarioId,
      currentQuote: state.currentQuote,
      currentOrder: state.currentOrder,
      lastAction: state.lastAction,
      lastError: state.lastError,
      presentationMode: state.presentationMode,
      metrics: state.metrics,
      orderActionMode: state.orderActionMode,
      quoteInputMode: state.quoteInputMode,
      demoFlowMode: state.demoFlowMode,
    }),
  );
}

function loadState() {
  const raw = localStorage.getItem(LS_STATE_KEY);
  if (!raw) return;
  try {
    const saved = JSON.parse(raw);
    state.currentScenarioId = saved.currentScenarioId || state.currentScenarioId;
    state.currentQuote = saved.currentQuote || null;
    state.currentOrder = saved.currentOrder || null;
    state.lastAction = saved.lastAction || 'idle';
    state.lastError = saved.lastError || null;
    state.presentationMode = Boolean(saved.presentationMode);
    state.metrics = saved.metrics || state.metrics;
    state.orderActionMode = saved.orderActionMode || state.orderActionMode;
    state.quoteInputMode = saved.quoteInputMode || state.quoteInputMode;
    state.demoFlowMode = saved.demoFlowMode || state.demoFlowMode;
    if (state.currentQuote) {
      renderQuoteSummary(state.currentQuote);
      renderQuotePill(state.currentQuote.status || 'active');
    }
    if (state.currentOrder) {
      renderOrderSummary();
    }
    renderKpiBar();
    renderPresentationMode();
    setOrderActionMode(state.orderActionMode);
    setQuoteInputMode(state.quoteInputMode);
    setDemoFlowMode(state.demoFlowMode);
  } catch {
    // ignore invalid saved state
  }
}

function scenarioIcon(id) {
  const icons = {
    'buy-100-usdt': '↗',
    'get-10000-thb': '฿',
    'vip-tier': '★',
    'stale-rate': '⏱',
    'expired-quote': '⌛',
    'insufficient-liquidity': '⚠',
  };
  return icons[id] || '•';
}

function renderLocalError(message) {
  setStatus('warn', 'local validation');
  responseNode.textContent = JSON.stringify({ error: message }, null, 2);
}

function asTrimmed(value) {
  return String(value ?? '').trim();
}

function toIso(value) {
  if (!value) return '—';
  const number = Number(value);
  if (Number.isNaN(number)) return String(value);
  return new Date(number * 1000).toISOString();
}

async function safeJson(resp) {
  const raw = await resp.text();
  if (!raw) return {};
  try {
    return JSON.parse(raw);
  } catch {
    return { raw };
  }
}
