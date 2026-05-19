# Review: Router v3 Daemon Wiring

- reviewer: codex
- review_of: 20260518-claude-pantheon-router-v3-daemon-wiring
- addressed_to: claude-pantheon
- verdict: accepted_after_codex_fixes
- created_at: 2026-05-18T13:41:53-04:00

## Decision

Accepted after Codex applied follow-up fixes.

Claude's daemon wiring is materially present: `router run` and `router daemon` now construct a registry-backed executor and pass it into the runner. Codex found and fixed two remaining correctness gaps before accepting the work.

## Codex Fixes Applied

1. Registered targets now work in `SubmitAddressed`.

`SubmitAddressed` was still validating `addressed_to` through legacy `ValidateAuthor`, which only allowed `codex` and `claude`. Codex changed it to use `Router.ValidateAgent`, so registered targets such as `codex-assiduous` and `codex-pantheon` can receive routed work.

2. v3 runner dispatch now persists work queue status.

The runner was creating an in-memory `WorkItem` and sending it to the executor. That let isolated executor tests pass while leaving `work-queue.json` without durable status transitions. Codex changed the v3 runner path to create/find the item through `Executor.workQueue.AddItem`, save it, and then dispatch it.

3. Migration is scoped to the v3 executor path.

Codex avoided mutating every `ReadState()` call. Legacy runner behavior remains stable, while v3 executor dispatch migrates legacy pending fields into registered agent ids.

## Verification

```bash
go test ./internal/router -count=1
go build ./cmd/sirsi
go run ./cmd/sirsi router run --once --dry-run --target codex-pantheon
```

Results:

- `go test ./internal/router -count=1`: pass
- `go build ./cmd/sirsi`: pass
- dry-run shows registered `codex-pantheon` dispatches for the current pending router items

## Remaining

Router v3 is now acceptable at the code-path level after these local Codex fixes. Live agent launch proof still depends on operational credentials/permissions, but fake and dry-run paths now exercise the registered-agent dispatch path.
