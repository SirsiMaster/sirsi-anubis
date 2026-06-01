# Phase-1 Step 2 — `mobile/*.go` / Go-to-Swift IPC Audit

**Lane:** B / `pantheon-mac-native-cli-pivot`
**Thread:** `thr-4990a8df4cbd1468` (claude-pantheon)
**Status:** Draft — awaiting codex-pantheon review before any Phase-1 code.

## Purpose

Decide the IPC mechanism between the planned SwiftUI Mac app and the Go business core: **gomobile bindings**, **local HTTP/dashboard endpoint**, **subprocess CLI**, or **a hybrid**. Per codex's Phase-1 step 1 ack, this is a written-only audit; no code in this step.

## Surface Inventory

`mobile/` is 1,032 LOC of non-test code (and 1,501 LOC of tests). All exported functions follow one pattern: **JSON-in, JSON-string-out**, wrapped in a `Response { ok, data, error }` envelope from `mobile/mobile.go`.

| File | Non-test LOC | Exported functions | Underlying package |
| :--- | ---: | :--- | :--- |
| `mobile.go` | 34 | `Version` + envelope helpers | (none — utility) |
| `anubis.go` | 65 | `AnubisScan`, `AnubisCategories` | `internal/jackal` |
| `brain.go` | 194 | `BrainClassify`, `BrainClassifyBatch`, `BrainModelInfo`, `BrainInstallModel` | `internal/brain` |
| `horus.go` | 112 | `HorusParseDir`, `HorusFileOutline`, `HorusContextFor`, `HorusMatchSymbols` | `internal/horus` |
| `ka.go` | 38 | `KaHunt`, `KaEnumerateApps` | `internal/ka` |
| `rtk.go` | 95 | `RtkFilter`, `RtkDefaultConfig` | `internal/rtk` |
| `seba.go` | 24 | `SebaDetectHardware`, `SebaDetectAccelerators` | `internal/seba` |
| `seshat.go` | 110 | `SeshatIngest`, `SeshatListSources`, `SeshatListTargets`, `SeshatListKnowledgeItems` | `internal/seshat` |
| `stele.go` | 158 | `SteleReadRecent`, `SteleStats`, `SteleVerify` | `internal/stele` |
| `thoth.go` | 64 | `ThothInit`, `ThothSync`, `ThothCompact`, `ThothDetectProject` | `internal/thoth` |
| `vault.go` | 138 | (vault functions) | `internal/vault` |

**Build target today:** `gomobile bind -target=ios -o PantheonCore.xcframework ./mobile/` (Makefile). Also has `-target=android` for an .aar. No macOS target.

**Consumer today:** `ios/Pantheon/` (Swift app, 9 deity views) — production user of the xcframework.

## What's Already Working

`internal/dashboard/server.go` runs an HTTP server on **port 9119** with these JSON endpoints already wired:

```
/api/stats              /api/findings          /api/horus/report
/api/notifications      /api/clean             /api/horus/scan
/api/stele              /api/ghosts            /api/horus/query
/api/events             /api/ghosts/clean      /api/vault/search
/api/run                /api/doctor            /api/vault/stats
/api/run/status         /api/slay              /api/vault/prune
                        /api/guard/stats       /api/ra/status
                        /api/guard/renice      /api/ra/scopes
```

This dashboard is already started **inside `cmd/sirsi-menubar/`** (`main.go` line 115). The Pantheon process today is — by accident of evolution — a Go HTTP server with a systray UI. The SwiftUI Mac app inherits the server unchanged.

## The Four Options

### Option A — gomobile bindings (extend to macOS)

**How:** `gomobile bind -target=macos,maccatalyst -o PantheonCore.xcframework ./mobile/`. Mac app imports the same xcframework iOS already uses.

**Pros:**
- Code reuse with `ios/Pantheon/` is maximal. Same JSON envelope, same function names.
- Sync function calls from Swift, no network stack.
- No port allocation, no localhost binding, no firewall prompts.

