#!/bin/bash
# Wake helper for the pull-model router. Fired by launchd WatchPaths on
# changes to state.json, items/, or proposals/. This script OBSERVES the
# queue and logs activity — it does NOT spawn agents or mutate items.
# That matches Codex's review condition: the plist is a wake helper, not
# the source of truth. Real work happens when an agent next opens an
# interactive session and reads its inbox.
#
# If you want auto-spawn on item arrival, write a sibling script that
# calls `sirsi router pull <id>` and invokes your agent of choice. Keep
# that separate from this observer so a broken spawner can't make the
# queue look broken.

set -euo pipefail

REPO="/Users/thekryptodragon/Development/sirsi-pantheon"
LOG_DIR="${REPO}/.agents/idea-router/logs"
WAKE_LOG="${LOG_DIR}/wake.log"

mkdir -p "${LOG_DIR}"

{
  echo "=== wake $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="
  cd "${REPO}"
  if [[ -x ./sirsi ]]; then
    ./sirsi router status --stale 6 2>&1 || true
  else
    echo "  (./sirsi binary not present — run 'go build ./cmd/sirsi' to enable status)"
  fi
} >> "${WAKE_LOG}"
