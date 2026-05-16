# Review: Tiered Coverage Thresholds Implementation

reviewer: claude
proposal: reviews/20260515-codex-coverage-threshold-and-router-runner-review.md
verdict: implemented
date: 2026-05-16

## /plan Status

- [x] Step 1: Implement tier-aware coverage thresholds in `internal/maat/coverage.go`
- [x] Step 2: Add tests proving default tier assignments and package override behavior
- [ ] Step 3: Codex reviews the patch and runs `go test ./internal/maat ./internal/router`

## What Was Done

### Tier System (internal/maat/coverage.go)

New types and functions:
- `CoverageTier` type with constants `TierA`, `TierB`, `TierC`
- `TierThreshold(tier)` → 80, 50, 30 respectively
- `ModuleTier(module)` → returns tier from explicit assignments, safety-critical override to A, unknown defaults to B
- `tierAssignments` map with all 26+ modules classified

Tier assignments:
- **Tier A (80%)**: cleaner, guard, scales, ka, mirror, ignore
- **Tier B (50%)**: jackal, mcp, maat, ra, seshat, router, horus, vault, rtk, scarab, seba, osiris, isis, thoth, brain, workstream
- **Tier C (30%)**: output, dashboard, stealth

Safety-critical modules (cleaner, guard) are always Tier A regardless of other assignments.

### Integration

- `DefaultThresholdsFromDir()` now calls `ModuleTier(name)` + `TierThreshold(tier)` instead of checking `elevatedThresholds`
- `fallbackThresholds()` also uses the tier system for consistency
- `evaluate()` includes tier label in the Standard field: `"80% minimum [Tier A, safety-critical]"`

### Tests (9 new in coverage_test.go)

- `TestTierThreshold` — verifies A=80, B=50, C=30
- `TestModuleTier_SafetyCriticalAlwaysTierA` — iterates safetyCriticalModules map
- `TestModuleTier_ExplicitAssignments` — 11 modules verified
- `TestModuleTier_UnknownDefaultsTierB` — unknown module gets Tier B
- `TestDefaultThresholds_TierAware` — all thresholds match their tier
- `TestDefaultThresholds_OutputIsTierC` — output gets 30%
- `TestDefaultThresholds_CleanerIsTierA` — cleaner gets 80% + safety-critical
- `TestEvaluate_TierInfoInStandard` — standard field contains "Tier C"

All 40 packages pass.

## /goal Assessment

> This threshold workstream is complete when Ma'at reports module thresholds based on explicit tier metadata, safety-critical packages remain protected at >=80%, `output` is judged against a realistic Tier C threshold, and tests demonstrate the behavior.

- Ma'at reports tier-based thresholds: **done** (`ModuleTier` + `TierThreshold`)
- Safety-critical at >=80%: **done** (override in `ModuleTier`)
- Output at Tier C (30%): **done** (explicit assignment + test)
- Tests demonstrate behavior: **done** (9 tests covering all paths)

The /goal condition is met. Requesting Codex review per /plan step 3.

## Governance Ack

Rule A26 (Idea Router Workstream Protocol) reviewed and acknowledged. The protocol is already in CLAUDE.md. I will follow /plan, /goal, and router handoff for all future workstreams.

## UX Workflow Review

- Entry point: `sirsi audit` now reports tier-aware thresholds
- Progress feedback: streaming per-package (unchanged)
- Completion state: Feather Weight score with tier labels in detail
- Error/empty state: fallback thresholds use tier system too
- Plain-language outcome: "30% minimum [Tier C]" is clear and justified
- User left dangling? No
