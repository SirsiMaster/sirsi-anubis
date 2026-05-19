# Decision: Codex Interim Universal Router Responder

- author: codex
- status: active
- topic: router-v3-multi-agent-queue
- created_at: 2026-05-18

## Decision

Until the multi-agent response fabric is implemented, Codex is the universal responder for router requests.

Claude agents and other workstream agents may place review, triage, approval, rejection, routing, and next-action requests into Codex's router queue. Codex will read and respond as those items arrive.

## Boundary

This is a coordination mandate, not blanket cross-repo implementation authority.

- Codex may review, triage, route, approve, reject, and define next work across router items.
- Codex may kick off or route independent work to registered repo-scoped agents when requested or when parallelism materially accelerates completion.
- Implementation remains repo-segmented by default.
- A single agent must not edit across repos unless a written super-agent mandate names the repos and grants that scope.
- `/plan` and `/goal` remain required for non-trivial workstreams.
- Parallel work must include explicit ownership, expected writeback artifact, ETA, and verification requirements.

## Dispatch Posture

Codex should not perform independent work serially when separate registered agents can safely own disjoint slices.

Default dispatch preference:

1. Fan out independent cross-repo tasks to each repo's registered agent.
2. Split same-repo implementation only when file ownership is disjoint and merge risk is low.
3. Keep Codex as coordinator/reviewer unless a written super-agent mandate names another coordinator.
4. If live launch is unavailable, write addressed router items instead of waiting for manual shuttle.

## Transition

When Ra's multi-agent response fabric is available, this interim rule should be replaced by registered agent ids, dynamic inboxes, and explicit executor profiles in `.agents/idea-router/agents.json`.
