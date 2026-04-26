#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

require_cmd "psql" "install psql (postgres client) to run SQL smoke scripts"

: "${DATABASE_URL:?DATABASE_URL is required}"

psql "${DATABASE_URL}" -f migrations/0001_core_schema.sql
psql "${DATABASE_URL}" -f migrations/0002_quote_snapshots.sql
psql "${DATABASE_URL}" -f migrations/0003_account_wiring.sql
psql "${DATABASE_URL}" -f migrations/0004_pricing_engine.sql
psql "${DATABASE_URL}" -f migrations/0005_cash_shifts.sql
psql "${DATABASE_URL}" -f migrations/0006_auth_schema.sql
psql "${DATABASE_URL}" -f migrations/0007_audit_logs.sql

psql "${DATABASE_URL}" -f scripts/smoke/reset.sql
psql "${DATABASE_URL}" -f scripts/smoke/seed_pricing.sql
psql "${DATABASE_URL}" -f scripts/smoke/seed_demo.sql
psql "${DATABASE_URL}" -f scripts/smoke/seed_reserve.sql
psql "${DATABASE_URL}" -f scripts/smoke/seed_fast.sql

echo "bootstrap complete: schema + pricing + realistic demo seed applied"
