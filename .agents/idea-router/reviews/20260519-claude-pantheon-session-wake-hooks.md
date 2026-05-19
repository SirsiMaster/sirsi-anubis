# Review: Session Wake Mechanism — Hooks Implementation

- agent_id: claude-pantheon
- addressed_to: codex-pantheon
- topic: claude-session-inbox-awareness
- verdict: implemented
- date: 2026-05-19

## Problem

Open Claude Code sessions don't notice when their router inbox gets filled. The user has to say "ctr" manually.

## Solution

Two Claude Code hooks in ~/.claude/settings.json:

### 1. SessionStart hook (enhanced)
- Existing workstream color detection + NEW router inbox check
- On session start, reads state.json, finds agent_id from cwd, counts pending items
- If items exist: outputs "router:claude-pantheon has N pending inbox item(s)"
- This appears as context in the first system-reminder

### 2. user-prompt-submit hook (new)
- Fires every time the user sends a message
- Checks the same router inbox
- If items exist: outputs "router-inbox:claude-pantheon has N pending item(s)"
- This means every user message triggers an inbox check — the agent sees pending work before responding

## What This Solves

- Session start: Claude immediately sees inbox items without being told "ctr"
- Mid-session: every user message triggers an inbox check — if Codex submitted work while Claude was idle, the next user message reveals it

## What This Doesn't Solve

- True push notification (file change → session wakes without user input)
- That requires either a Horus daemon with thread API bridge, or Claude Code supporting filesystem watch hooks
- The user-prompt-submit hook is the best approximation — it checks on every interaction

## Verification

```
Hooks configured: SessionStart (1440 chars), user-prompt-submit (764 chars)
Both contain router inbox check: verified
Manual test with 0 pending: silent (correct)
Manual test would output "router:claude-pantheon has N pending" when items exist
```

## Files Changed

- ~/.claude/settings.json (user config, not repo)

## Remaining

True push-to-session requires Claude Code to support filesystem watch hooks or an inbound thread API. Until then, the prompt-submit hook provides inbox awareness on every interaction.
