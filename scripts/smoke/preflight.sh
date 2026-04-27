#!/usr/bin/env bash
set -euo pipefail

ok() { printf '✅ %s\n' "$1"; }
warn() { printf '⚠️  %s\n' "$1"; }
fail() { printf '❌ %s\n' "$1"; }

missing=0

require_cmd() {
  local cmd="$1"
  local hint="$2"
  if command -v "$cmd" >/dev/null 2>&1; then
    ok "found command: ${cmd}"
  else
    fail "missing command: ${cmd} (${hint})"
    missing=1
  fi
}

echo "== Smoke preflight for finfin =="

require_cmd "go" "install Go toolchain"
require_cmd "curl" "install curl"
require_cmd "docker" "install Docker (required for local postgres/pilot stack)"
require_cmd "psql" "install postgres client for bootstrap/check SQL scripts"
require_cmd "grpcurl" "install grpcurl for Reserve/Complete/Cancel wrappers"

if [[ -n "${DATABASE_URL:-}" ]]; then
  ok "DATABASE_URL is set"
else
  fail "DATABASE_URL is not set (required for SQL bootstrap and checks)"
  missing=1
fi

if command -v docker >/dev/null 2>&1; then
  if docker info >/dev/null 2>&1; then
    ok "docker daemon is reachable"
  else
    fail "docker is installed but daemon is not reachable"
    missing=1
  fi
fi

if command -v psql >/dev/null 2>&1 && [[ -n "${DATABASE_URL:-}" ]]; then
  if psql "${DATABASE_URL}" -c "select 1;" >/dev/null 2>&1; then
    ok "database is reachable"
  else
    fail "database is not reachable via DATABASE_URL"
    missing=1
  fi
fi

if command -v grpcurl >/dev/null 2>&1 && [[ -n "${GRPC_ADDR:-}" ]]; then
  if grpcurl -plaintext "${GRPC_ADDR}" list >/dev/null 2>&1; then
    ok "gRPC endpoint is reachable (${GRPC_ADDR})"
  else
    fail "gRPC endpoint is not reachable (${GRPC_ADDR})"
    missing=1
  fi
else
  warn "GRPC_ADDR is not set (set it to enable gRPC reachability check)"
fi

if [[ -n "${HTTP_BASE_URL:-}" ]]; then
  if curl -fsS "${HTTP_BASE_URL}/healthz" >/dev/null 2>&1; then
    ok "HTTP endpoint is reachable (${HTTP_BASE_URL}/healthz)"
  else
    fail "HTTP endpoint is not reachable (${HTTP_BASE_URL}/healthz)"
    missing=1
  fi
else
  warn "HTTP_BASE_URL is not set (set it to enable HTTP reachability check)"
fi

if command -v go >/dev/null 2>&1; then
  if GOPROXY=https://proxy.golang.org go mod download \
    github.com/jackc/pgpassfile@v1.0.0 \
    github.com/jackc/pgservicefile@v0.0.0-20240606120523-5a60cdf6a761 \
    golang.org/x/crypto@v0.40.0 \
    golang.org/x/text@v0.27.0 >/dev/null 2>&1; then
    ok "critical go module download via proxy.golang.org works"
  else
    warn "critical go module download from proxy.golang.org is blocked (go test/build likely fail)"
  fi
fi

if [[ "${missing}" -eq 1 ]]; then
  echo
  fail "preflight failed: environment is not ready for full smoke flow"
  exit 1
fi

echo
ok "preflight passed: you can proceed with smoke bootstrap and full cycle"
