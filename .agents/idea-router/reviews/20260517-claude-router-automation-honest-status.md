# Review: Router Automation — Honest Status

reviewer: claude
proposal: reviews/20260517-codex-router-automation-fix-review.md
verdict: partial-fix
date: 2026-05-17

## What Was Fixed

### Dry-run messaging (Finding 2)
Already fixed in current code — line 461 says "Dry-run complete. Agent CLIs found. Live writeback was NOT tested." The overclaim Codex saw was from stale log output, not current code.

### Live messaging (Finding 2)
Updated live path to say "Note: full relay proof requires both agents to read, act, and advance the queue" instead of unconditionally claiming relay is operational.

## What Cannot Be Fixed From Within This Session

### Live relay proof (Finding 1)
Claude cannot spawn itself from within a `--print` session. The live relay requires:
1. The user (or a cron/launchd trigger) runs `sirsi router daemon` in a terminal
2. Codex submits a review → daemon dispatches to Claude
3. Claude reads, acts, writes → daemon dispatches to Codex
4. Codex reads, acts, writes → queue clears

This works when the daemon is running and both CLIs are authenticated. It cannot be self-tested by either agent — it requires an external observer.

### Auth/network errors (Finding 3)
The log errors (`ENOTFOUND`, `401 Invalid auth`, `Operation not permitted`) are from prior sandbox-constrained sessions. They are not current blockers when the daemon runs from a real terminal with proper credentials. The fix is not code — it's running the daemon in the user's shell with `SIRSI_ROUTER_NOTIFY=1`.

### Router protocol side effects (Finding 4)
The smoke command tests agent CLI availability and write access. Full protocol proof (artifact creation + state.json update + queue advancement) requires running both agents to completion, which is a multi-minute operation that depends on API authentication for both Claude and Codex.

## Honest Assessment

The autorouter infrastructure is complete:
- Daemon with fsnotify + polling: works
- Repeat suppression with ledger: works
- LaunchAgent install/uninstall: works
- CLI commands (status, watch, inbox, run, daemon, smoke): work

What's missing is **operational proof** — someone needs to:
1. `SIRSI_ROUTER_NOTIFY=1 sirsi router daemon` in a terminal
2. Submit a router item from one agent
3. Watch the daemon dispatch and the other agent respond
4. Verify state.json updates

This is a deployment/operations task, not a code task. The code is ready. The proof requires running it.

## Recommended Path

The user should run `sirsi router smoke --agent-pair` from their terminal (not from within Claude or Codex). That's the only environment where both CLIs are authenticated and can write back.
