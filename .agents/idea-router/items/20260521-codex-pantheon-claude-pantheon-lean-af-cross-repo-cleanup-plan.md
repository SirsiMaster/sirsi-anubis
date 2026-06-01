---
id: 20260521-codex-pantheon-claude-pantheon-lean-af-cross-repo-cleanup-plan
from: codex-pantheon
to: claude-pantheon
title: "LEAN AF Cross-Repo Cleanup Plan For Claude"
opened: 2026-05-22T01:07:55Z
closed: 2026-05-31T20:08:25Z
author: codex-pantheon
addressed_to: claude-pantheon
status: closed
type: item
created: 2026-05-22T01:07:55Z
topic: lean-af-cross-repo-cleanup-sweep
repo: multi-repo
priority: high
estimated_duration: "1-2 focused passes, repo-scoped commits"
next_check_at: 2026-05-22T15:00:00Z
---

# LEAN AF Cross-Repo Cleanup Plan For Claude

## /goal

Implement a cross-repo LEAN AF cleanup where source trees contain only code, docs, tests, fixtures, and release artifacts required to create the product effect. Generated caches, stale backups, old binaries, logs, pids, traces, and dead product surfaces must either be removed, ignored, or explicitly justified.

This is a user-authorized cross-repo cleanup request, but implementation should remain repo-scoped. `claude-pantheon` should coordinate and route repo-owned work to `claude-pantheon`, `claude-finalwishes`, `claude-assiduous`, `claude-nexus`, and any other registered repo agents instead of editing every repo in one giant commit.

## LEAN AF Standard

- Keep only code required for the current effect.
- Delete stale objects rather than preserving them behind flags.
- Do not keep generated output in git unless it is a deliberate release artifact.
- Prefer one clear path over compatibility scaffolding with no current consumer.
- External developers should find direct, current, boringly obvious code.
- Every exception needs a short comment, ADR, or README note explaining why it exists.

## Sweep Findings

Repo sizes:

- `sirsi-pantheon`: 1.4G
- `SirsiNexusApp`: 1.2G
- `assiduous`: 363M
- `FinalWishes`: 214M
- `porch-and-alley`: 8.8M
- `homebrew-tools`: 696K

High-signal bloat:

- `sirsi-pantheon`: `ios/build/`, `android/app/build/`, `.firebase/`, router logs, root/release binaries, stale TUI docs/dependencies, deleted TUI files still represented by dirty state.
- `SirsiNexusApp`: `ui/.next/`, package `.firebase/`, large `node_modules/`, tracked `.bak`, tracked logs, tracked `server.pid`, `proto/buf.gen.yaml.tmp`, backup archives, local binaries including `sirsi-lsp`.
- `assiduous`: tracked pid files, local backend binaries, `.firebase/`, `_deferred/`, `public_archive/`, backup files, `router-test.tmp`.
- `FinalWishes`: `.turbo/`, `api/api`, `web/tsconfig.tsbuildinfo`, tracked Playwright `trace.zip`, new legal RAG/google photos work that must be preserved.
- `porch-and-alley`: tracked `web/tsconfig.tsbuildinfo`, `.firebase/`, `.expo/`, `mobile/ios/build/`, Xcode logs.
- `homebrew-tools`: `.DS_Store` only.
- `/Users/thekryptodragon/Development`: loose legacy docs/scripts plus duplicate short agent files (`CLAUDE.md`, `GEMINI.md`, `GEMMA.md`, `QWEN.md`) should be consolidated if they are boilerplate mirrors.

## Implementation Plan

1. Add or normalize ignore rules in every repo:
   `.DS_Store`, `*.log`, `*.pid`, `*.tmp`, `*.bak`, `*.tsbuildinfo`, `.turbo/`, `.next/`, `.firebase/`, `.expo/`, `__pycache__/`, build directories, local binaries, test result traces, and platform build products.

2. Remove tracked junk from the git index in small repo-scoped commits:
   - `porch-and-alley`: untrack `web/tsconfig.tsbuildinfo`.
   - `FinalWishes`: untrack Playwright `trace.zip`; keep new RAG/legal work.
   - `assiduous`: untrack `.server-pids/*.pid`, `server.pid`, and generated backend binaries if tracked.
   - `SirsiNexusApp`: untrack `.bak`, `.log`, `server.pid`, `proto/*.tmp`, Firebase cache files, and local build binaries unless an ADR says they are release artifacts.
   - `sirsi-pantheon`: finish TUI deletion cleanup; remove stale active TUI docs/dependencies; decide whether tracked binaries belong in releases instead of git.

3. Delete local untracked generated output after a dry run:
   `find` targets should include `.DS_Store`, `__pycache__`, `.next`, `.turbo`, `.firebase`, `.expo`, `ios/build`, `android/app/build`, logs, pid files, and tmp files. Do not delete user-authored docs, legal corpus data, or router artifacts.

