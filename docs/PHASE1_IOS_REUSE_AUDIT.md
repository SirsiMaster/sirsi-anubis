# Phase-1 Step 3 — `ios/Pantheon/` File-Level Reuse Audit

**Lane:** B / `pantheon-mac-native-cli-pivot`
**Thread:** `thr-4990a8df4cbd1468` (claude-pantheon)
**Status:** Draft — awaiting codex-pantheon review before any Phase-1 code.

## Purpose

Decide which files in `ios/Pantheon/` port verbatim to the macOS app, which need rewrites, and which are iOS-only and should not be carried over. Honest framing this time — where the seams are, not just LOC totals.

## Headline Decision Already Locked

Step 2 chose **HTTP-primary over a unix socket** for Mac IPC; gomobile bindings stay iOS-only. That single choice forces the largest concrete change in the iOS code: **`Services/PantheonBridge.swift` is rewritten on Mac**, even though its public Swift API (`anubisScan(rootPath:categories:)`, `kaHunt(includeSudo:)`, etc.) stays nearly identical. Internal swap: `MobileAnubisScan(...)` synchronous function calls become `URLSession.shared.data(for: URLRequest(url: socketURL.appending("/api/anubis/scan")))`. Same JSON envelope. Different transport.

This means: the **Models** and **Views** can be lifted with discipline, but **Services** is essentially new code.

## Surface Inventory

4,517 LOC of Swift across 28 files in `ios/Pantheon/`:

```
App/                        568 LOC (4 files)
Models/                     565 LOC (10 files)
Services/PantheonBridge     291 LOC (1 file)
Theme/                       45 LOC (1 file)
Views/Deities/             2533 LOC (8 deity views)
Views/Shared/               232 LOC
Views/Previews/              94 LOC
Views/TUI/                  229 LOC
Info.plist / .entitlements      (URL scheme + App Group)
```

## Disposition By File

### Models (565 LOC) — all 10 files keep verbatim

Pure `Codable` structs mirroring the Go JSON envelopes from `mobile/*.go`. They describe the same wire format the Mac app will consume from `/api/anubis/*`, `/api/horus/*`, etc. **No changes needed.**

| File | LOC | Notes |
| :--- | ---: | :--- |
| `AnubisModels.swift` | 47 | `ScanCategory`, `ScanResult`, `Finding`, `RuleError`, `CategorySummary` |
| `KaModels.swift` | 49 | `GhostApp`, `InstalledApp` |
| `ThothModels.swift` | 30 | `ProjectInfo`, sync envelopes |
| `SebaModels.swift` | 44 | `HardwareProfile`, `AcceleratorProfile` |
| `SeshatModels.swift` | 47 | Knowledge sources/targets |
| `BrainModels.swift` | 73 | Classification responses |
| `HorusModels.swift` | 57 | Symbol graph, outlines |
| `RTKModels.swift` | 36 | Filter config + response |
| `SteleModels.swift` | 106 | Event ledger entries |
| `VaultModels.swift` | 36 | Vault search results |

**Verification before port:** confirm `internal/dashboard/api.go` response shapes still match these structs. If `internal/dashboard` JSON ever drifts from `mobile/*.go` JSON, the audit chose the wrong wire format. (Spot check during step 4 / implementation; not blocking this audit.)

### Theme (45 LOC) — verbatim

`Theme/PantheonTheme.swift` is pure SwiftUI `Color` + `Font` definitions plus a hex initializer. Identical on Mac. **Keep.**

### Services (291 LOC) — **rewrite**

`Services/PantheonBridge.swift` is `import PantheonCore` + 20+ wrapper functions calling `MobileAnubisScan(...)`, `MobileKaHunt(...)`, etc. synchronously and wrapping each in `Task.detached` for async ergonomics.

Mac port: same public API surface (`anubisScan`, `kaHunt`, `thothSync`, …), but each method's body becomes:

```swift
let req = URLRequest(url: socket.appending(path: "/api/anubis/scan"))
let (data, _) = try await urlSession.data(for: req)
let envelope = try JSONDecoder().decode(BridgeResponse<ScanResult>.self, from: data)
```

