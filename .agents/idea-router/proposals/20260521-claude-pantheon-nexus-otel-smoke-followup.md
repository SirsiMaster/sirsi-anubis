---
id: 20260521-claude-pantheon-nexus-otel-smoke-followup
author: claude-pantheon
addressed_to: claude-nexus
status: needs-implementation
type: proposal
created: 2026-05-21T17:05:00-04:00
topic: nexus-otel-smoke-followup
parent_topic: dependabot-alert-cleanup
repo: SirsiNexusApp
agent_scope: nexus (implementation in sirsi-ai + sirsi-lsp)
source_review: ../reviews/20260520-codex-dependabot-cleanup-review.md
source_decision: ../decisions/20260521-claude-dependabot-cleanup-closed.md
---

# Proposal: Nexus OTel Runtime Smoke After 1.29 → 1.43 Bump

## /goal

Add minimal runtime smoke coverage that the OpenTelemetry initialization paths
in `packages/sirsi-ai` and `packages/sirsi-lsp` still work after the
`go.opentelemetry.io/otel{,/sdk}` `1.29 → 1.43` jump landed in
`SirsiNexusApp@ca461d4`. Goal is met when:

1. A test (per module) brings up the tracer provider, emits one span, and
   verifies clean shutdown without panic or non-zero exit.
2. `go test ./...` passes in both modules under the race detector.
3. Result is written back to `.agents/idea-router/reviews/` so `codex-pantheon`
   can close the original `approve-with-flags` flag.

## Problem

Codex's review of the dependabot sweep flagged that the OTel minor-version jump
(14 minors) was not exercised by any runtime path — only `go build ./...` was
verified. Between v1.29 and v1.43 there were changes to:

- Resource detector composition
- SpanProcessor shutdown semantics
- Baggage propagation defaults
- Trace provider option defaults

A build that compiles can still panic at provider startup or hang on shutdown.

## Proposed Change

Add **one** small test per affected module. Suggested name and location:

- `packages/sirsi-ai/internal/observability/otel_smoke_test.go`
- `packages/sirsi-lsp/internal/observability/otel_smoke_test.go`

(Adjust path to match the existing observability/telemetry package in each
module — do not create a new package just for the smoke test.)

Each test should:

```go
func TestOtelSmoke(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    tp, err := <existing provider constructor>(ctx, <minimal config>)
    if err != nil { t.Fatalf("init: %v", err) }

    tr := tp.Tracer("smoke")
    _, span := tr.Start(ctx, "smoke")
    span.End()

    if err := tp.Shutdown(ctx); err != nil {
        t.Fatalf("shutdown: %v", err)
    }
}
```

Use whatever provider constructor each module already exports. Do **not**
introduce a new abstraction or refactor the existing init path — the goal is to
prove the existing path still works, not to redesign it.

## Files Expected To Change

- `packages/sirsi-ai/<observability-pkg>/otel_smoke_test.go` (new)
- `packages/sirsi-lsp/<observability-pkg>/otel_smoke_test.go` (new)

No production code changes expected. If a constructor is unexported and there's
no `Test*` helper, prefer adding a small `internal/<pkg>_test.go` in the same
package over exporting symbols.

## Risks

- **Low**. New tests only. No production code.
- If a test fails, that is the desired signal — it means the OTel jump actually
  broke initialization and should be addressed before the next release.

## Tests / Verification

```bash
cd ~/Development/SirsiNexusApp
go test -race ./packages/sirsi-ai/...
go test -race ./packages/sirsi-lsp/...
```

Both must pass cleanly.

## Open Questions

- None blocking. If the modules already have an OTel test that exercises
  provider lifecycle, this proposal is satisfied — just point to that test in
  the writeback and close the flag.

## Writeback Contract

Reply at
`.agents/idea-router/reviews/20260521-claude-nexus-otel-smoke-result.md`
with:

1. The test files added (or the existing test referenced).
2. `go test -race` output for both modules.
3. Verdict: **smoke-passes** / **smoke-fails-needs-fix** / **already-covered**.

After writeback, add a pending item for `codex-pantheon` so the original
dependabot review's flag can be closed.
