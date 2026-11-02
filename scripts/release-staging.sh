#!/usr/bin/env bash

set -euo pipefail

die() {
  printf '[ERROR] %s\n' "$1" >&2
  exit 1
}

info() {
  printf '\n==> %s\n' "$1"
}

usage() {
  cat <<'EOF'
Usage:
  scripts/release-staging.sh [options]

Build and deploy the current workspace to the staging server in one command.

Options:
  --id <release-id>       Use a custom release id.
  --label <label>         Add a short label to the generated release id.
  --host <ssh-host>       SSH host alias. Default: tree-server.
  --skip-build            Reuse existing local build artifacts.
  --with-miniapp          Also build miniapp/dist/mp-weixin for upload in WeChat DevTools.
  --smoke-phone-route     Smoke test the WeChat phone login route with an invalid code.
  -h, --help              Show this help.

Examples:
  scripts/release-staging.sh
  scripts/release-staging.sh --label phone-login --with-miniapp --smoke-phone-route
EOF
}

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ssh_host="${STAGING_SSH_HOST:-tree-server}"
staging_root="${STAGING_ROOT:-/opt/tree-staging}"
service_name="${STAGING_SERVICE_NAME:-tree-api-staging.service}"
health_url="${STAGING_HEALTH_URL:-https://tapi.bigbigboy.cn/api/health}"
api_base="${STAGING_API_BASE:-https://tapi.bigbigboy.cn/api}"
label="${RELEASE_LABEL:-manual}"
release_id="${RELEASE_ID:-}"
skip_build=0
with_miniapp=0
smoke_phone_route=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --id)
      [[ $# -ge 2 ]] || die '--id requires a value'
      release_id="$2"
      shift 2
      ;;
    --label)
      [[ $# -ge 2 ]] || die '--label requires a value'
      label="$2"
      shift 2
      ;;
    --host)
      [[ $# -ge 2 ]] || die '--host requires a value'
      ssh_host="$2"
      shift 2
      ;;
    --skip-build)
      skip_build=1
      shift
      ;;
    --with-miniapp)
      with_miniapp=1
      shift
      ;;
    --smoke-phone-route)
      smoke_phone_route=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "Unknown option: $1"
      ;;
  esac
done

sanitize_label() {
  printf '%s' "$1" |
    tr '[:upper:]' '[:lower:]' |
    sed -E 's/[^a-z0-9._-]+/-/g; s/^-+//; s/-+$//' |
    cut -c 1-48
}

label="$(sanitize_label "$label")"
[[ -n "$label" ]] || label="manual"

if [[ -z "$release_id" ]]; then
  release_id="$(date +%Y%m%d-%H%M%S)-$label"
fi

[[ "$release_id" =~ ^[A-Za-z0-9._-]+$ ]] || die "Unsafe release id: $release_id"

backend_binary="$repo_root/backend/bin/tree-api-staging-linux-amd64"
backend_migrations="$repo_root/backend/migrations"
admin_dist="$repo_root/admin-web/dist"
web_dist="$repo_root/web/dist"
incoming="$staging_root/incoming/$release_id"
ops_dir="$repo_root/scripts/ops"

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

command_exists ssh || die 'ssh is required'
command_exists rsync || die 'rsync is required'
command_exists pnpm || die 'pnpm is required'
command_exists curl || die 'curl is required'

cd "$repo_root"

info "Release id: $release_id"

if [[ "$skip_build" -eq 0 ]]; then
  info 'Build backend linux/amd64 binary'
  GOOS=linux GOARCH=amd64 "$repo_root/scripts/build-backend.sh" "$backend_binary"

  info 'Build admin web'
  pnpm --dir "$repo_root/admin-web" build:staging

  info 'Build public web'
  VITE_API_MODE=real pnpm --dir "$repo_root/web" build

  if [[ "$with_miniapp" -eq 1 ]]; then
    info 'Build WeChat mini program staging package'
    pnpm --dir "$repo_root/miniapp" build:mp-weixin:staging
  fi
else
  info 'Skip local builds'
fi

[[ -x "$backend_binary" ]] || die "Backend binary missing or not executable: $backend_binary"
[[ -d "$backend_migrations" ]] || die "Migrations missing: $backend_migrations"
[[ -d "$admin_dist" ]] || die "Admin dist missing: $admin_dist"
[[ -d "$web_dist" ]] || die "Web dist missing: $web_dist"
[[ -f "$ops_dir/backup-staging.sh" ]] || die "Backup operation script missing: $ops_dir/backup-staging.sh"
[[ -f "$ops_dir/health-staging.sh" ]] || die "Health operation script missing: $ops_dir/health-staging.sh"

info "Upload artifacts to $ssh_host:$incoming"
ssh "$ssh_host" "mkdir -p '$incoming/backend' '$incoming/admin-web' '$incoming/web' '$incoming/ops'"
rsync -az --delete "$backend_binary" "$ssh_host:$incoming/backend/tree-api"
rsync -az --delete "$backend_migrations/" "$ssh_host:$incoming/backend/migrations/"
rsync -az --delete "$admin_dist/" "$ssh_host:$incoming/admin-web/dist/"
rsync -az --delete "$web_dist/" "$ssh_host:$incoming/web/dist/"
rsync -az --delete "$ops_dir/" "$ssh_host:$incoming/ops/"

info 'Install versioned staging operation scripts'
ssh "$ssh_host" "set -eu; install -m 0750 '$incoming/ops/backup-staging.sh' '$staging_root/shared/backup-staging.sh'; install -m 0750 '$incoming/ops/health-staging.sh' '$staging_root/shared/health-staging.sh'; install -m 0644 '$incoming/ops/tree-staging-health.cron' /etc/cron.d/tree-staging-health; install -m 0644 '$incoming/ops/tree-staging-logrotate.conf' /etc/logrotate.d/tree-staging"

info 'Run remote deployment'
ssh "$ssh_host" \
  "APP_ENV=staging \
  STAGING_DEPLOY_CONFIRM=DEPLOY_STAGING \
  STAGING_ROOT='$staging_root' \
  RELEASE_ID='$release_id' \
  BACKEND_BINARY='$incoming/backend/tree-api' \
  BACKEND_MIGRATIONS='$incoming/backend/migrations' \
  ADMIN_DIST='$incoming/admin-web/dist' \
  WEB_DIST='$incoming/web/dist' \
  STAGING_BACKUP_SCRIPT='$staging_root/shared/backup-staging.sh' \
  STAGING_MIGRATE_SCRIPT='$staging_root/shared/migrate-staging.sh' \
  STAGING_SERVICE_NAME='$service_name' \
  STAGING_HEALTH_URL='$health_url' \
  '$staging_root/shared/deploy-staging.sh'"

info 'Smoke test health endpoint'
curl --fail --silent --show-error --max-time 15 "$health_url"
printf '\n'

if [[ "$smoke_phone_route" -eq 1 ]]; then
  info 'Smoke test WeChat mini-program phone login route with invalid code'
  phone_response="$(
    curl --silent --show-error --max-time 15 \
      -X POST "$api_base/auth/wechat-mini/phone-login" \
      -H 'Content-Type: application/json' \
      --data '{"code":"invalid-release-wechat-code","phoneCode":"invalid-release-smoke-code","clientType":"WECHAT_MINI_PROGRAM"}'
  )"
  printf '%s\n' "$phone_response"
  printf '%s' "$phone_response" | grep -Eq '"code":40307|"message":"微信手机号获取失败"' ||
    die 'wechat-phone-login smoke response was not the expected invalid-code error'
fi

info 'Remote service status'
ssh "$ssh_host" "readlink '$staging_root/current'; systemctl is-active '$service_name'"

cat <<EOF

Release completed.
Current release: $release_id
Rollback command:
  APP_ENV=staging STAGING_ROLLBACK_CONFIRM=ROLLBACK_STAGING STAGING_ROOT=$staging_root STAGING_SERVICE_NAME=$service_name STAGING_HEALTH_URL=$health_url scripts/rollback-staging.sh
EOF
