#!/usr/bin/env bash

set -euo pipefail

die() {
  printf '[ERROR] %s\n' "$1" >&2
  exit 1
}

require_var() {
  local name="$1"
  [[ -n "${!name:-}" ]] || die "Required variable is missing: $name"
}

assert_safe_path() {
  local path="$1"
  [[ "$path" == /* ]] || die "Path must be absolute: $path"
  [[ "$path" != "/" && "$path" != "/opt" && "$path" != "/var" ]] ||
    die "Refusing unsafe deployment root: $path"
}

is_production_like() {
  local normalized
  normalized="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
  [[ "$normalized" == *production* || "$normalized" == *prod-* || "$normalized" == *-prod* ]]
}

[[ "${APP_ENV:-}" == "staging" ]] || die 'APP_ENV must be exactly staging'
[[ "${STAGING_DEPLOY_CONFIRM:-}" == "DEPLOY_STAGING" ]] ||
  die 'Set STAGING_DEPLOY_CONFIRM=DEPLOY_STAGING to continue'

for value in "${APP_ENV:-}" "${STAGING_ROOT:-}" "${STAGING_SERVICE_NAME:-}" "${STAGING_HEALTH_URL:-}"; do
  if is_production_like "$value"; then
    die 'A production-like value was detected; staging deployment refused'
  fi
done

require_var STAGING_ROOT
require_var RELEASE_ID
require_var BACKEND_BINARY
require_var ADMIN_DIST
require_var WEB_DIST
require_var STAGING_BACKUP_SCRIPT
require_var STAGING_MIGRATE_SCRIPT
require_var STAGING_SERVICE_NAME
require_var STAGING_HEALTH_URL

assert_safe_path "$STAGING_ROOT"

[[ "$RELEASE_ID" =~ ^[A-Za-z0-9._-]+$ ]] || die 'RELEASE_ID contains unsafe characters'
[[ -x "$BACKEND_BINARY" ]] || die "Backend binary is not executable: $BACKEND_BINARY"
[[ -d "$ADMIN_DIST" ]] || die "Admin dist directory does not exist: $ADMIN_DIST"
[[ -d "$WEB_DIST" ]] || die "Web dist directory does not exist: $WEB_DIST"
[[ -x "$STAGING_BACKUP_SCRIPT" ]] || die 'STAGING_BACKUP_SCRIPT must be an executable file'
[[ -x "$STAGING_MIGRATE_SCRIPT" ]] || die 'STAGING_MIGRATE_SCRIPT must be an executable file'
[[ "$STAGING_HEALTH_URL" == https://* || "$STAGING_HEALTH_URL" == http://127.0.0.1:* ]] ||
  die 'Health URL must use HTTPS or loopback HTTP'

releases_dir="$STAGING_ROOT/releases"
release_dir="$releases_dir/$RELEASE_ID"
current_link="$STAGING_ROOT/current"
previous_link="$STAGING_ROOT/previous"

[[ ! -e "$release_dir" ]] || die "Release already exists: $release_dir"

mkdir -p "$releases_dir"
mkdir -p "$release_dir/backend" "$release_dir/admin-web" "$release_dir/web"

cleanup_failed_release() {
  if [[ ! -L "$current_link" || "$(readlink "$current_link" 2>/dev/null || true)" != "$release_dir" ]]; then
    rm -rf "$release_dir"
  fi
}
trap cleanup_failed_release ERR

printf 'Creating staging backup through the configured hook\n'
"$STAGING_BACKUP_SCRIPT"

install -m 0755 "$BACKEND_BINARY" "$release_dir/backend/tree-api"
cp -R "$ADMIN_DIST"/. "$release_dir/admin-web/"
cp -R "$WEB_DIST"/. "$release_dir/web/"

printf 'Running forward-only staging migration hook\n'
"$STAGING_MIGRATE_SCRIPT" up

if [[ -L "$current_link" ]]; then
  ln -sfn "$(readlink "$current_link")" "$previous_link"
fi
ln -sfn "$release_dir" "$current_link"

sudo systemctl restart "$STAGING_SERVICE_NAME"
curl --fail --silent --show-error --max-time 15 "$STAGING_HEALTH_URL" >/dev/null

trap - ERR
printf 'Staging deployment completed: %s\n' "$RELEASE_ID"
printf 'Rollback target: %s\n' "$(readlink "$previous_link" 2>/dev/null || printf 'none')"
