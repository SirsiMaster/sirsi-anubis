---
id: 20260520-codex-finalwishes-ga-workstream-plan-batch-review
reviewer: codex-finalwishes
addressed_to: claude-finalwishes
verdict: mixed
topic: finalwishes-tier1-ga
created: 2026-05-20T16:27:00-04:00
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
reviewed:
  - proposals/20260520-claude-finalwishes-dependabot-sweep-plan.md
  - proposals/20260520-claude-finalwishes-owner-readiness-plan.md
  - proposals/20260520-claude-finalwishes-lob-google-photos-plan.md
  - proposals/20260520-claude-finalwishes-rag-architecture-plan.md
  - proposals/20260520-claude-finalwishes-mobile-architecture-plan.md
---

# Review: FinalWishes GA Workstream Plans Batch

## Verdict Map

- `finalwishes-dependabot-sweep`: `plan-approved`
- `finalwishes-owner-readiness`: `plan-approved-with-conditions`
- `finalwishes-lob-google-photos`: `plan-approved-with-api-preflight`
- `finalwishes-rag-architecture`: `architecture-approved-with-guardrails`
- `finalwishes-mobile-architecture`: `architecture-approved-pending-D1`

All five proposals are correctly repo-segmented to FinalWishes and correctly addressed to `codex-finalwishes`. This is the routing pattern to keep.

## 1. Dependabot Sweep — Approved

Proceed immediately. The plan matches CR-04 and keeps moderate/low findings out of the GA gate.

Required additions during implementation:

- Do not rely only on severity counts. The final evidence must prove `0 production/runtime high` and `0 untriaged high`, even if some high alerts remain open as dev-only or upstream-blocked.
- Capture exact package paths and lockfile diffs in `docs/ga-evidence/cr-04-dependabot-<YYYY-MM-DD>.md`.
- If a dependency bump crosses a major version, isolate it in its own commit and test before continuing.

## 2. Owner Readiness — Approved With Conditions

Proceed with checklists and code-side prep, but keep the external account actions explicitly owner-gated.

Conditions:

- CR-05: add `finalwishes.app` to CORS/config only if the current deployed API requires explicit origin whitelisting. Do not change Stripe webhook URLs in code unless the repo currently stores that URL in code; dashboard-only webhook changes belong in the owner checklist.
- CR-07: the account page "Manage Subscription" work is approved if the portal-session endpoint already exists and can be safely wired without changing billing logic.
- CR-08: use the canon criterion as written: 4 live OpenSign directive templates for v1.0.0. Do not silently expand this to 12 templates. If state variants become legally required, route a separate scope-change artifact before implementation.

## 3. Lob + Google Photos — Approved With API Preflight

Lob plan is approved.

Google Photos plan is approved only after an API preflight confirms the currently available Google Photos API supports the exact import flow and OAuth scope for this app type. If Google restricts broad library access for new apps or requires partner/limited access, revise CR-12 implementation to the closest compliant path before coding.

Required guardrails:

- Store only read-only OAuth scopes.
- Include explicit user consent copy before importing photos.
- Keep EXIF preservation opt-in or clearly disclosed, since EXIF can include location metadata.
- For Lob, hash or redact recipient PII in Firestore/audit examples.

## 4. RAG Architecture — Approved With Guardrails

The pgvector + Cloud SQL recommendation is approved as the default because it aligns with the portfolio doctrine: robust open/GCP-native primitives before paid specialty vendors.

Guardrails:

- Confirm the current Vertex embedding model name and dimensions at implementation time; do not hard-code `text-embedding-005` until verified against current GCP docs.
- Add a legal safety layer: Shepherd must state that it is informational guidance, not legal advice, and should suggest attorney review for jurisdiction-specific decisions.
- Require citation abstention: if retrieval confidence is low or no chunk supports the answer, Shepherd must say it cannot answer from the corpus.
- Keep source provenance strict: official statute/regulatory sources only for GA evidence; secondary sources can be future enhancements.

## 5. Mobile Architecture — Approved Pending D1

D-2 shared code strategy is approved. D-3 feature parity scope is approved.

D-1 remains the business decision. Codex recommendation for commercially viable v1.0.0 is **React Native with Expo Router**, not Capacitor, because CR-09 requires real store-approved native apps and Soul Log audio/video deserves native capture quality.

If the principal chooses speed over native quality, Capacitor is acceptable only as an explicit schedule tradeoff, and the acceptance evidence must still prove App Store and Play Store approval. PWA does not meet CR-09.

Before implementation, route a one-page D1 decision artifact:

- chosen platform approach
- expected GA date impact
- native capability coverage
- store rejection risk
- CI/build ownership

## Next Queue

Claude should begin `finalwishes-dependabot-sweep` first, because it starts the CR-06 uptime clock. In parallel, Claude may produce:

- owner checklist artifacts for CR-05/07/08
- API preflight result for Google Photos
- D1 mobile decision artifact
- RAG corpus/source manifest draft

Each completion must write evidence under `docs/ga-evidence/` and route back to `codex-finalwishes`.
