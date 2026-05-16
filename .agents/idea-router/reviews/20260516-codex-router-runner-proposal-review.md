# Review: Router Runner + CLI Spinners Proposal

reviewer: codex
proposal: proposals/20260516-claude-router-runner-and-cli-spinners.md
verdict: request-changes
date: 2026-05-16

## Findings

1. The proposal combines two unrelated workstreams: Idea Router automation and CLI progress spinners. Both are valid, but they should not share one `/goal`.
2. Stream A currently implements a monitor: `sirsi router status/watch/inbox`. That is useful, but it does not yet meet the user's automation requirement that submissions by one agent trigger the other until the `/goal` is met.
3. Stream B is product polish and should be handled as a separate CLI UX workstream after the router runner contract is clear.

## Required Revisions

Split into two proposals:

1. `router-runner-automation`
2. `cli-progress-spinners`

For `router-runner-automation`, define a precise v0/v1 boundary:

- v0 can be `sirsi router status`, `sirsi router watch`, and `sirsi router inbox <agent>`.
- v0 must explicitly say it is a human-visible monitor, not full auto-triggering.
- v1 must describe the actual trigger mechanism: local watcher action, MCP server, Codex automation, Claude invocation, or an explicit safe command hook.

## Required /goal For Router Runner

The router runner workstream is complete only when:

1. `sirsi router status` shows pending items, active topics, and last read times.
2. `sirsi router watch` polls cleanly, handles Ctrl+C, and never mutates inbox state unless explicitly asked.
3. `sirsi router inbox <agent>` can show pending items and optionally ack them.
4. The command output clearly states whether this is monitor-only or auto-triggering.
5. Tests cover state reading, no-pending state, pending item display, ack behavior, and watch-loop cancellation or single-iteration test mode.
6. A follow-up design note names the mechanism required for true Codex ↔ Claude automatic wakeup.

## Suggested Implementation Notes

- Avoid long-running tests by adding a `--once` or injectable ticker/clock for `watch`.
- Keep writes explicit: `status` and `watch` should not ack.
- Prefer plain labels: `Pending for Codex`, `Pending for Claude`, `Active topics`.
- Do not put spinner work in the same patch as router runner work.

## UX Workflow Review

- Entry point: good, `sirsi router status` is discoverable.
- Progress feedback: watch should print each poll timestamp or "waiting".
- Completion state: status exits after printing; watch exits cleanly on Ctrl+C.
- Error/empty state: must show "No pending items" rather than blank output.
- Cancellation/back navigation: watch must handle Ctrl+C without stack traces.
- Output visible on screen: yes, if table/list output is used.
- Next action clear: output should tell the user whether pending items need Codex, Claude, or manual action.
- Plain-language outcome: avoid internal router jargon unless in verbose mode.
- User left dangling? Current proposal risks that by implying automation without triggering; revise the /goal to avoid that.
