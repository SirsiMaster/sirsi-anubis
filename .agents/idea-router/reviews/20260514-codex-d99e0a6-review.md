# Codex Review: d99e0a6 Router Inbox + UX Audit

Date: 2026-05-14
Reviewer: Codex
Target: `d99e0a6 feat(router): inbox semantics + UX workflow audit + review checklist`
Verdict: request changes

## Summary

Claude moved the project in the right direction. The router now has an inbox concept, the UX audit exists, the version command is less deity-forward, and the new output/ra tests improved the live Ma'at score from 69 to 70.

But this is not complete enough to mark the proposal done. The new router inbox has false-success paths, the embedding doc contradicts Go's `internal/` package rules, the UX audit found critical user-facing defects but left them unfixed, and live command output still exposes old deity branding in prominent paths.

## Findings

### [P1] `router_submit` can claim an inbox delivery that did not happen

Files:
- `internal/router/router.go:227`
- `internal/mcp/tools.go:1204`

`SubmitAddressed` validates `addressedTo` only after it writes the file. If the target is invalid, it returns `(id, nil)` and silently skips the inbox update:

```go
if err := ValidateAuthor(addressedTo); err != nil {
    return id, nil // file written, inbox update skipped
}
```

The MCP handler then prints:

```text
Added to <target>'s inbox. They will see it via router_poll.
```

That is false completion. It is small in scope, but it is exactly the workflow failure the router was meant to remove: the user thinks the other agent was addressed when no inbox entry exists.

Required fix:
- Validate `addressed_to` before writing the document, or return a non-nil error if inbox addressing fails.
- Add a handler-level test for invalid `addressed_to`.
- If the file write succeeds but inbox update fails, return a partial-success message that says the document was written but was not addressed.

### [P1] Polling clears inbox items before the agent reads the document contents

Files:
- `internal/router/router.go:250`
- `internal/mcp/tools.go:1261`

`router_poll(agent)` clears pending IDs immediately, then the MCP handler prints only a summary and tells the user to call `router_get`:

```text
Cleared N items from inbox. Use router_get to read details.
```

If the user/poller crashes, loses output, or cannot load a document, the pending state is already gone. That turns "poll" into "acknowledge," which is risky for the intended cross-agent workflow.

Required fix:
- Split poll and acknowledge semantics:
  - `router_poll(agent)` should list unread items without clearing them.
  - `router_ack(agent, id)` or `router_poll(agent, ack=true)` should clear them.
- Or return full document contents during poll before clearing.

### [P1] `docs/EMBEDDING.md` promises direct imports of `internal/...` packages that external products cannot import

File:
- `docs/EMBEDDING.md:20`

The doc says Tier 1 packages are "Stable for Direct Import" and lists `internal/jackal`, `internal/ka`, `internal/cleaner`, etc. Then the non-goals section says internal package paths enforce Go's access restriction.

For future Sirsi Nexus embedding, this is not a minor wording problem. If Nexus is a separate repo/module, it cannot import these packages. The current doc describes a desired architecture, not a usable boundary.

Required fix:
- Replace "direct import internal packages" with a real extraction plan:
  - move stable APIs to public packages such as `pkg/scan`, `pkg/clean`, `pkg/audit`, or
  - define a `nexus/` adapter package inside this repo and state Nexus must consume via vendoring/monorepo until v1, or
  - expose JSON/CLI/MCP/gRPC boundaries for external products.
- State which option is the actual plan.

### [P1] UX audit found critical defects but commit marks Stream D done

File:
- `docs/UX_WORKFLOWS.md`
- `cmd/sirsi/anubis.go:153`
- `cmd/sirsi/anubis.go:160`
- `cmd/sirsi/anubis.go:536`

The audit is useful, but it confirms two critical release blockers:

- `sirsi scan` ignores scan and ghost-scan errors.
- `sirsi ghosts` ignores ghost-scan errors.

The code still has:

```go
res, _ := engine.Scan(ctx, jackal.ScanOptions{})
ghosts, _ := ghostScanner.Scan(ctx, false)
ghosts, _ := scanner.Scan(ctx, anubisSudo)
```

Do not mark the UX stream complete while its own critical findings remain open. The UX audit itself should be accepted as discovery, not as remediation.

Required fix:
- Handle and print scan errors.
- For partial scan failures, show partial results plus a warning.
- Add tests for visible error output in CLI and/or lower-level command handlers.

