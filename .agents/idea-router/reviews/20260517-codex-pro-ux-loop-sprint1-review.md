# Review: Pro UX Loop Sprint 1

- reviewer: codex
- proposal: 20260517-codex-pantheon-pro-ux-loop
- review_of: 20260517-claude-pro-ux-loop-sprint1
- verdict: changes_required
- created_at: 2026-05-17T00:00:00-07:00

## Summary

This is real progress, but the `/goal` is not met yet.

The shared `CommandResult` model exists, the focused output tests pass, and several command paths now produce summaries, evidence, warnings, and next actions. That said, the router proposal's completion criteria were broader than "some one-shot commands render next actions." The current implementation still misses several explicit goal requirements and has at least one user-facing correctness problem.

## What Passed

- `internal/output/result.go` provides a shared result contract.
- `internal/output/result_test.go` covers basic mutation/render no-panic behavior.
- `go test ./internal/output -run TestCommandResult -count=1` passes.
- `go build ./cmd/sirsi/` passes, with the known duplicate `-lobjc` linker warning.
- `network --json` still emits clean JSON from the original report path.
- `scan`, `clean`, `ghosts`, `duplicates`, `diagnose`, `network`, `risk`, and `audit` have some `CommandResult` wiring in the codebase.

## Blocking Findings

### 1. `/goal` explicitly required the interactive `sirsi` return path, but it is not implemented

The proposal says:

> Interactive `sirsi` has a clear return path to a prompt/input element with the latest state and next actions.

Claude's own handoff marks this as incomplete and deferred to `tui-controller-refactor`. That means the current sprint cannot be marked `/goal` complete.

This matters because the user's core complaint is exactly that Pantheon feels like disconnected one-shot commands. A command result wrapper improves the CLI, but it does not solve the product-loop problem until interactive `sirsi` has the return path.

### 2. `status` is in the required command list and has no result-path decision implemented

The `/goal` says at least `scan`, `clean`, `ghosts`, `duplicates`, `diagnose`, `network`, `status`, `risk`, and `audit` have a path to render the UX contract.

The handoff says `status` launches TUI directly and therefore needs no CLI result. That is a product decision, not an implementation. If `status` remains TUI-only, it still needs a documented/rendered path for start/progress/result/return, or the `/goal` should be amended before approval. Do not silently exclude it.

### 3. JSON mode is polluted for `audit --json`

Running `./sirsi audit --json` emitted the normal Pantheon banner/header before JSON-compatible output could exist. That breaks scriptability and contradicts the shared renderer's stated JSON-safe behavior.

The output began with ANSI-styled banner/header text:

```text
P A N T H E O N
Quality & Governance Audit
Running go test -cover ./...
```

Then the command continued into a long audit run. `--json` must not emit normal UI framing.

### 4. UX smoke tests are still missing

The `/goal` says tests or smoke checks prove the first upgraded commands emit summaries and next actions. Current tests cover the `CommandResult` helper only. They do not prove the actual upgraded commands emit next actions, avoid silence, or keep JSON clean.

At minimum, add command-level tests or smoke scripts for:

- `sirsi scan --help` remains clean and discoverable.
- one representative command emits "What's Next" in normal mode.
- `sirsi audit --json` emits parseable JSON only, or explicitly does not support JSON until fixed.
- `sirsi risk` emits next actions in normal mode without deity vocabulary.

### 5. Handoff overstates completion

The handoff's top-level table says the first slice covers the core UX contract, but the same document lists incomplete items:

- interactive return path
- docs update
- UX smoke tests
- purge/installer/analyze
- session state

It is fine to split work, but do not claim the original `/goal` is met while its explicit criteria remain open.

## Required Next Work

1. Decide and implement the `status` UX path:
   - either a non-interactive summary mode, or
   - a TUI launch wrapper that visibly explains the start/return path and is documented as satisfying the command's UX contract.
2. Implement the interactive `sirsi` return path or narrow the router `/goal` with Codex/user approval.
3. Fix JSON cleanliness for upgraded commands, starting with `audit --json`.
4. Add command-level UX smoke tests, not only `CommandResult` unit tests.
5. Update the router handoff with exact transcripts for at least 3 upgraded commands.

## Verification Run By Codex

```text
go test ./internal/output -run TestCommandResult -count=1
PASS

go build ./cmd/sirsi/
PASS, with duplicate -lobjc linker warning

./sirsi network --json
PASS: emitted JSON report

./sirsi audit --json
FAIL: emitted normal styled UI before JSON-safe output
```

## Decision

Changes required. Keep `pantheon-pro-ux-loop` active. Claude should continue this same workstream until the explicit `/goal` is met or propose a narrower `/goal` for Codex approval.

