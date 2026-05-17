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

## Mandatory Workstream Protocol

Every non-trivial Codex/Claude workstream MUST use:

- `/plan` before implementation.
- `/goal` as the explicit completion flag.
- One repo-scoped agent per repository.
- A written super-agent mandate before any one agent coordinates or edits across repos.
- Router handoff files that keep the other agent queued until the `/goal` is met.

### Repo Segmentation

Work on repositories is segmented by default. A normal agent owns exactly one repository. It may inspect another repo only for read-only context and must not edit outside its assigned repo.

A super agent is allowed only when the `/plan` says:

1. Which repositories are in scope.
2. Whether the super agent may edit files or only coordinate.
3. Which repo-scoped agents own implementation.
4. What evidence is required before the `/goal` is complete.

### Goal Relay

Submissions by one agent should trigger the other:

1. Codex writes proposal/review/decision and adds a pending item for Claude.
2. Claude reads the pending item, works or reviews, then writes its own router artifact.
3. Claude adds a pending item for Codex.
4. The relay continues until the `/goal` completion condition is met.

## Full Automation

The commercial path is the autorouter daemon, not manual polling.

- Preview without launching agents:
  `sirsi router daemon --dry-run`
- Run live in the foreground:
  `SIRSI_ROUTER_NOTIFY=1 sirsi router daemon`
- Install and start the resident macOS launch agent:
  `sirsi router install-agent --load`
- Check whether it is installed and loaded:
  `sirsi router service-status`
- Stop and remove it:
  `sirsi router uninstall-agent`

The daemon watches `.agents/idea-router/state.json`, `proposals/`, `reviews/`, and `decisions/` with `fsnotify`, with a one-second fallback poll. It dispatches pending inbox items immediately, keeps a persistent `dispatch-ledger.json` so restarts do not relaunch unchanged work, and never acknowledges inbox items for an agent.

If no automation runner is active, the pending item is still the source of truth. Agents must check `state.json` and the latest router files at session start.
