#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

require_cmd "python3" "python3 is required for JSON assertions"

stamp_ms() { date +%s%3N; }

run_step() {
  local name="$1"
  shift
  local started ended elapsed
  started="$(stamp_ms)"
  "$@"
  ended="$(stamp_ms)"
  elapsed=$((ended - started))
  echo "timing.${name}=${elapsed}ms" >&2
}

assert_json_status() {
  local json="$1"
  local expected="$2"
  JSON_PAYLOAD="$json" EXPECTED_STATUS="$expected" python3 - <<'PY'
import json
import os
import sys

payload = os.environ['JSON_PAYLOAD']
expected = os.environ['EXPECTED_STATUS'].upper()
obj = json.loads(payload)
status = str(obj.get('status', '')).upper()
if status != expected:
    print(f"status assertion failed: expected={expected} got={status}", file=sys.stderr)
    sys.exit(1)
if not obj.get('orderId'):
    print("orderId is empty", file=sys.stderr)
    sys.exit(1)
PY
}

assert_order_status_sql() {
  local order_id="$1"
  local expected_status="$2"
  local got
  got="$(psql "$DATABASE_URL" -Atqc "select status from core.orders where id='${order_id}'::uuid")"
  if [[ "${got}" != "${expected_status}" ]]; then
    echo "order status assertion failed for ${order_id}: expected=${expected_status}, got=${got}" >&2
    exit 1
  fi
}

echo "== assertive smoke full-cycle =="
run_step bootstrap "${SCRIPT_DIR}/bootstrap.sh"

reserve_json="$(run_step reserve "${SCRIPT_DIR}/reserve.sh")"
assert_json_status "${reserve_json}" "RESERVED"

complete_json="$(run_step complete env ORDER_ID="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" EXPECTED_VERSION="1" "${SCRIPT_DIR}/complete.sh")"
assert_json_status "${complete_json}" "COMPLETED"

cancel_json="$(run_step cancel env ORDER_ID="bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" EXPECTED_VERSION="1" "${SCRIPT_DIR}/cancel.sh")"
assert_json_status "${cancel_json}" "CANCELLED"

assert_order_status_sql "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" "completed"
assert_order_status_sql "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" "cancelled"

# Forbidden transition sanity check: cancel after complete should fail.
set +e
cancel_after_complete="$(env ORDER_ID="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" EXPECTED_VERSION="2" "${SCRIPT_DIR}/cancel.sh" 2>&1)"
rc=$?
set -e
if [[ $rc -eq 0 ]]; then
  echo "expected cancel-after-complete to fail, but it succeeded" >&2
  exit 1
fi
echo "forbidden-transition check OK (cancel after complete failed as expected)"

echo "assertive smoke flow complete"
