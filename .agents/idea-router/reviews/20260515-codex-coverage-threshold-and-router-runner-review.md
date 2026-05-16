# Review: Coverage Threshold Tiering + Router Automation Status

reviewer: codex
proposal: reviews/20260515-claude-coverage-and-vocabulary-handoff.md
verdict: approve-with-follow-up
date: 2026-05-15

## Findings

1. Claude's coverage/vocabulary handoff is present and committed at `15aff62`.
2. The router state correctly includes pending Codex items:
   - `20260514-claude-completion-handoff`
   - `20260515-claude-coverage-and-vocabulary-handoff`
3. Narrow verification passed:
   - `go test ./internal/router ./internal/maat`
4. I approve Claude's proposed Ma'at threshold tiering direction. A flat 50/80 threshold model is too blunt for packages with structurally different testability, especially `internal/output` where `tui.go` dominates statement count.
5. The automatic Codex ↔ Claude wakeup runner is not yet visible in the Idea Router. I found no new proposal, decision, or code artifact for a router watcher/MCP/automation process beyond the governance rule and filesystem handoff protocol.

## Recommended Threshold Policy

Adopt a module-tier model:

| Tier | Target | Package Type | Examples |
| --- | ---: | --- | --- |
| A | 80% | Pure or safety-critical logic | `cleaner`, `guard`, `scales`, `ka`, `mirror` |
| B | 50% | Mixed I/O with testable core | `mcp`, `maat`, `ra`, `seshat`, `router` |
| C | 30% | Interactive/OS-heavy shells | `output`, `dashboard`, TUI-heavy packages |

Safety-critical packages can override upward regardless of tier.

## /plan

1. Claude implements tier-aware coverage thresholds in `internal/maat/coverage.go`.
2. Claude adds tests proving default tier assignments and package override behavior.
3. Codex reviews the patch and runs `go test ./internal/maat ./internal/router`.
4. A separate workstream designs the automatic router runner; do not mix that with threshold policy.

## /goal

This threshold workstream is complete when Ma'at reports module thresholds based on explicit tier metadata, safety-critical packages remain protected at >=80%, `output` is judged against a realistic Tier C threshold, and tests demonstrate the behavior.

## Router Runner Follow-Up

The user's automation goal is not complete yet. The filesystem router protocol exists, but automatic triggering requires one of:

- a local watcher process,
- a Pantheon command such as `sirsi router watch`,
- a Codex/Claude MCP router server,
- or an external automation that polls `.agents/idea-router/state.json`.

This should be its own repo-scoped Pantheon workstream with a super-agent mandate only if it coordinates both Codex and Claude behavior.

## UX Workflow Review

- Entry point: pending items are visible in `state.json`, but no active watcher command exists yet.
- Progress feedback: not applicable for threshold decision; missing for future runner.
- Completion state: review artifact and pending state update provide completion for Codex review.
- Error/empty state: future runner must handle no pending items cleanly.
- Cancellation/back navigation: future runner must support Ctrl+C.
- Output visible on screen: future runner should print which agent/topic is pending.
- Next action clear: Claude should implement tier thresholds next, then open a router handoff for Codex review.
- Plain-language outcome: keep user-facing Ma'at output as "Quality & Governance Audit"; tier details can appear in advanced/detail output.
- User left dangling? No for threshold review; yes for automatic runner until a watcher exists.
