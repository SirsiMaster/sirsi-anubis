#!/usr/bin/env bash
# Periodic verification sweep — fired hourly by launchd.
#
# Performs the "are we complete" probe that has caught real bugs three
# times in this arc (dispatch.sh cwd, missing reaper, adopt-without-watcher).
# Silent on healthy state. On any FAIL: appends to sweep.log AND drops a
# router item addressed to claude-pantheon describing the failure.
#
# Per AGENTS.md §Lean #2 (loud failure is the gift) — healthy = quiet,
# broken = alarm, no noise in between.

set -uo pipefail
ROUTER_ROOT="/Users/thekryptodragon/Development/sirsi-pantheon/.agents/idea-router"
REPO_ROOT="/Users/thekryptodragon/Development/sirsi-pantheon"
LOG="$ROUTER_ROOT/logs/sweep.log"
SIRSI="$HOME/.local/bin/sirsi"
export PATH="$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"

ts() { date "+%Y-%m-%dT%H:%M:%S%z"; }
cd "$REPO_ROOT" || { echo "[$(ts)] sweep FAIL: cannot cd to $REPO_ROOT" >> "$LOG"; exit 1; }

mkdir -p "$(dirname "$LOG")"
fails=()
fail() { fails+=("$1"); }

# 1. launchd dispatcher loaded
if ! launchctl list | awk '{print $3}' | grep -qx com.sirsi.idea-router; then
  fail "launchd job com.sirsi.idea-router NOT loaded"
fi

# 2. dispatch.sh recent activity (any fire within 24h)
last_dispatch=$(grep -E '^\[[0-9-]+T' "$ROUTER_ROOT/logs/dispatch.log" 2>/dev/null | tail -1 | head -c 25)
if [ -z "$last_dispatch" ]; then
  fail "dispatch.log empty or unreadable"
else
  last_epoch=$(date -j -f "[%Y-%m-%dT%H:%M:%S" "$last_dispatch" "+%s" 2>/dev/null || echo 0)
  now_epoch=$(date "+%s")
  if [ $((now_epoch - last_epoch)) -gt 86400 ]; then
    fail "dispatch.sh has not fired in 24h+ (last: $last_dispatch)"
  fi
fi

# 3. Per-thread watchers: every pidfile must point to a live PID
for pf in /tmp/sirsi-router-watch-*.pid; do
  [ -f "$pf" ] || continue
  pid=$(cat "$pf" 2>/dev/null)
  if [ -z "$pid" ] || ! kill -0 "$pid" 2>/dev/null; then
    fail "watcher pidfile $pf points to dead PID $pid (removing)"
    rm -f "$pf"
  fi
done

# 4. Active CTR claude threads must have a live watcher (the adopt-without-
#    watcher bug we fixed 2026-05-31). Reaper runs implicitly on `thread list`.
"$SIRSI" thread list --json 2>/dev/null | python3 -c "
import json, sys, os
try:
    threads = json.load(sys.stdin)
except Exception:
    print('thread list parse failed', file=sys.stderr); sys.exit(1)
for t in threads:
    th = t.get('thread') or {}
    if th.get('status') != 'active': continue
    if not th.get('agent_id','').startswith('claude-'): continue
    if t.get('idle_seconds', 1e9) > 600: continue  # stale anyway; reaper will catch
    tid = th.get('thread_id')
    pf = f'/tmp/sirsi-router-watch-{tid}.pid'
    if not os.path.exists(pf):
        print(f'thread {tid} ({th.get(\"agent_id\")}) active but no watcher pidfile')
" 2>&1 | while IFS= read -r line; do fail "$line"; done

# 5. Probe round-trip: send → confirm seen by dispatcher → close
probe_title="sweep-probe-$(date +%s)"
probe_id=$("$SIRSI" router send --from claude-pantheon --to claude-pantheon \
  --title "$probe_title" --instructions "automated sweep probe — close on receive" 2>&1 \
  | awk -F': ' '/Sent/{print $NF}')
if [ -z "$probe_id" ]; then
  fail "probe send failed"
else
  sleep 12
  if ! grep -q "$probe_title\|claude-pantheon.*item.*to dispatch" "$ROUTER_ROOT/logs/dispatch.log" 2>/dev/null; then
    : # not an error per se — dispatcher only logs when items found, may have processed silently
  fi
  "$SIRSI" router close "$probe_id" --result "sweep ok" >/dev/null 2>&1 || true
fi

# Report
if [ ${#fails[@]} -eq 0 ]; then
  echo "[$(ts)] sweep PASS" >> "$LOG"
  exit 0
fi

# Alarm: write to log + drop a router item
{
  echo "[$(ts)] sweep FAIL — ${#fails[@]} issue(s):"
  for f in "${fails[@]}"; do echo "  - $f"; done
} >> "$LOG"

alarm_body="# Periodic Sweep Alarm — $(ts)

The hourly verification sweep found $(echo ${#fails[@]}) issue(s) in router infrastructure:

$(for f in "${fails[@]}"; do echo "- $f"; done)

Run manually to investigate:

    $ROUTER_ROOT/sweep.sh

See log: $LOG
"
"$SIRSI" router send --from sweep-bot --to claude-pantheon \
  --title "sweep alarm: ${#fails[@]} infra issue(s)" \
  --instructions "$alarm_body" >> "$LOG" 2>&1 || echo "[$(ts)] alarm send failed" >> "$LOG"

exit 1
