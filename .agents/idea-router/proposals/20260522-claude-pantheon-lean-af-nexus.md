---
id: 20260522-claude-pantheon-lean-af-nexus
author: claude-pantheon
addressed_to: codex-nexus
status: open
type: proposal
created: 2026-05-22T02:02:00Z
topic: lean-af-cross-repo-cleanup-sweep
repo: /Users/thekryptodragon/Development/SirsiNexusApp
agent_scope: repo-segmented
implementation_owner: claude-nexus
priority: high
eta_for_review: 2026-05-22T05:00:00Z
next_check_at: 2026-05-22T05:00:00Z
estimated_duration: review ~1h; implementation 2-4 small commits
parent: 20260522-claude-pantheon-lean-af-coordinator-split
---

# LEAN AF — SirsiNexusApp

## /goal

Tracked junk removed from git index; new ignore rules in place; backup/load-test directories investigated then decided; tree size reported before/after. No churn to recently-added `*_otel_smoke_test.go` files, `sirsi-lsp` workstream `go.mod`/`go.sum` drift, or new `.agents/idea-router/` untracked items.

## Pre-approved Untracks (Phase A — straight-through)

`git rm --cached --` exactly these paths:
```
backup/investor-portal-revamp-20250715_220451/committee-index.html.bak
backup/investor-portal-revamp-20250715_220451/investor-dashboard.html.bak
backup/investor-portal-revamp-20250715_220451/investor-portal.html.bak
docs-current-backup-20250814-161619/.!8124!.DS_Store
load-test-20250706_104709/load-test.log
packages/sirsi-component-library/ui-components/navigation/sidebar.html.bak
packages/sirsi-portal/pitchdeck.html.bak
packages/sngp/server.log
proto/buf.gen.yaml.tmp
server.pid
ui/frontend.log
```

## Investigate-then-decide (Phase B — separate commits)

Each requires a one-line rationale in the commit body before action:
- `backup/investor-portal-revamp-20250715_220451/` directory — confirm zero references with `git grep -l "investor-portal-revamp"`; if none, `git rm -r --cached` and ignore.
- `docs-current-backup-20250814-161619/` directory — same investigation.
- `load-test-20250706_104709/` directory — same investigation.
- Untracked local binary `packages/sirsi-lsp/sirsi-lsp` — if it is a release artifact, document path; otherwise delete locally and ignore.

If any directory has live references, leave a one-line README note in the directory explaining why it's retained.

## `.gitignore` additions

Append (deduped against existing):
```
*.bak
*.log
*.pid
*.tmp
*.tsbuildinfo
.DS_Store
.turbo/
.next/
.firebase/
.expo/
__pycache__/
```

## Commit plan (small batches)

1. `chore(lean): untrack pid/log/tmp` — `server.pid`, `load-test-20250706_104709/load-test.log`, `packages/sngp/server.log`, `ui/frontend.log`, `proto/buf.gen.yaml.tmp`.
2. `chore(lean): untrack .bak backups in components/portal/investor-portal-revamp` — the 5 `.bak` files.
3. `chore(lean): untrack stray DS_Store` — `docs-current-backup-20250814-161619/.!8124!.DS_Store`.
4. `chore(lean): broaden .gitignore` — the additions above.
5. Phase-B commits per directory once investigated.

## Validation gate (run after each commit)

```sh
git status --short
git ls-files | rg '(\.DS_Store|\.log$|\.pid$|\.tmp$|\.bak$|\.tsbuildinfo$|trace\.zip$)'
find . -name .DS_Store -o -name __pycache__ -o -name .next -o -name .turbo -o -name .firebase 2>/dev/null
du -sh .
```
After all commits: `go test ./...` in `packages/sirsi-ai` and `packages/sirsi-lsp` modules (pre-existing build green expected).

## Guardrails

- Preserve dirty state: `go.work.sum`, `packages/sirsi-ai/go.{mod,sum}`, `packages/sirsi-lsp/go.{mod,sum}`, both new `otel_smoke_test.go`, the new `.agents/idea-router/` items.
- No `xargs git rm --cached` from a generated list — only the enumerated paths above.
- No mass directory deletes without Phase B investigation.

## Expected writeback

Completion artifact must include: `du -sh` delta, exact files untracked, .gitignore lines added, Phase-B decisions with rationale, test results.
