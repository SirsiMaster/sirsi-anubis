# Proposal: FinalWishes Repo Hygiene & Audit

- agent_id: claude-finalwishes
- repo: /Users/thekryptodragon/Development/FinalWishes
- topic: finalwishes-repo-hygiene
- addressed_to: codex-pantheon
- created_at: 2026-05-18T13:45:00-04:00

## /plan

Full repo audit: compare FinalWishes state against memory, CANONICAL_DEVELOPMENT_PLAN.md, and CHANGELOG. Fix all discrepancies found. Verify zero-error claim with tsc, npm audit, go vet, go test, and vitest.

## /goal

1. VERSION file matches CHANGELOG (0.9.0)
2. All orphaned [Unreleased] CHANGELOG sections resolved
3. AGENTS.md files committed (root + web/)
4. 0 npm vulnerabilities
5. 0 TypeScript errors (tsc --noEmit)
6. 0 Go vet issues
7. All Go tests passing (11/11 packages)
8. All Vitest tests passing (168/168)

## What Changed (claude-finalwishes session, 2026-05-18)

### Completed:
- **VERSION**: `0.1.0-alpha` → `0.9.0` (was stale since March)
- **CHANGELOG**: Removed 2 orphaned `[Unreleased]` sections. First (lines 281-308) retagged under `[0.9.0] (cont.)` as pre-session 9 work. Second (lines 731-734) removed entirely — all planned items were completed months ago.
- **AGENTS.md**: Committed root (216 lines, full operational directive) and web/ (15 lines, router stub). Both were untracked.
- **npm audit**: Fixed 7 vulnerabilities (5 moderate protobufjs, 2 high protobufjs code injection + express-rate-limit). Now 0 vulns.
- **Go tests**: Fixed 11 failing tests in `internal/service/estate/`. Root cause: `checkEstateAccess()` requires user ID in context, but tests used bare `context.Background()`. Fix: added `testCtx()` helper using existing `auth.InjectUserIDForTest()`. All 11 packages now pass.

### Verification results:
| Check | Before | After |
|-------|--------|-------|
| tsc --noEmit | 0 errors | 0 errors |
| npm audit | 7 vulns | 0 vulns |
| go vet | 0 issues | 0 issues |
| go test ./... | 10/11 pass | 11/11 pass |
| vitest | 168/168 pass | 168/168 pass |

### Audit findings (informational, no action taken):
- **SMS provider**: `functions/index.js` has 2 TODOs for real SMS integration (placeholder only)
- **DNS switchover**: `finalwishes.app` not live — still at `finalwishes-prod.web.app` (owner action)
- **Google Photos API**: Dropped from plan in favor of Cloud Storage uploads, but CANONICAL_DEVELOPMENT_PLAN.md not updated to reflect this
- **Cloud Functions tests**: Zero test files for Firebase Functions (tested manually only)

### Commit:
`0f06d7b` — `chore: repo hygiene — VERSION sync, CHANGELOG cleanup, test fixes, vuln remediation`

## Tests/builds run
- `npx tsc --noEmit` — 0 errors
- `npm audit` — 0 vulnerabilities
- `npm test -- --run` — 168 passed, 0 failed
- `go vet ./...` — 0 issues
- `go test ./...` — 11/11 packages passed

## Failures or blockers
None. All /goal criteria met.

## Next action for Codex
- Review and acknowledge this hygiene pass
- Optionally: update CANONICAL_DEVELOPMENT_PLAN.md to note Google Photos API was replaced by Cloud Storage uploads (decision was made but plan not updated)
- Optionally: add `finalwishes-repo-hygiene` to `completed_topics` in `state.json`
