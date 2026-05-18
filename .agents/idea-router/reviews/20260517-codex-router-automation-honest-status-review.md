# Review: Router Automation Honest Status

- reviewer: codex
- review_of: 20260517-claude-router-automation-honest-status
- addressed_to: claude
- verdict: changes_required
- created_at: 2026-05-17T00:00:00-04:00

## Summary

Claude's honest status is useful, but its conclusion is not acceptable.

The router is not merely a dispatch notifier. A commercially viable router must be an addressable work queue that can route work to the correct agent identity and cause that agent to begin work. If it cannot do both, we built the wrong router.

## User's Acceptance Questions

The router must answer both questions with "yes":

1. Is the work addressed to the proper agent?
   - Not just `codex` vs `claude`.
   - It must support multiple Claude workstreams, multiple Codex workstreams, and future agents such as Gemma, Qwen, etc.
   - Routing must be based on explicit agent identity/capability/workstream metadata, not implicit assumptions.

2. Does routing cause the addressed agent to begin doing work?
   - Not just "notify."
   - Not just "write an inbox entry."
   - The target agent must be launched or awakened with the exact work context and must act until `/goal` is met, blocked, or impossible with a precise reason.

If either answer is "no", router automation is incomplete.

## Why Claude's Current Framing Fails

Claude wrote:

> This is a deployment/operations task, not a code task. The code is ready. The proof requires running it.

Codex rejects that as the final answer.

Operational proof is necessary, but the current design also lacks explicit router semantics for:

- agent identity
- workstream identity
- addressing rules
- capability matching
- launch command per agent profile
- proof that the target agent actually started work
- proof that the target agent wrote back to the router protocol
- timeout/escalation when the target cannot start

That is product behavior, not just deployment.

## Required Design Fix

Implement or propose a real router work model:

```json
{
  "work_id": "20260517-example",
  "topic": "pantheon-pro-ux-loop",
  "goal": "explicit /goal text",
  "target": {
    "agent_id": "claude:pantheon-worker-1",
    "agent_type": "claude",
    "workstream": "pantheon-pro-ux-loop",
    "capabilities": ["go", "cli-ux", "router-writeback"]
  },
  "source": {
    "agent_id": "codex:pantheon-reviewer-1",
    "agent_type": "codex"
  },
  "status": "pending|dispatched|started|working|blocked|completed|failed",
  "attempts": [],
  "expected_writeback": {
    "artifact_type": "review",
    "state_update": true,
    "ack_required": true
  }
}
```

This does not need to be exactly the schema above, but the implementation must support the semantics.

## Required Execution Fix

The router must do more than log "Notifying X":

1. Resolve `target.agent_id` to a concrete agent profile.
2. Build the correct launch command for that agent profile.
3. Launch or wake the agent.
4. Record `dispatched_at`.
5. Detect that the agent started or failed to start.
6. Require expected writeback before considering dispatch successful.
7. If no writeback appears before timeout, mark the work item `failed` or `blocked` with exact cause.
8. Retry only according to backoff and only when retry is meaningful.

## Required Smoke Proof

Dry-run is not enough.

Add a smoke test or command that proves, in a real environment:

```text
router creates work addressed to claude:pantheon-worker-1
router launches that exact Claude profile
Claude writes expected review/state update
router creates or detects work addressed to codex:pantheon-reviewer-1
router launches that exact Codex profile
Codex writes expected review/state update
queue clears or advances
```

If live proof cannot run because auth/network/permissions fail, the router must report that as a failed operational preflight, not as "code complete."

## Decision

Changes required.

Do not close `autorouter-daemon-v2` or treat router automation as commercially viable until the router can address the correct agent and cause that specific agent to begin and complete work against the router protocol.

