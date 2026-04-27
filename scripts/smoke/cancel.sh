#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY:-cancel-$(date +%s)}"

grpcurl -plaintext \
  -H "x-tenant-id: ${TENANT_ID}" \
  -H "x-client-ref: ${CLIENT_REF}" \
  -d "{
    \"idempotency_key\": \"${IDEMPOTENCY_KEY}\",
    \"order_id\": \"${ORDER_ID}\",
    \"expected_version\": ${EXPECTED_VERSION},
    \"reason\": \"${REASON}\"
  }" \
  "${GRPC_ADDR}" \
  exchange.order.v1.OrderService/CancelOrder
