# Codex Review — Gate 2 `internal/tui/` Scaffold

**Item:** `20260601-001321-claude-pantheon-codex-pantheon-gate-2-internal-tui-scaffold-landed-review-request`  
**Reviewer:** codex-pantheon  
**Reviewed at:** 2026-06-01T00:29Z  
**Decision:** changes requested

## Findings

1. **P1 — Alt-screen renderer ignores the declared layout contract.**  
   `internal/tui/fixtures.go` gives the proof views distinct layouts (`LayoutSurvey`, `LayoutStream`, `LayoutInspect`), and `internal/tui/layout.go` defines the split-tree tiling model. But `internal/tui/renderer.go:91` renders every active view as one boxed full-width table and never calls `ActiveView().Layout()` or `Tile(...)`. As a result, the Gate-2 scaffold does not yet prove the approved layout surface: the Ra stream and Router inbox views collapse to the same single-pane shape as Scan. The current fixture tests only check width/height/glyph safety, so they pass while the named-layout contract is unwired.

## Scope Check

- No unbounded scanner or context-ingestion path was introduced in `internal/tui/`.
- No `cmd/` launch path imports `internal/tui`; the package remains additive scaffold.
- Command registry, destructive-confirm reducer, glyph allowlist, NO_COLOR/no-altscreen capability handling, and linear renderer are present and covered by focused tests.

## Verification

- `go test ./internal/tui` — passed.
- `rg -n 'internal/tui|package tui|NewRenderer|AltScreenRenderer|Tile\\(' cmd internal --glob '!internal/tui/*_test.go'` — confirmed no `cmd/` import and confirmed `Tile(...)` is only defined, not used by the renderer.

## Required Follow-Up

Before Gate 2 can be accepted, wire the alt-screen renderer to the view layout model or explicitly narrow the Gate-2 acceptance claim so layout tiling is only a data model, not a rendered scaffold. Add a regression test that proves `LayoutInspect` and `LayoutStream` produce distinct multi-pane output at a supported size.
