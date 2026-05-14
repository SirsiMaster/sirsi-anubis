# Codex Proposal: Router Workflow + UX Review Contract

Date: 2026-05-14
Author: Codex
Status: proposal

## Problem

The idea-router is useful as a shared mailbox, but it does not yet create a smooth cross-agent workflow. In practice, the user still has to say "ball in your court" to move work between Claude and Codex.

That is acceptable for now, but it should be made explicit and improved. The user wants to approve decisions, not ferry messages between agents.

Separately, every Claude update should be reviewed not only for code hygiene and tests, but for actual product usability:

- Does the workflow start clearly?
- Does it show visible progress?
- Does it finish with a clear result?
- Does it tell the user what happened and what to do next?
- Does it leave the user dangling?
- Does it expose internal mythology when the user needs plain outcomes?
- Does it recover cleanly from errors, cancellation, empty results, denied permissions, and narrow terminals?

## Suggested Direction

Treat the idea-router as an inbox protocol first, not a process-spawning system.

The safest durable model is:

1. Agents write proposals/reviews/decisions to `.agents/idea-router/`.
2. The router state tracks unread work per agent.
3. `router_poll` returns pending work for the current agent.
4. `router_submit` writes a response and updates state.
5. Optional notification is a convenience layer only, never required for correctness.

This keeps the system portable across Codex, Claude, local CLIs, CI, and future Sirsi Nexus embedding.

## Proposed Solution

### Stream A: Router Inbox Semantics

Own:
- `internal/router`
- `.agents/idea-router/state.json`
- router docs

Implement:
- Per-agent inbox state:
  - `pending_for_codex`
  - `pending_for_claude`
  - `read_by`
  - `last_seen`
- `router_poll(agent)` returns unread documents addressed to that agent.
- `router_submit(...)` can optionally mark a response as addressed to the other agent.
- No process spawn is required for the core loop.

Acceptance:
- Claude can submit a review addressed to Codex.
- Codex can poll and see that review without the user naming the file.
- Codex can submit a response addressed to Claude.
- Claude can poll and see the response.

### Stream B: Notification As Optional Add-On

Own:
- `internal/router/notify.go`
- `internal/mcp/tools.go`
- tests

Implement:
- Keep `router_notify` opt-in behind `SIRSI_ROUTER_NOTIFY=1`.
- Make notify write an artifact first, then optionally spawn.
- Notification failure must not corrupt inbox state.
- Tool output should say exactly whether it wrote the inbox item, whether it attempted notification, and whether notification succeeded.

Acceptance:
- With notify disabled, the router still works fully through polling.
- With notify enabled, failures are visible and non-destructive.

### Stream C: UX Review Checklist In Every Claude Handoff

Own:
- `.agents/idea-router/README.md`
- `.agents/idea-router/DESIGN.md`
- future review template

Add a mandatory UX section to every Claude implementation response:

```md
## UX Workflow Review

- Entry point:
- Progress feedback:
- Completion state:
- Error/empty state:
- Cancellation/back navigation:
- Output visible on screen:
- Next action clear:
- Plain-language outcome:
- Internal/module names hidden or justified:
- User left dangling? yes/no:
```

Acceptance:
- Every future Claude pass includes this section.
- Codex reviews it independently instead of assuming tests imply usability.

### Stream D: Pantheon Workflow Audit

Own:
- TUI flow docs
- command help text
- existing TUI tests if present

Audit these workflows end to end:
- `sirsi scan`
- `sirsi clean`
- `sirsi purge`
- `sirsi installer`
- `sirsi diagnose`
- `sirsi analyze`
- `sirsi audit`
- `sirsi risk`
- interactive TUI tabs/actions

For each workflow, answer:
- Can a regular user discover it?
- Does it show progress?
- Does it finish with a clear result?
- Is the result actionable?
- Are dangerous actions reversible or clearly confirmed?
- Are empty states helpful?
- Are errors visible and understandable?
- Does it work in a narrow terminal?
- Does it avoid deity/module vocabulary unless the user is in advanced mode?

Acceptance:
- Produce a compact `docs/UX_WORKFLOWS.md` with pass/fail status per workflow.
- File follow-up implementation streams only for failed workflows.

## Codex Standing Review Rule

From this point forward, when reviewing Claude updates for Pantheon, Codex should inspect both:

1. Engineering correctness: safety, tests, architecture, races, release gates.
2. Product usability: workflow start/progress/completion, screen output, errors, empty states, next actions, and whether users are left dangling.

The product can have deity-named internal modules. The user-facing experience should lead with outcomes.

## Why This Matters

Pantheon is intended to be a free Apache 2.0 infrastructure and AI IDE hygiene tool for regular users and power developers, and later embeddable inside Sirsi Nexus for enterprise use.

That means shippability is not only "does it compile?" It is:

- safe by default,
- understandable on first run,
- visible while working,
- decisive when done,
- honest when blocked,
- portable beyond the standalone CLI/TUI.

