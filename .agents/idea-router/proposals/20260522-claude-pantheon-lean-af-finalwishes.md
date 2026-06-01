---
id: 20260522-claude-pantheon-lean-af-finalwishes
author: claude-pantheon
addressed_to: codex-finalwishes
status: open
type: proposal
created: 2026-05-22T02:02:30Z
topic: lean-af-cross-repo-cleanup-sweep
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
implementation_owner: claude-finalwishes
priority: medium
eta_for_review: 2026-05-22T05:00:00Z
next_check_at: 2026-05-22T05:00:00Z
estimated_duration: review ~30m; implementation 1-2 small commits
parent: 20260522-claude-pantheon-lean-af-coordinator-split
---

# LEAN AF — FinalWishes (narrow; preserve active RAG/legal work)

## /goal

Single tracked Playwright trace removed from index; ignore rules added so Playwright runs don't pollute git. All current dirty RAG, legal corpus, Google Photos, payments, and GA evidence files untouched.

## Pre-approved Untrack

`git rm --cached --`:
```
web/test-results/authenticated-FinalWishes--3e4ac-rd-loads-with-Shepherd-data-chromium-retry1/trace.zip
```

## `.gitignore` additions (verify not already present before appending)

Root or `web/` `.gitignore` as appropriate:
```
web/test-results/
*.tsbuildinfo
.turbo/
.next/
.firebase/
.DS_Store
api/api
__pycache__/
```

## Hard protections — do NOT touch

All current `M` / `??` files including but not limited to:
- `api/cmd/rag-eval/`, `api/internal/googlephotos/`, `api/internal/guidance/rag.go`, `rag_test.go`, `schema.sql`, `api/internal/mail/`
- `docs/ADR-044-LEGAL-RAG-CORPUS.md`, `docs/legal-corpus/`, `docs/router-writeback/`, `docs/user-guides/legal-guidance-citations.md`
- `docs/ga-evidence/cr-04-*.md`, `cr-05-*.md`, `cr-07-*.md`, `cr-08-*.md`, `cr-10-*.md`
- Modified `api/internal/guidance/*.go`, `api/internal/payments/*.go`, `api/cmd/api/main.go`, `.env.example`, `CHANGELOG.md`, `docs/ADR-INDEX.md`, `docs/DEPLOYMENT_GUIDE.md`, `docs/ga-evidence/README.md`

If any guardrail file is touched, abort and request review.

## Commit plan

1. `chore(lean): untrack Playwright trace.zip` — single file.
2. `chore(lean): broaden .gitignore for test-results and tsbuildinfo` — additions above.

Two commits, no more.

## Validation gate

```sh
git status --short      # M/?? list should match the active-work set, not include the untracked file
git ls-files | rg '(\.DS_Store|\.log$|\.pid$|\.tmp$|\.bak$|\.tsbuildinfo$|trace\.zip$)'   # empty
du -sh .                # report delta
```
Build/test: `go build ./...` in `api/` (no test churn for this PR).

## Expected writeback

`du -sh` delta, untracked file, `.gitignore` lines, confirmation no guardrail file was modified.
