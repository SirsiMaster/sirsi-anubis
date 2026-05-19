# Proposal: Assiduous v1.1.0 "Production Complete"

- agent_id: claude-assiduous
- repo: /Users/thekryptodragon/Development/assiduous
- topic: assiduous-v110-completion
- addressed_to: codex
- created_at: 2026-05-18T00:00:00-04:00

## /plan

Complete Assiduous from v1.0.54 (shipped MVP) to v1.1.0 (production complete). 7 tracks covering security, billing, email, RPCs, testing, polish, and external dependencies. Full plan at `CODEX_TASKS.md` in the repo root.

## /goal

Ship v1.1.0 with:
1. 0 npm vulnerabilities (web)
2. CSP tightened (no unsafe-eval), HSTS + Permissions-Policy headers
3. Stripe billing lifecycle complete (checkout + portal + webhooks)
4. 4 notification Cloud Functions wired to SendGrid email
5. Playwright tests running in CI with 0 skipped
6. axe-core a11y tests on 10 key pages
7. Custom analytics events for key user actions
8. All .env.example files documented

## What Changed (Claude session, 2026-05-18)

### Completed by claude-assiduous:
- **Track 1a**: Fixed protobufjs + postcss vulns. web: 0 vulns. functions: 9 low transitive (unfixable without breaking firebase-admin).
- **Track 1b**: Removed `'unsafe-eval'` from CSP, added `object-src 'none'` + `base-uri 'self'`.
- **Track 1c**: Added `Strict-Transport-Security` (HSTS preload) + `Permissions-Policy` headers.
- **Track 1d**: Updated 3 `.env.example` files, created `backend/.env.example` with all 15+ vars.
- **Track 2a**: Added `POST /api/stripe/create-portal-session` endpoint. `go build` clean, `go test ./pkg/billing/` passes.
- **Track 4a**: Audited all 31 RPCs — ALL are implemented (not 39 stubs as previously believed). Frontend calls 21/31. 10 are backend-only/admin ops. Track 4 eliminated from scope.

### Drafted for Codex:
- `CODEX_TASKS.md` in repo root — 6 self-contained Codex tasks:
  1. Billing management page + portal session integration
  2. Wire notification Cloud Functions to SendGrid email
  3. Add offer + subscription email templates
  4. Playwright test data seeding + unskip tests
  5. Accessibility (axe-core) tests
  6. Custom analytics events

## Tests/Builds Run
- `go build ./cmd/api/` — clean
- `go vet ./pkg/billing/` — clean
- `go test ./pkg/billing/ -v` — 2/2 pass
- `npm audit` (web) — 0 vulnerabilities

## Failures or Blockers
- functions: 9 low transitive vulns in firebase-admin chain (requires breaking downgrade, not actionable)
- Stripe portal requires Customer Portal configuration in Stripe Dashboard (manual step)
- Bright MLS credentials still pending from vendor

## Exact Next Action Requested from Codex

Pick up `CODEX_TASKS.md` in the Assiduous repo. Tasks 1-3 are highest priority (billing UI, email wiring, email templates). Tasks 4-6 (testing, a11y, analytics) follow. Each task is self-contained with files to modify, exact code snippets, and verification steps.
