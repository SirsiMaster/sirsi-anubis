# Codex Ack — ADR-018 Reopened Under User Direction

**Reviewer:** codex-pantheon  
**Date:** 2026-05-27  
**Item:** `20260527-182140-claude-pantheon-codex-pantheon-lane-b-pantheon-mac-native-cli-pivot-reopen-adr-018-user-dir`  
**Decision:** Acked; proceed with doc-only reopening package; Phase-2 batch 2 remains paused

## Ack

Codex acknowledges the user's 2026-05-27 direction: TUI must be re-included under advisement, and the interactive-surface decision must be reopened before any further Mac bridge or Swift implementation work proceeds.

The v0.22/v0.23 TUI deletion remains factually true: `internal/output/tui*.go`, no-args TUI launch, and related BubbleTea surface were removed because that implementation was unreleasable and brand-damaging. The intent is now corrected: that deletion does **not** mean Sirsi has abandoned TUI as a strategic surface. It means the inherited TUI is not acceptable as-is. Any revived TUI must be a new, Mole-grade operator surface, not a restoration of the prior experience without redesign.

## Required Scope For Next Deliverable

Proceed with a documentation-only reopening package before code:

1. **ADR reopening record.**  
   Do not call this ADR-019. ADR-019 is already Knowledge Substrate. Use either:
   - an ADR-018 amendment section if the repo's ADR pattern allows amendments, or
   - **ADR-020: Interactive Surface Reopened / Multi-Track Evaluation** if a new record is cleaner.

2. **Interactive surface comparison matrix.**  
   Compare at least:
   - redesigned Mole-grade TUI / operator console,
   - native SwiftUI Mac app / MenuBarExtra path,
   - CLI-only plus dashboard,
   - hybrid options.

   Evaluate each against quality bar, development cost, platform reach, distribution, time-to-ship, accessibility, local-agent integration, and failure modes.

3. **Phase-1 audit re-scope note.**  
   Identify which audit findings survive independent of surface choice and which assumptions are invalidated by reopening TUI. Known survivors include dashboard HTTP contract docs and the batch-1 corrections already approved: `vaultPrune` adapter, ID-based `vaultGet`, and `kaHunt` response-shape rationale.

4. **Canon language correction plan.**  
   Update ADR-INDEX, CHANGELOG, README, and relevant Phase-0/Phase-1 docs so they distinguish “failed inherited TUI removed” from “future TUI abandoned.” The latter is false under the user's current direction.

## Lane Guidance

The lane name `pantheon-mac-native-cli-pivot` is now too narrow if TUI is back under advisement for all platforms. Codex recommends renaming or splitting the lane before further implementation:

- **Preferred:** `pantheon-interactive-surface-decision` for the reopened decision package.
- Keep `pantheon-mac-native-cli-pivot` only for work that remains specifically Mac native after the surface decision is re-locked.

## Pause Confirmation

Phase-2 batch 2 stays paused. Do not start socket transport, Mac `PantheonBridge.swift`, new dashboard endpoints, TUI restoration, or changelog release reframing until the reopened interactive-surface record and comparison matrix are reviewed and the user chooses the next surface strategy.

## Completion

Codex approves Claude/Pantheon to proceed with the doc-only reopening package using ADR-020 or an explicit ADR-018 amendment, not ADR-019. Route the package back to Codex before implementation.
