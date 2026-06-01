// Package tui is the Phase-2 batch-2 Gate-2 scaffold for the Pantheon
// operator console, "Horus" 𓂀 (ADR-015 deity hierarchy, ADR-020 Hybrid C).
//
// Scope (gate-bounded). This package is a SCAFFOLD only, approved by
// codex-pantheon's "TUI Design Proof Gate 1" review and the user's Gate-2
// sign-off (2026-05-31). It contains:
//
//   - the five layout primitives (Frame, Pane, Table, Palette, Toast)
//   - the binary split-tree layout model and the three named layouts
//   - the Command registry with data-driven status-bar hints
//   - the app state model and the View/reducer contract
//   - the renderer contract, with a first-class linear (no-altscreen) renderer
//   - the three proof screens (Scan, Ra deploy, Router inbox) as fixtures
//
// What this package is NOT:
//
//   - It is NOT a functional resurrection of the deleted v0.22 BubbleTea TUI
//     (ADR-018). The deleted code is not the foundation (ADR-020 §"What Stays
//     True"); this is a new design grown from docs/TUI_DESIGN_PROOF.md.
//   - It wires NO operator-facing launch path. `sirsi` with no arguments still
//     prints help (docs/CLI_COMPATIBILITY.md). A default-surface decision is a
//     separate gate and is explicitly out of scope here.
//
// Governance. Glyph policy follows Rules G1–G3 (see glyph.go) so that no
// terminal-font tofu can break the cell grid — the failure mode that helped
// sink v0.22. Deity attribution is table-driven from the Deity Registry
// (Rule A25). Destructive verbs are never wired to a single keystroke; the
// scaffold models confirmation as state, not as an immediate command.
package tui
