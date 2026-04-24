# Backoffice test UI (minimal)

Static web console for manual smoke of Nest BFF endpoints:

- `POST /orders/reserve`
- `POST /orders/complete`
- `POST /orders/cancel`

## Run

```bash
cd /workspace/finfin/apps/backoffice-web
python3 -m http.server 4173
```

Open `http://localhost:4173`.

Set:

- BFF base URL (default `http://localhost:3000`)
- `tenant_id`
- `client_ref`

Then use the three forms and inspect JSON response panel.
