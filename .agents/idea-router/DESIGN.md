# Idea Router Design — Codex ↔ Claude Collaboration

## Goal

Create a shared, low-friction collaboration channel where Codex and Claude can independently propose, critique, revise, and converge on implementation plans before presenting recommendations to the user for authorization.

The user should not be the message bus.

## Recommended v0: Filesystem Router

Use a repo-local folder:

```text
.agents/idea-router/
  README.md
  state.json
  proposals/
  reviews/
  decisions/
  transcripts/
```

This is intentionally simple. Both agents can read and write plain Markdown/JSON. No server, auth, or protocol work is needed for v0.

## Message Types

### Proposal

Path:

```text
.agents/idea-router/proposals/YYYYMMDD-HHMM-agent-topic.md
```

Template:

```markdown
# Proposal: <topic>

author: codex|claude
status: draft|needs-review|accepted|rejected|superseded
created: <iso timestamp>

## Problem

## Proposed Change

## Files Expected To Change

## Risks

## Tests / Verification

## Open Questions
```

### Review

Path:

```text
.agents/idea-router/reviews/YYYYMMDD-HHMM-reviewer-topic.md
```

Template:

```markdown
# Review: <proposal topic>

reviewer: codex|claude
proposal: <proposal file>
verdict: approve|request-changes|reject

## Findings

## Suggested Revisions

## Residual Risk

## UX Workflow Review

For every implementation review, include this checklist:

- Entry point: [how does the user discover/start this?]
- Progress feedback: [spinner, streaming output, progress bar?]
- Completion state: [clear result printed?]
- Error/empty state: [what happens when nothing found or operation fails?]
- Cancellation/back navigation: [Ctrl+C, Esc, back work?]
- Output visible on screen: [does it print something the user can read?]
- Next action clear: [does it tell the user what to do next?]
- Plain-language outcome: [no deity/module jargon in user output?]
- Internal names hidden or justified: [module names only in advanced mode?]
- User left dangling? [yes/no — does the flow end cleanly?]
```

### Decision

Path:

```text
.agents/idea-router/decisions/YYYYMMDD-HHMM-topic.md
```

Template:

```markdown
# Decision: <topic>

deciders: codex, claude
status: ready-for-user|approved-by-user|rejected-by-user|implemented

## Final Recommendation

## Why This Is The Best Path

## User Authorization Needed

## Implementation Checklist
```

## State File

`.agents/idea-router/state.json`:

```json
{
  "version": 1,
  "active_topics": [],
  "last_codex_read": null,
  "last_claude_read": null,
  "rules": {
    "no_feature_expansion": true,
    "safety_first": true,
    "plain_user_language": true,
    "deities_are_modules": true
  }
}
```

## Agent Protocol

1. Before starting work, read `state.json`, latest proposals, latest reviews, and latest decisions.
2. If proposing work, write a proposal and mark it `needs-review`.
3. The other agent reviews the proposal.
4. The proposing agent revises or writes a decision.
5. Only decisions with `status: ready-for-user` should be presented to the user for authorization.
6. After implementation, the reviewing agent performs a code review and writes a review file.

## Conflict Rules

- If Codex and Claude disagree, preserve both positions in a decision file.
- Safety objections block implementation until resolved.
- Failing tests block release recommendations.
- User-facing terminology must prefer plain outcomes over deity/module names.
- Deity names may remain in advanced namespaces and attribution.

## What v0 Enables

- Codex can be the outside product/architecture judge.
- Claude can implement.
- Claude can critique Codex proposals.
- Codex can review Claude patches.
- The user sees converged decisions, not raw back-and-forth.

## Future v1: MCP Router

Build a local MCP server if the filesystem router proves useful.

Tools:

- `router_create_proposal`
- `router_list_open_topics`
- `router_add_review`
- `router_record_decision`
- `router_get_context`

Resources:

- `idea-router://state`
- `idea-router://proposal/{id}`
- `idea-router://decision/{id}`

The MCP server should still persist to `.agents/idea-router/` so the collaboration history remains inspectable in git.

## First Topics To Route

1. Cleanup safety reset.
2. Command vocabulary reset.
3. TUI architecture decomposition.
4. Thoth as "remember" product surface.
5. Ma'at as CI/CD and PR triage product surface.

