# Codex Review — ADR-020 Canon-Correction Follow-up

**Reviewer:** codex-pantheon  
**Date:** 2026-05-31  
**Item:** `20260531-191051-claude-pantheon-codex-pantheon-lane-b-pantheon-interactive-surface-decision-canon-correctio`  
**Decision:** Changes required before full ack / before `docs/TUI_DESIGN_PROOF.md`

## Summary

The follow-up pass fixes the four major misses from the 2026-05-29 review:

- `README.md` no longer advertises the deleted v0.22 TUI as live and now frames the new TUI as ADR-020 / Hybrid C design work.
- `docs/CLI_COMPATIBILITY.md` now opens with the correct v0.22 BubbleTea implementation-removal framing and cites ADR-020 / Hybrid C.
- `docs/ADR-INDEX.md` now includes ADR-019 and ADR-020, updates ADR-018 to partially-in-force / amended-by-ADR-020, advances next available to ADR-021, and updates the footer to May 29.
- `docs/ADR-018-NATIVE-MAC-APP.md` now scopes “TUI Sunset” to the v0.22 implementation and points readers to ADR-020 as current direction.

The reported commit-scope accounting is also directionally acceptable: this follow-up touched 4 files, and the broader canon-correction scope is 11 files when prior-pass overlap is counted as described.

## Remaining Corrections

1. **`docs/ADR-INDEX.md` still has the ADR-010 status mismatch.**  
   The Master Registry and `docs/ADR-010-MENUBAR-APPLICATION.md` say ADR-010 is Accepted, but Numbering History still says:

   `ADR-010 | Proposed — Menu Bar Application`

   This was one of the stale lower-section items called out in the prior review. Change it to Active/Accepted consistently before claiming the ADR index is fully corrected.

2. **`docs/CLI_COMPATIBILITY.md` still has one stale Mac-native-first note.**  
   In the `status` row, the Notes column still says:

   `Was already non-TUI; "live dashboard" rendering moves to the Mac app.`

   Under ADR-020 / Hybrid C, the live dashboard/operator flow should be described as returning in the new Mole-grade TUI first, with Mac native SwiftUI later. Please reword that note to avoid implying Mac app is the immediate replacement surface.

## Accepted

- The README corrections are accepted.
- The CLI compatibility opening/governing-decision correction is accepted, subject to the single `status` row note above.
- The ADR-018 title/scoping correction is accepted.
- The 11-file canon-correction scope statement is accepted, with the caveat that unrelated dirty `.agents`, Lane A, and Phase-0 files remain out of Lane B commit scope.
- The next Lane B router item can be `docs/TUI_DESIGN_PROOF.md` after these two small canon nits are patched and re-routed or included in the next handoff.

## Verification Performed

- `sirsi router show <item>`
- `rg 'TUI eliminat|no interactive TUI|interactive TUI was removed|Mac native (app|SwiftUI) is the (interactive )?surface|sole interactive surface|click to open TUI|sirsi.*Opens interactive TUI|Type scan in the TUI|three interfaces|Full CLI \\+ menu bar \\+ TUI' README.md docs/CLI_COMPATIBILITY.md docs/ADR-INDEX.md docs/ADR-018-NATIVE-MAC-APP.md`
- `git diff -- README.md docs/CLI_COMPATIBILITY.md docs/ADR-INDEX.md docs/ADR-018-NATIVE-MAC-APP.md`
- Direct reads of `README.md`, `docs/CLI_COMPATIBILITY.md`, `docs/ADR-INDEX.md`, `docs/ADR-018-NATIVE-MAC-APP.md`, and `docs/ADR-010-MENUBAR-APPLICATION.md`.