`URLSession` configured with a unix-socket connection delegate (Foundation supports this on macOS 13+). Same `BridgeResponse<T>` envelope decoder.

**Estimated Mac Services LOC:** ~250–300 (similar size; different internals). Keeping the public API stable means **callers in `Views/` don't need to change** — the discipline of the iOS bridge pays off here.

### App (568 LOC) — mixed

| File | LOC | Disposition | Notes |
| :--- | ---: | :--- | :--- |
| `App/PantheonApp.swift` | 86 | **Rewrite** | iOS-only: `WindowGroup`, `WidgetKit` import, `host_statistics64` page-size assumption (16 KB ARM64), `WidgetCenter.shared.reloadAllTimelines()` on launch. Mac app uses `MenuBarExtra` scene; system stats come from the dashboard server's `/api/stats`, not from in-app Mach calls. No widget rebuild. |
| `App/AppState.swift` | 90 | **Port with deletions** | `ActiveDeity` enum (8 cases + glyph/subtitle/icon) keeps verbatim. `viewMode: ViewMode.gui/.tui` is **dead** — no TUI on Mac, no TUI in the iOS app post-ADR-018 either; delete the enum and the `.tui` case throughout. `handleDeepLink` (`sirsi://deity/{name}`) keeps verbatim — URL schemes work on Mac. |
| `App/ContentView.swift` | 208 | **Rewrite navigation, keep view children** | iPhone uses `TabView` with `tabItem` for deity selection (5–8 tabs is the iOS sweet spot). Mac equivalent is a `NavigationSplitView` sidebar or, for `MenuBarExtra`, a vertical list. The eight `AnubisView()`, `KaView()`, … children render identically; only the chrome around them changes. |
| `App/PantheonIntents.swift` | 184 | **Port with substitutions** | `AppIntents` framework is available on macOS 13+, so the framework imports survive. The intents themselves (`AnubisScanIntent`, `ThothSyncIntent`, `SebaHardwareIntent`) all call `MobileAnubisScan(...)` etc. directly — must route through the new HTTP bridge. Replace with `await bridge.anubisScan(...)`. The Siri-Shortcuts surface is less compelling on Mac; instead, these intents become Spotlight / menubar-menu actions. Same code structure. |

### Views/Deities (2,533 LOC) — mostly port, two need closer look

| File | LOC | Disposition | iOS-only deps |
| :--- | ---: | :--- | :--- |
| `AnubisView.swift` | 185 | **Port with one deletion** | Lines 80–83 write scan results to App Group for widgets (`SharedDataManager.saveScanResults(...)`). Mac doesn't have iPhone widgets; drop. (macOS has desktop widgets via WidgetKit, but that's a separate product decision — out of scope for step 3.) |
| `KaView.swift` | 154 | **Verbatim port** | No widget glue, no UIKit. SwiftUI primitives only. |
| `ThothView.swift` | 607 | **Port, biggest** | Uses `NavigationStack` + `navigationTitle` — works on Mac. The size warrants its own read during implementation but no obvious iOS-only deps. |
| `SebaView.swift` | 181 | **Port with one deletion** | Lines 71–81 write hardware JSON to App Group for widgets. Same drop as `AnubisView`. |
| `SeshatView.swift` | 170 | **Verbatim port** | No iOS-only deps. |
| `BrainView.swift` | 630 | **Port, second biggest** | No App Group writes, but `WidgetKit` imported (likely vestigial). Confirm during port; drop import if unused. |
| `HorusView.swift` | 240 | **Verbatim port** | No iOS-only deps. |
| `SteleView.swift` | 366 | **Verbatim port** | No iOS-only deps. |

### Views/Shared (232 LOC) — verbatim

`SharedComponents.swift` — buttons, badges, severity pills. SwiftUI primitives only. **Keep.**

### Views/Previews (94 LOC) — verbatim

`DeityPreviews.swift` — `#Preview` macros. **Keep.**

### Views/TUI (229 LOC) — **discard**

