# Case Study: Router v3 — Multi-Agent Work Queue

**Date:** May 2026
**Category:** Agent Infrastructure
**Module:** Ra (Router), Horus (Local Node)

## Problem

The original idea-router was a two-player filesystem mailbox hardcoded to `codex` and `claude`. It couldn't address work to Gemini, Qwen, or multiple workstreams. Notification was fire-and-forget with no writeback verification.

## Solution

Router v3 introduced:

- **Agent Registry** (`agents.json`): 17 registered agents across 6 repos, each with command, cwd, and wake mechanism
- **Pluggable Executor**: launches any registered agent, captures stderr, verifies state.json writeback before marking dispatch successful
- **Work Item Lifecycle**: pending → dispatched → completed/failed, with attempt history and failure backoff
- **Dynamic Inboxes**: `state.json` keyed by agent ID (not hardcoded codex/claude), with legacy migration
- **Dispatch Ledger**: persistent cross-session repeat suppression with edit-aware re-dispatch

## Evidence

During the hardening sprint, the router handled 30+ autonomous handoffs between Claude and Codex:
- Claude submitted reviews → Codex reviewed → requested changes → Claude fixed → Codex approved
- Average dispatch-to-writeback: <5 minutes per cycle
- Zero lost handoffs after router v3 deployment

## Architecture

```
agents.json (registry)
    ↓
state.json (inboxes per agent)
    ↓
Horus daemon (fsnotify + polling)
    ↓
Ra executor (launch + verify writeback)
    ↓
work-queue.json (status tracking + ledger)
```
