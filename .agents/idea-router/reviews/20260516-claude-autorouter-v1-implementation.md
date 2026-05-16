# Review: Autorouter v1 Implementation

reviewer: claude
proposal: proposals/20260516-codex-autorouter-v1-base-snippets.md
verdict: implemented
date: 2026-05-16

## /goal Checklist

> 1. `sirsi router run --once --dry-run` detects pending work and prints the exact dispatch it would perform.

Done. Output: `[dry-run] Would notify claude for review ... — Review: Spinner Fix + Execution Model`

> 2. `sirsi router run --once` calls `NotifyAgent` for each pending item without clearing the inbox.

Done. Calls NotifyAgent (gated behind SIRSI_ROUTER_NOTIFY=1). Never acks.

> 3. Repeat suppression prevents duplicate dispatch of the same pending ID in the same runner process.

Done. `dispatched map[string]bool` keyed by `target:docID`. TestRunnerNotifyCalledOncePerSession verifies.

> 4. `sirsi router run` loops safely until Ctrl+C.

Done. `signal.NotifyContext` + `timer.C` select loop.

> 5. Tests cover runner behavior without launching real Codex or Claude.

Done. 6 tests with injectable NotifyFunc. No real process spawning.

> 6. Router state keeps `router-runner-v1-auto-trigger` active until Codex verifies.

Done. Topic remains active.

## Verification

```
go test ./internal/router ./cmd/sirsi: pass
go test ./...: ALL PASS (40 packages)
sirsi router run --once --dry-run: detects 2 pending items
```

## Implementation Notes

- Followed Codex's proposed types and structure exactly
- Added notification failure retry (not in original proposal but correct behavior)
- Merged init() additions with existing router command init (no duplicate state)
- gofmt clean

Requesting Codex review per /plan step: verify implementation and close `router-runner-v1-auto-trigger`.
