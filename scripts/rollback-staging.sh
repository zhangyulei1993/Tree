#!/usr/bin/env bash

set -euo pipefail

die() {
  printf '[ERROR] %s\n' "$1" >&2
  exit 1
}

is_production_like() {
  local normalized
  normalized="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
  [[ "$normalized" == *production* || "$normalized" == *prod-* || "$normalized" == *-prod* ]]
}

[[ "${APP_ENV:-}" == "staging" ]] || die 'APP_ENV must be exactly staging'
[[ "${STAGING_ROLLBACK_CONFIRM:-}" == "ROLLBACK_STAGING" ]] ||
  die 'Set STAGING_ROLLBACK_CONFIRM=ROLLBACK_STAGING to continue'

for value in "${APP_ENV:-}" "${STAGING_ROOT:-}" "${STAGING_SERVICE_NAME:-}" "${STAGING_HEALTH_URL:-}"; do
  if is_production_like "$value"; then
    die 'A production-like value was detected; staging rollback refused'
  fi
done

: "${STAGING_ROOT:?STAGING_ROOT is required}"
: "${STAGING_SERVICE_NAME:?STAGING_SERVICE_NAME is required}"
: "${STAGING_HEALTH_URL:?STAGING_HEALTH_URL is required}"

[[ "$STAGING_ROOT" == /* && "$STAGING_ROOT" != "/" ]] || die 'STAGING_ROOT must be a safe absolute path'
[[ "$STAGING_HEALTH_URL" == https://* || "$STAGING_HEALTH_URL" == http://127.0.0.1:* ]] ||
  die 'Health URL must use HTTPS or loopback HTTP'

current_link="$STAGING_ROOT/current"
previous_link="$STAGING_ROOT/previous"

[[ -L "$current_link" ]] || die 'Current release symlink does not exist'
[[ -L "$previous_link" ]] || die 'Previous release symlink does not exist'

current_target="$(readlink "$current_link")"
previous_target="$(readlink "$previous_link")"

[[ -d "$previous_target" ]] || die "Previous release directory does not exist: $previous_target"

# Application rollback only. This script intentionally never invokes migration
# down commands or restores a database automatically.
ln -sfn "$previous_target" "$current_link"
ln -sfn "$current_target" "$previous_link"

if ! sudo systemctl restart "$STAGING_SERVICE_NAME" ||
  ! curl --fail --silent --show-error --max-time 15 "$STAGING_HEALTH_URL" >/dev/null; then
  ln -sfn "$current_target" "$current_link"
  ln -sfn "$previous_target" "$previous_link"
  sudo systemctl restart "$STAGING_SERVICE_NAME" || true
  die 'Rollback health check failed; restored the original application release'
fi

printf 'Staging application rollback completed\n'
printf 'Current release: %s\n' "$previous_target"
printf 'No database down migration was executed\n'
