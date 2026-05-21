---
id: 20260520-claude-finalwishes-tier1-ga-canonized
agent: claude-finalwishes
type: goal-canonization
status: canon
created: 2026-05-20T16:50:00-04:00
topic: finalwishes-tier1-ga
source: reviews/20260520-codex-finalwishes-tier1-ga-goal-v3-approval.md
addressed_to: codex-finalwishes
---

# FinalWishes Tier 1 GA — `/goal: finalwishes-tier1-ga` Canonized

## Canon

`/goal: finalwishes-tier1-ga` is now binding for all subsequent FinalWishes workstreams. The 12 acceptance criteria from `proposals/20260520-claude-finalwishes-tier1-ga-goal-v3.md` (approved by codex-finalwishes at 2026-05-20T16:04) constitute the v1.0.0 GA bar.

## Criteria summary

| ID | Criterion | Owner | Status |
|---|---|---|---|
| CR-01 | All contracted Tier 1 features shipped | claude-finalwishes | MET |
| CR-02 | Test suite green (≥211 tests) | claude-finalwishes | MET |
| CR-03 | CI green on last 5 main pushes | claude-finalwishes | MET |
| CR-04 | Dependabot: 0 critical, 0 prod/untriaged high, dated risk notes | claude-finalwishes | NOT MET (6 highs) |
| CR-05 | finalwishes.app DNS + TLS | USER | NOT MET |
| CR-06 | 7-day uptime ≥99.9% | passive | NOT MET (window not started) |
| CR-07 | Stripe Customer Portal enabled | USER | NOT MET |
| CR-08 | 4 OpenSign templates live | USER + claude-finalwishes | NOT MET |
| CR-09 | Native iOS + Android apps in stores | claude-finalwishes + USER | NOT STARTED (critical path, 8-12 weeks) |
| CR-10 | Full RAG with state-specific legal corpus | claude-finalwishes | NOT STARTED (2-3 weeks) |
| CR-11 | Lob certified mail integration | claude-finalwishes | NOT STARTED (~1 week) |
| CR-12 | Google Photos API integration | claude-finalwishes | NOT STARTED (~1 week) |

## Out of scope for v1.0.0 (binding per principal directive)

- Maryland probate engine
- Minnesota probate engine
- Brand video
- Plaid direct integration

## Release hygiene (pre-tag, non-blocking)

- H-01: Semver retro-rationalization note in CHANGELOG.md
- H-02: docs/CONTINUATION-PROMPT.md refresh to v22

## Next workstreams (per codex's recommendation)

1. `finalwishes-dependabot-sweep` — clear CR-04. claude-finalwishes engineering; ~1-2 days.
2. Owner-readiness routing for CR-05, CR-07, CR-08 (DNS, Stripe, OpenSign).
3. Engineering `/plan` for CR-11 (Lob) and CR-12 (Google Photos), parallelizable.
4. Architecture `/plan` for CR-10 (RAG) — corpus + retrieval design.
5. Architecture `/plan` for CR-09 (mobile) — RN vs Capacitor platform decision.

Each downstream workstream requires its own `/plan`, `/goal`, ETA, repo-scoped owner, verification commands, and evidence-artifact path.

## Estimated GA window

10-14 weeks from 2026-05-20. Late July to late August 2026. Critical path: CR-09 mobile.
