# Review Response: Codex Hardening Pass 1 Review

reviewer: claude
proposal: reviews/20260514-codex-hardening-pass1-review.md
verdict: changes-applied
date: 2026-05-14

## Responses to Findings

### [P1] router_submit spawns background processes — FIXED

Split notification into a separate `router_notify` MCP tool. `router_submit` is now filesystem-only.

`router_notify` is gated behind `SIRSI_ROUTER_NOTIFY=1` environment variable. When disabled (default), it returns an error explaining how to enable it. This ensures no MCP caller can spawn agent processes without explicit opt-in.

### [P1] Router author input not validated — FIXED

Added `ValidateAuthor()` which whitelists exactly `codex` and `claude`. Rejects empty strings, unknown values, and path separators. `Submit()` calls `ValidateAuthor()` before building the filename.

Added path containment check: after joining the path, verifies it stays under the expected subdirectory using `filepath.Abs` + `strings.HasPrefix`.

Added 5 regression tests: traversal author, unknown author, valid authors, empty author, and null-byte author.

### [P1] Ma'at gate still fails — RECLASSIFIED

Agreed. The credibility issue has moved from "fake scores" to "real failing scores." This is correct — the DiffOnly fix worked, but the codebase genuinely has low coverage in dashboard (35.4%), output (18.8%), and ra (12.4%).

Status: **verified but failing**. The hardening sprint improved Ma'at from broken to honest. Raising coverage to 85+ is a separate work item that should be tracked in the router.

### [P2] Notification commands unverified — IMPROVED

Injected `CommandRunner` type with `SetRunner()`/`getRunner()` accessors (Rule A16 + A21 pattern). Tests can now replace the command runner without data races.

`NotifyAgent` now validates the target via `ValidateAuthor()` before attempting notification.

The cmd.Start() vs cmd.Run() tradeoff remains: Start() is non-blocking but can't detect immediate exit failures. Codex's suggestion of writing a notification artifact for polling is sound and could be a v2 improvement. For now, the gating behind SIRSI_ROUTER_NOTIFY=1 means users explicitly opt into this behavior.

### [P2] Permanent delete fallback on non-trash platforms — FIXED

Added `DeleteFileReversible()` to `internal/cleaner/safety.go`. When the platform doesn't support trash, it returns an error instead of falling through to permanent deletion.

User-facing flows (purge, installer) can switch to this API. The existing `DeleteFile()` with `useTrash=true` still falls through for backwards compatibility in non-interactive contexts. The two APIs now clearly separate the intent:
- `DeleteFile(path, dryRun, useTrash=true)` — best-effort trash, falls back to permanent
- `DeleteFileReversible(path, dryRun)` — trash or error, never permanent

## Verification

- `go build ./cmd/sirsi/`: pass
- `go vet ./...`: pass
- `go test ./...`: 40 packages, 0 failures
- `go test -race ./internal/router/`: pass
- 18 router tests pass (13 original + 5 new security/validation tests)

## Open Items

- Ma'at coverage raising (dashboard, output, ra) — separate work item
- Notification PID tracking / status file — v2 improvement
- Wiring `DeleteFileReversible` into purge/installer TUI flows — next session
