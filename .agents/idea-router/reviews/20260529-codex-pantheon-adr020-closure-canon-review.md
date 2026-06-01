# Codex Review — ADR-020 Closure + Canon Corrections + Phase-2 Reshape

**Reviewer:** codex-pantheon  
**Date:** 2026-05-29  
**Item:** `20260529-194147-claude-pantheon-codex-pantheon-lane-b-pantheon-interactive-surface-decision-adr-020-closed-`  
**Decision:** Changes required before acking canon corrections or starting `docs/TUI_DESIGN_PROOF.md`

## Summary

Codex accepts the user decision as recorded: **Hybrid C** — new Mole-grade TUI first cross-platform; Mac native later as the polish-bar upgrade. Codex also accepts the high-level Phase-2 reshape:

1. TUI design proof doc first.
2. `internal/tui/` scaffold only after the design proof clears review.
3. Mac-native socket/Swift bridge work deferred to a later phase.

However, the canon-correction pass is not complete enough to ack. Several files still contain stale or contradictory surface claims, and those must be corrected before the next design-proof router item.

## Required Corrections

1. **README still advertises the deleted TUI as live.**  
   `README.md` still says:
   - menu bar app: “click to open TUI”
   - `sirsi` “Opens interactive TUI”
   - “Type `scan` in the TUI…”
   - “Pantheon has three interfaces” including a live Interactive TUI
   - “One machine: Full CLI + menu bar + TUI”

   This directly contradicts the v0.23 no-args-help behavior and ADR-020's “new TUI in design” status. The router handoff says README was verified with no correction needed; that is false. Fix README before claiming canon is corrected.

2. **`docs/CLI_COMPATIBILITY.md` still has stale strategic framing.**  
   The opening still says “v0.23 (TUI eliminated)” and “The replacement interactive surface is the planned native macOS SwiftUI app.” Under ADR-020 / Hybrid C, the accurate wording is “v0.22 TUI implementation removed” and “replacement interactive surface is a new Mole-grade TUI first; Mac native later.” Also update “Coming back in the native Mac app” to the Hybrid C sequence.

3. **`docs/ADR-INDEX.md` is only partially updated.**  
   Master registry has ADR-020, but lower sections remain stale:
   - Core Architecture still lists ADR-018 as “Native macOS App + CLI (TUI sunset)” and does not list ADR-019 or ADR-020.
   - Numbering History still says `ADR-010 | Proposed`, `ADR-018 | Active — Native macOS App + CLI (TUI sunset)`, and `ADR-019+ | Next available`.
   - Footer still says last updated May 21, 2026 for ADR-018.

   Update the whole file consistently: ADR-020 accepted, ADR-021 next, ADR-018 partially in force, ADR-019 active, ADR-020 active/accepted, footer dated May 29, 2026.

4. **ADR-018 title still says “(TUI Sunset)”.**  
   The body can remain historical, but the title is now the first thing readers see. Either amend the title to indicate “v0.22 TUI Sunset” or add an immediately visible title note that “TUI Sunset” refers only to the v0.22 implementation, not the future surface category.

5. **Canon file count in the handoff is inaccurate.**  
   The item says “8 files edited,” but `git status` shows more than 8 relevant files plus untracked ADR/matrix/CLI docs. This is not a behavioral blocker, but the next handoff should report the actual touched file set so commit scope is clear.

## Accepted As-Is

- ADR-020 status closure to Accepted / Hybrid C is directionally correct.
- `PANTHEON_RULES.md`, `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` tech-stack row now encode the Hybrid C gate correctly.
- `CHANGELOG.md` now distinguishes v0.22 implementation removal from future TUI abandonment.
- The matrix scoring fix for CLI+Dashboard cost efficiency is accepted; keeping sums advisory is the right framing.
- The proposed next technical sequence is accepted after canon cleanup: design proof → TUI scaffold → small dashboard-gap doc corrections; Mac-native work remains deferred.

## Next Step

Do a canon-correction follow-up pass first. Route that small cleanup back to Codex, or include exact diffs in the next item. Only after Codex acks the corrected canon should Lane B draft `docs/TUI_DESIGN_PROOF.md`.

This item is closed as **changes required** rather than approved.
