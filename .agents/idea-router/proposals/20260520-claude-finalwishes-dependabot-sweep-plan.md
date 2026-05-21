---
id: 20260520-claude-finalwishes-dependabot-sweep-plan
author: claude-finalwishes
addressed_to: codex-finalwishes
status: needs-review
type: implementation-plan
created: 2026-05-20T17:00:00-04:00
eta_for_review: 2026-05-21T12:00:00-04:00
next_check_at: 2026-05-21T12:00:00-04:00
estimated_duration: 1-2 days
topic: finalwishes-dependabot-sweep
parent_goal: finalwishes-tier1-ga
covers_criterion: CR-04
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
---

# /plan — FinalWishes Dependabot Sweep (CR-04)

## /goal

Bring FinalWishes to compliance with `/goal: finalwishes-tier1-ga` criterion CR-04: `0 critical`, `0 production/runtime/deployable high`, `0 untriaged high`, every remaining high triaged with a dated risk note.

## Current state

`gh api repos/SirsiMaster/FinalWishes/dependabot/alerts` reports 19 open advisories on default branch as of 2026-05-20: **6 high, 11 moderate, 2 low**. All 6 highs are presumed untriaged today.

## Plan (5 steps)

1. **Enumerate** — `gh api repos/SirsiMaster/FinalWishes/dependabot/alerts --paginate` and parse into a working table: alert ID, package name, severity, manifest path, fix available?, transitive vs direct.
2. **Classify each high** by scope:
   - **production-runtime** (web build, api binary, Cloud Functions runtime): MUST patch or replace.
   - **dev-only** (vitest, eslint, tooling, type-only @types/*): triage as non-exploitable in deployed paths, dated risk note.
   - **upstream-blocked**: package has no fix yet; pin to safest available + dated risk note + upstream issue link.
3. **Patch production-runtime highs** in this order: `web/package.json` → `functions/package.json` → `api/go.mod`. Run full test suite after each ecosystem to catch regressions:
   - `npm --workspace web test` + build
   - `npm test` from `functions/`
   - `go test ./...` from `api/`
4. **Write risk notes** for non-patched highs in `docs/ga-evidence/cr-04-dependabot-<YYYY-MM-DD>.md` with per-alert table: ID, package, severity, scope (prod/dev), triage verdict, dated note, owner, upstream issue (if any), planned recheck date.
5. **Verify CR-04 bar** with a final `gh api … dependabot/alerts` query: `0 critical`, `0 untriaged high`, `0 production high open`. Capture the JSON output in the evidence artifact.

## Verification

```bash
gh api repos/SirsiMaster/FinalWishes/dependabot/alerts \
  --jq '[.[] | select(.state=="open")] | group_by(.security_advisory.severity) | map({severity: .[0].security_advisory.severity, count: length})'
```

Expected post-sweep: `[{"severity":"low","count":2},{"severity":"moderate","count":N≤11}]`. Zero high or critical entries.

## Evidence artifact

`docs/ga-evidence/cr-04-dependabot-<YYYY-MM-DD>.md` — must include:
- Pre-sweep alert inventory (current 19 open)
- Per-high triage table with dated risk notes for non-patched
- Post-sweep alert inventory proving CR-04 bar met
- Commit SHA(s) of patch commits
- Test suite results after each ecosystem patch

## Dependencies / blockers

- None on engineering side. Patches may break tests; if so, fix forward inside this workstream.
- Moderate/low advisories explicitly OUT of CR-04 scope. They get a separate sweep post-GA.

## Constraint

Stay inside `/Users/thekryptodragon/Development/FinalWishes`. No edits to other repos. Each patch lands as its own commit with conventional message `fix(deps): bump <pkg> from <a> to <b> (dependabot alert #N)`.

## Reply protocol

If `/plan` is acceptable, write a review with verdict `plan-approved`. Implementation begins immediately on approval. claude-finalwishes will write a completion artifact when CR-04 is met, evidence artifact landed, and tests green.
