#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "${SCRIPT_DIR}/common.sh"

"${SCRIPT_DIR}/bootstrap.sh"

"${SCRIPT_DIR}/reserve.sh"
ORDER_ID="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" EXPECTED_VERSION="1" "${SCRIPT_DIR}/complete.sh"
ORDER_ID="bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" EXPECTED_VERSION="1" "${SCRIPT_DIR}/cancel.sh"

echo "smoke flow complete: Reserve -> Complete -> Cancel"