`TUIContainerView.swift` is an in-app terminal emulator (gold-on-black monospace, command input, `TerminalState` model) that mimicked the BubbleTea TUI inside the iOS app. Per ADR-018 (TUI eliminated), this view's premise no longer holds. The iOS app should also delete it eventually — out of Lane B scope, but flag.

The `AppState.ViewMode.tui` case noted above is the toggle that revealed this view. Deleting both is one consistent change.

### Info.plist / Entitlements — adapt

- `Info.plist` — `CFBundleURLTypes` for `sirsi://` URL scheme. **Keep** — deep links work on Mac.
- `Pantheon.entitlements` — `com.apple.security.application-groups = group.ai.sirsi.pantheon`. **Drop** for the Mac app's first cut (no widget extension). Add back if/when a Mac widget extension is built. The Mac app will need different entitlements (sandbox, login-item via `SMAppService`, network client for the unix socket).

## Reuse Tally

- **Verbatim Swift (no edits):** ~1,295 LOC — all Models, Theme, Views/Shared, Views/Previews, half of `AppState.swift`, ~75% of the deity views.
- **Port with minor edits (delete App Group writes, switch bridge calls):** ~1,800 LOC — most deity views, `AppState.swift`, `PantheonIntents.swift`.
- **Rewrite (HTTP transport over unix socket, MenuBarExtra scene, sidebar navigation):** ~500 LOC — `PantheonBridge.swift`, `PantheonApp.swift`, `ContentView.swift` navigation.
- **Discard:** ~229 LOC — `Views/TUI/TUIContainerView.swift` and the `ViewMode` enum it depended on.

Net for the Mac app: **~3,095 LOC carried** (Models + Theme + most Views) + **~750 LOC new or rewritten** (Bridge HTTP transport, app shell, navigation, Mac-shaped intents). Total Mac SwiftUI ~3,850 LOC.

## Where I Was Wrong-Headed Before (Honest Notes Per Codex's Step-2 Feedback)

In the step-2 audit I framed the unix-socket switch as "swap `net.Listen("tcp", ...)` for `net.Listen("unix", ...)`." Codex flagged that as too LEAN. Same risk here: the Bridge rewrite **is** new code, even though the file count stays low. The Models reuse is real; the Bridge reuse is a *contract* reuse, not a *code* reuse. Calling the Bridge "ported" would be a soft lie.

Concrete framing for the Bridge work:

- New `URLSession.Configuration` for unix-socket connections (macOS 13+ supports this via `URLSession`'s `connectionProxyDictionary` or a custom `URLProtocol`; the exact mechanism needs validation in implementation).
- Stale-socket fallback: if the dashboard server isn't running, the app needs to start it (the same way `cmd/sirsi-menubar/main.go:115` does today) or instruct the user.
- Background task lifecycle differs on Mac — no app-extension model for the menubar; the app itself runs as long as the user wants it in the menubar.

## Step 4 Authorization Ask

If codex approves this disposition, step 4 is the **read-only Mole.app inspection** (Rule A19 — no bundle mutation). The output of step 4 is a UX-pattern reference for the Mac app's polish bar — typography, animation, density, navigation idioms. No code derives from Mole.

## /goal

Codex review of the file-level disposition. Specifically: (a) confirm the verbatim/port/rewrite/discard split, (b) confirm dropping the App Group writes in `AnubisView` / `SebaView` for the Mac app's first cut, (c) confirm `Views/TUI/TUIContainerView.swift` and `AppState.ViewMode.tui` are discarded together, (d) ack that "Services/PantheonBridge.swift is rewritten" is honest framing despite identical public API.

## References

- Step 1 audit (`cmd/sirsi-menubar/`): `docs/PHASE1_MENUBAR_REUSE_AUDIT.md`
- Step 2 audit (`mobile/*.go` IPC): `docs/PHASE1_MOBILE_GOMOBILE_AUDIT.md`
- Step 2 review (codex, approved-with-conditions): `.agents/idea-router/reviews/20260522-codex-pantheon-mobile-gomobile-ipc-audit-review.md`
- ADR-018 (TUI sunset, native Mac app adopted)
