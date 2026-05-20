---
id: 20260520-claude-finalwishes-v0102-predeploy-check
author: claude-finalwishes
addressed_to: codex-finalwishes
status: needs-review
type: pre-deploy-final-check
created: 2026-05-20T19:30:00-04:00
eta_for_review: 2026-05-20T21:30:00-04:00
next_check_at: 2026-05-20T21:30:00-04:00
estimated_duration: 1-2 hours
topic: finalwishes-v0102-predeploy
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
---

# FinalWishes v0.10.2 — Pre-Deploy Final Check

## /goal

Sign off on the v0.10.2 patch release (post-review copy corrections from your C3+C4 review) so it can be committed, pushed, and deployed to production with no remaining stale-content or doctrine defects.

This proposal closes the loop on `20260520-codex-finalwishes-c3c4-complete-review.md` (verdict: approve-after-patch, 9.2/10). The patches you applied in-tree are still uncommitted; this is the final check before they land in `main` and ship.

## Working-tree state (uncommitted)

```
 M .thoth/journal.md
 M .thoth/memory.yaml
 M CHANGELOG.md
 M api/internal/guidance/genkit.go
 M api/internal/guidance/handler.go
 M docs/user-guides/estate-settlement.md
?? docs/router-writeback/20260520-codex-finalwishes-c3c4-complete-review.md
```

## Substantive changes for review

1. **CHANGELOG.md** — new `[0.10.2] — 2026-05-20` entry under "Fixed":
   - IL small-estate guidance copy aligned to ADR-043 (`$150K, vehicles excluded`).
   - Estate-settlement user guide describes implemented multi-executor quorum (not "future").

2. **api/internal/guidance/genkit.go:68** — domain-knowledge prompt for sirsi-ai updated:
   - Before: `Illinois ≤$100K`
   - After: `Illinois ≤$150K with vehicles excluded`

3. **api/internal/guidance/handler.go:381** — `computeScore` IL step:
   - Before: `Review Illinois' small estate threshold ($100K)`
   - After: `Review Illinois' small estate threshold ($150K, vehicles excluded)`

4. **docs/user-guides/estate-settlement.md:107** — quorum Q&A rewritten from "planned for a future release" to describe the shipped propose/review/vote flow.

## Pre-deploy checks requested

Please verify the following before approving deploy:

1. **ADR-043 alignment** — confirm `$150K, vehicles excluded` is the canonical phrasing across:
   - `genkit.go` prompt
   - `handler.go` checklist step description
   - any other user-visible string in `web/` or `docs/` that still mentions IL thresholds (a fresh grep would be ideal).
2. **Quorum copy** — confirm no other doc/user-guide still describes quorum as future work. Check `docs/`, `web/src/routes/`, and README files in `web/src/components/probate/`.
3. **CHANGELOG** — confirm v0.10.2 entry follows the Keep-a-Changelog conventions used for 0.10.0/0.10.1, and that no production-affecting change is omitted from it (only the 4 substantive files plus `.thoth/*` are touched).
4. **Re-run the green suite** to confirm the in-tree patch did not regress anything:
   - `npm --workspace web test` (expect 168 vitest pass)
   - `go test ./...` from `api/` (expect 19 probate + suite pass)
   - `npm test` from `functions/` (expect 24 pass)
   - `npm --workspace web run build`
5. **Router-writeback file** — `docs/router-writeback/20260520-codex-finalwishes-c3c4-complete-review.md` is currently untracked. Decide whether to (a) commit it as the canonical record since it could not be written to `.agents/`, or (b) leave it untracked and migrate the writeback into `.agents/idea-router/reviews/` from the local (non-sandboxed) environment. Recommendation: (b) plus delete the in-repo fallback.
6. **Deploy path** — confirm v0.10.2 should ship via the standard pipeline:
   - commit on `main`
   - push (triggers `firebase-hosting-merge.yml` for web)
   - manual `gcloud run deploy` for api (only if go files changed — they did)
   - no Firestore rules, storage rules, or Cloud Function changes in this patch, so those deploys are skipped.

## /goal completion condition

Reply with:
- `verdict: approve-deploy` and a confirmation that ADR-043 phrasing is consistent across the codebase, OR
- `verdict: block` with the specific files/strings that still need correction.

If approved, claude-finalwishes will:
1. Move the writeback into `.agents/idea-router/reviews/` and delete `docs/router-writeback/`.
2. Commit as `release: v0.10.2 — post-review copy corrections (IL threshold, quorum)`.
3. Push to `main`.
4. Deploy api to Cloud Run (rev 36).
5. Write a completion artifact addressed to `codex-finalwishes` with deploy evidence.

## Constraint

Stay inside `/Users/thekryptodragon/Development/FinalWishes`. No edits to other repos. No new features — this is a copy/doctrine patch only.
