# Review: Hardening Sprint Completion Handoff

reviewer: claude
proposal: decisions/20260513-codex-product-reset.md
verdict: ready-for-review
date: 2026-05-14

## What Was Done (commits 07f8cc7..4c588e7)

### Safety (Rule A1)
- All purge/installer deletion routes through cleaner.DeleteFileReversible
- No silent permanent deletion on trashless platforms (errors instead)
- Protected paths validated before every deletion
- moveToTrash() dangerous fallback removed entirely

### Error Visibility
- engine.Scan() errors now shown via output.Warn (partial results preserved)
- ghostScanner.Scan() errors captured in both runWeigh and runKa
- TUI streaming errors propagated through streamLineMsg.err
- Ghost cleanup reports partial failures instead of silent skip
- Completion banner shows "Failed" when deityState == stateFailed

### Vocabulary (Outcome-First)
- 52 output.Header() calls converted from deity names to plain outcomes
- Zero deity headers remain in any output.Header() call
- Shared Banner: "One Install. Everything Clean." (no deity row)
- Root help sections: "Clean My Machine", "Fix My Environment", "Keep Shipping"
- Version: module names are Fleet, Memory, Quality, Health, etc.
- Custom help renamed to "guides" (eliminates duplicate help entry)
- README, CLAUDE.md, getting-started, anubis guide, isis guide, index.html updated

### Router
- Inbox semantics: PendingForCodex/PendingForClaude state tracking
- SubmitAddressed validates before writing (no false-success)
- PollInbox peeks without clearing; AckInbox clears explicitly
- router_notify gated behind SIRSI_ROUTER_NOTIFY=1
- Author whitelist + path containment on all writes
- Injectable CommandRunner for test mocking

### Tests & Coverage
- 37+ new tests across output, ra, oplog, jackal, router, mcp
- Race detection clean on all critical packages
- 40 packages, 0 failures

### Docs
- EMBEDDING.md: 4-tier classification, nexus/ adapter as extraction path
- UX_WORKFLOWS.md: pass/fail audit of 9 CLI workflows
- UX review checklist added to router DESIGN.md template

## UX Workflow Review

- Entry point: `sirsi --help` shows outcome-first sections. All commands discoverable.
- Progress feedback: TUI has spinners. CLI commands still lack spinners (documented in UX_WORKFLOWS.md).
- Completion state: All commands print results with timing.
- Error/empty state: Scan/ghost errors now visible. Empty states handled.
- Cancellation: Ctrl+C works. TUI Esc works.
- Output visible: Yes — gold/black branded output on all commands.
- Next action clear: "What's Next" suggestions on most commands.
- Plain-language outcome: All headers are outcome labels. Zero deity names in output.
- Internal names hidden: Module names only in `sirsi <module> <verb>` advanced syntax.
- User left dangling? No — all flows complete with visible results or clear errors.

## What Remains

1. **Ma'at gate**: 70/100 (needs 85 for stable release). Coverage gap in output (21%) and ra (20%).
2. **CLI spinners**: Long operations lack progress feedback in CLI-only mode.
3. **TUI controller refactor**: tui.go is ~2K lines (Codex deferred this).
4. **Linux trash**: DeleteFileReversible errors on Linux — needs freedesktop trash spec.

## Recommendation

This is shippable as v0.21.1-beta. The safety, vocabulary, error handling, and router are all done. Remaining work is coverage polish and platform breadth — neither is safety-critical.

Requesting Codex review of commits 07f8cc7..4c588e7 (10 commits total).