**Cons:**
- gomobile's macOS target is *less mature* than iOS. `maccatalyst` works; pure `macos` target has historical rough edges with Apple Silicon and CGO.
- Doubles the Go runtime in the user's machine (one in the xcframework, one in `sirsi` CLI). ~12 MB extra binary.
- Every Go change requires xcframework rebuild and Xcode re-link. Slow inner loop.
- Function call model is request/response only — can't stream events or push notifications without a custom callback bridge (gomobile supports this but the patterns get awkward).
- gomobile is **maintenance-mode software**. Google has not deprecated it but does not actively develop it. Building the future on a frozen toolchain is a LEAN AF anti-pattern.

### Option B — Local HTTP / dashboard endpoint

**How:** Mac app starts (or attaches to) the existing `internal/dashboard.Server` on `127.0.0.1:9119`. SwiftUI uses `URLSession` + `Codable` against the JSON endpoints already wired.

**Pros:**
- **Zero new Go code.** The endpoints already exist and are exercised by `internal/dashboard/dashboard_test.go`.
- Same wire format the future web dashboard and any third-party tool will use. One contract, many clients.
- Streaming: `/api/events` and `/api/run/status` are already HTTP endpoints; SwiftUI can poll or upgrade to SSE/WebSocket later without changing the Go side.
- Inner loop: change a Go endpoint → restart the dashboard process → SwiftUI sees it. No Xcode rebuild.
- Decouples Swift app version from Go core version: one can be hot-fixed independently.

**Cons:**
- Loopback binding asks for `Local Network` permission on macOS Sonoma+ in sandboxed apps. Mitigation: bind to a **unix domain socket** at `~/Library/Application Support/ai.sirsi.pantheon/dashboard.sock` instead of `127.0.0.1:9119` — no entitlement needed.
- Port collision if user runs `sirsi` CLI dashboard simultaneously. Mitigation: lockfile via existing `platform.TryLock` pattern; first process wins, second attaches.
- Wire format is JSON over HTTP — overhead vs a direct function call. Negligible in practice (single-digit ms for the workloads here).

### Option C — Subprocess CLI

**How:** Mac app shells out to `sirsi <verb> --json` for each operation. Parses stdout as JSON.

**Pros:**
- Simplest possible mental model. The CLI is already the canonical surface.
- Zero coupling. If `sirsi` is updated independently, the app just sees newer output.
- Native to the macOS sandbox model (NSTask via Process API).

**Cons:**
- Process spawn overhead (~50–200 ms per invocation) — too slow for the 1 Hz stats refresh the menubar needs.
- No streaming. Long-running operations (`sirsi scan`, `sirsi guard`) require capturing stderr/stdout incrementally — doable but ugly.
- State held in the Go process is lost between invocations. `guard.StartBridge` can't run in a subprocess that exits after returning JSON.

### Option D — Hybrid (HTTP server + CLI subprocess)

**How:** SwiftUI app **owns** a long-lived `internal/dashboard.Server` child process for state-bearing operations (stats refresh, guard alerts, periodic scan). Falls back to `sirsi <verb> --json` subprocesses for one-shot operations where spawning a new Go process is appropriate (e.g., `sirsi diagram`, `sirsi version`).

**Pros:**
- Each operation uses the right transport.
- Reuses 100% of existing Go code: dashboard server stays as-is, CLI stays as-is.
- Server boundary is a unix socket → no Local Network entitlement.

**Cons:**
- Two transport mechanisms to maintain. Slightly more SwiftUI code (one URLSession config, one Process wrapper).

## Recommendation: **Option D — Hybrid, HTTP-primary**

LEAN AF analysis:

1. **Reuses the most existing code.** Option B already does ~95% of this; Option D just adds subprocess fallback for cases where it's obviously simpler. Option A doubles the Go runtime and locks the Mac app to a maintenance-mode toolchain. Option C alone can't carry the stats loop.

2. **One wire format.** The same JSON envelope shipped from `internal/dashboard/api.go` is what the future Pantheon web UI, Horus-as-fleet-dashboard, and any third-party integration will see. Picking Option B/D means the SwiftUI app is just **the first client of a contract that scales**.

3. **No gomobile rebuild cycle.** Phase-1 will iterate heavily; tying Go-side iteration to xcframework rebuilds + Xcode re-linking would slow every cycle by ~30 seconds.

4. **Unix socket sidesteps the macOS sandbox.** `127.0.0.1:9119` would trigger the Local Network permission prompt — a brand-damaging first-run experience. Binding to `~/Library/Application Support/ai.sirsi.pantheon/dashboard.sock` avoids it entirely and is a one-line change in `internal/dashboard/server.go` (swap `net.Listen("tcp", ...)` for `net.Listen("unix", ...)`).

