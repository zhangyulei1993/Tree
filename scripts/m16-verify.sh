#!/usr/bin/env bash

set -uo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

failures=0

run_check() {
  local name="$1"
  shift

  printf '\n== %s ==\n' "$name"
  if "$@"; then
    printf '[PASS] %s\n' "$name"
  else
    printf '[FAIL] %s\n' "$name" >&2
    failures=$((failures + 1))
  fi
}

run_check \
  'Docker Compose configuration' \
  docker compose -f docker/docker-compose.yml config

run_check \
  'Backend Go tests' \
  docker exec tree-dev bash -lc 'cd /workspace/backend && go test ./...'

run_check \
  'Admin web build' \
  docker exec -e CI=true tree-dev bash -lc 'cd /workspace/admin-web && pnpm build'

run_check \
  'PC/H5 web build' \
  docker exec -e CI=true tree-dev bash -lc 'cd /workspace/web && pnpm build'

run_check \
  'Mini program build' \
  docker exec -e CI=true tree-dev bash -lc 'cd /workspace/miniapp && pnpm build:mp-weixin'

run_check \
  'Repository security check' \
  bash scripts/security-check.sh

printf '\n'
if ((failures > 0)); then
  printf 'M16 verification result: FAIL (%d check(s) failed)\n' "$failures" >&2
  exit 1
fi

printf 'M16 verification result: PASS\n'
