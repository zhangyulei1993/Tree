#!/usr/bin/env bash
set -euo pipefail

root="${TREE_STAGING_ROOT:-/opt/tree-staging}"
service_name="${TREE_STAGING_SERVICE_NAME:-tree-api-staging.service}"
public_health_url="${TREE_STAGING_PUBLIC_HEALTH_URL:-https://tapi.bigbigboy.cn/api/health}"
disk_limit="${TREE_STAGING_DISK_LIMIT_PERCENT:-85}"
local_health_urls=(
  "${TREE_STAGING_API_HEALTH_URL:-http://127.0.0.1:18081/api/health}"
  "${TREE_STAGING_ADMIN_HEALTH_URL:-http://127.0.0.1:18083/api/health}"
  "${TREE_STAGING_PUBLIC_API_HEALTH_URL:-http://127.0.0.1:18084/api/health}"
)

check_health() {
  local url="$1"
  local body
  body="$(curl --fail --silent --show-error --max-time 10 "$url")"
  [[ "$body" == *'"status":"ok"'* ]] || {
    printf '[ERROR] unhealthy response from %s\n' "$url" >&2
    return 1
  }
}

for url in "${local_health_urls[@]}"; do
  check_health "$url"
done
check_health "$public_health_url"
systemctl is-active --quiet "$service_name"

disk_usage="$(df -P "$root" | awk 'NR == 2 { gsub(/%/, "", $5); print $5 }')"
[[ "$disk_usage" =~ ^[0-9]+$ ]] || {
  printf '[ERROR] unable to determine disk usage for %s\n' "$root" >&2
  exit 1
}
if (( disk_usage >= disk_limit )); then
  printf '[ERROR] disk usage is %s%%, limit is %s%%\n' "$disk_usage" "$disk_limit" >&2
  exit 1
fi

printf '[OK] %s service=%s disk=%s%%\n' "$(date -Is)" "$service_name" "$disk_usage"
