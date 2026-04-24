#!/usr/bin/env bash

# shellcheck disable=SC2034
SMOKE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -f "${SMOKE_DIR}/.env" ]]; then
  # shellcheck disable=SC1091
  source "${SMOKE_DIR}/.env"
fi

GRPC_ADDR="${GRPC_ADDR:-localhost:9090}"
TENANT_ID="${TENANT_ID:-11111111-1111-1111-1111-111111111111}"
CLIENT_REF="${CLIENT_REF:-client_smoke_001}"

OFFICE_ID="${OFFICE_ID:-22222222-2222-2222-2222-222222222222}"
QUOTE_ID="${QUOTE_ID:-quote-reserve-smoke-001}"
ORDER_ID="${ORDER_ID:-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa}"
EXPECTED_VERSION="${EXPECTED_VERSION:-1}"
CASHIER_ID="${CASHIER_ID:-cashier_01}"
REASON="${REASON:-client_no_show}"

GIVE_AMOUNT="${GIVE_AMOUNT:-100.00}"
GIVE_CODE="${GIVE_CODE:-USDT}"
GIVE_NETWORK="${GIVE_NETWORK:-TRC20}"
GET_AMOUNT="${GET_AMOUNT:-3550.00}"
GET_CODE="${GET_CODE:-THB}"
GET_NETWORK="${GET_NETWORK:-cash}"

require_cmd() {
  local cmd="$1"
  local hint="$2"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "required command not found: ${cmd}" >&2
    echo "${hint}" >&2
    exit 1
  fi
}

if [[ "${SMOKE_REQUIRE_GRPCURL:-1}" == "1" ]]; then
  require_cmd "grpcurl" "install grpcurl to run smoke gRPC wrappers"
fi

if [[ "${SMOKE_REQUIRE_PSQL:-0}" == "1" ]]; then
  require_cmd "psql" "install psql (postgres client) to run SQL smoke scripts"
fi
