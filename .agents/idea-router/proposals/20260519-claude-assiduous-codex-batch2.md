# Codex Batch 2: Assiduous v1.1.0 Remaining Tasks

- agent_id: claude-assiduous
- repo: /Users/thekryptodragon/Development/assiduous
- topic: assiduous-v110-completion
- addressed_to: codex-assiduous
- created_at: 2026-05-19T11:00:00-04:00
- estimated_duration: 2-3 hours total

## Context

All Claude tasks are done. These are the remaining Codex tasks from CODEX_TASKS.md + the plan. Working tree is clean at commit `65025169`.

## Tasks (priority order)

### 1. Wire PricingPlans to checkout (Track 2c)
`web/src/components/PricingPlans.tsx` already has the checkout call via `startSubscriptionCheckout()`. Verify it works end-to-end. If the API path doesn't match the Go route (`/api/stripe/create-checkout-session`), fix it.

### 2. Unskip + fix 35 Playwright tests (Track 5b)
Create `tests/seed-data.ts` with Firestore seed script. Unskip tests in `properties.spec.ts`, `contracts.spec.ts`, `navigation.spec.ts`. See CODEX_TASKS.md Task 4.

### 3. Go unit tests — 70% coverage (Track 5c)
Write tests for `backend/pkg/rpc/`, `backend/pkg/billing/`, `backend/pkg/mls/`. Currently 570 LOC across 7 test files. Target 70%+ on core packages.

### 4. axe-core accessibility tests (Track 5d)
Install `@axe-core/playwright`, create `tests/accessibility.spec.ts` for 10 key pages. See CODEX_TASKS.md Task 5.

### 5. ARIA attributes (Track 6a)
Add ARIA to remaining pages targeting WCAG 2.1 AA on 10 key pages.

### 6. Custom analytics events (Track 6b)
Create `web/src/lib/analytics.ts`, wire events into hooks. See CODEX_TASKS.md Task 6.

### 7. Landing page polish (Track 6c)
Add pricing section, testimonial placeholders, ROI messaging to LandingPage.tsx.

### 8. Wire Plaid Link Token (Track 6d)
Backend `pkg/plaid/plaid.go` is complete. Frontend components exist. Wire them together.

### 9. Wire Lob certified mail (Track 6e)
Backend `pkg/lob/lob.go` is complete. Frontend components exist. Wire them together.

## /goal
All 9 tasks implemented, `npm run build` and `go build` clean, verification per CODEX_TASKS.md.

## Verify
- `cd web && npm run build` — 0 errors
- `cd backend && go build ./cmd/api/` — clean
- `cd functions && npx tsc --noEmit` — clean
- `npx playwright test --list` — 0 skipped
