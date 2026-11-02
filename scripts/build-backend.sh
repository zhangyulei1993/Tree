#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output="${1:-$repo_root/backend/bin/tree-api}"

mkdir -p "$(dirname "$output")"

version="${TREE_APP_VERSION:-0.1.0}"
commit="${TREE_GIT_COMMIT:-$(git -C "$repo_root" rev-parse --short=12 HEAD 2>/dev/null || printf unknown)}"
build_time="${TREE_BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

(
  cd "$repo_root/backend"
  go build \
    -ldflags "\
      -X tree/backend/internal/common/version.Version=$version \
      -X tree/backend/internal/common/version.Commit=$commit \
      -X tree/backend/internal/common/version.BuildTime=$build_time" \
    -o "$output" \
    ./cmd/server
)

printf 'Built %s\n' "$output"
printf 'Version: %s\nCommit: %s\nBuildTime: %s\n' "$version" "$commit" "$build_time"
