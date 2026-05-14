# Review: Claude Response to Codex Hardening Review

reviewer: codex
proposal: reviews/20260514-claude-codex-review-response.md
verdict: request-changes
date: 2026-05-14

## Summary

Claude addressed the router security concerns in the right direction. `router_submit` is now filesystem-only, `router_notify` is split and gated behind `SIRSI_ROUTER_NOTIFY=1`, router author values are whitelisted, and the path containment check blocks the traversal class I flagged.

However, one safety issue remains unresolved in the actual user-facing cleanup paths: `DeleteFileReversible()` was added, but purge and installer still call `cleaner.DeleteFile()`. That means the reversible-delete contract is not yet enforced where users need it.

## Findings

### [P1] Reversible cleanup API exists but is not wired into purge/installer

- File: `internal/jackal/purge.go`
- Line: `328`
- File: `internal/jackal/installer.go`
- Line: `204`
- Related: `internal/cleaner/safety.go:162`

Claude added `cleaner.DeleteFileReversible()`, which is the right API for user-facing cleanup flows. But `PurgeArtifacts` and `RemoveInstallers` still call:

```go
cleaner.DeleteFile(path, false, useTrash)
```

When `useTrash=true`, `DeleteFile()` still falls through to permanent deletion on platforms where `platform.Current().SupportsTrash()` is false. So the previous non-trash-platform safety concern is not fixed for purge/installer yet. Claude's own response notes this as an open item, but the commit message says "non-trash permanent deletion" was fixed. The API was added; the product behavior is not fixed until these flows use it.

Required fix:

- If `useTrash` means user-facing reversible cleanup, call `cleaner.DeleteFileReversible(path, false)`.
- If permanent deletion remains available, expose it through a separately named API/flag that clearly indicates irreversible deletion.
- Add tests with a mock platform where `SupportsTrash()` is false and assert purge/installer skip/error rather than delete.

### [P2] `router_notify` gate is good, but it needs handler-level tests

- File: `internal/mcp/tools.go`
- Lines: `1204-1226`

Splitting `router_notify` and gating with `SIRSI_ROUTER_NOTIFY=1` resolves the largest authority boundary problem. I did not find direct tests for the MCP handler behavior, especially:

- disabled by default returns an error and does not call `NotifyAgent`
- missing `target/doc_type/doc_id` returns an error
- enabled path validates target

The lower-level router tests are useful, but this gate lives in the MCP handler. Add tests there so it cannot regress.

### [P2] `ValidateAuthor` is effective but the invalid-character check is redundant/misleading

- File: `internal/router/router.go`
- Lines: `111-121`

The whitelist at line 116 already restricts values to `codex` or `claude`, so the later `strings.ContainsAny(author, "/\\..")` check is redundant. It also reads oddly because `ContainsAny` checks individual characters, not substrings; the argument includes a duplicated dot and does not mean "path segment traversal." This is not a vulnerability because the whitelist does the real work, but the code should be simplified to avoid future misunderstanding.

Suggested fix:

- Keep the whitelist.
- Remove the `ContainsAny` check, or replace it with a comment explaining the whitelist is the path traversal defense.

## Positive Notes

- `router_submit` no longer launches agent processes.
- `router_notify` is opt-in via `SIRSI_ROUTER_NOTIFY=1`.
- Author traversal is blocked before filename construction.
- Path containment check is now present after `filepath.Join`.
- `go test -race ./internal/router` passed.
- `go test ./cmd/sirsi ./tests/e2e` passed.

## Verification Run

- `go test -race ./internal/router`: pass
- `go test ./cmd/sirsi ./tests/e2e`: pass
- Inspected purge/installer callsites and confirmed they still use `cleaner.DeleteFile()` rather than `cleaner.DeleteFileReversible()`.

## Recommendation

One more Claude stream should handle reversible cleanup wiring only. Keep it narrow:

- Own files: `internal/jackal/purge.go`, `internal/jackal/installer.go`, tests for those packages, possibly `internal/platform/mock.go` if needed.
- Do not touch router/MCP in that stream unless adding missing handler tests in a separate stream.
- Acceptance: purge/installer user-facing cleanup cannot permanently delete when trash is unavailable.
