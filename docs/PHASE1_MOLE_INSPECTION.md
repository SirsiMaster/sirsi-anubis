# Phase-1 Step 4 — Mole.app Read-Only Inspection

**Lane:** B / `pantheon-mac-native-cli-pivot`
**Thread:** `thr-4990a8df4cbd1468` (claude-pantheon)
**Status:** Draft — closes Phase-1 audits on codex ack.

## Rule A19 Compliance

This inspection is **read-only**: `ls`, `PlistBuddy -c "Print"`, `file`, `otool -L`, `du`. No `cp`, `mv`, `rm`, or write operation targeted `/Applications/Mole.app/**`. No code in this audit derives from Mole's source (there is no source — Mole's GitHub at github.com/tw93/MoleApp is the canonical source if/when we want to study patterns more deeply, *outside* this inspection).

**Purpose:** UX-pattern reference for Pantheon's Mac app polish bar — typography, animation density, framework choices, system integration patterns. **Not a feature list.** Not a UI clone.

## Bundle Identity

| Field | Value |
| :--- | :--- |
| `CFBundleName` | Mole |
| `CFBundleIdentifier` | `com.tw93.MoleApp` |
| `CFBundleShortVersionString` | 1.2.0 |
| `LSMinimumSystemVersion` | **14.0** (macOS Sonoma) |
| Built with | Xcode 17 / macOS 26.4 SDK |
| Bundle size | 18 MB |
| Architecture | Universal (x86_64 + arm64) |
| Copyright | © 2026 tw93 |
| Update mechanism | Sparkle 2.9.1, `SUFeedURL=https://mole.fit/appcast.xml`, Ed25519 (`SUPublicEDKey`) |

**`LSUIElement` is NOT set.** Mole runs as a **regular Dock-launched app**, not a menubar-only app. This is the first inversion vs. our current planning assumption: I had been treating Mole as the benchmark for a `MenuBarExtra` experience. Mole is actually a windowed app with a Dock presence. The implication for Pantheon is *not* that we should follow — it's that the "Mole quality bar" the user invoked applies to **either form factor**. Our choice (MenuBarExtra) remains valid, just confirmed as a deliberate divergence.

## Framework Stack (linked, read via `otool -L`)

Mac-system frameworks linked:
- **SwiftUI** (current version 7.4.27) — primary UI
- **AppKit** — fallback / window chrome
- **Combine** — reactive state plumbing
- **Observation** (`libswiftObservation.dylib`) — `@Observable` macro adopted
- **SceneKit** + `libswiftSceneKit.dylib` — the planet imagery is **3D scenes**, not static images
- **CryptoKit** — for Sparkle's Ed25519 signature verification
- **CoreServices**, **IOKit**, **CFNetwork**, **CoreGraphics**, **Security**, **Spatial** — system-level integration
- **Sparkle.framework** (2.9.1) — the only third-party framework. Bundled with helper apps: `Autoupdate`, `Updater.app`, `XPCServices/`

**No declared usage of:** Electron, React Native, Tauri, Catalyst, Mac Catalyst, GoMobile, or any cross-platform shim. **Pure native SwiftUI + AppKit + SceneKit.** This is the bar.

## Resources (read via `ls /Applications/Mole.app/Contents/Resources/`)

```
ActivityMonitor.icns   earth.jpg     mars.png         neptune.jpg    sun.jpg
Assets.car             jupiter.jpg   mercury.png      saturn_ring.png
Daemon.icns            mole-logo.png mole.icns        saturn.jpg
Setting.icns
```

Observations relevant to our work:

1. **No `.lproj` directories** — Mole ships English-only. Localization is *not* the polish bar.
2. **Custom `.icns` files for in-app affordances** (`ActivityMonitor.icns`, `Daemon.icns`, `Setting.icns`) — they don't reuse SF Symbols for every icon. There's hand-crafted icon work, but selectively.
3. **Planet textures as JPG / PNG** — they're SceneKit material maps. The "planet exploration" metaphor is a decorative SceneKit scene, not gameplay or interactive 3D. Pantheon does not need this.
4. **`Assets.car`** — compiled Xcode asset catalog. Standard.

## Permissions Profile (Info.plist usage descriptions)

```
NSDesktopFolderUsageDescription:        "Mole scans your Desktop for leftover installer files."
NSDocumentsFolderUsageDescription:      "Mole scans Documents for leftover installer files."
NSDownloadsFolderUsageDescription:      "Mole scans Downloads for leftover installer files."
NSSystemAdministrationUsageDescription: "Mole runs optional maintenance tasks that require administrator access."
```

Patterns:
- **Plain, scoped, one-sentence usage strings.** No marketing. No vague "improve your experience" copy. Each string names the user-visible folder and the verb. This is the standard our `Info.plist` should match when Pantheon requests file-system access.
- **Folder-scoped TCC entitlements** rather than `com.apple.security.files.user-selected.read-write`. Mole asks for the three concrete folders most installer leftovers live in. Pantheon should consider the same: prefer scoped Desktop/Documents/Downloads/Applications-support over blanket file access.

## What Pantheon Should Take From This (Pattern References Only)

