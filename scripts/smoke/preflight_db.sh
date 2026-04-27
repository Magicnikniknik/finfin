#!/usr/bin/env bash
set -euo pipefail

: "${DATABASE_URL:?DATABASE_URL is required}"

attempts="${SMOKE_DB_WAIT_ATTEMPTS:-30}"
sleep_s="${SMOKE_DB_WAIT_SLEEP:-2}"

for i in $(seq 1 "$attempts"); do
  if psql "$DATABASE_URL" -c "select 1;" >/dev/null 2>&1; then
    echo "database is reachable"
    exit 0
  fi
  echo "waiting for database (${i}/${attempts})..."
  sleep "$sleep_s"
done

echo "database did not become reachable in time" >&2
exit 1
