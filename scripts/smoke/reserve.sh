#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

CLIENT_REF="${CLIENT_REF:-client_reserve_smoke_001}"
IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY:-reserve-$(date +%s)}"
SIDE="${SIDE:-BUY}"

grpcurl -plaintext \
  -H "x-tenant-id: ${TENANT_ID}" \
  -H "x-client-ref: ${CLIENT_REF}" \
  -d "{
    \"idempotency_key\": \"${IDEMPOTENCY_KEY}\",
    \"office_id\": \"${OFFICE_ID}\",
    \"quote_id\": \"${QUOTE_ID}\",
    \"side\": \"${SIDE}\",
    \"give\": {\"amount\": \"${GIVE_AMOUNT}\", \"currency\": {\"code\": \"${GIVE_CODE}\", \"network\": \"${GIVE_NETWORK}\"}},
    \"get\": {\"amount\": \"${GET_AMOUNT}\", \"currency\": {\"code\": \"${GET_CODE}\", \"network\": \"${GET_NETWORK}\"}}
  }" \
  "${GRPC_ADDR}" \
  exchange.order.v1.OrderService/ReserveOrder
