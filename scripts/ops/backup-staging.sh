#!/usr/bin/env bash
set -euo pipefail

root="${TREE_STAGING_ROOT:-/opt/tree-staging}"
secrets="$root/.staging-secrets"
env_file="$root/backend/.env"
backup_dir="${TREE_BACKUP_DIRECTORY:-$root/shared/backups}"

[[ -f "$secrets" && -f "$env_file" ]] || { echo '[ERROR] missing staging environment' >&2; exit 1; }
set -a
# shellcheck disable=SC1090
. "$secrets"
# shellcheck disable=SC1090
. "$env_file"
set +a
: "${MYSQL_DATABASE:?MYSQL_DATABASE required}"
: "${MYSQL_USER:?MYSQL_USER required}"
: "${MYSQL_PASSWORD:?MYSQL_PASSWORD required}"
: "${MYSQL_HOST:=127.0.0.1}"
: "${MYSQL_PORT:=13306}"

cnf="$(mktemp)"
temp_backup=""
cleanup() { rm -f "$cnf" "$temp_backup"; }
trap cleanup EXIT
chmod 0600 "$cnf"
printf '[client]\nhost=%s\nport=%s\nuser=%s\npassword=%s\ndefault-character-set=utf8mb4\n' "$MYSQL_HOST" "$MYSQL_PORT" "$MYSQL_USER" "$MYSQL_PASSWORD" > "$cnf"

enabled=1 retention_days=14 max_storage_mb=512 max_disk_usage_percent=85
# Deployments run a backup before applying migrations, so a missing config table
# deliberately falls back to the conservative defaults above.
if config="$(mysql --defaults-extra-file="$cnf" --skip-column-names --batch -e "SELECT enabled, retention_days, max_storage_mb, max_disk_usage_percent FROM backup_operation_configs WHERE id = 1" "$MYSQL_DATABASE" 2>/dev/null)"; then
  read -r enabled retention_days max_storage_mb max_disk_usage_percent <<< "$config"
fi
for value in "$enabled" "$retention_days" "$max_storage_mb" "$max_disk_usage_percent"; do [[ "$value" =~ ^[0-9]+$ ]] || { echo '[ERROR] invalid backup configuration' >&2; exit 1; }; done
(( enabled == 1 )) || { echo '[SKIP] backup disabled by admin configuration'; exit 0; }
(( retention_days >= 1 && retention_days <= 90 )) || { echo '[ERROR] invalid retention days' >&2; exit 1; }
(( max_storage_mb >= 64 && max_storage_mb <= 102400 )) || { echo '[ERROR] invalid storage limit' >&2; exit 1; }
(( max_disk_usage_percent >= 60 && max_disk_usage_percent <= 95 )) || { echo '[ERROR] invalid disk limit' >&2; exit 1; }

mkdir -p "$backup_dir"
find "$backup_dir" -type f -name "${MYSQL_DATABASE}-*.sql.gz" -mtime +"$retention_days" -delete
disk_usage="$(df -P "$root" | awk 'NR == 2 { gsub(/%/, "", $5); print $5 }')"
(( disk_usage < max_disk_usage_percent )) || { printf '[SKIP] disk usage %s%% reaches configured limit %s%%\n' "$disk_usage" "$max_disk_usage_percent" >&2; exit 0; }

temp_backup="$(mktemp "${TMPDIR:-/tmp}/${MYSQL_DATABASE}.backup.XXXXXX.sql.gz")"
mysqldump --defaults-extra-file="$cnf" --single-transaction --no-tablespaces --routines --triggers --databases "$MYSQL_DATABASE" | gzip -9 > "$temp_backup"
current_bytes="$(find "$backup_dir" -type f -name "${MYSQL_DATABASE}-*.sql.gz" -printf '%s\n' | awk '{sum += $1} END {print sum + 0}')"
new_bytes="$(stat -c '%s' "$temp_backup" 2>/dev/null || stat -f '%z' "$temp_backup")"
limit_bytes=$((max_storage_mb * 1024 * 1024))
(( current_bytes + new_bytes <= limit_bytes )) || { printf '[SKIP] backup storage cap reached: current=%s bytes, cap=%s bytes\n' "$current_bytes" "$limit_bytes" >&2; exit 0; }

out="$backup_dir/${MYSQL_DATABASE}-$(date +%Y%m%d-%H%M%S).sql.gz"
mv "$temp_backup" "$out"
temp_backup=""
chmod 0600 "$out"
printf 'backup created: %s\n' "$out"
