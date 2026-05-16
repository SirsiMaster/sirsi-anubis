# Decision: Idea Router Workstream Governance

deciders: codex
status: ready-for-claude-review
created: 2026-05-15T00:00:00-07:00

## Final Recommendation

Adopt Rule A26 as the shared Codex/Claude workstream protocol across Sirsi repositories.

## /plan

1. Add the protocol to Pantheon canonical agent rules.
2. Mirror the protocol into Codex and Claude instruction files.
3. Extend Idea Router documentation so both agents know how to hand off work.
4. Update router state so Claude has an explicit pending review item.

## /goal

The workstream is complete when Claude can read the Idea Router and Pantheon rules, identify the repo-segmentation mandate, see that `/plan` and `/goal` are required for every non-trivial workstream, and pick up a pending router item to review or implement next.

## Why This Is The Best Path

This avoids relying on chat memory. The router becomes the shared filesystem contract, while Pantheon rules make repo segmentation and goal completion enforceable.

## User Authorization Needed

No repository moves or destructive cleanup are authorized by this decision. It only establishes collaboration protocol.

## Implementation Checklist

- [x] Add Rule A26 to PANTHEON_RULES.md.
- [x] Mirror Rule A26 to AGENTS.md.
- [x] Mirror Rule A26 to CLAUDE.md.
- [x] Update .agents/idea-router/README.md.
- [x] Update .agents/idea-router/DESIGN.md.
- [x] Update .agents/idea-router/state.json with pending Claude review.
