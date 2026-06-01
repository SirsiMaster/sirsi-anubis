# ADR-020: Interactive Surface Reopened — Multi-Track Evaluation

| Field | Value |
| :--- | :--- |
| **Status** | Accepted — Hybrid C (TUI first cross-platform; Mac native later) — closed 2026-05-29 |
| **Date** | 2026-05-29 |
| **Author** | claude-pantheon (`thr-1ca491d095768e1a`, lane: `pantheon-interactive-surface-decision`) |
| **Reviewer** | codex-pantheon |
| **Supersedes** | None (ADR-018 remains *partially* in force — see §"Relationship to ADR-018") |
| **Related** | [ADR-016](ADR-016-TUI-PRIMARY-INTERFACE.md), [ADR-018](ADR-018-NATIVE-MAC-APP.md), [ADR-010](ADR-010-MENUBAR-APPLICATION.md) |

## Context

ADR-018 (2026-05-21) ended the v0.22 BubbleTea TUI and adopted native macOS SwiftUI + CLI-on-Windows/Linux as Pantheon's interactive surface. The trigger was a user judgement that the shipped v0.22 TUI was *"utterly unreleasable… would damage Sirsi's reputation forever if released to the public."* The decision shipped: `internal/output/tui*.go` (~4,800 LOC, 20 files) was deleted, no-args TUI launch was removed, `charm.land/bubbletea/v2` dropped from `go.mod`, and Phase-0 closed clean.

On 2026-05-27 the user reopened that direction:

> *"I really do believe that TUIs are the wave and it would behoove us to build one. In fact if we can't build one, it calls into question our ability to build Sirsi overall."*

The same day, codex-pantheon independently surfaced the same correction in router item `20260527-175206-codex-pantheon-claude-pantheon-pantheon-tui-direction-correction-…`. Two independent paths converged.

## The False Dichotomy ADR-018 Locked

ADR-018's framing collapsed two distinct propositions into one decision:

1. **The v0.22 TUI implementation is unreleasable.** True. Confirmed by user evidence and ratified.
2. **Therefore Pantheon should not have a TUI.** Does not follow.

The deletion of `internal/output/tui*.go` was the right tactical move (removed a brand-damaging surface that was reachable). The strategic claim that fell out of it — Sirsi abandons TUI as a product surface — was a step too far. ADR-020 corrects that.

## Decision

**The interactive-surface decision is reopened.** Pantheon's surface strategy is now under formal multi-track evaluation. No single surface is canonical until the evaluation closes.

Tracks under evaluation (see `docs/INTERACTIVE_SURFACE_COMPARISON.md` for the matrix):

1. **Redesigned Mole-grade TUI / operator console** — new build, not a restoration of the v0.22 code.
2. **Native SwiftUI Mac app / MenuBarExtra** — ADR-018's chosen direction.
3. **CLI + dashboard** — the surface as-shipped today, minus any reanimation of interactive chrome.
4. **Hybrid** — combinations of the above (e.g., native Mac shell with a TUI operator mode for power users; CLI + TUI; native + dashboard).

The user picks the path after reviewing the matrix. Codex-pantheon reviews before lock.

## What Stays True From ADR-018

These survive surface-independent and are **not** under re-evaluation:

- The v0.22 BubbleTea TUI was unreleasable. **The deleted code does not come back as the foundation.** Any revived TUI is a new design at Mole-grade quality.
- The `internal/dashboard` HTTP contract (the work captured in `docs/DASHBOARD_API.md`) is the canonical transport boundary regardless of surface choice. The Phase-2 batch-1 corrections (`vaultPrune` adapter, ID-based `vaultGet`, `kaHunt` response-shape rationale) survive.
- `cmd/sirsi-menubar/`'s 590 LOC of systray→TUI glue stays deleted (it depended on the v0.22 TUI specifically).
- The CLI surface (`sirsi <verb>`) is unchanged.

## What ADR-018 Got Wrong (Now Corrected)

- The framing "Pantheon's interactive surface moves from TUI to native SwiftUI Mac app, benchmarked against Mole" implied the *idea* of a TUI was abandoned. **This is rescinded.** TUI returns as a first-class candidate surface.
- The plan for Windows/Linux to be "CLI only. No native GUI planned. No TUI" was a default-from-elimination, not an evaluated choice. **Rescinded.** TUI may serve those platforms; the comparison matrix evaluates that explicitly.
- The CHANGELOG `[Unreleased]` framing was corrected after ADR-020: the v0.22 implementation was eliminated, but the TUI surface category remains active under Hybrid C.

## Relationship to ADR-018

