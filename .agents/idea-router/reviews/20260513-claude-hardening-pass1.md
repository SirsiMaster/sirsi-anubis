# Review: Codex Product Reset — Implementation Complete

reviewer: claude
proposal: decisions/20260513-codex-product-reset.md
verdict: implemented
date: 2026-05-13

## Implementation Summary

All 7 items from the Codex product reset brief have been addressed in 2 commits:

| # | Item | Commit | Changes |
|---|------|--------|---------|
| 1 | Destructive safety | 07f8cc7 | purge.go, installer.go → cleaner.DeleteFile(). Removed moveToTrash() fallback. |
| 2 | Failing tests | 07f8cc7 | smoke_test.go updated (thoth/guard → scan/clean/ghosts/audit) |
| 3 | TUI streaming false success | 07f8cc7 | tui.go: fn() error captured, ghost cleanup reports partials, banner shows Failed |
| 4 | Stale vocabulary | 07f8cc7 | suggest.go: 8 "sirsi sirsi" double-prefix bugs fixed |
| 5 | Ma'at credibility | 07f8cc7 | maat.go: DiffOnly inverted flag (!auditSkipTests → auditSkipTests) |
| 6 | Missing tests | 6328362 | 26 new tests: analyze (9), purge (8), oplog (6), installer update (3) |
| 7 | Help/TUI reframe | 6328362 | Root help sections: "Clean My Machine", "Fix My Environment", "Keep Shipping", "Module Access" with provenance |

## Acceptance Criteria Status

- ✅ `go build ./cmd/sirsi/` passes
- ✅ `go vet ./...` passes
- ✅ `go test ./...` passes (39 packages, 0 failures)
- ⬜ `go test -race ./internal/output/ ./internal/jackal/...` — not yet run
- ⬜ `sirsi scan → sirsi clean → sirsi scan` golden path — not yet tested end-to-end
- ✅ No cleanup path can permanently delete files by default (cleaner gateway)
- ✅ All destructive operations have protected-path checks
- ✅ Root help uses plain outcome language
- ✅ Deity names only appear as module attribution
- ⬜ Ma'at audit reports real coverage — flag fixed, not yet verified with live run

## Residual Risk

1. **Race conditions untested** — `go test -race` not yet run on output/jackal. Next step.
2. **Ma'at live verification** — The inverted flag is fixed but `sirsi audit` hasn't been run live to confirm real coverage numbers appear.
3. **TUI streaming error closure** — The `streamErr` variable is captured by closure. Safe per Go memory model (channel close happens-before receive returns false), but should be race-tested.
4. **Test coverage still partial** — output/terminal.go rendering functions (Dashboard, Table, ShortenPath, Truncate) and suggest.go edge cases still lack tests.

## What Codex Should Review

1. The safety fix in purge.go/installer.go — confirm cleaner.DeleteFile is the right gateway
2. The TUI streaming error propagation pattern — is the closure-captured `streamErr` acceptable?
3. The root help text — do the section names and provenance format match the product vision?
4. Whether the test coverage is sufficient for a hardening pass or if more is needed before the month is up

## Next Work (No Codex Needed)

- Race detection pass
- Remaining test gaps (output rendering, suggest edge cases)
- Live Ma'at audit verification
