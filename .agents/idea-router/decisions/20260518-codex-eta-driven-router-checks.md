# Decision: ETA-Driven Router Checks

- author: codex
- status: active
- topic: router-v3-multi-agent-queue
- created_at: 2026-05-18T13:40:00-04:00

## Decision

Replace fixed high-frequency Codex router polling with ETA-driven check-backs.

The 60-second heartbeat was deleted because sixty checks per hour wastes tokens when no agent has indicated that work is ready.

## Rule

Every routed task must include at least one of:

- `eta_for_review`: absolute expected completion/review time.
- `next_check_at`: absolute time the responder should check again.
- `estimated_duration`: approximate duration if an absolute timestamp is not practical.

Use ISO-8601 timestamps when possible.

## Universal Responder Behavior

Codex remains the interim universal responder, but it should check the router:

1. At or near `next_check_at` / `eta_for_review`.
2. When the user explicitly says `ctr`.
3. When an agent writes an urgent/blocking item.
4. Opportunistically during active user work.

Codex should not run fixed high-frequency polling unless the user explicitly re-enables it for a short burst.

## Handoff Quality

Router items missing ETA/check-back fields are incomplete unless they are urgent, already complete, or explicitly marked as no-check-needed with a reason.
