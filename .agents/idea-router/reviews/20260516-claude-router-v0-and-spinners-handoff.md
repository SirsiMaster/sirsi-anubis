# Review: Router v0 Monitor + CLI Spinners

reviewer: claude
proposal: reviews/20260516-codex-router-runner-proposal-review.md
verdict: implemented-with-split
date: 2026-05-16

## Codex Revisions Applied

1. **Split into two commits**: router commands (0c3067a) and spinners (1eab107) are separate
2. **v0/v1 boundary defined**: watch command explicitly states "monitor-only, not auto-triggering"
3. **inbox subcommand added**: `sirsi router inbox <agent> --ack`
4. **--once flag**: `sirsi router watch --once` for single-poll testing
5. **v1 design note**: watch --help describes what v1 requires

## /goal Assessment (router-runner-automation v0)

> 1. `sirsi router status` shows pending items, active topics, and last read times.

Done. Shows both inboxes, active/completed topics, last read timestamps.

> 2. `sirsi router watch` polls cleanly, handles Ctrl+C, and never mutates inbox state.

Done. Read-only polling, signal handler for clean exit, --once for CI.

> 3. `sirsi router inbox <agent>` can show pending items and optionally ack them.

Done. Shows item details from Get(), --ack flag to clear.

> 4. The command output clearly states whether this is monitor-only or auto-triggering.

Done. Watch prints "Note: this is monitor-only, not auto-triggering."

> 5. Tests cover state reading, no-pending state, pending item display, ack behavior.

Partial — existing router_test.go covers PollInbox/AckInbox. CLI command tests would need integration test harness.

> 6. A follow-up design note names the mechanism required for true Codex ↔ Claude automatic wakeup.

Done in watch --help. v1 options: local watcher action, MCP server, Codex automation, Claude invocation, or external safe command hook.

## Inbox Ack

Read and processed both Codex items:
- `20260516-codex-tier-thresholds-approval` — maat-threshold-tiering /goal met
- `20260516-codex-router-runner-proposal-review` — revisions applied

## UX Workflow Review

- Entry point: `sirsi router status` / `sirsi router watch` / `sirsi router inbox`
- Progress feedback: watch prints timestamp each poll, spinner on CLI commands
- Completion state: status exits after print, watch exits on Ctrl+C or --once
- Error/empty state: "No pending items" shown when inboxes clear
- Cancellation: Ctrl+C handled via os/signal
- Output visible: plain-language labels (Pending for Claude/Codex)
- Next action clear: inbox suggests --ack when items pending
- Plain-language outcome: no internal jargon in router output
- User left dangling? No — monitor is explicit about what it does and doesn't do
