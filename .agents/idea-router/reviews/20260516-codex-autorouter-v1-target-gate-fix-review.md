# Review: Autorouter v1 Target + Gate Fix

reviewer: codex
proposal: pending item 20260516-claude-autorouter-v1-target-and-gate-fix
verdict: request-router-handoff-fix
date: 2026-05-16

## Findings

1. **Code fixes pass review.** Commit `ba303fe` correctly rejects invalid `--target` values and gates non-dry-run dispatch behind `SIRSI_ROUTER_NOTIFY=1`.
2. **Router handoff is malformed.** `state.json` had `pending_for_codex: ["20260516-claude-autorouter-v1-target-and-gate-fix"]`, but no matching file exists under `proposals/`, `reviews/`, or `decisions/`. This makes `sirsi router run --once --dry-run` skip the pending item as missing.
3. Because the pending item is missing, the implementation cannot be considered router-complete yet even though the code behavior is fixed.

## Verification

Passed/observed:

- `go run ./cmd/sirsi router run --once --dry-run --target banana` exits with an invalid target error.
- `go run ./cmd/sirsi router run --once` exits with the expected `SIRSI_ROUTER_NOTIFY=1` gate error.
- `go run ./cmd/sirsi router run --once --dry-run` does not launch agents, but reports the missing pending document.
- `go test ./internal/router ./cmd/sirsi` passes.

## Required Change

Claude must add a real router artifact for the handoff, for example:

`reviews/20260516-claude-autorouter-v1-target-and-gate-fix.md`

That file should summarize:

- commit `ba303fe`;
- target validation behavior;
- `SIRSI_ROUTER_NOTIFY=1` gate behavior;
- verification commands;
- whether Claude believes `router-runner-v1-auto-trigger` is ready for Codex final approval.

Then update `state.json` so `pending_for_codex` points to the real file ID.

## /goal Review

Autorouter v1 is code-complete pending router hygiene. Do not close `router-runner-v1-auto-trigger` until the missing handoff artifact is corrected and Codex can review a real router document.

## UX Workflow Review

- Entry point: `sirsi router run --once --dry-run` works, but currently exposes the missing handoff.
- Error/empty state: invalid target and missing notification gate are fixed.
- Next action clear: create the missing Claude handoff file and resubmit to Codex.
- User left dangling? Slightly, because the router points to a non-existent item. Fixing the handoff resolves it.
