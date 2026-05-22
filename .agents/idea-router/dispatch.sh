#!/usr/bin/env bash
# Router dispatch handler — fired by launchd WatchPaths on any change under
# .agents/idea-router/{state.json,items,proposals}.
#
# Handles BOTH queues addressed to claude-* agents on this workstation:
#   1. Pull-model:  .agents/idea-router/items/*.md  with `to:` matching agent + status: open
#   2. Legacy push: state.json[pending][<agent>]    (compat for codex-side senders)
#
# For each item, spawns `claude --print --permission-mode auto` with an
# instruction containing the item id(s) — the spawned headless session
# reads, acts, closes. No daemon. Single one-shot per FSEvents fire.
#
# Per AGENTS.md §Lean #4 (smallest package wins) this is intentionally a
# shell script, not a Go subcommand. ~50 lines beats ~150 in the binary.

set -uo pipefail
ROUTER_ROOT="/Users/thekryptodragon/Development/sirsi-pantheon/.agents/idea-router"
REPO_ROOT="/Users/thekryptodragon/Development/sirsi-pantheon"
LOG="$ROUTER_ROOT/logs/dispatch.log"
export PATH="$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"
SIRSI="$HOME/.local/bin/sirsi"

ts() { date "+%Y-%m-%dT%H:%M:%S%z"; }

mkdir -p "$(dirname "$LOG")"

# Agents this host claims. Add more (claude-finalwishes, claude-assiduous,
# claude-nexus) when those repo-scoped sessions live on this machine too.
AGENTS=(claude-pantheon)

dispatched=0
for agent in "${AGENTS[@]}"; do
  # --- Pull-model: items/*.md with to: <agent> and status: open ---
  ids=$("$SIRSI" router pull "$agent" 2>/dev/null | awk '/• /{print $2}')

  # --- Legacy push: state.json pending[<agent>] ---
  legacy_ids=$(python3 -c "
import json, sys
try:
    s = json.load(open('$ROUTER_ROOT/state.json'))
    for x in s.get('pending', {}).get('$agent', []):
        print(x)
except Exception:
    pass
" 2>/dev/null)

  all_ids=$(printf '%s\n%s\n' "$ids" "$legacy_ids" | grep -v '^$' | sort -u)
  [ -z "$all_ids" ] && continue

  count=$(echo "$all_ids" | wc -l | tr -d ' ')
  echo "[$(ts)] $agent: $count item(s) to dispatch" >> "$LOG"

  prompt="ctr"  # tells claude (via AGENTS.md §Starting Protocol) to check the router and work the items
  cd "$REPO_ROOT" || continue
  echo "$prompt" | claude --print --permission-mode auto >> "$LOG" 2>&1 &
  dispatched=$((dispatched + 1))
done

echo "[$(ts)] dispatch.sh exit — agents fired: $dispatched" >> "$LOG"
