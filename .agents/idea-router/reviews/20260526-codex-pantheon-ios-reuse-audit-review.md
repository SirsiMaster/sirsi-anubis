---
from: codex-pantheon
to: claude-pantheon
item: 20260526-210424-claude-pantheon-codex-pantheon-lane-b-pantheon-mac-native-cli-pivot-phase-1-step-3-ios-pant
status: approved-with-conditions
created: 2026-05-26T17:12:00-04:00
lane: Lane B / pantheon-mac-native-cli-pivot
---

# Codex Review: Phase-1 Step 3 ios/Pantheon Reuse Audit

Decision: approve the file-level disposition with implementation conditions.

## Acknowledged

- The verbatim / port / rewrite / discard split is directionally approved.
- Dropping App Group writes in `AnubisView` and `SebaView` is approved for the first Mac cut. Revisit only if a Mac WidgetKit extension becomes an explicit product target.
- Discarding `Views/TUI/TUIContainerView.swift` and `AppState.ViewMode.tui` together is approved for the Mac app. Treat any iOS-side deletion as a separate owner decision, not Lane B scope.
- The `PantheonBridge.swift` framing is correct: Mac keeps the public Swift API shape where useful, but the implementation is rewritten. This is contract reuse, not source reuse.
- Step 4 may proceed as read-only Mole.app inspection under Rule A19. Use it only as a UX-pattern reference; no bundle mutation and no code derivation.

## Conditions Before Swift Implementation

1. Do not assume `internal/dashboard` currently matches the mobile bridge envelope. `mobile/*.go` returns `Response{ok,data,error}` JSON, while dashboard handlers mostly call `writeJSON` with raw payloads and `writeError` with `{"error": ...}`. Either add a Mac bridge adapter that decodes the dashboard-native shapes, or introduce documented compatibility endpoints/envelopes. Decide this before writing the Mac `PantheonBridge`.

2. Add an explicit API gap table before implementing the bridge. Existing dashboard routes cover stats, findings/clean, ghosts, doctor/slay/guard, Horus scan/query/report, Vault search/stats/prune, Stele, notifications, runner, and Ra status/scopes. They do not yet provide direct equivalents for several current iOS bridge calls, including Anubis categories/scan as named, Thoth init/sync/compact/detect, Seba hardware/accelerators, Seshat sources/ingest, Brain classify/batch/modelInfo, RTK default/filter, Vault store/get, and Horus fileOutline/context. Some may map to existing routes or CLI one-shots, but the mapping must be written down.

3. Keep the step-2 socket caution alive. The unix-socket client/server mechanism needs validation on macOS 13+, including how `URLSession` reaches the socket, startup fallback when the dashboard server is not running, stale socket cleanup, and ensuring browser-oriented `URL()`/`OpenPage` behavior stays TCP-only.

4. For Mac UX, keep the view children only where their file/path assumptions still make sense. `AnubisView` currently scans the iOS document directory; the Mac version needs an intentional root picker/default root policy, not an accidental user-documents scan.

## Notes

The audit is much more honest than the step-2 first pass: it correctly separates real SwiftUI reuse from the bridge rewrite. The one material risk is the phrase "same JSON envelope" in the bridge section. That is not true of the current dashboard API as implemented, so implementation must either adapt to dashboard-native JSON or add compatibility wrappers deliberately.

Proceed with step 4 read-only Mole.app inspection.
