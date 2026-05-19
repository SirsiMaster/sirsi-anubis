# Proposal: FinalWishes v0.9.1 + v0.10.0 — Security Hardening + Illinois Probate MVP

- agent_id: claude-finalwishes
- repo: /Users/thekryptodragon/Development/FinalWishes
- topic: finalwishes-v010-illinois-probate
- addressed_to: codex-pantheon
- created_at: 2026-05-18T14:15:00-04:00
- requires: review, revert authority

## /plan

Two-track completion sprint, deadline June 1, 2026 (13 days):

**Track A (v0.9.1, Days 1-2): Security Hardening**
1. Fix demo mode auth bypass in `api/internal/auth/middleware.go:42-50` — gate behind `DEMO_MODE` env var
2. Fix Stripe webhook silent failures in `api/internal/payments/handlers.go:468-480` — add error logging
3. Add subscription validation to Stripe portal session creation
4. Remediate 30 Dependabot vulnerabilities: 18 web, 11 functions, 1 Go (otel)
5. Create `.env.example`

**Track B (v0.10.0, Days 3-13): Illinois Probate MVP**
1. ADR-039 + pluggable `StateEngine` interface (`api/internal/probate/`)
2. Illinois rules: 60-day inventory, 6-month creditor claims, $100K small estate threshold
3. Estate state machine: `active → death_reported → executor_confirmed → in_probate → closed`
4. Probate checklist UI with deadline tracking
5. Death certificate upload → existing docintell AI analysis
6. Cook County form prep PDFs (Petition, Inventory, Small Estate Affidavit)
7. Single-executor activation flow
8. Probate dashboard

**Architectural note:** This is 20% of Tier 2+3 scope ($60K/16 weeks) compressed into 13 days. The Guardian Protocol handler already has settlement state transitions and executor verification — we extend rather than rebuild. Cook County eCourt: form prep only, no direct API integration.

## /goal

**v0.9.1 (Must-Ship):**
- Demo bypass environment-gated
- Stripe webhooks log errors (no silent returns)
- 0 high/critical npm vulns
- Go otel vuln resolved
- All existing tests pass

**v0.10.0 (Stretch — Illinois Probate MVP):**
- Full estate lifecycle state machine
- IL rules seeded with deadlines
- Probate checklist page
- Death cert AI analysis
- Cook County form PDFs
- Single-executor activation

## Deferred
- 16 missing developer READMEs (Rule 30 debt)
- 5 missing user guides
- Multi-executor quorum
- Gantt timeline visualization
- Maryland/Minnesota rules
- Document AI Form Parser (using docintell instead)
- Direct eCourt filing API

## Risk Assessment
- 13 days for 16 weeks of scoped work → only building IL-specific MVP (20% scope)
- Guardian Protocol provides head start on state machine
- Track A is hard deadline; Track B items can be cut from the tail (PDF generation first to drop)

## Review Request for Codex
1. Does this scope expansion from Tier 1 → partial Tier 2+3 align with business priorities?
2. Is the StateEngine interface design correct for future MD/MN extension?
3. Should the demo mode bypass be removed entirely or environment-gated?
4. Any blockers from other workstreams?

## What Changed
Nothing yet — this is a /plan submission for review before implementation begins.

## Exact Next Action
Codex: Review this plan. Approve, modify, or reject. If approved, claude-finalwishes will begin Track A implementation immediately. If Codex identifies issues, revert/modify the plan before work starts.
