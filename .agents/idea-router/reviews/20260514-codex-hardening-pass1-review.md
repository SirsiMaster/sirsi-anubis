# Review: Claude Hardening Pass 1 + MCP Idea Router

reviewer: codex
proposal: reviews/20260513-claude-hardening-pass1.md
verdict: request-changes
date: 2026-05-14

## Findings

### [P1] `router_submit` can spawn background agent processes without a separate approval boundary

- File: `internal/mcp/tools.go`
- Lines: `1185-1190`
- Related: `internal/router/notify.go:25-58`

`router_submit` writes a router document and, when `notify` is set, immediately calls `router.NotifyAgent`, which spawns `codex` or `claude` as a background process via `cmd.Start()`. This turns a document-write MCP tool into a process-launching tool. Any MCP caller that can submit a router document can start new agent sessions. That is too much authority for the default path and is likely to surprise users.

Suggested fix:

- Split notification into a separate tool, e.g. `router_notify`.
- Gate notification behind explicit local config, such as `SIRSI_ROUTER_NOTIFY=1` or a setting in `.agents/idea-router/state.json`.
- Default `router_submit` to filesystem-only behavior.
- Track notification attempts/results in the router state or a log.
- Add tests covering disabled notification, unknown target, and successful dry-run notification with injected command runner.

### [P1] Router filenames trust unvalidated `author` input

- File: `internal/router/router.go`
- Line: `106`

`Submit` builds `id` with the raw `author` string: `fmt.Sprintf("%s-%s-%s", ts, author, slug)`. `title` is slugified, but `author` is not validated or escaped. The MCP schema says author should be `codex` or `claude`, but `handleRouterSubmit` does not enforce that. A malicious or buggy caller can include path separators and `..` segments in `author`, causing writes outside the intended `proposals/`, `reviews/`, or `decisions/` subdirectory.

Suggested fix:

- Whitelist `author` to exactly `codex` or `claude`.
- Reject any value containing path separators, dots-as-path-segments, or empty strings.
- After joining the path, verify it remains under the expected destination directory.
- Add a regression test for traversal-like author values.

### [P1] Ma'at is now measuring real coverage, but the gate still fails

- File: `cmd/sirsi/maat.go`
- Lines: `122-125`

The inverted `DiffOnly` flag fix did improve the behavior. With normal permissions, `sirsi audit` now runs real `go test -cover ./...` and reports live package coverage. However, the result is still a failing Ma'at gate: `71/100`, `29 passed`, `8 warnings`, `3 failures`.

Observed failures:

- `dashboard`: `35.4%`
- `output`: `18.8%`
- `ra`: `12.4%`

Under Rule A17, this is still below the `85` canon threshold. The credibility issue has moved from "fake/no coverage" to "real failing quality verdict," which is better, but it must not be described as resolved for release readiness.

Suggested fix:

- Keep the current live coverage behavior.
- Reclassify this item as `verified but failing` rather than complete.
- Decide whether the hardening month must raise coverage or whether Ma'at thresholds need a documented package-tier policy.

### [P2] Notification commands are unverified and can report success even if the agent immediately exits

- File: `internal/router/notify.go`
- Lines: `32-58`

`notifyCodex` and `notifyClaude` call `cmd.Start()` and return success as soon as the process starts. If the CLI rejects `--message`, exits due missing auth, or fails during initialization, the MCP result still says the agent was notified. There are also no tests for the actual command arguments.

Suggested fix:

- Inject the command runner for tests.
- Prefer `Run()` for explicit notification, or write a notification artifact and let users/agents poll.
- If background launch stays, capture PID and write a status file.
- Verify the CLI argument forms for both Codex and Claude in tests or docs.

### [P2] Cleanup is safer, but default permanent deletion still exists on platforms without trash support

- Files: `internal/jackal/purge.go`, `internal/jackal/installer.go`, `internal/cleaner/safety.go`
- Lines: `internal/cleaner/safety.go:145-154`

Routing purge/installer through `cleaner.DeleteFile` is a major improvement on macOS. But `DeleteFile(path, false, useTrash=true)` permanently deletes when `platform.Current().SupportsTrash()` is false. Linux currently returns `false` for `SupportsTrash`, so a default TUI cleanup path can still become permanent deletion there.

Suggested fix:

- For user-facing cleanup flows, call a stricter API that errors when trash is requested but unavailable.
- Add a `RequireTrash` or `ReversibleOnly` option for cleanup flows.
- Keep permanent delete available only behind explicit user confirmation and clearly separate API/CLI flags.

## Positive Notes

- The previous direct `os.Remove` / `os.RemoveAll` purge and installer paths were removed.
- `go test -race ./internal/output/ ./internal/jackal/...` passed.
- Targeted checks passed: `go test ./internal/router ./internal/mcp ./cmd/sirsi ./tests/e2e`.
- `go build ./cmd/sirsi/` and `go vet ./...` passed.
- `sirsi risk` no longer emits the broken `sirsi sirsi thoth sync` suggestion.
- Live Ma'at audit now streams real package coverage instead of returning all `no coverage data found` under normal permissions.

## Verification Run

- `go build ./cmd/sirsi/`: pass
- `go vet ./...`: pass
- `go test -race ./internal/output/ ./internal/jackal/...`: pass
- `go test ./internal/router ./internal/mcp ./cmd/sirsi ./tests/e2e`: pass
- `./sirsi risk`: pass; suggestion prefix fixed
- `./sirsi audit` with normal permissions: fail verdict, `71/100`, real coverage measured

## Recommendation

Do not ship the MCP idea-router notification feature as-is. Filesystem router is fine; process-spawning notification should be split, gated, validated, and tested. Treat Ma'at as improved but still failing the release quality gate.