1. **Native pure-SwiftUI + AppKit is achievable at 18 MB.** Sparkle and a SceneKit scene push it to that size; without the planets it's smaller. Our Mac app should target a similar order of magnitude.
2. **Sparkle 2.x + Ed25519 SUFeedURL is the canonical OSS update path.** If Pantheon's Mac app wants out-of-band updates (not Homebrew-only), Sparkle is the reference. Defer adopt/no-adopt to a later product decision; do not block Phase-1 implementation on it.
3. **Permission strings should be concrete, scoped, and single-sentence.** Reuse the pattern when Pantheon asks for Documents/Downloads/Applications-support access for cache cleaning, ghost detection, etc.
4. **Selective custom iconography over universal SF Symbols.** Pantheon already uses Egyptian glyphs as deity identifiers; that aesthetic alignment can land at the same fidelity with hand-crafted `.icns` for menubar/in-app affordances.
5. **macOS 14+ as `LSMinimumSystemVersion` is defensible for a 2026 app.** Pantheon currently targets macOS 13+ per the step-1 decision (`MenuBarExtra` API). Mole confirms a 14+ target is also LEAN — most Mac users upgrade promptly. If we hit any 13+ API friction, raising to 14+ is supported by precedent. *Recommendation: hold at 13+ for the menubar API breadth; revisit if a 14+ API materially simplifies something.*

## What Pantheon Should NOT Take

1. **3D planet UI.** It's Mole's signature aesthetic, not a transferable pattern. Pantheon's brand language is Egyptian glyphs + gold-on-black, codified in `Theme/PantheonTheme.swift`. SceneKit is not on our shortest path to "Mole-grade quality" — quality there means typography, hierarchy, animation timing, *not* literal 3D.
2. **Dock-launched windowed app form factor.** ADR-018 commits to MenuBarExtra. Don't redecide.
3. **Localization.** Out of scope for v1.

## Phase-1 Audits Closure

Steps 1–4 are now drafted:

| Step | Subject | Status |
| :--- | :--- | :--- |
| 1 | `cmd/sirsi-menubar/` reuse audit | ✅ approved-with-conditions |
| 2 | `mobile/*.go` IPC audit (Option D, unix socket) | ✅ approved-with-conditions |
| 3 | `ios/Pantheon/` file-level reuse audit | ✅ approved-with-conditions |
| 4 | Mole.app read-only inspection | ⏳ this document, awaiting ack |

**On step-4 ack, Phase-1 audits close.** Implementation begins under Phase 2, gated by per-batch codex reviews. The first implementation batch carries the consolidated condition list from steps 1–4 (recap below).

## Consolidated Conditions Carried Into Implementation

From step 1:
- `MenuBarExtra` + macOS 13+ confirmed default.
- Module home (`cmd/sirsi-menubar/` vs `internal/menustats/`) decided in step 2 → not moved.
- `findSirsiBinary` ADR-016 comment cleanup with menubar batch.
- `SIRSI_HEADLESS=1` deletion verified safe; ship with menubar batch.
- LaunchAgent removal distinguishes app login item (replace with `SMAppService`) from Idea Router watcher (Lane A, untouched).

From step 2:
- Unix-socket `Config.Socket string` is **additive only**; TCP stays zero-config default.
- Tests for: TCP default, socket mode listen, deliberate stale-socket cleanup, `URL()`/browser-open NOT used in socket mode.
- Socket path: `~/Library/Application Support/ai.sirsi.pantheon/`, restrictive permissions, stale-only-after-confirming-no-owner removal. Single-user; defer auth/mTLS.
- `docs/DASHBOARD_API.md` is a prerequisite before Swift depends on the contract.

From step 3:
- **API gap table required** before writing the Mac `PantheonBridge.swift`. Existing dashboard endpoints cover stats/findings/clean/ghosts/doctor/slay/guard/horus-{scan,query,report}/vault-{search,stats,prune}/stele/notifications/runner/ra; they do **not** yet provide direct equivalents for several iOS bridge calls (Anubis categories/scan as named, Thoth init/sync/compact/detect, Seba hardware/accelerators, Seshat sources/ingest, Brain classify/batch/modelInfo, RTK default/filter, Vault store/get, Horus fileOutline/context). Write the mapping (existing endpoint, new endpoint, or CLI one-shot) before the bridge.
- **JSON envelope mismatch:** `mobile/*.go` returns `Response{ok,data,error}`; `internal/dashboard` mostly uses bare `writeJSON` payloads and `writeError{error: …}`. Choose: add a Mac bridge adapter that decodes dashboard-native shapes, or introduce documented compatibility envelopes. Decide before the bridge code.
- `AnubisView` Mac port needs an intentional root-picker / default-root policy, not an accidental user-documents scan inherited from iOS.

From step 4 (this doc):
- Match Mole's permission-string style (concrete, scoped, single-sentence).
- Hand-crafted `.icns` for menubar/in-app affordances; keep SF Symbols for everything else.
- Sparkle 2.x + Ed25519 noted as a future option; do not block Phase-1 on update mechanism.

## /goal

Codex ack of: (a) this inspection's findings, (b) the consolidated condition list above as the Phase-2 implementation gate. On ack, I draft a Phase-1 completion decision and submit the first Phase-2 implementation batch proposal (likely: API gap table + dashboard envelope decision + `docs/DASHBOARD_API.md`, ahead of any Swift code).

## References

- Step 1: `docs/PHASE1_MENUBAR_REUSE_AUDIT.md`
- Step 2: `docs/PHASE1_MOBILE_GOMOBILE_AUDIT.md`
- Step 3: `docs/PHASE1_IOS_REUSE_AUDIT.md`
- Step 3 review (codex, approved-with-conditions): `.agents/idea-router/reviews/20260526-codex-pantheon-ios-reuse-audit-review.md`
- Rule A19 — `PANTHEON_RULES.md` §2.16 (No Application Bundle Mutations — ABSOLUTE PROHIBITION)
- ADR-018 (TUI sunset, native Mac app adopted)
