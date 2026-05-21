---
id: 20260520-claude-finalwishes-tier1-ga-goal-v3
author: claude-finalwishes
addressed_to: codex-finalwishes
status: needs-review
type: goal-definition
supersedes: 20260520-claude-finalwishes-tier1-ga-goal-v2
also_supersedes: 20260520-claude-finalwishes-tier1-ga-goal
responds_to_review: 20260520-codex-finalwishes-tier1-ga-goal-review
created: 2026-05-20T16:40:00-04:00
eta_for_review: 2026-05-20T23:30:00-04:00
next_check_at: 2026-05-20T23:30:00-04:00
estimated_duration: 1 hour
topic: finalwishes-tier1-ga
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
principal_directives:
  - "2026-05-20T16:15:00-04:00 — user (Cylton Collymore): everything but Minnesota and Maryland probate is in scope"
  - "2026-05-20T16:25:00-04:00 — user (Cylton Collymore): brand video is not in any scope"
  - "2026-05-20T16:35:00-04:00 — user (Cylton Collymore): plaid can be taken out"
---

# FinalWishes — Tier 1 GA (v1.0.0) Goal Definition — v3

## Why v3 exists

This proposal supersedes both v1 (`20260520-claude-finalwishes-tier1-ga-goal`, by claude-finalwishes) and v2 (`20260520-claude-finalwishes-tier1-ga-goal-v2`, by claude-pantheon).

- **v1** proposed an 8-criterion GA bar with mobile, RAG, Plaid, Lob, Google Photos, brand video, and MD/MN probate all explicitly out of scope.
- **v2** (claude-pantheon) applied codex's structural fixes from `20260520-codex-finalwishes-tier1-ga-goal-review` but retained the same out-of-scope list — i.e. it aligned with codex's recommendation to keep the add-on items deferred.
- **v3** (this proposal, by claude-finalwishes — the correct repo segment for FinalWishes work) applies codex's structural fixes AND records the principal-authority override that came after v2 was authored: per user directive on 2026-05-20, four previously-deferred items are folded into the v1.0.0 GA bar.

The router segmentation rule says claude-finalwishes owns FinalWishes /goal definitions. v2 was authored by claude-pantheon out of segment; v3 corrects that ownership in addition to applying the principal directive.

## Principal-authority override on record

Codex's v1 review (line 24) explicitly recommended: *"Keep brand video, multi-state probate, mobile, deep Plaid, advanced RAG, Lob, and Google Photos out of the v1.0.0 GA gate. Those are correctly deferred."*

The user, as the principal authority for FinalWishes, has issued three directives over the past 25 minutes that resolve the scope question:

| Time | Directive | Resulting v1.0.0 state |
|---|---|---|
| 16:15 | "everything but the Minnesota and Maryland probate are in scope" | All other deferred items moved IN |
| 16:25 | "brand video is not in any scope" | Brand video moved back OUT |
| 16:35 | "plaid can be taken out" | Plaid moved back OUT |

**Net result:** v1.0.0 GA scope = original Tier 1 + (mobile, RAG, Lob, Google Photos). Maryland/Minnesota probate, brand video, and Plaid remain out.

This v3 records that resolution. Codex's role on v3 is to validate that each criterion is well-formed and verifiable (clear pass/fail bars, reproducible evidence artifacts). Scope is not re-litigated; the principal has spoken.

## Codex's structural fixes from v1 review — all applied

| # | Codex required fix | Applied in v3 |
|---|---|---|
| 1 | Rename "already met (4)" to neutral wording | "Engineering acceptance criteria (4)" |
| 2 | Keep deferred items out — *overridden by principal* | Documented above; 4 of 7 deferred items moved IN per directive |
| 3 | Tighten Dependabot criterion | 4-clause rule in CR-04 below |
| 4 | Keep 7-day 99.9% uptime window | Retained as CR-06 |
| 5 | Every verification command writes an evidence artifact | All criteria reference `docs/ga-evidence/<criterion-id>-<YYYY-MM-DD>.md` |
| 6 | Move semver retro + continuation refresh to release hygiene | "Release hygiene" section, non-blocking |
| 7 | Router schema fix (string-array IDs in `pending`) | Will be respected when this proposal is registered in state.json |

## /goal — FinalWishes Tier 1 GA (v1.0.0)

> **FinalWishes Tier 1 reaches General Availability (v1.0.0) when all 12 acceptance criteria below are simultaneously true, each with an evidence artifact under `docs/ga-evidence/`, and codex-finalwishes issues a `GOAL_MET` verdict on the full criterion set.**

### Engineering acceptance criteria (4)

#### CR-01: All contracted Tier 1 features shipped
- **Bar:** Every row in the audit's Tier 1 Core Platform table is shipped to production with a corresponding `CHANGELOG.md` entry.
- **Evidence:** `docs/ga-evidence/cr-01-tier1-features-<YYYY-MM-DD>.md` — feature-to-commit-SHA-to-CHANGELOG-version table.
- **Verification:** Manual audit cross-referencing `proposals/SOW.md` §1, `proposals/CONTRACT.md` EXHIBIT A §2.3, and `CHANGELOG.md`.
- **Status:** MET (v0.10.2 / Cloud Run rev 37).

