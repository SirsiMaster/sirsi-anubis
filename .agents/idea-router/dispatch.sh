#!/usr/bin/env bash
# Router dispatch handler — fired by launchd WatchPaths on any change under
# .agents/idea-router/{state.json,items,proposals}.
#
# Handles queues for local agents this workstation can actually wake.
# Pull-model items live in .agents/idea-router/items/*.md; legacy pending
# queues live in state.json[pending][<agent>].
#
# Dispatch stays intentionally small: resolve each local agent, then spawn the
# matching CLI once with a router-focused prompt. No daemon, no polling loop.

set -uo pipefail
ROUTER_ROOT="/Users/thekryptodragon/Development/sirsi-pantheon/.agents/idea-router"
REPO_ROOT="/Users/thekryptodragon/Development/sirsi-pantheon"
LOG="$ROUTER_ROOT/logs/dispatch.log"
export PATH="$HOME/.local/bin:/Applications/Codex.app/Contents/Resources:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"
SIRSI="$HOME/.local/bin/sirsi"
CLAUDE_BIN="$HOME/.local/bin/claude"
CODEX_BIN="/Applications/Codex.app/Contents/Resources/codex"

ts() { date "+%Y-%m-%dT%H:%M:%S%z"; }

mkdir -p "$(dirname "$LOG")"

# Agents this host claims. Keep this explicit so launchd does not pretend to
# deliver to agents that have no local headless wake path.
AGENTS=(claude-pantheon codex-pantheon)

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

  prompt="ctr

You are $agent on this workstation.
Read $ROUTER_ROOT/state.json and the open router item(s) addressed to $agent:
$all_ids

Work only those addressed items. Write the required router review/decision/completion artifacts, update state.json, and stop. If a blocker prevents work, write a blocker artifact instead of looping."

  cd "$REPO_ROOT" || continue
  case "$agent" in
    claude-*)
      if [ ! -x "$CLAUDE_BIN" ]; then
        echo "[$(ts)] $agent blocked: claude CLI not executable at $CLAUDE_BIN" >> "$LOG"
        continue
      fi
      echo "$prompt" | "$CLAUDE_BIN" --print --permission-mode auto >> "$LOG" 2>&1 &
      ;;
    codex-*)
      if [ ! -x "$CODEX_BIN" ]; then
        echo "[$(ts)] $agent blocked: Codex CLI not executable at $CODEX_BIN" >> "$LOG"
        continue
      fi
      echo "$prompt" | "$CODEX_BIN" exec -C "$REPO_ROOT" --sandbox workspace-write - >> "$LOG" 2>&1 &
      ;;
    *)
      echo "[$(ts)] $agent blocked: no dispatch branch for agent family" >> "$LOG"
      continue
      ;;
  esac
  if ! "$SIRSI" router ack "$agent" $all_ids >> "$LOG" 2>&1; then
    echo "[$(ts)] $agent warning: failed to ack legacy pending ids" >> "$LOG"
  fi
  dispatched=$((dispatched + 1))
done

echo "[$(ts)] dispatch.sh exit — agents fired: $dispatched" >> "$LOG"