ADR-018 is **partially in force**: the v0.22 TUI deletion stands; the multi-platform surface strategy is reopened. ADR-020 does not supersede ADR-018 (which would imply ADR-018 is wholly wrong). Instead, ADR-020 amends ADR-018's strategic scope while preserving its tactical deletion.

When ADR-020 closes with a surface decision, that closure either:

- **Confirms ADR-018's direction** (native Mac + CLI). ADR-020 becomes a "checked our work" record. ADR-018 stays accepted.
- **Adopts TUI as primary or hybrid.** ADR-020 supersedes ADR-018's surface claim while preserving its v0.22-deletion ruling.
- **Adopts CLI-only.** Both ADR-016 and ADR-018 surface claims are superseded.

## Constraints That Bind Any Outcome

These hold regardless of which surface(s) the user picks:

- **Mole-grade quality bar.** Whatever ships must clear the typography, hierarchy, animation, and density bar the user invoked. This applies to TUI revival, Mac native, and any hybrid alike. Quality is the gate; medium is the choice.
- **`sirsi-pantheon` rule A19 (no application bundle mutation)** continues. Mole inspection per `docs/PHASE1_MOLE_INSPECTION.md` was read-only and remains the only allowed engagement with `/Applications/Mole.app`.
- **No restoration of v0.22 TUI code as foundation.** A revived TUI is a new design.
- **Phase-2 batch-2 (socket transport + Mac PantheonBridge.swift + new dashboard endpoints) stays paused** until ADR-020 closes.

## "Why This TUI Will Be Different" Gate (Codex Condition 4)

If the user picks any track that includes a TUI (1, 4, 5, 6), the first deliverable is **not code**. It is a short design-proof document (`docs/TUI_DESIGN_PROOF.md`, ~5–10 pages) that must clear codex review before any `internal/tui/` package is created. The proof must specify:

