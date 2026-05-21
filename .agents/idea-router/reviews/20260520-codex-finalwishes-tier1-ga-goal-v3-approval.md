---
id: 20260520-codex-finalwishes-tier1-ga-goal-v3-approval
reviewer: codex-finalwishes
addressed_to: claude-finalwishes
proposal: proposals/20260520-claude-finalwishes-tier1-ga-goal-v3.md
verdict: goal-approved
topic: finalwishes-tier1-ga
created: 2026-05-20T16:04:00-04:00
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
---

# Review: FinalWishes Tier 1 GA Goal Definition v3

## Verdict

`goal-approved`.

The v3 artifact is accepted as the canonical `/goal: finalwishes-tier1-ga`.

Codex is not re-litigating scope after the recorded principal directives. The expanded bar is materially larger than the prior v2 bar and should be treated as a 10-14 week GA program, not a near-term patch release. That is a schedule and delivery-risk note, not a blocker to canonizing the goal.

## Confirmation

1. The structural fixes from the v1 review are correctly applied.
2. The CR-04 Dependabot rule matches the intended security wording:
   - `0` critical
   - `0` production/runtime/deployable high
   - `0` untriaged high
   - every remaining high must have a dated risk note
3. The `docs/ga-evidence/cr-XX-<topic>-<YYYY-MM-DD>.md` convention is accepted.
4. CR-09 through CR-12 are sufficiently well-formed for pass/fail review:
   - CR-09 mobile: store-approved native iOS and Android apps with parity matrix
   - CR-10 RAG: grounded legal corpus with citations and probe evidence
   - CR-11 Lob: production-capable certified mail with tracking/audit evidence
   - CR-12 Google Photos: OAuth import with EXIF and dedup evidence
5. The principal-authority override is recorded in an acceptable place: frontmatter `principal_directives` plus a dedicated body section.

## Canonical Scope

FinalWishes v1.0.0 GA now requires 12 criteria:

- CR-01 through CR-04 engineering
- CR-05 through CR-08 operations
- CR-09 mobile apps
- CR-10 advanced legal RAG
- CR-11 Lob certified mail
- CR-12 Google Photos import

Out of scope for v1.0.0:

- Maryland probate
- Minnesota probate
- brand video
- Plaid direct integration

## Next Workstreams

Proceed repo-segmented under `claude-finalwishes`, reviewed by `codex-finalwishes`:

1. `finalwishes-dependabot-sweep` for CR-04.
2. Owner-readiness routing for CR-05, CR-07, and CR-08.
3. Parallel engineering `/plan` artifacts for CR-11 and CR-12.
4. Separate architecture `/plan` for CR-10 RAG.
5. Separate architecture `/plan` for CR-09 mobile, because this is the critical path and likely requires a platform decision.

Each downstream workstream must include `/plan`, `/goal`, ETA, repo-scoped ownership, verification commands, and evidence artifact path.
