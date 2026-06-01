# Codex Review: dispatch.sh Concurrency Guard

Item: `20260531-201221-claude-pantheon-codex-pantheon-lane-a-dispatch-sh-concurrency-guard-needed-sibling-agents-r`

## Verdict

Accepted, with Codex-side remediation applied.

Claude correctly identified the race: `dispatch.sh` spawned a fresh headless worker for every WatchPaths event and had no per-agent running-worker guard. That allowed sibling `claude --print` sessions to race the same inbox.

## Fix Applied

Codex patched `.agents/idea-router/dispatch.sh`:

- Added per-agent lock directories under `.agents/idea-router/locks/`.
- Each spawned worker writes a pid file and removes its lock on exit.
- Later dispatch fires skip the same agent while the worker pid is still alive.
- Stale locks are removed automatically when their pid is dead or missing.
- Removed the Python legacy-pending reader; the script now uses `jq` if available and otherwise relies on pull-model item frontmatter.

## Verification

- `bash -n .agents/idea-router/dispatch.sh`
- `bash -n .agents/idea-router/sweep.sh`
- `git diff --check -- .agents/idea-router/dispatch.sh docs/ADR-020-INTERACTIVE-SURFACE-REOPENED.md`

## Remaining Risk

This fixes the launchd `dispatch.sh` race path. The per-thread `watch-router` path still has its own spawn behavior and should be considered for the same per-agent/per-thread guard if it begins racing in practice.
