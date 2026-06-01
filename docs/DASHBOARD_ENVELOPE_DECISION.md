# Dashboard JSON Envelope Decision

**Lane:** B / `pantheon-mac-native-cli-pivot`
**Status:** Decision — awaiting codex-pantheon review with `DASHBOARD_API.md` and `DASHBOARD_API_GAP.md` in one batch.
**Decides:** how the Mac `PantheonBridge.swift` reconciles the envelope mismatch between `mobile/*.go` (`Response{ok,data,error}`) and `internal/dashboard` (bare resource body + `{"error":"..."}` on non-2xx).

## Background

`mobile/*.go` returns:

```json
{ "ok": true,  "data": <T-as-raw-message> }
{ "ok": false, "error": "<msg>" }
```

`internal/dashboard` returns:

```json
<T-as-root>                              // 2xx
{ "error": "<msg>" }                     // non-2xx
```

Plus two edge cases on the dashboard side that return **200 OK with an `error` field in the body** when the underlying resource is empty (`/api/findings` when no scan has run; `/api/vault/stats` when the vault is unavailable). Documented in `DASHBOARD_API.md`.

The Mac app cannot consume both shapes without a choice.

## Options

### Option A — Swift-side adapter (decode dashboard-native)

The Mac `PantheonBridge.swift` decodes the dashboard's native shapes directly. Each method:

```swift
func anubisScan(...) async throws -> ScanResult {
    let (data, resp) = try await urlSession.data(for: req)
    if let http = resp as? HTTPURLResponse, http.statusCode >= 400 {
        let err = try JSONDecoder().decode(ErrorBody.self, from: data)
        throw BridgeError.server(err.error, status: http.statusCode)
    }
    return try JSONDecoder().decode(ScanResult.self, from: data)
}
```

**Pros:**
- **Zero Go changes** for envelope concerns. Existing dashboard endpoints, current tests, current browser dashboard, and CLI `sirsi dashboard` users keep working unchanged.
- The HTTP status code is the canonical error signal — idiomatic for HTTP clients.
- The two 200-OK-with-in-body-error edge cases are visible and handled per-endpoint where they live, not hidden in a global wrapper.
- Web dashboard and any third-party tool see the same shape. **One contract, many clients** — the point of choosing HTTP in step 2.

**Cons:**
- Swift `PantheonBridge.swift` looks slightly different from iOS `PantheonBridge.swift` (which decodes `BridgeResponse<T>` via the envelope). Confirmed acceptable in step 3: same public API, different internals. The iOS-Mac code-sharing claim was always contract-sharing, not implementation-sharing.
- Per-endpoint adapter for the two soft-state cases (`/api/findings`, `/api/vault/stats`) — small Swift code carrying the in-body-error knowledge.

### Option B — Dashboard compat envelope layer (opt-in via header or path)

Add `Response{ok,data,error}` wrapping to `internal/dashboard` for clients that request it, e.g.:

- **Header gate:** `Accept: application/vnd.sirsi.envelope+json` → wrap response.
- **Path prefix:** `/api/v2/...` → mirror endpoint with wrapped shape.

Mac bridge calls request the envelope; CLI/browser stay on the bare shape.

**Pros:**
- iOS-Mac Swift bridges look almost identical internally — both decode `BridgeResponse<T>`.
- A single Swift `BridgeResponse<T>: Decodable` survives across both platforms.

**Cons:**
- **New Go code in every handler** — every `writeJSON(w, v)` must become `writeJSON(w, maybeWrap(r, v))` and every `writeError` must check the request and emit the wrapped shape. Affects 28+ handlers. Touching every endpoint to support an opt-in wrapper is the opposite of LEAN.
- **Two contracts to maintain.** The bare-shape and wrapped-shape paths must stay synchronized forever. Tests double.
- **Drift risk.** Future endpoints may forget the `maybeWrap` and silently break the Mac client.
- **The wrapper buys nothing over HTTP status codes.** `{ok: false, error: "x"}` with HTTP 200 is strictly worse than `{error: "x"}` with HTTP 500 for cache layers, monitoring, and any reverse proxy.

