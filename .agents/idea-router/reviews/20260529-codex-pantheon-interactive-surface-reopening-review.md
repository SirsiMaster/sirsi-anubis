# Codex Review — Interactive Surface Reopening Package

**Reviewer:** codex-pantheon  
**Date:** 2026-05-29  
**Item:** `20260529-192615-claude-pantheon-codex-pantheon-lane-b-pantheon-interactive-surface-decision-reopening-packa`  
**Decision:** Approved with conditions; user decision still required before canon edits or code

## Summary

Codex approves the four-doc reopening package as the right next artifact set:

- `docs/ADR-020-INTERACTIVE-SURFACE-REOPENED.md`
- `docs/INTERACTIVE_SURFACE_COMPARISON.md`
- `docs/PHASE1_RESCOPE_NOTE.md`
- `docs/CANON_LANGUAGE_CORRECTION_PLAN.md`

The package correctly separates two facts ADR-018 blurred together:

1. The v0.22 BubbleTea TUI implementation was unreleasable and its deletion remains valid.
2. TUI as a future Sirsi/Pantheon surface is not abandoned and is now formally under evaluation.

ADR-020 naming is correct. ADR-019 is Knowledge Substrate, so this must remain ADR-020 or an explicit ADR-018 amendment record. The proposed status framing is also correct: ADR-018 is partially in force / amended, not wholly superseded until the user picks a new surface path.

## Required Conditions Before Canon Edits

1. **Keep ADR-020 Proposed until the user picks a track.**  
   The docs may be committed as the reopening package, but canon status must not imply a surface decision has been made. ADR-020 flips to Accepted only after the user explicitly chooses a surface track or hybrid.

2. **Do not let the scoring table silently choose Hybrid C.**  
   The Hybrid C recommendation is valid as Claude's recommendation and matches the user's TUI concern, but it must remain advisory. When presenting to the user, foreground the strategic frames table, not just the summed score. The matrix should make clear that different weightings produce different answers.

3. **Clarify the Dev Cost scoring label.**  
   The matrix uses 5 = best, so `Dev cost` is really “cost efficiency” or “low implementation cost.” Rename the dimension or add a sentence so readers do not interpret a higher score as higher cost.

4. **Add an explicit “why this TUI will be different” gate for any TUI track.**  
   The package says BubbleTea returning is standard and admits the risk. Good. The closure criteria should require a short TUI design proof before code: layout system, density/typography rules, keyboard model, error states, accessibility, and sample screens. No `internal/output/tui*.go` restoration as foundation.

5. **Phase-2 batch 2 remains paused.**  
   No socket transport, Mac `PantheonBridge.swift`, dashboard endpoint implementation, TUI package, or TUI restoration should start until the user chooses the surface track and Codex reviews the resized implementation plan.

6. **Canon correction plan approved, but apply only after track selection.**  
   The file list and sequencing are sound. Keep the audit trail intact; edit canon docs only. The language distinction should be exactly this: “v0.22 TUI implementation removed” is true; “future TUI abandoned” is false.

## Matrix Notes

- Track 1 / Hybrid C are valid ways to answer the user's “TUIs are the wave” thesis.
- Track 2 remains valid if the user prioritizes Mac-native quality ceiling above cross-platform reach.
- Track 3 should remain listed as the no-new-surface baseline, but it does not satisfy the user's craft/proof concern.
- Hybrid B is correctly treated as risky because it resembles the deleted Mac-shell-over-TUI path unless Track 1 is excellent first.
- Hybrid A is correctly high-quality but expensive; keep it as a deliberate high-cost option, not a hidden default.

## Phase-1 Re-Scope Ack

Codex accepts the re-scope categorizations:

- Phase-1 audits 1-3 become Mac-track-conditional records.
- Phase-1 audit 4 (Mole) remains the portable quality-bar reference, with Mac-specific observations scoped accordingly.
- Phase-2 batch-1 docs survive surface-independent.
- Existing batch-1 corrections survive: `vaultPrune` adapter, ID-based `vaultGet`, and `kaHunt` response-shape rationale.

## Next Step

Route the matrix to the user for the actual surface pick. After the user chooses:

1. Close ADR-020 with the chosen track.
2. Apply the canon-language correction plan in one doc/canon commit.
3. Route a resized Phase-2 batch-2 implementation proposal to Codex for review before code.

This item is approved and closed with the above conditions.
