# Phase-1 Re-Scope Note

**Governing ADR:** ADR-020 (Interactive Surface Reopened)
**Purpose:** State which Phase-1 audit findings survive surface-independent and which are invalidated by the reopening. Codex-pantheon's required step before any further Phase-2 work.

## Audits in Question

| Audit | File |
| :--- | :--- |
| 1 | `docs/PHASE1_MENUBAR_REUSE_AUDIT.md` — `cmd/sirsi-menubar/` |
| 2 | `docs/PHASE1_MOBILE_GOMOBILE_AUDIT.md` — Go↔Swift IPC choice |
| 3 | `docs/PHASE1_IOS_REUSE_AUDIT.md` — `ios/Pantheon/` file-level |
| 4 | `docs/PHASE1_MOLE_INSPECTION.md` — Mole UX reference |

Plus the Phase-2 batch-1 docs (`docs/DASHBOARD_API.md`, `docs/DASHBOARD_API_GAP.md`, `docs/DASHBOARD_ENVELOPE_DECISION.md`), included here for completeness because the codex ack-with-conditions on the reopen explicitly named them as survivors.

## Audit 1 — `cmd/sirsi-menubar/` Reuse

### Survives surface-independent

- The disposition table for `stats.go` business logic (`StatsSnapshot`, `CollectStats`, `collectDeities`, `collectRa`). These call `internal/dashboard` / `internal/vitals` / `internal/deity` directly. Same business surface regardless of UI track. **Reusable by TUI, Mac native, or CLI status views alike.**
- The deletion of the 590 LOC of systray→TUI AppleScript glue. That code was specific to spawning the v0.22 bubble-tea TUI through Terminal/iTerm — the deleted artifact stays deleted under every track (a revived TUI is a redesign, not a restore).
- `findSirsiBinary` and the `TryLock` pattern.
- `Info.plist` / `LaunchAgent` discard reasoning — only valid if the Mac app is chosen as one of the tracks.

### Invalidated or re-scoped

- The audit's headline framing ("with the TUI eliminated, every case is a dangling target") was correct for v0.22-TUI elimination, but the broader "TUI eliminated" claim is now false under ADR-020. **Reword in the canon correction batch.**
- The recommendation to rewrite `liveState`/`updateTitle` as Swift only stands if Track 2/4/A is picked. Under Tracks 1/5/6, that priority logic lives in Go (TUI status bar) or both.
- The `MenuBarExtra`-as-default conclusion is Mac-specific and presumes Mac native is chosen.

### Action for codex

Hold this audit's body intact as a Mac-track-conditional record. Add a header note saying "applies only if Track 2 or any hybrid that includes Mac native is chosen."

## Audit 2 — `mobile/*.go` / Go↔Swift IPC

### Survives surface-independent

- **The HTTP-over-unix-socket choice (Option D) is the correct boundary regardless of surface.** TUI, Mac native, and any hybrid all consume the same `internal/dashboard` JSON contract. The transport pick only matters in the Mac-app case (where the unix-socket avoids the Local Network permission). For TUI, the same Go process *is* the dashboard; no IPC needed.
- The "gomobile stays iOS-only" decision. iOS continues to use `mobile/*.go`. No track in ADR-020 changes this.
- The `SIRSI_HEADLESS=1` verified-single-site finding.

### Invalidated or re-scoped

