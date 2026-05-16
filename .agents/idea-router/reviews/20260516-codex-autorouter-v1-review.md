# Review: Autorouter v1 Implementation

reviewer: codex
proposal: reviews/20260516-claude-autorouter-v1-implementation.md
verdict: request-changes
date: 2026-05-16

## Findings

1. **Invalid `--target` is silently accepted.** `sirsi router run --once --dry-run --target banana` exits successfully with `No pending dispatches.` That should be rejected. The command contract says target is `codex|claude|all`; accepting arbitrary values makes operator mistakes look like a clean queue.
2. **Notification gating docs/behavior are inconsistent.** The handoff and command help say notification uses `NotifyAgent` and is gated behind `SIRSI_ROUTER_NOTIFY=1`, but `internal/router/notify.go` does not check that environment variable. If autorouter v1 is intended to actually launch agents by default, remove the gating claim and make the help explicit. If the gate is intended, implement it in `routerRunCmd` or `NotifyAgent` and add a test.

## What Looks Good

- `internal/router/runner.go` is simple and testable.
- `RunnerOptions.Notify` avoids launching real agents in tests.
- `--dry-run` prints the dispatch clearly.
- Inbox items are not auto-acked.
- Repeat suppression is session-local and covered by tests.
- Notification failure does not mark an item dispatched, so retry behavior is sane.
- `router-runner-v1-auto-trigger` remains active pending Codex approval.

## Verification

Passed:

- `gofmt -l internal/router/runner.go internal/router/runner_test.go cmd/sirsi/routercmd.go`
- `go test ./internal/router ./cmd/sirsi`
- `go test ./internal/maat ./internal/router`
- `go test ./...`
- `go run ./cmd/sirsi router run --once --dry-run`
- `go run ./cmd/sirsi router run --once --dry-run --target codex`
- `go run ./cmd/sirsi router run --once --dry-run --target claude`

Observed:

- `--target codex` shows the pending Codex dispatch.
- `--target claude` reports no pending dispatches, as expected.
- `--target banana` currently reports no pending dispatches instead of failing.

## Required Changes

1. Validate `runTarget` before constructing the runner. Accept only `all`, `codex`, or `claude`.
2. Resolve the notification-gating mismatch:
   - either implement `SIRSI_ROUTER_NOTIFY=1` gating and test it;
   - or update help/handoff language to say `sirsi router run` launches the target CLI by default, while `--dry-run` is the safe preview mode.
3. Add tests for whichever notification-gating behavior is chosen.

## /goal Review

Autorouter v1 is very close, but not complete until target validation and notification semantics are unambiguous. Do not close `router-runner-v1-auto-trigger` yet.

## UX Workflow Review

- Entry point: `sirsi router run --once --dry-run` is good.
- Progress feedback: one-shot dispatch output is clear.
- Completion state: once mode exits cleanly.
- Error/empty state: invalid target currently looks like empty state; fix required.
- Cancellation/back navigation: loop uses context/signal path.
- Output visible on screen: yes.
- Next action clear: fix target validation and notification gating/docs, then resubmit to Codex.
- Plain-language outcome: good once the gate wording is aligned.
- User left dangling? Not operationally, but v1 should remain open until the two fixes land.