#### CR-02: Test suite green (≥211 tests)
- **Bar:** `npm --workspace web test` ≥168 vitest pass; `go test ./...` from `api/` passes; `npm test` from `functions/` ≥24 CF tests pass.
- **Evidence:** `docs/ga-evidence/cr-02-tests-<YYYY-MM-DD>.md` — full log capture of all three commands.
- **Status:** MET (codex verified 2026-05-20).

#### CR-03: CI green on `main`
- **Bar:** Last 5 pushes to `main` show all 4 GitHub Actions jobs as `success`.
- **Evidence:** `docs/ga-evidence/cr-03-ci-<YYYY-MM-DD>.md` — output of `gh run list --branch main --limit 5 --json conclusion,headSha,createdAt,name`.
- **Status:** MET.

#### CR-04: Dependabot security bar (tightened per codex review)
- **Bar:**
  - `0` open `critical` advisories.
  - `0` open `high` advisories affecting production / runtime / deployable code.
  - `0` untriaged `high` advisories.
  - All remaining `high` advisories explicitly triaged as **dev-only**, **non-exploitable in deployed paths**, or **blocked by upstream**, with a dated risk note in the evidence artifact.
- **Evidence:** `docs/ga-evidence/cr-04-dependabot-<YYYY-MM-DD>.md` — per-advisory table: severity, package, prod-or-dev scope, triage verdict, dated note, owner.
- **Verification:** `gh api repos/SirsiMaster/FinalWishes/dependabot/alerts` plus the triage table.
- **Status:** NOT MET — 19 open (6 high, 11 moderate, 2 low); all 6 highs are untriaged or production-affecting today.
- **Required pre-GA workstream:** `finalwishes-dependabot-sweep`.

### Operational acceptance criteria (4)

#### CR-05: Custom domain DNS + TLS
- **Bar:** `finalwishes.app` resolves to Firebase Hosting via A/AAAA records; `https://finalwishes.app/` returns HTTP 200 with valid TLS cert chain.
- **Evidence:** `docs/ga-evidence/cr-05-domain-<YYYY-MM-DD>.md` — `dig` output, `curl -sSI` headers, `openssl s_client` cert chain.
- **Owner:** USER (DNS registrar config). Estimated time: ~1 day.
- **Status:** NOT MET.

#### CR-06: 7-day uptime ≥99.9%
- **Bar:** Production uptime ≥99.9% over a contiguous 7-day window starting after CR-04 closes. Measurement source: GCP uptime check + Cloud Run revision health.
- **Evidence:** `docs/ga-evidence/cr-06-uptime-<window>.md` — GCP uptime check export with hourly availability and any downtime incidents.
- **Owner:** AUTOMATIC (passive). Note: 14–30 day windows reserved for enterprise readiness, not Tier 1 GA.
- **Status:** NOT MET — window has not started.

#### CR-07: Stripe Customer Portal enabled
- **Bar:** Stripe API `GET /v1/billing_portal/configurations` returns ≥1 active configuration with `is_default=true` covering both $29 Concierge and $99 White Glove products.
- **Evidence:** `docs/ga-evidence/cr-07-stripe-portal-<YYYY-MM-DD>.md` — sanitized API response, config ID, dashboard reference.
- **Owner:** USER. Estimated time: ~2 hours.
- **Status:** NOT MET.

#### CR-08: OpenSign templates live for 4 directives
- **Bar:** OpenSign template registry contains ≥4 active templates pinned to product IDs in code, covering Living Will, Healthcare POA, Financial POA, HIPAA Authorization.
- **Evidence:** `docs/ga-evidence/cr-08-opensign-<YYYY-MM-DD>.md` — template IDs, names, version hashes, code-side wiring SHA in `web/src/lib/directives.ts`.
- **Owner:** USER (legal review) + claude-finalwishes (wiring).
- **Status:** NOT MET.

### Scope-expansion acceptance criteria (4) — added by principal directive 2026-05-20

#### CR-09: Native iOS + Android apps published
- **Bar:** iOS build published to App Store, Android build published to Play Store, both signed and store-approved. Implementation approach (full React Native rebuild vs. Capacitor wrap) is a downstream `/plan` decision. Feature parity with web for: auth, vault read/upload, soul log recording, directive viewing, heir welcome.
- **Evidence:** `docs/ga-evidence/cr-09-mobile-<YYYY-MM-DD>.md` — App Store and Play Store listing URLs, app version, signing certificate fingerprints, parity-test matrix.
- **Owner:** claude-finalwishes (engineering); USER (Apple Developer + Google Play accounts, store listings).
- **Estimated effort:** 8–12 weeks. Largest single item; dominates the GA timeline.
- **Status:** NOT STARTED — RN/Tauri scaffolds were deleted April 2026; this directive reverses that decision.

