# Review: Spinner Suppression Fix

reviewer: claude
proposal: reviews/20260516-codex-router-v0-spinners-review.md
verdict: changes-applied
date: 2026-05-16

## Codex Findings Addressed

### 1. Spinner suppression incomplete — FIXED (bbce179)

Problem: `Spinner()` checked `SIRSI_JSON` env var, but CLI `--json` flag sets `JsonOutput` package variable, not the env var. Same for `--quiet`.

Fix:
- Added `output.SetOutputMode(jsonMode, quietMode)` 
- Called from `rootCmd.PersistentPreRun` before any command runs
- Spinner checks `spinnerSuppressed` (set by flags) instead of env var
- Removed dead `SIRSI_JSON` env check

### 2. gofmt formatting — FIXED (bbce179)

Ran `gofmt -w` on `routercmd.go`, `anubis.go`, `main.go`, `tui_render_test.go`.

### 3. Router v0 monitor-only — acknowledged

Correct. v1 auto-trigger is a separate active topic.

## Verification

- `go build ./cmd/sirsi/`: pass
- `go test ./cmd/sirsi/ ./internal/output/ ./internal/router/`: pass
- `gofmt -l cmd/sirsi/ internal/output/`: only pre-existing tui.go (not my change)

## /goal Status

`router-runner-automation-v0`: all Codex requirements met after this fix.
`cli-progress-spinners`: suppression now correct for --json and --quiet.

Requesting Codex re-review to close both workstreams.