5. **`mobile/*.go` stays useful — for iOS.** This audit does **not** discard `mobile/*.go`. It stays the iOS bridge (and Android, if that target is still alive). For Mac, we don't load the xcframework.

## What Changes In Go

The Mac app changes are mostly Swift. Go-side work is small and additive:

| Change | Where | Size | Notes |
| :--- | :--- | ---: | :--- |
| Add `unix://...sock` listener mode to dashboard.Server | `internal/dashboard/server.go` | ~20 LOC | Config-flag-driven: `Config.Socket string`. TCP path stays default for CLI use. |
| Document the JSON envelope contract | `docs/DASHBOARD_API.md` (new) | ~120 LOC | Lists every endpoint, request shape, response shape. Required before the Swift app starts consuming. |
| Confirm `SIRSI_HEADLESS` is unreferenced outside `cmd/sirsi-menubar/` | (audit only, no code) | 0 LOC | **✅ Confirmed in this audit.** `grep -rn SIRSI_HEADLESS` returns only `cmd/sirsi-menubar/main.go:37`, this doc, and the prior audit reviews. Safe to delete with the menubar batch (codex condition 4). |

## What Stays Out Of Scope

- **iOS bindings.** `mobile/*.go` continues to ship `PantheonCore.xcframework` for iOS. This audit is Mac-only.
- **Android.** Out of scope for Lane B.
- **Authentication / mTLS / IPC hardening.** Unix-socket file-mode `0600` under the user's home is sufficient for a single-user Mac app. Defer until/unless a multi-user use case appears.
- **Streaming protocol (SSE vs WebSocket vs long-poll).** The current `/api/events` is a simple poll endpoint. Decide streaming format when the first real-time view in SwiftUI needs it.

## Closing Codex's Step 1 Conditions

Re-confirming the five conditions from the prior review:

1. **`MenuBarExtra` + macOS 13+** — confirmed default. Carry only AppKit/`NSStatusItem` if a separate ADR forces it later.
2. **Module home** — with HTTP-primary, kept business logic (`StatsSnapshot`, `CollectStats`, `collectDeities`, `collectRa`) does **not** need to move to `internal/menustats/`. It already lives in `cmd/sirsi-menubar/stats.go` *and* is consumed by `internal/dashboard/api.go`. The dashboard's `StatsFn` callback (`main.go:120`) is the boundary. Recommend: extract to `internal/menustats/` only if/when the dashboard server moves out of `cmd/sirsi-menubar/` — separate item.
3. **`findSirsiBinary` comment** — stale ADR-016 reference flagged for cleanup in the menubar batch.
4. **`runHeadless` / `SIRSI_HEADLESS`** — ✅ verified single-site. Safe to delete with menubar batch.
5. **LaunchAgent removal** — distinguishing note: `cmd/sirsi-menubar/bundle/ai.sirsi.pantheon.plist` is the **app login item** (replace with `SMAppService`); `com.sirsi.idea-router.plist` is the **Idea Router FSEvents watcher** (Lane A, owned by codex, **untouched**).

## Step 3 Authorization Ask

If codex approves Option D, step 3 is the `ios/Pantheon/` file-level audit — specifically, which SwiftUI views and view models port to macOS with minor changes vs. which are iOS-specific (Siri Shortcuts, WidgetKit, touch gestures). After that, step 4 is the read-only Mole.app inspection.

## /goal

Codex review of: (a) the Option D / HTTP-primary recommendation, (b) the unix-socket transport choice for sandboxing, (c) the confirmation that `SIRSI_HEADLESS=1` is unreferenced and can be deleted with the menubar batch. Step 3 (`ios/Pantheon/` audit) starts on ack.

## References

- Codex step 1 review: `.agents/idea-router/items/20260522-024232-…cmd-sirs.md` `## Result`
- Step 1 audit: `docs/PHASE1_MENUBAR_REUSE_AUDIT.md`
- Dashboard server: `internal/dashboard/server.go`, port `9119` per `internal/dashboard/colors.go:23`
- Existing iOS consumer: `ios/Pantheon/` (audited in step 3)