- The audit was framed as "Mac IPC choice." Under ADR-020, the IPC choice is Mac-conditional. If Mac native is not chosen, the socket transport is not required (the dashboard's TCP-on-127.0.0.1 default suffices for the browser dashboard).
- Codex's conditions for the socket implementation (additive `Config.Socket`, stale-socket cleanup, no accidental `URL()`/`OpenPage` use in socket mode) only bind if Mac native is chosen.

### Action for codex

Same as audit 1 — header note marking it conditional on Track 2/4/A/C(later phase).

## Audit 3 — `ios/Pantheon/` Reuse

### Survives surface-independent

- The Models layer (`ios/Pantheon/Models/*.swift`, 565 LOC) describes the same JSON shapes from `internal/dashboard`. They are usable as a Swift consumer's `Codable` mirror whenever a Swift surface is built — but they describe the **wire format**, which is what matters.
- The principle that the dashboard JSON contract is the boundary, not the bridge implementation.
- The discard of `Views/TUI/TUIContainerView.swift` *as an iOS-side cleanup* — it was the embedded bubble-tea-emulator inside the iOS app, depending on v0.22 TUI. It stays discarded regardless of track. Codex's step-3 ack noted iOS-side deletion is out of Lane B; that holds.

### Invalidated or re-scoped

- The "port to Mac" disposition (Views/Shared, Views/Deities, Theme, AppState, ContentView, PantheonApp) is Mac-conditional. Under non-Mac tracks, the iOS code stays iOS-only.
- The Bridge rewrite plan (HTTP over unix socket) is Mac-conditional.

### Action for codex

Same as audits 1 and 2. Conditional on a Mac native track.

## Audit 4 — Mole.app Inspection

### Survives surface-independent

- **The Mole-grade quality bar applies to whatever ships.** Typography, animation, density, scoped TCC permission strings, selective custom iconography, Sparkle as a future option, English-only for v1 — all transferable to a TUI build too (the typography/density notes especially).
- Rule A19 compliance precedent (read-only bundle inspection only).
- The "Mole is Dock-launched, not menubar-only" correction stays factual.

### Invalidated or re-scoped

- The audit's framing of Mole as "Pantheon's Mac app polish bar reference" is Mac-conditional. Under TUI-only tracks, Mole becomes a reference for *the quality bar generally*, not the Mac UX specifically.

### Action for codex

This audit's findings are the most surface-portable. Header note: "the Mole quality bar applies to any surface; the Mac-specific patterns apply only if a Mac native track is chosen."

## Phase-2 Batch-1 Docs — Full Survivors

The codex ack on the ADR-018 reopen explicitly named these as surface-independent:

- `docs/DASHBOARD_API.md` — describes the existing HTTP transport contract. Every track consumes it.
- `docs/DASHBOARD_API_GAP.md` — the 19 new endpoints + 6 adapters are still required for any surface that wants typed access to the deities. TUI, Mac native, dashboard front-end — all benefit.
- `docs/DASHBOARD_ENVELOPE_DECISION.md` (Option A — Swift-side adapter) — the Swift-side language remains Mac-conditional, but the **principle** (clients adapt to the dashboard's HTTP-idiomatic shape rather than rewrapping it) generalizes. For a TUI consuming the dashboard from the same Go process, no adapter is needed (direct function calls); the principle still says "don't introduce a wrapper layer."
- Batch-1 corrections (vaultPrune Adapter, vaultGet ID-based, kaHunt response-shape rationale). Codex pre-approved these in the reopen ack.

No changes needed to the batch-1 docs from the reopening. The vaultPrune / vaultGet / kaHunt corrections are still pending implementation against `docs/DASHBOARD_API_GAP.md` — apply them as small edits in the next doc pass, separate from this re-scope note.

## What Phase-2 Batch-2 Becomes Under Each Track

Sized in advance so the user sees the implementation footprint per choice:

| Track | Phase-2 batch-2 scope |
| :--- | :--- |
| **Track 1 (TUI)** | New `internal/tui/` package (~2–4K LOC Go). Reuses `internal/dashboard` business logic directly (no IPC layer needed). New `cmd/sirsi-tui/` or fold into `cmd/sirsi/`. Charm libs (`lipgloss`, `bubbletea` returning) are the standard. **Risk to call out:** "we just deleted bubbletea." |
| **Track 2 (Mac Native)** | Per Phase-1 audits 1–3 as written. Socket transport in `internal/dashboard.Server` + Mac `PantheonBridge.swift` + 19 new dashboard endpoints + Xcode project + signing. |
| **Track 3 (CLI+Dashboard)** | No new work. Phase-1 audits archive, Phase-2 batch-1 corrections still land. |
| **Hybrid A** | Tracks 1 + 2 combined. Two surfaces, two release pipelines. |
| **Hybrid B** | Track 1 + thin SwiftUI Mac wrapper around the TUI (~few hundred LOC). |
| **Hybrid C** | Track 1 now; Track 2 deferred to a later phase. |

## Action Items Captured For Codex's Review

1. Hold Phase-1 audits 1–3 as conditional records (Mac-track-conditional). Add header notes when the surface decision closes.
2. Phase-1 audit 4 (Mole) is the most portable; only minor scoping note needed.
3. Phase-2 batch-1 docs stay; vaultPrune/vaultGet/kaHunt corrections still pending as separate small edits.
4. Phase-2 batch-2 reshapes per the user's surface pick (matrix in `docs/INTERACTIVE_SURFACE_COMPARISON.md`).

## /goal

Codex ack of the re-scope categorizations as part of the doc-only reopening bundle.
