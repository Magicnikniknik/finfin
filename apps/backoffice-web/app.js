const responseNode = document.getElementById('responseJson');
const statusBadge = document.getElementById('statusBadge');

const fields = {
  baseUrl: document.getElementById('baseUrl'),
  tenantId: document.getElementById('tenantId'),
  clientRef: document.getElementById('clientRef'),
};

const LS_KEY = 'finfin.backoffice.config';
loadConfig();

for (const input of Object.values(fields)) {
  input.addEventListener('change', saveConfig);
}

document.getElementById('reserveForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  const f = new FormData(e.target);
  const payload = {
    idempotency_key: f.get('idempotency_key'),
    office_id: f.get('office_id'),
    quote_id: f.get('quote_id'),
    side: f.get('side'),
    give: {
      amount: f.get('give_amount'),
      currency: { code: f.get('give_code'), network: f.get('give_network') },
    },
    get: {
      amount: f.get('get_amount'),
      currency: { code: f.get('get_code'), network: f.get('get_network') },
    },
  };
  await sendOrder('reserve', payload);
});

document.getElementById('completeForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  const f = new FormData(e.target);
  const payload = {
    idempotency_key: f.get('idempotency_key'),
    order_id: f.get('order_id'),
    expected_version: Number(f.get('expected_version')),
    cashier_id: f.get('cashier_id'),
  };
  await sendOrder('complete', payload);
});

document.getElementById('cancelForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  const f = new FormData(e.target);
  const payload = {
    idempotency_key: f.get('idempotency_key'),
    order_id: f.get('order_id'),
    expected_version: Number(f.get('expected_version')),
    reason: f.get('reason'),
  };
  await sendOrder('cancel', payload);
});

async function sendOrder(action, payload) {
  const baseUrl = fields.baseUrl.value.trim().replace(/\/$/, '');
  const tenantId = fields.tenantId.value.trim();
  const clientRef = fields.clientRef.value.trim();

  if (!baseUrl) return renderLocalError('Set BFF base URL first');
  if (!tenantId) return renderLocalError('Set tenant_id first');
  if (!clientRef) return renderLocalError('Set client_ref first');

  const url = `${baseUrl}/orders/${action}`;
  setStatus('idle', `sending ${action}`);

  try {
    const resp = await fetch(url, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        'x-tenant-id': tenantId,
        'x-client-ref': clientRef,
      },
      body: JSON.stringify(payload),
    });

    const json = await safeJson(resp);
    const tag = resp.ok ? 'ok' : resp.status >= 500 ? 'err' : 'warn';
    setStatus(tag, `${resp.status} ${resp.statusText}`);
    responseNode.textContent = JSON.stringify({ request: payload, response: json }, null, 2);
  } catch (error) {
    setStatus('err', 'network error');
    responseNode.textContent = JSON.stringify(
      {
        request: payload,
        error: String(error),
      },
      null,
      2,
    );
  }
}

function renderLocalError(message) {
  setStatus('warn', 'local validation');
  responseNode.textContent = JSON.stringify({ error: message }, null, 2);
}

async function safeJson(resp) {
  try {
    return await resp.json();
  } catch {
    return { raw: await resp.text() };
  }
}

function setStatus(kind, text) {
  statusBadge.className = `status ${kind}`;
  statusBadge.textContent = text;
}

function saveConfig() {
  localStorage.setItem(
    LS_KEY,
    JSON.stringify({
      baseUrl: fields.baseUrl.value,
      tenantId: fields.tenantId.value,
      clientRef: fields.clientRef.value,
    }),
  );
}

function loadConfig() {
  const raw = localStorage.getItem(LS_KEY);
  if (!raw) return;
  try {
    const cfg = JSON.parse(raw);
    fields.baseUrl.value = cfg.baseUrl || fields.baseUrl.value;
    fields.tenantId.value = cfg.tenantId || '';
    fields.clientRef.value = cfg.clientRef || '';
  } catch {
    // ignore invalid saved state
  }
}
