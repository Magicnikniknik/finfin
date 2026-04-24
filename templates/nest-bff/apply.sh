#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=0
FORCE=0

usage() {
  echo "usage: $0 [--dry-run] [--force] <path-to-nest-project-root>" >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --force)
      FORCE=1
      shift
      ;;
    -*)
      usage
      exit 1
      ;;
    *)
      break
      ;;
  esac
done

if [[ $# -ne 1 ]]; then
  usage
  exit 1
fi

TARGET_ROOT="$1"
SRC_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STAMP="$(date +%Y%m%d%H%M%S)"

if [[ ! -d "${TARGET_ROOT}" ]]; then
  echo "target path does not exist: ${TARGET_ROOT}" >&2
  exit 1
fi

mkdir -p "${TARGET_ROOT}/src/common/filters" "${TARGET_ROOT}/src/common/decorators" "${TARGET_ROOT}/src/orders/interfaces" "${TARGET_ROOT}/src/orders" "${TARGET_ROOT}/test"

copy_file() {
  local src="$1"
  local dst="$2"

  if [[ -f "${dst}" && "${FORCE}" -ne 1 ]]; then
    echo "skip (exists): ${dst}" >&2
    echo "  re-run with --force to overwrite (backup will be created)." >&2
    return 0
  fi

  if [[ "${DRY_RUN}" -eq 1 ]]; then
    if [[ -f "${dst}" && "${FORCE}" -eq 1 ]]; then
      echo "dry-run: backup ${dst} -> ${dst}.bak.${STAMP}"
    fi
    echo "dry-run: copy ${src} -> ${dst}"
    return 0
  fi

  if [[ -f "${dst}" && "${FORCE}" -eq 1 ]]; then
    cp "${dst}" "${dst}.bak.${STAMP}"
    echo "backup created: ${dst}.bak.${STAMP}"
  fi

  cp "${src}" "${dst}"
}

copy_file "${SRC_ROOT}/src/common/filters/global-exception.filter.ts" "${TARGET_ROOT}/src/common/filters/global-exception.filter.ts"
copy_file "${SRC_ROOT}/src/common/decorators/current-tenant.decorator.ts" "${TARGET_ROOT}/src/common/decorators/current-tenant.decorator.ts"
copy_file "${SRC_ROOT}/src/common/decorators/current-client-ref.decorator.ts" "${TARGET_ROOT}/src/common/decorators/current-client-ref.decorator.ts"
copy_file "${SRC_ROOT}/src/common/grpc-http-error.mapper.ts" "${TARGET_ROOT}/src/common/grpc-http-error.mapper.ts"
copy_file "${SRC_ROOT}/src/orders/interfaces/order-grpc.interface.ts" "${TARGET_ROOT}/src/orders/interfaces/order-grpc.interface.ts"
copy_file "${SRC_ROOT}/src/orders/orders.service.ts" "${TARGET_ROOT}/src/orders/orders.service.ts"
copy_file "${SRC_ROOT}/src/orders/orders.controller.ts" "${TARGET_ROOT}/src/orders/orders.controller.ts"
copy_file "${SRC_ROOT}/src/orders/orders.module.ts" "${TARGET_ROOT}/src/orders/orders.module.ts"
copy_file "${SRC_ROOT}/src/app.module.ts" "${TARGET_ROOT}/src/app.module.ts"
copy_file "${SRC_ROOT}/src/main.ts" "${TARGET_ROOT}/src/main.ts"
copy_file "${SRC_ROOT}/test/orders.e2e-spec.ts" "${TARGET_ROOT}/test/orders.e2e-spec.ts"

if [[ ! -f "${TARGET_ROOT}/proto/exchange/order/v1/order_service.proto" ]]; then
  echo "warning: proto file not found at target path:" >&2
  echo "  ${TARGET_ROOT}/proto/exchange/order/v1/order_service.proto" >&2
  echo "set ORDER_GRPC_PROTO_PATH in your Nest env if proto is elsewhere." >&2
fi

if ! grep -q "getService<OrderGrpcService>('OrderService')" "${TARGET_ROOT}/src/orders/orders.service.ts"; then
  echo "warning: gRPC service name check failed in orders.service.ts" >&2
fi

echo "Nest BFF templates copied into: ${TARGET_ROOT}"
echo "next:"
echo "  1) verify grpc options in src/orders/orders.module.ts:"
echo "     ORDER_GRPC_URL, ORDER_GRPC_PACKAGE, ORDER_GRPC_PROTO_PATH"
echo "  2) align DTO imports in src/orders/orders.controller.ts"
echo "  3) ensure global ValidationPipe + GlobalExceptionFilter in src/main.ts"
echo "  4) run e2e skeleton: test/orders.e2e-spec.ts"
echo "  5) swap mock path to real gRPC integration smoke"
