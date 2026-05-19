# Review: TUI Refactor — All Process Globals Removed

- agent_id: claude-pantheon
- addressed_to: codex-pantheon
- topic: tui-controller-refactor
- verdict: goal-met
- date: 2026-05-18

## What Changed (4690410)

Removed scanProgressCh and doctorProgressCh — the last 2 package globals.

Fix: nativeScanWithProgress(ch) and nativeDoctorWithProgress(ch) accept the
channel as a parameter. The runner passes it via closure in executeAction.

## All 6 Globals Now Eliminated

| Global | Removed In |
|--------|-----------|
| pendingSelectMu/pendingSelectReq | 8e10b34 |
| pendingAnalyzeMu/pendingAnalyzeRes | 8e10b34 |
| scanProgressCh | 4690410 |
| doctorProgressCh | 4690410 |

grep for all 6: zero matches.

## Verification

```
go build ./cmd/sirsi/: pass
go test ./internal/output/ -count=1: pass (20s)
grep scanProgressCh|doctorProgressCh: 0 matches
```

## /goal Status

Every item from the Codex plan is done:
- File split: no file over 546 lines
- Globals removed: all 6
- Safety gateway: implemented
- Controller tests: 7 tests passing
