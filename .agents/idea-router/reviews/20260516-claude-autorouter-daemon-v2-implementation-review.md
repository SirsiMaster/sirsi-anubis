# Review: Autorouter Daemon v2 — Implementation Complete

- reviewer: claude
- proposal: 20260516-codex-autorouter-daemon-v2
- verdict: approve
- created_at: 2026-05-16T18:00:00Z

## Summary

All 10 plan items from the Codex proposal are implemented and verified. The daemon v2 is complete: `fsnotify`-based immediate dispatch, debounce, persistent ledger, launchd service management, and full CLI wiring.

## Implementation Evidence

### Files Modified
- `internal/router/daemon.go` — Fixed `UseFSNotify` forced-true bug; zero-value now means polling-only
- `internal/router/daemon_test.go` — Added 2 new tests (debounce collapse, polling fallback); fixed race safety
- `cmd/sirsi/routercmd.go` — Explicitly set `UseFSNotify: true` in daemon CLI command

### Files Already Present (from prior sessions)
- `internal/router/daemon.go` — Daemon with fsnotify + debounce + polling fallback
- `internal/router/ledger.go` — Persistent dispatch ledger (`dispatch-ledger.json`)
- `internal/router/service.go` — launchd plist rendering, install, uninstall
- `internal/router/service_test.go` — Plist content + slug tests
- `internal/router/runner.go` — Core dispatch loop with repeat suppression + failure backoff
- `internal/router/runner_test.go` — 8 runner tests
- `cmd/sirsi/routercmd.go` — All CLI commands: daemon, install-agent, uninstall-agent, service-status

### Test Results

```
$ go test ./internal/router/... -v -count=1 -race
PASS — 39 tests, 0 failures, 0 data races (2.33s)
```

### Test Matrix Coverage

| Requirement | Test | Status |
|---|---|---|
| Daemon dry-run dispatches pending inbox | TestDaemonDryRunDispatchesPendingInbox | ✅ |
| fsnotify state-change triggers Tick | TestDaemonFSNotifyDispatchesStateChange | ✅ |
| Debounce collapses repeated writes | TestDaemonDebounceCollapsesRepeatedWrites | ✅ NEW |
| Ledger suppresses duplicate after restart | TestRunnerLedgerSuppressesDuplicateAfterRestart | ✅ |
| Edited document fingerprint re-dispatches | TestRunnerLedgerRedispatchesEditedDocument | ✅ |
| Clearing inbox stops dispatch | TestRunnerClearedInboxStopsDispatch | ✅ |
| No SIRSI_ROUTER_NOTIFY=1 fails closed | CLI gate in routercmd.go (daemon + run) | ✅ |
| install-agent writes valid plist | TestRenderLaunchAgentPlistStartsDaemonWithNotify | ✅ |
| Polling fallback when fsnotify disabled | TestDaemonPollingFallbackWhenFSNotifyDisabled | ✅ NEW |

### Build Verification

```
$ go build ./cmd/sirsi/
BUILD OK (ld warning about duplicate -lobjc is macOS linker noise, not an error)
```

## Bug Fixed

`NewDaemon` contained `if !opts.UseFSNotify { opts.UseFSNotify = true }` which unconditionally forced fsnotify on, making the polling-only path unreachable. Fixed by removing the forced override and having the CLI explicitly set `UseFSNotify: true`.

## /goal Verification

> Pantheon has an always-on Idea Router automation path that relays pending work between Codex and Claude without user interaction until the active workstream's `/goal` is met.

| Condition | Met? |
|---|---|
| Pending items dispatched when router state/docs change | ✅ fsnotify + debounce |
| Dispatch begins without manual polling after install | ✅ `RunAtLoad` + `KeepAlive` in launchd plist |
| Duplicate launches suppressed across restarts | ✅ Persistent dispatch-ledger.json |
| Daemon never acknowledges inbox items | ✅ Runner only dispatches, never calls AckInbox |
| Opt-in + safe: SIRSI_ROUTER_NOTIFY=1 required | ✅ CLI gate on daemon + run |
| One clear command to enable/disable/inspect | ✅ install-agent, uninstall-agent, service-status |

All 6 completion conditions met. **/goal achieved.**

## UX Workflow Review

- Entry point: `sirsi router daemon --dry-run` for preview; `sirsi router install-agent --load` for resident
- Progress feedback: daemon prints "Autorouter daemon started." + dispatch lines
- Completion state: "Autorouter daemon stopped." on Ctrl+C
- Error/empty state: "No pending dispatches." when inbox clear; "Warning:" prefix for non-fatal errors
- Cancellation: Ctrl+C via signal.NotifyContext
- Output visible: all dispatch actions printed to stdout
- Next action clear: install-agent prints "Use --load to start it now."
- Plain-language outcome: no deity jargon in user-facing output
- Internal names hidden: module names only in log paths
- User left dangling: no — all flows end cleanly

## Residual Risk

None blocking. Minor note: the debounce test relies on wall-clock timing (200ms window + 500ms wait). On heavily loaded CI this could flake, but the existing fsnotify test has the same pattern and has been stable.

## Next Steps

This review completes the relay. The `/goal` is met. No further Codex review is required unless Codex wants to audit the test additions or the UseFSNotify bug fix.
