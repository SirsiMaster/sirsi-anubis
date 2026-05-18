# Proposal: TUI Controller Refactor

- agent_id: codex-pantheon
- addressed_to: codex-pantheon
- topic: tui-controller-refactor
- status: ready-for-plan

## /plan

Codex to write the /plan for decomposing internal/output/tui.go (~2K lines) into focused controllers. This was explicitly deferred during the hardening sprint with "do not start with a large rewrite."

Now that the hardening sprint core is complete, this is the right time.

Suggested direction from the original Codex product reset:
- Introduce a small action runner interface
- Move each workflow into a focused controller: scan, clean, status, analyze, ci, memory
- Remove process-global pending state for scan/analyze/select flows
- Keep renderers dumb: input state → view string
- Make destructive actions go through a single safety gateway

## /goal

tui-controller-refactor is complete when:
- tui.go is split into focused controller files
- Each controller owns one workflow's state and transitions
- No single file exceeds ~500 lines
- All existing TUI tests pass
- Ma'at coverage for output package improves (structural ceiling removed)

## Implementation Boundary

codex-pantheon writes the /plan. claude-pantheon implements after plan approval.
