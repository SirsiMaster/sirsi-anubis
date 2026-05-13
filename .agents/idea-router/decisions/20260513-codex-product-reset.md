# Claude Implementation Brief — Sirsi Pantheon Product Reset

## Context

Codex reviewed the Pantheon TUI v0.21.0 session and found that the app is not merely suffering from hygiene bugs. The deeper issue is product coherence, safety trust, and usability.

The target product is:

> A free Apache 2.0 infrastructure and AI IDE hygiene tool that is fast and easy for regular users, while still powerful for developers with advanced AI and machine hygiene needs.

The deities are module names aligned to the work they do. They are not the user-facing mental model and must not become show stoppers.

## Product Direction

Stop adding features for at least one month. The goal is not breadth. The goal is a coherent, safe, trustworthy core.

Primary user jobs:

- Clean my machine.
- Show what is wasting space.
- Fix my AI/dev environment.
- Keep my workflow from degrading.
- Help my agents remember context and keep shipping.
- Triage stalled CI/CD and PR work.

The interface should expose outcomes first:

- `sirsi scan`
- `sirsi clean`
- `sirsi fix`
- `sirsi status`
- `sirsi remember`
- `sirsi ci`

Deity names should appear as provenance, not navigation requirements:

```text
Context memory stale
Handled by Thoth

3 stalled PRs found
Handled by Ma'at

RAM pressure detected
Handled by Isis
```

## Critical Review Findings To Fix First

1. Destructive cleanup bypasses safety.

   `internal/jackal/purge.go` and `internal/jackal/installer.go` delete paths directly. `moveToTrash` falls back to permanent deletion when Finder/osascript fails. This violates Rule A1. Replace with cleaner-backed APIs, protected-path validation, explicit dry-run behavior, operation logging, and no permanent fallback unless the user has explicitly confirmed that mode.

2. Full test suite is failing.

   `go test ./...` failed outside the sandbox. `cmd/sirsi` killed `ghosts` after 30 seconds. `tests/e2e` still expects old root help entries. Release work is blocked until tests pass.

3. TUI streaming can report false success.

   Native streaming actions call `fn()` and ignore returned errors. Errors must flow through completion messages so the TUI cannot mark a failed command as successful.

4. Ma'at gate is not meaningful right now.

   `sirsi audit` reported `50/100` and "no coverage data found" for every module. Either fix coverage ingestion or stop treating that output as a real quality verdict.

5. Command language is inconsistent.

   Old terms remain in user-facing docs, suggestions, tests, and help: `weigh`, `judge`, `ka`, `doctor`. Compatibility aliases can remain hidden, but normal users should see plain verbs only.

6. Broken command suggestions exist.

   `internal/suggest/suggest.go` emits commands like `sirsi sirsi thoth sync`. Fix all generated command strings with tests.

## Architecture Direction

The current TUI has too much feature-specific logic in one place. `internal/output/tui.go` is doing routing, command orchestration, global mailbox state, rendering decisions, and workflow transitions.

Refactor direction:

- Introduce a small action runner interface.
- Move each workflow into a focused controller: scan, clean, status, analyze, ci, memory.
- Remove process-global pending state for scan/analyze/select flows.
- Keep renderers dumb: input state to view string.
- Make destructive actions go through a single safety gateway.

Do not start with a large rewrite. Start with safety and tests, then peel off controllers.

## Usability Direction

The first-run experience should be a plain-language control surface, not a mythology menu.

Suggested primary tabs:

- Storage
- Performance
- AI IDE
- Projects
- CI
- Memory
- Settings

The golden path must be excellent:

```text
scan -> explain -> preview -> clean -> verify
```

Every cleanup result should answer:

- What will happen?
- What will never happen?
- How much space is affected?
- Is it recent or risky?
- Is it moved to Trash?
- Where is the operation log?
- How do I verify afterward?

## Thoth And Ma'at Product Roles

Thoth should become the understandable context promise:

> Never lose project context again.

Useful commands:

- `sirsi remember`
- `sirsi remember status`
- `sirsi remember compact`
- `sirsi remember sync`

Ma'at should become operational CI/CD and PR triage:

> Your pipeline has 4 blocked PRs, 2 flaky checks, and 1 stale branch.

Useful commands:

- `sirsi ci`
- `sirsi ci prs`
- `sirsi ci failures`
- `sirsi ci fix-plan`

Ma'at remains quality sovereignty internally, but the user-facing job is pipeline clarity.

## Acceptance Criteria For The Hardening Month

- `go build ./cmd/sirsi/` passes.
- `go vet ./...` passes.
- `go test ./...` passes consistently.
- `go test -race ./internal/output/ ./internal/jackal/...` passes.
- `sirsi scan -> sirsi clean -> sirsi scan` works without stale aliases in the visible flow.
- No cleanup path can permanently delete files by default.
- All destructive operations have dry-run semantics and protected-path checks.
- Root help and TUI navigation use plain outcome language.
- Deity names only appear as module attribution or advanced namespaces.
- Ma'at audit either reports real coverage or clearly states it cannot assess.

## Implementation Order

1. Fix destructive safety paths in purge and installer.
2. Fix failing tests and stale smoke expectations.
3. Fix TUI streaming error propagation.
4. Clean the command vocabulary and suggestions.
5. Add missing tests for analyze, purge, installer trash paths, oplog, vitals, and suggestions.
6. Reframe root help and TUI tabs around user jobs.
7. Only after that, consider deeper TUI controller refactoring.

## Collaboration Request

Claude should implement only after producing a small plan and confidence rating. Codex will independently review each change through the idea router described in `IDEA_ROUTER_DESIGN.md`.

