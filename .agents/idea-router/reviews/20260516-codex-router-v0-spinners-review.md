# Review: Router v0 Monitor + CLI Spinners

reviewer: codex
proposal: reviews/20260516-claude-router-v0-and-spinners-handoff.md
verdict: request-changes
date: 2026-05-16

## Findings

1. **Spinner suppression is incomplete.** `output.Spinner` suppresses for `SIRSI_TUI=1` and `SIRSI_JSON=1`, but the CLI `--json` flag sets the package variable `JsonOutput`; it does not set `SIRSI_JSON`. The global `--quiet` flag also is not reflected in `Spinner`. This means long commands can still write spinner frames to stderr during JSON or quiet output.
2. **Formatting cleanup is needed.** `gofmt -l` reports:
   - `cmd/sirsi/routercmd.go`
   - `cmd/sirsi/anubis.go`
   - `cmd/sirsi/main.go`
3. **Router v0 behavior is correctly monitor-only.** `status`, `watch --once`, and `inbox codex` work from the repo root and do not mutate inbox state without `--ack`.
4. **Router v1 auto-trigger is still open.** The active topic `router-runner-v1-auto-trigger` is correct. v0 should not be represented as satisfying the user's automatic trigger goal.

## Verification

Passed:

- `go test ./cmd/sirsi ./internal/output ./internal/router`
- `go test ./internal/maat ./internal/router`
- `go run ./cmd/sirsi router status`
- `go run ./cmd/sirsi router watch --once`
- `go run ./cmd/sirsi router inbox codex`

Notes:

- The shell emitted `/Users/thekryptodragon/.rvm/scripts/rvm:29: operation not permitted: ps` under the sandbox, but commands completed.
- macOS linker warnings about duplicate `-lobjc` appeared under `go run`/tests, but builds/tests completed.

## Required Changes

1. Fix spinner suppression by either:
   - setting an output package state from root `PersistentPreRun`, for example `output.SetMode(JsonOutput, quietMode)`, or
   - setting environment variables from the CLI flags before any spinner can start.
2. Run `gofmt` on the touched Go files.
3. Keep `router-runner-v1-auto-trigger` active until a real trigger mechanism exists.

## /goal Review

`router-runner-automation-v0` is functionally close but not fully review-clean until the spinner suppression and formatting issues are fixed. `cli-progress-spinners` should remain pending cleanup.

## UX Workflow Review

- Entry point: `sirsi router status/watch/inbox` works.
- Progress feedback: `watch --once` gives visible pending-state output.
- Completion state: `status` and `inbox` exit cleanly.
- Error/empty state: inbox has clear no-pending behavior.
- Cancellation/back navigation: watch code has Ctrl+C handling; not manually exercised in this review.
- Output visible on screen: yes.
- Next action clear: fix spinner suppression + gofmt, then ask Codex for re-review.
- Plain-language outcome: good; v0 monitor boundary is clear.
- User left dangling? Not for v0 monitor; the true auto-trigger goal remains a separate v1 workstream.
