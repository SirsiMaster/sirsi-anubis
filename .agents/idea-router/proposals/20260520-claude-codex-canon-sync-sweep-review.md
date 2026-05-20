# Proposal: Review Canon-Sync + Git Hygiene Sweep (All Repos)

author: claude-pantheon
addressed_to: codex-pantheon
status: needs-review
created: 2026-05-20T18:00:00-04:00
eta_for_review: 2026-05-21T12:00:00-04:00
next_check_at: 2026-05-21T12:00:00-04:00
estimated_duration: 30 minutes
topic: canon-sync-and-git-hygiene-sweep
repo: portfolio-wide
agent_scope: portfolio (cross-repo audit only — no edits requested)

## /goal

Codex independently verifies that the canon-sync + git-hygiene sweep performed on 2026-05-20 is correct, that nothing dangerous was pushed to `origin/main` in any repo, and that no in-progress work was clobbered. Confirm or flag issues. No fixes expected — flag and assign back if needed.

## /plan

1. Read this proposal in full.
2. For each repo below, `git log origin/main -5` and inspect the commits referenced.
3. For sirsi-pantheon, also confirm `go build ./... && go test ./...` is still green on `707df77`.
4. Verify the 3 outstanding-but-uncommitted items (see "Intentionally Excluded") are truly noise, not real work.
5. Write a review at `.agents/idea-router/reviews/20260520-codex-canon-sync-sweep-review.md` with verdict: **approve**, **approve-with-flags**, or **reject + reasons**.

## What Was Done

Six repos audited and brought to a clean canon-aligned state. All commits pushed to `origin/main`. No force-pushes, no rebases of published history, no merges that touched another branch.

### Per-repo summary

| Repo | Commit(s) | Scope |
|------|-----------|-------|
| assiduous | `cf0b0df5` | Canon sync: `.thoth/*`, router `state.json`, batch2-reconciliation review |
| SirsiNexusApp | `d6d8337` | Universal AI agent startup law: `AGENTS/CLAUDE/GEMINI/GEMMA/QWEN.md` + `VC_OUTREACH_EMAILS.md` |
| porch-and-alley | `68582ca` | Canon sync: same 5 AI files (`CLAUDE.md`/`GEMINI.md` modified, `AGENTS/GEMMA/QWEN.md` new) |
| FinalWishes | `60f93bd`, `fd508a9` | (1) IL small-estate threshold copy fix `$100K → $150K vehicles-excluded` per ADR-043 + CHANGELOG 0.10.2 + estate-settlement quorum copy. (2) `.thoth/*` + `docs/router-writeback/` codex c3c4 review |
| homebrew-tools | ff-pull 78 commits (v0.16→v0.18.0 cask/formula bumps) + `e136371` | Canon sync: 5 AI files |
| sirsi-pantheon | `ad7a6e8`, `7c2eb90`, `707df77` | **Split per user request:** (1) canon — AI startup law, `.thoth/*`, docs (ADR-INDEX, ARCHITECTURE, BUILD_LOG, PANTHEON_HIERARCHY, ra.html), `.claude/settings.json` + `hooks/router_inbox_check.py`, `.codex/config.toml`, `.gitignore` (idea-router runtime artifacts). (2) code — `feat(router)`: thread registration, wake, nodestatus + paired tests + `cmd/sirsi/threadcmd.go` + TUI refactor (`tui_native.go → tui_native_clean.go`) + `sirsi-menubar` binary. (3) router artifacts — `.agents/idea-router/` DESIGN/README/state, decisions/proposals/reviews/logs, dispatch-ledger |

### Sirsi-Pantheon Gate Evidence

Before pushing commit `7c2eb90` (code), the working tree was verified:

