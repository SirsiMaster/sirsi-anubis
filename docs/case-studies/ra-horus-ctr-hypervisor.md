# Case Study: Ra/Horus CTR Hypervisor — From Manual Shuttle to Autonomous Relay

**Date:** May 2026
**Category:** Architecture Innovation
**Module:** Ra (Fleet), Horus (Local)

## Problem

Building Pantheon required coordinating two AI agents (Claude and Codex) across multiple repositories. Each agent session was isolated — neither knew what the other had done, reviewed, or blocked. The user became the message bus, manually saying "check the router" to move work between agents.

This created three failures:
1. **Manual shuttle fatigue** — the user spent time ferrying context between agents instead of doing product work
2. **Lost state** — when conversations compacted or sessions ended, pending work disappeared
3. **No accountability** — neither agent knew if the other had read, acted on, or completed its assigned work

## Solution: Ra/Horus Split

### Ra 𓇶 — Sirsi-Wide Orchestration

Ra owns everything that coordinates across machines, repos, and agents:

- **Idea Router (CTR)**: filesystem-based work queue at `.agents/idea-router/`
- **Agent Registry**: `agents.json` with 8+ registered agent profiles across 6 repos
- **Pluggable Executor**: launches any registered agent, verifies writeback, tracks status
- **Autorouter Daemon**: fsnotify + polling, dispatch ledger, repeat suppression
- **Work Items**: pending → dispatched → completed/failed with attempt history

### Horus 𓂀 — Per-Desktop Runtime Node

Horus owns what's happening on THIS machine:

- **Daemon health**: is the autorouter running? launch agent status?
- **Agent visibility**: which agents are active, their windows, their state
- **Repo status**: uncommitted work, last scan, last action
- **Operator dashboard**: TUI status tab, menu bar, workstation metrics

### Supporting Layer

- **Thoth** preserves router state across context compactions
- **Ma'at** validates that /plan and /goal governance is followed
- **Net** checks plan-to-implementation alignment

## Evidence

The hardening sprint that built this system ran autonomously for 30+ commits:

1. Codex wrote a product reset decision → Claude implemented safety fixes
2. Claude submitted for review → Codex requested changes
3. Claude fixed → Codex approved → Claude continued
4. Loop repeated 8 times without the user ferrying messages (after router was built)

The autorouter daemon detects state.json changes via fsnotify and dispatches within 200ms. Repeat suppression prevents duplicate dispatch. Failed dispatches retry with exponential backoff.

## Result

| Before | After |
|--------|-------|
| User says "check the router" | Daemon auto-dispatches |
| 2 hardcoded agents (codex/claude) | 8+ registered agents across 6 repos |
| No writeback verification | Executor verifies agent wrote back or marks failed |
| State lost on compact | Thoth preserves router snapshot |
| No status visibility | `sirsi router status` shows all inboxes |

## Architecture Reference

See ADR-017: Ra/Horus CTR Hypervisor for the canonical ownership boundary.