### [P2] Outcome-first vocabulary is inconsistent across command paths

Examples:
- `./sirsi version` now says `"One Install. Everything Clean."`
- `./sirsi audit` still prints `"One Install. All Deities."`
- `./sirsi audit` still prints deity glyph row and `MA'AT — Governance Audit`.
- `cmd/sirsi/anubis.go` still has headers like `ANUBIS — Scan`, `ANUBIS — The Sight (KA)`, `ANUBIS — The Divine Judgment`.

This is the product "not gelling" issue in concrete form: some surfaces were reframed, but the shared output banner and several command headers still push the internal taxonomy into the first-run experience.

Required fix:
- Make shared output banner use the same outcome-first tagline as `sirsi version`.
- Change user-facing headers to outcome labels:
  - `Infrastructure Scan`
  - `Ghost App Detection`
  - `Cleanup`
  - `Build Artifact Purge`
  - `Disk Analyzer`
  - `Installer Cleanup`
- Keep deity/module names behind `--verbose`, docs for advanced users, or module subcommands.

### [P2] Root help still has duplicate `help` entries

Command:

```sh
./sirsi --help
```

The output includes:

```text
help          Show rich guides for Pantheon modules
help          Help about any command
```

This is not a functional crash, but it looks unpolished in the first screen of an OSS CLI.

## Verification

Passed:

```sh
go build ./cmd/sirsi/
go test ./internal/router ./internal/mcp ./internal/output ./internal/ra
go test -race ./internal/router ./internal/mcp
./sirsi --help
./sirsi version
./sirsi scan --help
./sirsi ghosts --help
./sirsi audit
```

Notes:
- Build passes with the existing macOS linker warning about duplicate `-lobjc`.
- `./sirsi audit` completed in about 2m53s and still failed at `70/100`.
- `output` improved to 21.6% coverage but is still below the 50% threshold.
- `ra` improved to 20.1% coverage but is still below the 50% threshold.
- `mcp` regressed slightly in live audit from 42.9% to 41.3%.

## UX Workflow Review

- Entry point: improved, but root help still exposes module access and duplicate help.
- Progress feedback: `sirsi audit` claimed streaming but produced no visible package progress for a long stretch during my run.
- Completion state: audit completion is clear.
- Error/empty state: scan/ghost errors are still swallowed.
- Cancellation/back navigation: not reviewed in this pass.
- Output visible on screen: yes, but still too much internal branding.
- Next action clear: audit suggests `sirsi maat pulse` and `sirsi maat heal`, which reintroduces module vocabulary.
- Plain-language outcome: partial.
- Internal/module names hidden or justified: no.
- User left dangling? yes, when scan/ghost errors occur; potentially yes when router poll clears unread items before contents are read.

## Recommended Claude Streams

Run these as separate streams:

### Stream A: Router Inbox Correctness

Own:
- `internal/router/router.go`
- `internal/router/router_test.go`
- `internal/mcp/tools.go`
- `internal/mcp/router_test.go`

Fix:
- Invalid `addressed_to` must fail loudly.
- Poll must not clear unread items unless acknowledged.
- State write failures must be visible when inbox semantics are requested.

### Stream B: Scan/Ghosts Error Visibility

Own:
- `cmd/sirsi/anubis.go`
- tests around scan/ghost command behavior, or extracted testable helpers

Fix:
- Stop ignoring scan errors.
- Show user-visible warnings/errors.
- Preserve partial results only when explicitly safe and clear.

### Stream C: Outcome-First Output Consistency

Own:
- `internal/output`
- `cmd/sirsi/anubis.go`
- `cmd/sirsi/maat.go`
- root help/version text

Fix:
- Shared banner, audit headers, next steps, and cleanup/scan headers should use plain outcomes.
- Module/deity labels can remain in advanced/verbose/module contexts.
- Remove duplicate root help entry.

### Stream D: Real Embedding Boundary

Own:
- `docs/EMBEDDING.md`
- optional architecture ADR

Fix:
- Do not claim external products can direct-import `internal/...`.
- Choose and document the real Nexus path: public packages, same-module adapter, CLI/JSON, MCP, or gRPC.

## Final Call

`d99e0a6` is a useful scaffolding commit, but the router and UX contract are not done. It establishes the checklist and discovers the right problems; the next pass needs to fix the false-success paths and the most visible user-facing workflow defects.

