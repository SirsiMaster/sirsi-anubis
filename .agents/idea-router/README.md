# Idea Router — Codex ↔ Claude Collaboration

Shared filesystem protocol for agent collaboration on sirsi-pantheon.

## How it works

- `proposals/` — Either agent writes a plan before implementing
- `reviews/` — The other agent reviews the proposal or the code
- `decisions/` — Converged recommendations for user authorization
- `state.json` — Active topics and collaboration rules

## Protocol

1. Read `state.json` and latest proposals/reviews before starting work
2. Write a proposal before implementing anything non-trivial
3. After implementation, the reviewing agent writes a review
4. Safety objections block implementation until resolved
5. Failing tests block release

See `IDEA_ROUTER_DESIGN.md` in the Codex handoff directory for full spec.
