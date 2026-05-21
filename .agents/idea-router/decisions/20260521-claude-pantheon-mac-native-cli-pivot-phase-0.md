---
id: 20260521-claude-pantheon-mac-native-cli-pivot-phase-0
deciders: codex-pantheon, claude-pantheon
status: ready-for-user
created: 2026-05-21T17:50:00-04:00
topic: pantheon-mac-native-cli-pivot
repo: sirsi-pantheon
references:
  - .agents/idea-router/proposals/20260521-claude-pantheon-mac-native-cli-pivot.md
  - .agents/idea-router/reviews/20260521-codex-pantheon-mac-native-cli-pivot-review.md
  - docs/ADR-018-NATIVE-MAC-APP.md
  - docs/sprints/SPRINT-NATIVE-MAC-APP-PHASE-0.md
---

# Decision: Pantheon Mac-Native Pivot — Phase 0 ready for user

## Final Recommendation

Adopt **ADR-018 — Native macOS App + CLI as Pantheon's Interactive Surfaces**, which:

1. Pivots Pantheon's interactive surface to a **native macOS SwiftUI app** (`cmd/sirsi-app/`, standalone) with `cmd/sirsi-menubar/` as a companion status surface.
2. Keeps the CLI (`sirsi <verb>`) as the cross-platform automation backbone.
3. **Supersedes ADR-016 (TUI as Primary Interface).** Freezes the broken TUI now (pre-v0.23), hides it from default entry points, and removes it from `internal/output/tui*.go` only after Mac app v1.0 user-UAT parity.
4. Carries every Codex-conditioned constraint verbatim: native SwiftUI not Catalyst, standalone + menubar companion (not menubar-extended), reuse audit before code, first vertical slice = Status, no Mole asset copying, honest Windows/Linux CLI labeling via a new `docs/CLI_COMPATIBILITY.md`.

## Why This Is The Best Path

- **Matches the user's stated quality bar (Mole).** Mole is a 5 MB native macOS SwiftUI/AppKit binary; no terminal TUI can reach that ceiling. Adopting the same medium removes the wrong-medium fight.
- **Codex review converged with the proposal** (`approve-with-conditions`) on every architectural axis: Catalyst rejected, standalone primary + companion menubar, TUI hidden now then removed at v1.0, Sparkle + Developer ID distribution, honest CLI compatibility matrix.
- **Reuses existing assets.** `cmd/sirsi-menubar/` (Session 18), `ios/Pantheon/` SwiftUI views, `mobile/*.go` gomobile bindings, `PantheonCore.xcframework` v0.17.0 — the path is connection, not greenfield.
- **Honors prior user feedback memory.** `feedback_menubar_broken.md` and `feedback_mole_quality.md` both warned of this exact failure mode before v0.21.0's "complete" claim. ADR-018 codifies the corrective.
- **Phase-0 is no-code-by-design.** Only foundation artifacts (ADR, sprint plan, ADR-INDEX update, CLI compatibility matrix, TUI deprecation header, default-route hide). Real code begins in Phase 3 (Status vertical slice) after user UAT gates Phase 1 and Phase 2.

## User Authorization Needed

User approval is required on **three** items before any further Phase-0 task or any Phase-1 work begins:

1. **ADR-018 as written** (`docs/ADR-018-NATIVE-MAC-APP.md`). Object to any specific clause and Claude will revise.
2. **Phase-0 sprint plan** (`docs/sprints/SPRINT-NATIVE-MAC-APP-PHASE-0.md`), in particular:
   - Task 6 — change default `sirsi` (no-args) behavior to print help instead of launching the broken TUI; preserve TUI behind `--experimental-tui`. This is the only code change in Phase 0.
   - The TUI cadence: freeze + hide now, delete only at Mac app v1.0.
3. **ADR-016 supersession** — accept that "TUI as Primary Interface" becomes historical context, not active doctrine.

If all three are approved, Claude proceeds with Phase-0 tasks 4–8 (`docs/CLI_COMPATIBILITY.md`, TUI deprecation header, default-route hide, router state update), then opens Phase 1 (reuse audit) with a new router proposal for Codex review.

## Implementation Checklist (Phase 0 only, post-approval)

- [x] ADR-018 authored (this work item)
- [x] Phase-0 sprint plan authored (this work item)
- [x] ADR-INDEX updated (ADR-018 added; ADR-016 marked superseded)
- [ ] `docs/CLI_COMPATIBILITY.md` authored from current verb registry
- [ ] Deprecation notice header added to `internal/output/tui.go`
- [ ] Default-route hide: `sirsi` no-args prints help; TUI preserved behind `--experimental-tui`
- [ ] Router state updated (close `claude-pantheon` pending item; record this decision)
- [ ] Phase-1 router proposal opened to `codex-pantheon` (reuse audit scope)

## Goal status

The proposal's `/goal` has three completion conditions. Status:

1. ✅ Codex wrote a review (`approve-with-conditions`).
2. ✅ Review explicitly addressed all four required topics (platform split, reuse strategy, TUI sunset, open architecture questions).
3. ⏸ ADR drafted and Phase-0 sprint plan drafted; **user approval pending**. No app scaffold code written or planned in Phase 0 — first scaffold begins in Phase 3 after Phase-1 and Phase-2 gates.

Goal is **not yet met**; the relay continues to **user** (not Codex) for ADR approval, per the proposal's /goal step 3.
