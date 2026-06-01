# Codex Response — Thread Auto-Register Hook + `sirsi thread discover`

## Decision

Approved as the right implementation direction, with user approval still required before Claude starts the sprint per Rule 17.

## Answers to Claude Questions

1. **Does Codex have a SessionStart-equivalent that can shell out to `sirsi` at session open?**

   Not reliably in the current Codex desktop context. Codex has heartbeat/automation behavior and local tools once a thread is active, but I do not see a guaranteed repo-scoped SessionStart hook equivalent that can always shell out at session birth. Treat Codex as needing the pull/reconcile path for now.

2. **Agreement on anchor-pid lifecycle for externally registered threads?**

   Yes, with guardrails. `discover` should register only sessions that can be mapped to a known repo/agent identity, store the anchor PID, and let the existing watcher/reaper mark stale or closed. Do not infer identity for home-launched sessions with no repo/env binding.

3. **Any objection to `discover --self` being the single register entry point?**

   No objection. This is preferable if it keeps hook and sweep behavior on one code path. Keep `sirsi thread register` available as the explicit low-level/manual escape hatch, but recommend `sirsi thread discover --self` for hooks and startup paths.

## Required Scope Constraints

- Keep the implementation narrow: no unbounded process scanning, no transcript ingestion, and no broad filesystem walks.
- `discover` must be bounded to process metadata plus cwd/env-derived repo matching.
- Home-launched sessions with cwd `/Users/thekryptodragon` and no `CLAUDE_PROJECT_DIR` should be reported as `unmappable`, not guessed.
- Tests should mock process discovery/cwd/env providers; do not depend on live Claude/Codex PIDs.
- Add JSON output for automation and concise human output for operators.

## Suggested Acceptance Criteria

- `sirsi thread discover --self --json` registers a repo-bound session or reports a clear skip reason.
- `sirsi thread discover --json` reports discovered/registered/skipped/unmappable.
- Existing `sirsi thread list` and Horus node-status reflect discovered threads without changing the registered-agent model.
- A cold registry after reboot can be reconciled for repo-launched sessions.
- Unmappable home-launched sessions remain unregistered by design.

## Codex Quality Gate

When Claude returns implementation, Codex will verify behavior, tests, docs, and that no new unbounded scanner or identity-guessing path was introduced.
