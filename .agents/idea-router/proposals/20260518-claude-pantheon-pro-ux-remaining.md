# Proposal: Pantheon Pro UX Loop — Remaining Items

- agent_id: claude-pantheon
- addressed_to: claude-pantheon
- topic: pantheon-pro-ux-loop
- status: ready-for-execution

## /plan

Complete the remaining Pro UX loop items from the Codex proposal:

1. Wire remaining commands to CommandResult (status already launches TUI — skip)
2. Session state persistence: last scan result, last cleanup, pending recommendations stored in ~/.config/pantheon/tui-state.json, restored on TUI launch
3. UX smoke tests: automated tests that verify key commands emit CommandResult fields (summary, evidence, next_actions)
4. Update README quickstart and docs/UX_WORKFLOWS.md to reflect current behavior

## /goal

pantheon-pro-ux-loop is complete when:
- Session state persists and restores across TUI sessions
- At least 3 UX smoke tests prove CommandResult output
- README quickstart matches current first-run experience
- UX_WORKFLOWS.md reflects all passing commands

## Implementation Boundary

claude-pantheon owns this. sirsi-pantheon repo only.
