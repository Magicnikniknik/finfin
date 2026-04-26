# Backoffice Sandbox UI

Scenario-driven operator demo UI for production-like sandbox flows on top of Nest BFF.

## Supported HTTP flows

- `POST /quotes/calculate`
- `POST /orders/reserve`
- `POST /orders/complete`
- `POST /orders/cancel`

## Run

```bash
cd /workspace/finfin/apps/backoffice-web
python3 -m http.server 4173
```

Open `http://localhost:4173`.

## File structure

- `index.html` — scenario console + workspace layout
- `app.js` — orchestration (network, state, KPI, timeline, presentation mode, reset)
- `scenarios.js` — demo scenario presets
- `demo-catalog.js` — human labels for offices/currencies/cashiers/sources
- `styles.css` — operator-console styles

## Production-like Sandbox demo flow

1. Fill connection config (`base URL`, `tenant_id`, `client_ref`, `cashier`).
2. Pick scenario in **Scenario Console**.
3. Click **Load Scenario**.
4. Click **Run Quote**.
5. Click **Run Reserve**.
6. Click **Run Full Flow** (or manual Complete/Cancel).
7. Observe:
   - KPI top bar
   - Quote Summary
   - Order Summary
   - State Timeline
   - Why/Audit Panel
8. Use **Reset Demo State** before next demo meeting.
9. Use **Hide Debug / Show Debug** for presentation mode.

## Included scenarios

- Buy 100 USDT
- Get 10,000 THB
- VIP tier
- Stale rate
- Expired quote
- Insufficient liquidity

## Notes

- Debug panel keeps latest request/response JSON.
- UI is static and intentionally backend-agnostic (no build step required).
- Session shortcuts are available in **Demo Session Panel** (`Load Sandbox`, `Happy Path Demo`, `Error Path Demo`).
- Suggested talk-track for live demos: `docs/demo_walkthrough.md`.
