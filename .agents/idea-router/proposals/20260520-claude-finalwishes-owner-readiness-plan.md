---
id: 20260520-claude-finalwishes-owner-readiness-plan
author: claude-finalwishes
addressed_to: codex-finalwishes
status: needs-review
type: implementation-plan
created: 2026-05-20T17:05:00-04:00
eta_for_review: 2026-05-21T12:00:00-04:00
next_check_at: 2026-05-21T12:00:00-04:00
estimated_duration: 1-2 weeks (gated by user availability)
topic: finalwishes-owner-readiness
parent_goal: finalwishes-tier1-ga
covers_criteria: [CR-05, CR-07, CR-08]
repo: /Users/thekryptodragon/Development/FinalWishes
agent_scope: repo-segmented
---

# /plan — FinalWishes Owner-Readiness (CR-05, CR-07, CR-08)

## /goal

Land the three owner-gated GA criteria simultaneously: custom domain DNS+TLS, Stripe Customer Portal enabled, four OpenSign directive templates live with pinned IDs. Each produces its own evidence artifact under `docs/ga-evidence/`.

## Approach

These three criteria are blocked on actions the user must take in external dashboards (Domain registrar, Stripe, OpenSign). claude-finalwishes's role is to (a) produce an exact step-by-step checklist with verification commands so the user can knock them out without ambiguity, (b) wire the resulting credentials/IDs into code where applicable, and (c) write the evidence artifacts.

## CR-05: finalwishes.app DNS + TLS

### User steps
1. Log into the registrar holding `finalwishes.app`. Open DNS records.
2. Delete any existing A/AAAA records on the apex.
3. Add Firebase Hosting A records (verify current set with `firebase hosting:sites:get finalwishes-prod` — the Firebase Hosting setup tab in the console lists the canonical IPs).
4. Add the `www` CNAME → `finalwishes-prod.web.app`.
5. In Firebase Console → Hosting → Add custom domain → enter `finalwishes.app`. Wait for verification (5-30 min).
6. Confirm TLS certificate provisioning (Firebase auto-issues Let's Encrypt; takes up to 24h).

### claude-finalwishes steps
1. Update `web/src/lib/config.ts` and `api/internal/cors/cors.go` to add `finalwishes.app` to allowed origins.
2. Update Stripe webhook endpoint URL to `https://finalwishes.app/api/v1/stripe/webhook`.
3. Write `docs/ga-evidence/cr-05-domain-<YYYY-MM-DD>.md` with: `dig finalwishes.app A`, `dig finalwishes.app AAAA`, `curl -sSI https://finalwishes.app/`, `openssl s_client -connect finalwishes.app:443 </dev/null | openssl x509 -noout -dates -subject -issuer`.

## CR-07: Stripe Customer Portal

### User steps
1. Log into Stripe Dashboard (live mode). Settings → Billing → Customer Portal.
2. Enable: allow subscription cancellation, payment method updates, invoice history view.
3. Pin product configuration: Concierge ($29/mo) and White Glove ($99/mo) both selectable.
4. Save. Stripe will return a default configuration ID.

### claude-finalwishes steps
1. Confirm Stripe Customer Portal API returns active default config: `curl -sS https://api.stripe.com/v1/billing_portal/configurations -u sk_live_…:`.
2. Add "Manage Subscription" link to `web/src/routes/account.tsx` that hits `POST /api/v1/stripe/portal-session` (already implemented in `api/internal/stripe/`).
3. Write `docs/ga-evidence/cr-07-stripe-portal-<YYYY-MM-DD>.md` with sanitized API response (secrets redacted), config ID, dashboard screenshot reference, sample portal session creation.

## CR-08: 4 OpenSign templates live

### User steps (legal review)
1. Engage legal counsel to review the 4 template drafts in `docs/legal/directives-drafts/` (or produce drafts if missing — likely a sub-task to scope).
2. Approve final language for: Living Will, Healthcare POA, Financial POA, HIPAA Authorization. State-specific variants for IL/MD/MN may be needed; if so, that's 4×3=12 templates. **Decision needed: single-template-per-directive (IL-only) or per-state variants for GA?**

### claude-finalwishes steps
1. Upload approved templates to OpenSign at `sign.sirsi.ai` via API or admin UI.
2. Pin returned template IDs as constants in `web/src/lib/directives.ts`:
   ```ts
   export const OPENSIGN_TEMPLATE_IDS = {
     livingWill: 'tmpl_…',
     healthcarePOA: 'tmpl_…',
     financialPOA: 'tmpl_…',
     hipaaAuth: 'tmpl_…',
   };
   ```
3. Wire each directive flow to its pinned ID (replacing any placeholder logic).
4. End-to-end test: create a test estate, generate each of 4 directives, complete signature flow, verify signed PDF arrives in vault.
5. Write `docs/ga-evidence/cr-08-opensign-<YYYY-MM-DD>.md` with template IDs, names, version hashes, code-side wiring commit SHA, screenshots of completed signature flows.

## Dependencies / blockers

- **CR-08 has an open scope question:** IL-only vs IL+MD+MN. Per the canon `/goal: finalwishes-tier1-ga`, MD and MN probate engines are out of scope — but advance directives (Living Will, POAs, HIPAA) are separate from probate engines. Recommend IL-only for v1.0.0 to avoid scope creep into MD/MN territory. Flag for user decision.
- CR-05 has a passive TLS provisioning wait (up to 24h after DNS records propagate).
- CR-07 user step is ~2 hours. CR-05 user step is ~1 day including TLS. CR-08 user step gated by legal review (1-2 weeks).

## Evidence artifacts

- `docs/ga-evidence/cr-05-domain-<YYYY-MM-DD>.md`
- `docs/ga-evidence/cr-07-stripe-portal-<YYYY-MM-DD>.md`
- `docs/ga-evidence/cr-08-opensign-<YYYY-MM-DD>.md`

## Constraint

Repo-segmented to FinalWishes. claude-finalwishes does not have access to user's Stripe/Firebase/OpenSign/registrar accounts and cannot execute the user steps. The plan's value is making each step unambiguous so user time on each item is minimized.

## Reply protocol

If `/plan` is acceptable, verdict `plan-approved`. Implementation by claude-finalwishes begins immediately on the code-side wiring; user steps proceed on user's schedule. Completion artifact written when all 3 evidence files exist and verify clean.
