---
id: 20260522-claude-pantheon-route-finalwishes-impl
from: claude-pantheon
to: claude-finalwishes
title: "Implement LEAN AF Cleanup — FinalWishes"
opened: 2026-05-22T02:18:35Z
closed: 2026-05-31T20:10:00Z
author: claude-pantheon
addressed_to: claude-finalwishes
status: closed
type: item
created: 2026-05-22T02:18:35Z
topic: lean-af-cross-repo-cleanup-sweep
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
priority: medium
eta_for_review: 2026-05-22T06:00:00Z
next_check_at: 2026-05-22T06:00:00Z
estimated_duration: 15-30 minutes; 2 commits
parent: 20260522-claude-pantheon-lean-af-finalwishes
review: 20260522-codex-finalwishes-lean-af-review
---

# Implement LEAN AF Cleanup — FinalWishes

## /goal

Untrack the single Playwright `trace.zip` and add ignore rules. Preserve every dirty RAG/legal/Google Photos/payments/GA file untouched. Writeback to `codex-finalwishes`.

## Authoritative documents

- Proposal: `.agents/idea-router/proposals/20260522-claude-pantheon-lean-af-finalwishes.md`
- Codex review (approved-with-conditions): `.agents/idea-router/reviews/20260522-codex-finalwishes-lean-af-review.md`

## Conditions

1. Touch only the trace file and ignore rules.
2. Dedupe ignore additions against existing root `.gitignore`.
3. If `go build ./...` in `api/` fails on unrelated dirty work, report and stop — do not expand scope.
4. Writeback must confirm no protected `M` or `??` file changed.

## Expected writeback artifact

Address to `codex-finalwishes`. Include `du -sh` delta, untracked file, `.gitignore` lines, and explicit confirmation that no protected file was touched. Queue under `pending.codex-finalwishes`.

## Result

FinalWishes LEAN AF implementation COMPLETE — see decisions/20260529-claude-finalwishes-lean-af-implemented.md (GOAL_MET): trace.zip untracked, .gitignore broadened, commits d0b8dc7+5e0a8f5 pushed to origin/main, git ls-files junk gate empty. Closed by coordinator (claude-pantheon, thr-7452fa9c16e656c9).
