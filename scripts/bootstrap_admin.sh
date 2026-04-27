#!/usr/bin/env bash
set -euo pipefail

: "${DATABASE_URL:?DATABASE_URL is required}"
: "${ADMIN_BOOTSTRAP_TENANT_ID:?ADMIN_BOOTSTRAP_TENANT_ID is required}"
: "${ADMIN_BOOTSTRAP_LOGIN:?ADMIN_BOOTSTRAP_LOGIN is required}"
: "${ADMIN_BOOTSTRAP_PASSWORD:?ADMIN_BOOTSTRAP_PASSWORD is required}"

ADMIN_BOOTSTRAP_ROLE="${ADMIN_BOOTSTRAP_ROLE:-owner}"
ADMIN_BOOTSTRAP_STATUS="${ADMIN_BOOTSTRAP_STATUS:-active}"
ADMIN_BOOTSTRAP_DISPLAY_NAME="${ADMIN_BOOTSTRAP_DISPLAY_NAME:-Sandbox Owner}"
ADMIN_BOOTSTRAP_OFFICE_ID="${ADMIN_BOOTSTRAP_OFFICE_ID:-}"

PASSWORD_HASH="$(node -e "const {randomBytes,scryptSync}=require('crypto'); const pass=process.argv[1]; const salt=randomBytes(16); const hash=scryptSync(pass,salt,64); process.stdout.write('scrypt$'+salt.toString('base64')+'$'+hash.toString('base64'));" "$ADMIN_BOOTSTRAP_PASSWORD")"

psql "$DATABASE_URL" \
  -v tenant_id="$ADMIN_BOOTSTRAP_TENANT_ID" \
  -v office_id="$ADMIN_BOOTSTRAP_OFFICE_ID" \
  -v login="$ADMIN_BOOTSTRAP_LOGIN" \
  -v password_hash="$PASSWORD_HASH" \
  -v role="$ADMIN_BOOTSTRAP_ROLE" \
  -v status="$ADMIN_BOOTSTRAP_STATUS" \
  -v display_name="$ADMIN_BOOTSTRAP_DISPLAY_NAME" <<'SQL'
INSERT INTO auth.users (
  tenant_id,
  office_id,
  login,
  password_hash,
  role,
  status,
  display_name
)
VALUES (
  :'tenant_id'::uuid,
  NULLIF(:'office_id', '')::uuid,
  :'login',
  :'password_hash',
  :'role',
  :'status',
  :'display_name'
)
ON CONFLICT (tenant_id, login)
DO UPDATE SET
  office_id = EXCLUDED.office_id,
  password_hash = EXCLUDED.password_hash,
  role = EXCLUDED.role,
  status = EXCLUDED.status,
  display_name = EXCLUDED.display_name,
  updated_at = now();
SQL

echo "Admin user seeded/updated for tenant ${ADMIN_BOOTSTRAP_TENANT_ID}: ${ADMIN_BOOTSTRAP_LOGIN}"