### Option C — Dashboard rewrite to envelope-by-default (breaking change)

Replace `writeJSON(w, v)` with `writeEnvelope(w, v)` globally. Single shape forever.

**Pros:**
- Eliminates the divergence entirely.

**Cons:**
- **Breaks every existing client.** The browser dashboard, CLI scripts that piped `sirsi dashboard` output, and any external integration would all need updates simultaneously.
- Violates Rule 11 (Do No Harm) and Rule 12 (Additive-Only Changes) from Universal Rules.
- Solves a problem we don't actually have: HTTP status codes are already the canonical error signal.

**Rejected.**

## Decision

**Option A — Swift-side adapter.**

### Reasoning

1. **LEAN.** Option B adds Go code to every handler in `internal/dashboard` to support an envelope only the Mac client wants. The Mac client can do its own decoding in one file. Cost is in the place that benefits.
2. **HTTP-idiomatic.** Status codes already are the error signal. Wrapping them in a JSON `ok` field is a layer on top of a layer.
3. **Contract stability.** The dashboard's wire format is what the future web dashboard, Horus, and third-party tools will see. Wrapping for one client risks fossilizing a wrapper-only contract by accident.
4. **iOS parity argument is contract-deep, not code-deep.** The iOS bridge's public Swift API (`anubisScan(...) async throws -> ScanResult`) survives unchanged on Mac because the Swift signature *is* the contract. What's underneath is implementation, and step 3's review explicitly approved that framing.
5. **Soft-state errors are honest.** The two endpoints that return 200-OK-with-in-body-error (`/api/findings` empty, `/api/vault/stats` unavailable) describe a real product distinction: "the resource is reachable but currently empty/uninitialized." Per-endpoint handling expresses that more honestly than a global `ok` flag.

### Required Swift-side code

Approximate budget for the Mac `PantheonBridge.swift`:

```swift
struct ErrorBody: Decodable { let error: String }
enum BridgeError: Error { case server(String, status: Int); case decode(Error) }

// Generic decoder used by every bridge method.
func decode<T: Decodable>(_ data: Data, _ resp: URLResponse) throws -> T {
    if let http = resp as? HTTPURLResponse, http.statusCode >= 400 {
        let err = (try? JSONDecoder().decode(ErrorBody.self, from: data))?.error ?? "http \(http.statusCode)"
        throw BridgeError.server(err, status: http.statusCode)
    }
    do {
        return try JSONDecoder().decode(T.self, from: data)
    } catch {
        throw BridgeError.decode(error)
    }
}
```

Two per-endpoint adapters needed for the in-body-error cases (`/api/findings`, `/api/vault/stats`) — each is ~10 lines that check for an `error` field after a successful decode.

Total estimated Swift bridge envelope-handling code: **~80 LOC**. Compares favorably to Option B's per-handler Go changes across 28+ handlers.

## Implications For Phase-2 Batch Two

When the implementation batch begins:

- **No Go changes required for envelope handling.** `internal/dashboard.Server` stays as-is.
- **`Config.Socket` unix-socket transport** (step 2 condition) is the only Go change in batch two.
- **Swift `PantheonBridge.swift`** uses a generic `decode<T>` helper plus two per-endpoint soft-state adapters.
- **New endpoints from `DASHBOARD_API_GAP.md`** all use the existing `writeJSON`/`writeError` pattern. The envelope decision is consistent across legacy and new.

## Alternatives Documented And Rejected

- **Option B (compat envelope, opt-in).** Rejected for LEAN and contract-stability reasons.
- **Option C (envelope-by-default rewrite).** Rejected for Rule 11/12 violation.

## /goal

Codex review of the three docs (`DASHBOARD_API.md`, `DASHBOARD_API_GAP.md`, this one) as a single batch. Specifically: confirm Option A, ack the new-endpoint set in the gap table, and decide the single AMBIGUOUS case (`anubisScan` → new `POST /api/scan` for iOS parity vs runner+events path for long-running). On ack, the Phase-2 batch-two proposal opens with: (1) socket transport, (2) Mac PantheonBridge.swift scaffolding using Option A decoder, (3) the new endpoints in priority order.
