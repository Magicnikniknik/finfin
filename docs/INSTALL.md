# Finfin Pilot Install

## 1. Prepare env

```bash
cp .env.example .env
```

Update secrets in `.env`.

## 2. Start services

```bash
make pilot-up
```

## 3. Bootstrap database and admin

```bash
set -a
source .env
set +a
export PILOT_DATABASE_URL=postgres://postgres:postgres@localhost:5432/finfin?sslmode=disable
make pilot-bootstrap
```

## 4. Open apps

- BFF API: http://localhost:3000
- Backoffice: http://localhost:4173

> `localhost:3000` is an API service (not a separate web UI).  
> Quick check in browser or curl: `http://localhost:3000/healthz`

## 5. Verify that everything is actually running

```bash
docker compose -f docker-compose.pilot.yml --env-file .env ps
curl -sS http://localhost:3000/
curl -sS http://localhost:3000/healthz
curl -sS http://localhost:3000/readyz
```

Expected:
- `docker compose ... ps` shows `finfin-postgres`, `finfin-grpc`, `finfin-bff`, `finfin-backoffice` as running.
- `/healthz` returns `{"status":"ok"}`.
- `/readyz` returns `{"status":"ready"}`.

## 6. Run core API scenarios (auth + quote + order flow)

Use the bootstrap user from `.env` (default demo login shown below):

```bash
ACCESS_TOKEN=$(
  curl -sS -X POST http://localhost:3000/auth/login \
    -H 'content-type: application/json' \
    -d '{"tenant_id":"11111111-1111-1111-1111-111111111111","login":"owner_demo","password":"owner_demo_password"}' \
    | python3 -c 'import json,sys; print(json.load(sys.stdin)["access_token"])'
)

echo "$ACCESS_TOKEN" | head -c 24; echo
```

Calculate quote:

```bash
curl -sS -X POST http://localhost:3000/quotes/calculate \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -H 'content-type: application/json' \
  -d '{
    "tenant_id":"11111111-1111-1111-1111-111111111111",
    "office_id":"22222222-2222-2222-2222-222222222222",
    "base_asset":"USD",
    "quote_asset":"USDT",
    "side":"BUY",
    "amount":"100.00"
  }'
```

Then continue with order lifecycle:

- reserve: `POST /orders/reserve`
- complete: `POST /orders/complete`
- cancel: `POST /orders/cancel`

Use `docs/PILOT_ACCEPTANCE_CHECKLIST.md` as the full pass/fail matrix.

## 7. Login

Use the bootstrap admin credentials from `.env`.

## 8. Run pilot acceptance

Follow `docs/PILOT_ACCEPTANCE_CHECKLIST.md`.