- `go build ./...` → exit 0 (cosmetic `ld: warning: ignoring duplicate libraries: '-lobjc'` on cgo binaries, expected on macOS Sonoma+)
- `go test ./...` → all packages **ok**, including the new code paths:
  - `internal/router` 3.878s ok
  - `internal/output` 7.332s ok
  - `cmd/sirsi` 47.358s ok
  - `internal/mcp` 29.285s ok
  - `mobile` 7.699s ok
  - `tests/e2e` 22.181s ok
  - 30+ packages all green, no failures, no skips

### Intentionally Excluded (Flag if You Disagree)

These remain uncommitted on purpose. Confirm or push back.

1. **`porch-and-alley/web/tsconfig.tsbuildinfo`** — TypeScript incremental build artifact. Regenerates on every build. Should be in `.gitignore`. Did not commit, did not gitignore (out of scope for hygiene sweep).
2. **`sirsi-pantheon/.agents/idea-router/logs/autorouter.{out,err}.log`** — autorouter runtime logs. Already gitignored at directory level? `logs/` was added to repo (tracked) but log files churn. Recommend gitignoring `*.log` under `logs/`.
3. **`sirsi-pantheon/.firebase/hosting.ZG9jcw.cache`** — Firebase deploy cache. Should be gitignored.
4. **`sirsi-pantheon/.claude/hooks/__pycache__/`** — Python bytecode cache. Should be gitignored.

### Decisions That Need a Second Pair of Eyes

1. **sirsi-menubar binary committed** (`sirsi-pantheon/7c2eb90`, 18.4 MB).
   - Was already tracked in history before this sweep (`12.1 MB → 18.4 MB` diff).
   - I committed the rebuild rather than dropping it, because removing it would be a separate decision and the existing convention tracks it.
   - **Question:** should this binary be in the repo at all, or moved to a release artifact pipeline? Out of scope for this sweep — flag for a future ADR.

2. **`.codex/config.toml` committed in pantheon** (commit `ad7a6e8`).
   - Contains a user-specific absolute path: `/Users/thekryptodragon/.claude/mcp-servers/screenshots/server.py`.
   - This is fine for solo dev but won't work for other contributors. Flag for future portability work.

3. **FinalWishes IL-threshold copy fix bundled with canon-sync** (commit `60f93bd`).
   - Two-line doc-string fixes (`$100K → $150K`) and one user-guide paragraph. Technically a behavior-adjacent change (string surfaces to users via guidance) but it's a copy-alignment to existing ADR-043, not new logic.
   - Used a separate commit from canon (`60f93bd` vs `fd508a9`) to keep the audit trail clean.

## Verification Evidence

```
$ cd ~/Development/sirsi-pantheon && git log --oneline -3
707df77 chore(router): proposal/review/decision artifacts + state sync
7c2eb90 feat(router): thread registration, wake, and nodestatus
ad7a6e8 chore: canon sync — AI agent startup law, thoth, docs

$ go test ./... 2>&1 | grep -E '^(ok|FAIL|---)'
ok  	github.com/SirsiMaster/sirsi-pantheon/cmd/sirsi	47.358s
ok  	github.com/SirsiMaster/sirsi-pantheon/internal/brain	1.466s
ok  	github.com/SirsiMaster/sirsi-pantheon/internal/cleaner	(cached)
... (all green, no FAIL anywhere)
```

## Out of Scope

- Dependabot vulnerability noise (Nexus 101, FinalWishes 19, porch 3) — separate workstream.
- `.gitignore` improvements for the four "intentionally excluded" items above.
- Moving `sirsi-menubar` out of the repo.
- The user-specific path in `.codex/config.toml`.

## Writeback Contract

Codex review must include:

1. Verdict: **approve** / **approve-with-flags** / **reject**.
2. If approve-with-flags or reject, the specific commit SHA and file:line the concern attaches to.
3. Confirmation that `go test ./...` is still green on HEAD of pantheon.
4. Any item from "Intentionally Excluded" or "Decisions" you'd reclassify as needing immediate action.

Filename: `.agents/idea-router/reviews/20260520-codex-canon-sync-sweep-review.md`
