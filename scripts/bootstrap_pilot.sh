#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

POSTGRES_DB="${POSTGRES_DB:-finfin}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"

PILOT_DATABASE_URL="${PILOT_DATABASE_URL:-postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable}"
PILOT_ADMIN_DATABASE_URL="${PILOT_ADMIN_DATABASE_URL:-postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/postgres?sslmode=disable}"

: "${ADMIN_BOOTSTRAP_TENANT_ID:?ADMIN_BOOTSTRAP_TENANT_ID is required}"
: "${ADMIN_BOOTSTRAP_LOGIN:?ADMIN_BOOTSTRAP_LOGIN is required}"
: "${ADMIN_BOOTSTRAP_PASSWORD:?ADMIN_BOOTSTRAP_PASSWORD is required}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[error] required command not found: $1"
    exit 1
  fi
}

require_cmd psql
require_cmd bash

if [[ ! "$POSTGRES_DB" =~ ^[a-zA-Z0-9_]+$ ]]; then
  echo "[error] invalid POSTGRES_DB: $POSTGRES_DB"
  exit 1
fi

echo "[step] ensuring database exists: $POSTGRES_DB"
db_exists="$(psql "$PILOT_ADMIN_DATABASE_URL" -tAc "SELECT 1 FROM pg_database WHERE datname='${POSTGRES_DB}'")"
if [[ "$db_exists" != "1" ]]; then
  psql "$PILOT_ADMIN_DATABASE_URL" -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${POSTGRES_DB}"
fi

echo "[step] applying migrations"
psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/migrations/0001_core_schema.sql"
psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/migrations/0002_quote_snapshots.sql"
psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/migrations/0003_account_wiring.sql"
psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/migrations/0004_pricing_engine.sql"
psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/migrations/0005_cash_shifts.sql"
psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/migrations/0006_auth_schema.sql"
psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/migrations/0007_audit_logs.sql"

echo "[step] bootstrapping admin user"
bash "$ROOT_DIR/scripts/bootstrap_admin.sh"

if [[ "${SANDBOX_SEED_DEMO:-false}" == "true" ]]; then
  echo "[step] loading production-like sandbox demo seed"
  psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/scripts/smoke/reset.sql"
  psql "$PILOT_DATABASE_URL" -f "$ROOT_DIR/scripts/smoke/seed_demo.sql"
fi

echo
echo "[ok] finfin pilot box is ready"
echo "BFF:        http://localhost:${BFF_PORT:-3000}"
echo "Backoffice: http://localhost:${BACKOFFICE_PORT:-4173}"
echo "Login:      ${ADMIN_BOOTSTRAP_LOGIN}"