1. **Layout system.** What primitive(s) compose the screen? Grid, list, panes, modal stack. Why this primitive set, not the v0.22 set.
2. **Density and typography rules.** Font choices for the target terminals (SF Mono, JetBrains Mono, Iosevka — what's the bar). Character cells per logical region. Spacing and divider rules. Glyph budget (Egyptian deity glyphs vs Unicode box drawing — explicit).
3. **Keyboard model.** Command palette? Modal (vim-like)? Modeless? Chord bindings? What's reachable in 0/1/2 keystrokes? Conflict resolution with terminal emulator shortcuts.
4. **Error states.** How are errors surfaced? Inline, banner, modal, log pane? Recovery path for each.
5. **Accessibility.** Screen-reader expectations (TUI accessibility is real but easy to break). Color contrast. High-contrast mode toggle. Reduced motion.
6. **Sample screens.** At least three canonical views (e.g., scan results table, Ra deployment status, router inbox) rendered as ASCII/Unicode mocks. These mocks are the visual proof — they replace a SwiftUI Figma prototype.
7. **Explicit "different from v0.22" deltas.** What specifically caused v0.22 to be unreleasable, and what does this design do differently for each item: tabs were decorative → tabs are navigable how; key bindings did nothing → key bindings dispatch through what mechanism; wrong color semantics → color system specification; wrong deity attribution → deity registry binding rule; stale CLI verb → command surface binding to post-restructure verbs; glyph rendering failure → font fallback strategy.

The design proof is reviewed by codex-pantheon and by the user before any Go code lands. If the proof does not credibly clear the Mole-grade bar in its medium, the TUI track does not start. No second chance to discover "unreleasable" after the code is written.

For Tracks 2 and 3 (Mac native only, CLI-only): this gate does not apply. Mac native has its own design proof requirement covered by the existing Phase-1 audits and the Mole inspection. CLI-only requires no new design proof.

## Required Deliverables Before Closing

Tracked separately in this Phase-2 amendment batch:

1. **This ADR.** ← here.
2. `docs/INTERACTIVE_SURFACE_COMPARISON.md` — track-by-track matrix.
3. `docs/PHASE1_RESCOPE_NOTE.md` — what survives, what is invalidated.
4. `docs/CANON_LANGUAGE_CORRECTION_PLAN.md` — file-by-file edit plan for the "failed implementation removed" vs "future TUI abandoned" distinction.

Routed as one bundle to codex-pantheon for review before any code or canon edits land.

## Open Questions Left For The User

Recorded here so the comparison matrix doesn't have to relitigate them:

1. Is the TUI a **product surface** (the way users experience Pantheon) or a **proof-of-craft demonstration** (we can build one because Sirsi-grade software demands it)? The framing changes the bar.
2. If TUI ships, what's the target audience — operators/power users, or general developers?
3. If both Mac native and TUI ship, what's the **default** surface a new user sees on first launch?
4. How does the Windows/Linux story land — TUI by default? CLI only with TUI opt-in? Native ports (out of scope today)?

These are pivoted to the user via the comparison matrix's recommendation column, not decided here.

## Closure (2026-05-29)

**User decision:** Hybrid C — TUI first cross-platform, Mac native later as the polish-bar upgrade.

**Implications now binding:**

1. **First deliverable is `docs/TUI_DESIGN_PROOF.md`** per §"Why This TUI Will Be Different" Gate. No `internal/tui/` Go code lands before that proof clears codex review.
2. **Phase-2 batch-1 docs** (`docs/DASHBOARD_API.md`, `docs/DASHBOARD_API_GAP.md`, `docs/DASHBOARD_ENVELOPE_DECISION.md`) survive in full; the dashboard contract is what the TUI consumes from inside the same Go process (no IPC layer needed for TUI).
3. **Phase-1 audits 1–3** (`PHASE1_MENUBAR_REUSE_AUDIT.md`, `PHASE1_MOBILE_GOMOBILE_AUDIT.md`, `PHASE1_IOS_REUSE_AUDIT.md`) become **deferred records** — their findings apply when the Mac native track activates in a later phase, not now.
4. **Phase-1 audit 4** (`PHASE1_MOLE_INSPECTION.md`) stays active — Mole-grade quality bar applies to the TUI build as much as the Mac build.
5. **Canon-correction commit** per `docs/CANON_LANGUAGE_CORRECTION_PLAN.md` lands as a single doc commit, ahead of any TUI code.
6. **Phase-2 batch-2 reshape** (next router item) is: (a) TUI design proof doc as gate 1; (b) `internal/tui/` scaffold after codex acks the proof; (c) Mac native deferred to a Phase-3 once the TUI clears its v1 quality bar.

**ADR-018 final status:** Partially In Force / Amended By ADR-020. The v0.22 deletion stands. The Mac-native-as-sole-Mac-surface direction is **deferred, not cancelled** — Hybrid C explicitly carries Mac native as a later upgrade.

**ADR-016 status:** Already marked Superseded by ADR-018; that remains correct under Hybrid C.

## Amendment (2026-06-01): Surface Ladder + Surfaces Are Router Threads

> User directive + codex-pantheon review (router items `20260601-055029`, `20260601-145331`). Folded into ADR-020 rather than a new ADR (ADR-021/022 are taken; resident-surface canon is concise enough to live here + A27).

**The ladder.** Interactive surfaces rank, in priority/richness order:

```
CLI  >  menubar  >  TUI  >  SwiftUI        (+ IDE plugins, editor-embedded, parallel to menubar)
```

This is a *priority* ordering, not a strict build sequence. All surfaces share the in-process `internal/dashboard` contract; none forks core logic.

**Surfaces are router threads.** "Registration" is load-bearing: every surface that can initiate work or take operator interaction is a **router-registered thread** (A26/A27), not just a renderer — registered to its own PID, heartbeating from its native runloop on a bounded interval (≥60s; never a frequent tick — `mds_stores` guardrail), idempotent on `(agent_id, pid)`, closed on graceful shutdown with ADR-022 OS-truth reaping as the hard-kill fallback. Surface ids: `menubar`, `tui`, `vscode`/`jetbrains`/`cursor`, `macapp`. Codified in A27 (`PANTHEON_RULES.md`/`CLAUDE.md`/`AGENTS.md`, "Resident UI surfaces are nodes too").

**Menubar: Go-now, Swift-incorporate-later.** The menubar already exists in Go (`cmd/sirsi-menubar/`, `fyne.io/systray`). It is **not** rewritten in Swift now; it is hardened in place and later *incorporated* into the rung-4 SwiftUI `MenuBarExtra`, which consumes the same dashboard contract — built once in spirit. "Built ahead, then incorporated" (user).

**Status.**
- **Step 1 — menubar router-registration: DONE** (commit `543e959`; codex VERIFY PASS on `145331`). Resident `agent=sirsi-menubar surface=menubar` thread, 60s bounded heartbeat decoupled from the stats tick, SIGTERM close, idempotent across restarts (`thread list` shows exactly one active record).
- **Step 2 — replace menubar Terminal-spawn actions with dashboard/result-contract results: APPROVED, next** (codex). Same constraints: no `internal/tui/`, no frequent heartbeat/write amplification, no additional long-running surface loop.
- **IDE plugins, SwiftUI** — later rungs; SwiftUI absorbs the menubar.

## /goal

Closed. Next item from this thread: canon-correction commit + Phase-2 batch-2 reshape proposal to codex-pantheon.
