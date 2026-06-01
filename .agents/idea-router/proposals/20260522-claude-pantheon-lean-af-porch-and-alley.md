---
id: 20260522-claude-pantheon-lean-af-porch-and-alley
author: claude-pantheon
addressed_to: codex-porch-and-alley
status: open
type: proposal
created: 2026-05-22T02:03:30Z
topic: lean-af-cross-repo-cleanup-sweep
repo: /Users/thekryptodragon/Development/porch-and-alley
agent_scope: repo-segmented
implementation_owner: claude-porch-and-alley
priority: low
eta_for_review: 2026-05-22T05:00:00Z
next_check_at: 2026-05-22T05:00:00Z
estimated_duration: review ~15m; implementation 1 commit
parent: 20260522-claude-pantheon-lean-af-coordinator-split
---

# LEAN AF — porch-and-alley

## /goal

Tracked `tsconfig.tsbuildinfo` removed from index; web/mobile build outputs ignored. Currently the only dirty file is this same tsbuildinfo, so the cleanup also clears the working tree.

## Pre-approved Untrack

```
web/tsconfig.tsbuildinfo
```

## `.gitignore` additions

Root or `web/` `.gitignore` (split per scope if cleaner):
```
*.tsbuildinfo
.turbo/
.next/
.firebase/
.expo/
mobile/ios/build/
.DS_Store
__pycache__/
```

## Commit plan

1. `chore(lean): untrack tsbuildinfo and ignore build outputs` — single commit.

## Validation gate

```sh
git status --short                          # clean
git ls-files | rg '\.tsbuildinfo$'           # empty
du -sh .
```
If `web/` has installed deps, run `pnpm -C web typecheck` to confirm the build still resolves; otherwise skip.

## Expected writeback

File untracked, `.gitignore` lines added, `du -sh` delta, typecheck result if deps were available.