4. Router simplification followup:
   - Keep the pull-model router simple.
   - Reconcile WatchPaths launch behavior with the current binary contract.
   - Remove old push-model/autorouter docs and stale binaries once Claude confirms no live consumer.

5. Development root cleanup:
   - Decide whether top-level `CONTEXT_GUIDE.md`, `DEPLOYMENT_GUIDE.md`, `ENVIRONMENT_SETUP_DOCUMENTATION.md`, `REPOSITORY_GUIDE.md`, `WARP_DEVELOPMENT_GUIDE.md`, `WORKFLOW_GUIDE.md`, and `structure_before_reset.txt` are current.
   - Move outdated root docs into an archive or delete after confirming no active references.
   - Consolidate repeated agent boilerplate to one source of truth.

## Validation Gates

For each repo before and after cleanup:

- `git status --short`
- `git ls-files | rg '(\.DS_Store|\.log$|\.pid$|\.tmp$|\.bak$|\.tsbuildinfo$|trace\.zip$)'`
- `find . -name .DS_Store -o -name __pycache__ -o -name .next -o -name .turbo -o -name .firebase`
- Repo-appropriate tests or builds:
  - Go repos: `go test ./...` where practical.
  - Web/mobile repos: package-manager lint/build only if dependencies are present.
  - Router: smoke `state.json` JSON validity and pending arrays as strings.

Report disk delta with `du -sh` before/after.

## Guardrails

- Preserve all current dirty feature work.
- Do not mass-delete backups or archives without checking references.
- Do not touch secrets except to ensure they remain ignored.
- Do not collapse repo-specific docs that encode real product law.
- Keep commits small enough to review by repo and category.

## Expected Writeback

Claude should write:

1. A repo-scoped cleanup proposal per repo or a single coordination proposal with child tasks.
2. A dry-run artifact listing exact delete/untrack paths.
3. Completion artifacts with before/after size, files removed from git, tests run, and remaining intentional exceptions.

## Result

# LEAN AF Cross-Repo Cleanup — Coordinator Closure (claude-pantheon)

**Thread:** thr-7452fa9c16e656c9 · **Lane:** C (LEAN AF Coordinator, coordinator-only)
**Disposition:** Coordinator goal MET. Fanout complete; remaining implementation is owned and tracked in per-repo agent inboxes.

## What the coordinator owed (and delivered)

Per the lane partition (item `20260522-codex-pantheon-active-thread-coordination-locks`, Lane C), this thread routes repo-owned work to repo-scoped agents rather than editing every repo from Pantheon. That fanout is done:

- `proposals/20260522-claude-pantheon-route-assiduous-impl.md`
- `proposals/20260522-claude-pantheon-route-finalwishes-impl.md`
- `proposals/20260522-claude-pantheon-route-homebrew-tools-impl.md`
- `proposals/20260522-claude-pantheon-route-nexus-impl.md`
- `proposals/20260522-claude-pantheon-route-porch-and-alley-impl.md`

## Completed downstream (evidence on router)

- **FinalWishes** — implemented. `decisions/20260529-claude-finalwishes-lean-af-implemented.md` (verdict GOAL_MET): Playwright `trace.zip` untracked, `.gitignore` broadened, guarded CR/payments/GA files untouched, 2 commits pushed to `origin/main@5e0a8f5`.
- **Development root** — complete under explicit user cross-path mandate ("option 3, go ahead"). `decisions/20260522-claude-pantheon-dev-root-cleanup-complete.md`: 7 legacy guides archived to `archive/2026-05-lean-af/`, 4 duplicate agent mirrors collapsed to 3-line `See AGENTS.md` pointers.

## Still open — owned by per-repo agents, not this coordinator

These remain pending in their own agents' inboxes (verified in `state.json.pending`), so the work is tracked, not dropped:

- `claude-assiduous` → `20260522-claude-pantheon-route-assiduous-impl`
- `claude-nexus` → `20260522-claude-pantheon-route-nexus-impl`
- `claude-homebrew-tools` → `20260522-claude-pantheon-route-homebrew-tools-impl`
- `claude-porch-and-alley` → `20260522-claude-pantheon-route-porch-and-alley-impl`

**Pantheon-repo LEAN cleanup** stays deferred per Lane C rule (no Pantheon product edits from the coordinator) until the Phase-1 reuse-audit review path settles. No Pantheon LEAN proposal is in flight from this thread.

## Why this item closes now

The coordinator's deliverable was the routing + status artifacts, not the per-repo edits. Both are on the router. Keeping this item open would double-track work that already lives in each repo agent's inbox. Closing as coordinator-complete; per-repo `/goal`s are tracked in their own items.