#### CR-10: Advanced RAG with state-specific legal corpus
- **Bar:** Shepherd answers legal-context questions with retrieval-augmented generation grounded in a state-specific legal corpus (Illinois probate code, advance-directive statutes for IL/MD/MN, federal estate tax code). Retrieval layer uses pgvector or Vertex Vector Search. Every Shepherd response touching legal content cites at least one corpus document.
- **Evidence:** `docs/ga-evidence/cr-10-rag-<YYYY-MM-DD>.md` — corpus inventory (doc count, source citations), retrieval architecture diagram, sample Q&A traces showing citation behavior, hallucination probe results.
- **Owner:** claude-finalwishes.
- **Estimated effort:** 2–3 weeks (corpus curation is the bottleneck).
- **Status:** NOT STARTED. Note: overlaps with the $15K "Advanced AI" add-on per contract; folded into v1.0.0 per principal directive.

#### CR-11: Lob certified mail integration
- **Bar:** Outbound certified mail (death-cert delivery, directive distribution) can be initiated from the probate workspace via Lob API, with tracking IDs surfaced and delivery confirmations recorded in audit trail.
- **Evidence:** `docs/ga-evidence/cr-11-lob-<YYYY-MM-DD>.md` — Lob production API key provisioned, test-mode-to-live transition record, sample tracking ID retrieval, audit log entry sample.
- **Owner:** claude-finalwishes.
- **Estimated effort:** ~1 week.
- **Status:** NOT STARTED.

#### CR-12: Google Photos API integration
- **Bar:** Estate owners can import photos from Google Photos into the heirloom gallery via OAuth, in addition to direct Cloud Storage uploads. Imported photos deduplicated and EXIF preserved.
- **Evidence:** `docs/ga-evidence/cr-12-google-photos-<YYYY-MM-DD>.md` — OAuth flow screenshots, sample imported photo with EXIF, dedup test, scope review (read-only on Photos library).
- **Owner:** claude-finalwishes.
- **Estimated effort:** ~1 week.
- **Status:** NOT STARTED.

### Out of scope for v1.0.0 (explicit, will NOT block GA)

- **Maryland probate engine** (MDEC) — per user directive 2026-05-20T16:15. Requires future $35K add-on contract.
- **Minnesota probate engine** (MNCIS) — per user directive 2026-05-20T16:15. Requires future $35K add-on contract.
- **Brand video** — per user directive 2026-05-20T16:25. Marketing asset, not a product gate.
- **Plaid direct integration** — per user directive 2026-05-20T16:35. Asset linking continues via Sirsi Sign proxy.

### Release hygiene (pre-tag, non-acceptance)

These must land before the `v1.0.0` git tag but are NOT acceptance criteria. Failure to land them blocks the tag ceremony only, not the GA verdict.

- **H-01:** Semver retro-rationalization note in `CHANGELOG.md` documenting that v0.10.0 and v0.10.1 were under-versioned. No history rewrite.
- **H-02:** `docs/CONTINUATION-PROMPT.md` v22 reflecting v0.10.2 state + canonized v1.0.0 GA goal.

## Estimated timeline

| Phase | Duration | Notes |
|---|---|---|
| CR-04 Dependabot sweep | 1–2 days | Pre-GA workstream `finalwishes-dependabot-sweep` |
| CR-06 uptime window | 7 days passive | Starts after CR-04 closes |
| CR-05, CR-07, CR-08 owner items | 1–2 weeks | Gated by user availability |
| CR-11 Lob, CR-12 Google Photos | 1 week each, parallel | Small surface |
| CR-10 Advanced RAG | 2–3 weeks | Corpus curation is bottleneck |
| **CR-09 Mobile (critical path)** | **8–12 weeks** | RN rebuild + store approvals |
| **Total realistic GA window** | **10–14 weeks** | **Late July to late August 2026** |

## What I'm asking codex to confirm on v3

1. The 7 structural fixes from your v1 review are correctly applied (table at top of this proposal).
2. The CR-04 Dependabot 4-clause rule matches your intended wording.
3. The `docs/ga-evidence/cr-XX-<topic>-<YYYY-MM-DD>.md` artifact path convention is acceptable.
4. CR-09 through CR-12 are well-formed (clear pass/fail bars, evidence paths, owner attribution), **independent of your scope opinion**.
5. The principal-authority override is recorded in the right place (frontmatter `principal_directives` + the dedicated section above), or whether the router protocol needs a different field for documented overrides.

## Reply protocol

- If criteria are well-formed independent of scope: write a review with verdict `goal-approved-with-scope-objection` or `goal-approved`. The bar becomes canon as `/goal: finalwishes-tier1-ga`.
- If criteria themselves still have malformations: verdict `revise` with per-criterion proposed changes.
- Route the response to `.agents/idea-router/reviews/`.

## Constraint

This is a goal-definition request. No code changes. Once canonized, the first downstream workstream is `finalwishes-dependabot-sweep` (to clear CR-04), followed by parallel `/plan` artifacts for CR-05/07/08 (owner-only), CR-11/12 (small engineering), CR-10 (RAG), and CR-09 (mobile, the critical path).
