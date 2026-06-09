#!/usr/bin/env bash

set -uo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

failures=0

pass() {
  printf '[PASS] %s\n' "$1"
}

fail() {
  printf '[FAIL] %s\n' "$1" >&2
  failures=$((failures + 1))
}

tracked_files() {
  git ls-files -z
}

printf 'Tree M16 security check\n'
printf 'Repository: %s\n\n' "$ROOT_DIR"

tracked_env="$(
  git ls-files |
    awk '
      /(^|\/)\.env($|\.)/ &&
      $0 !~ /\.env\.example$/ &&
      $0 !~ /\.env\.sample$/ &&
      $0 !~ /\.env\.template$/ { print }
    '
)"
if [[ -n "$tracked_env" ]]; then
  printf '%s\n' "$tracked_env" >&2
  fail 'Real environment files appear to be tracked'
else
  pass 'No real .env files are tracked'
fi

tracked_sensitive_artifacts="$(
  git ls-files |
    grep -E '(^|/)(project\.private\.config\.json|id_rsa|id_ed25519)$|(\.p12|\.pfx|\.jks|\.keystore)$|(^|/)(backup|backups)/.*\.(sql|dump|gz)$' ||
    true
)"
if [[ -n "$tracked_sensitive_artifacts" ]]; then
  printf '%s\n' "$tracked_sensitive_artifacts" >&2
  fail 'Sensitive local configuration, key stores, or database backups are tracked'
else
  pass 'No private miniapp config, key stores, or database backups are tracked'
fi

private_key_markers="$(
  tracked_files |
    xargs -0 grep -IlE -- '-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----' 2>/dev/null ||
    true
)"
if [[ -n "$private_key_markers" ]]; then
  printf '%s\n' "$private_key_markers" >&2
  fail 'Private key material was found in tracked files'
else
  pass 'No private key material was found in tracked files'
fi

# Scan source and deployable configuration. Documentation, examples, generated
# output, migrations, lockfiles, and mock-only UI data are intentionally excluded.
hardcoded_secret_pattern='(JWT_(USER|ADMIN)_SECRET|WECHAT(_MINI)?_APP_SECRET|MYSQL_PASSWORD|REDIS_PASSWORD|ACCESS_TOKEN|REFRESH_TOKEN)[[:space:]]*[:=][[:space:]]*["'\''][^"'\'']{8,}["'\'']'
hardcoded_secrets="$(
  git grep -n -I -E "$hardcoded_secret_pattern" -- \
    'backend/**/*.go' \
    'admin-web/src/**' \
    'web/src/**' \
    'miniapp/src/**' \
    'scripts/**' \
    ':!**/*_test.go' \
    ':!**/mock/**' 2>/dev/null |
    grep -Ev '(os\.Getenv|GetString|viper|example_|mock_|change_me|replace_with|redacted|masked)' ||
    true
)"
if [[ -n "$hardcoded_secrets" ]]; then
  printf '%s\n' "$hardcoded_secrets" >&2
  fail 'Potential hard-coded credentials were found in source or deployment scripts'
else
  pass 'No obvious hard-coded credentials were found in source or deployment scripts'
fi

if git check-ignore -q backend/.env; then
  pass 'backend/.env is ignored'
else
  fail 'backend/.env is not ignored'
fi

for path in docker/mysql/data docker/redis/data admin-web/dist web/dist miniapp/dist; do
  if git check-ignore -q "$path"; then
    pass "$path is ignored"
  else
    fail "$path is not ignored"
  fi
done

if git grep -n -I -E 'WECHAT(_MINI)?_APP_SECRET' -- 'admin-web/src/**' 'web/src/**' 'miniapp/src/**' >/tmp/tree-m16-frontend-secret-scan.$$ 2>/dev/null; then
  cat /tmp/tree-m16-frontend-secret-scan.$$ >&2
  rm -f /tmp/tree-m16-frontend-secret-scan.$$
  fail 'A WeChat AppSecret reference exists in frontend source'
else
  rm -f /tmp/tree-m16-frontend-secret-scan.$$
  pass 'No WeChat AppSecret reference exists in frontend source'
fi

printf '\n'
if ((failures > 0)); then
  printf 'Security check result: FAIL (%d issue(s))\n' "$failures" >&2
  exit 1
fi

printf 'Security check result: PASS\n'
