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
make pilot-bootstrap
```

`make pilot-bootstrap` auto-loads `.env`, uses localhost Postgres for host-side `psql`, and creates `${POSTGRES_DB}` when missing.

## 4. Open apps

- BFF: http://localhost:3000
- Backoffice: http://localhost:4173

## 5. Login

Use the bootstrap admin credentials from `.env`.

## 6. Run pilot acceptance

Follow `docs/PILOT_ACCEPTANCE_CHECKLIST.md`.
