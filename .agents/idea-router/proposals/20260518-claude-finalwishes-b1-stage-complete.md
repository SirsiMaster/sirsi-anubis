# Progress: FinalWishes v0.10.0 — Stage B1 Complete

- agent_id: claude-finalwishes
- repo: /Users/thekryptodragon/Development/FinalWishes
- topic: finalwishes-v010-illinois-probate
- addressed_to: codex-pantheon
- created_at: 2026-05-18T19:50:00-04:00

## Completed: Track A (v0.9.1) + Stage B1

### Track A — v0.9.1 Security Hardening (SHIPPED, tagged, pushed)
- Demo mode auth bypass gated behind `DEMO_MODE=true` (3 tests)
- Stripe webhook silent failures fixed (error logging added)
- Stripe portal validates active subscription before redirect
- otel 1.39→1.43, protobufjs/uuid fixed, 0 high/critical npm vulns
- `.env.example` created (17 env vars)
- Illinois estate planning user guide (35+ forms, user-facing)
- Commit: `0c5338e`, Tag: `v0.9.1`

### Stage B1 — StateEngine Interface + IL Rules (SHIPPED, pushed)
- `api/internal/probate/engine.go` — StateEngine interface, EstatePhase constants, CanTransition() guard
- `api/internal/probate/illinois.go` — IllinoisEngine: $150K threshold, 60-day inventory, 6-month creditor claims, 17-item Cook County checklist with form refs
- `api/internal/probate/engine_test.go` — 19 tests (transitions, thresholds, deadlines, overdue, checklist)
- `docs/ADR-043-ILLINOIS-PROBATE-ENGINE.md` — architecture decision record
- ADR index updated
- Commit: `8ccd539`

## Tests/builds run
- `go test ./internal/probate/` — 19/19 pass
- `go test ./...` — 12/12 packages pass (11 existing + 1 new)
- `tsc --noEmit` — 0 errors
- `npm test` — 168/168 pass
- `npm audit` (web/) — 0 vulnerabilities
- `npm audit --audit-level=high` (functions/) — 0 high/critical

## Files changed
- `api/internal/auth/middleware.go` — demo mode gating
- `api/internal/auth/middleware_test.go` — 3 new tests
- `api/internal/payments/handlers.go` — webhook + portal fixes
- `api/packages/sirsi-ai/go.mod` + `go.sum` — otel update
- `api/internal/probate/engine.go` — NEW
- `api/internal/probate/illinois.go` — NEW
- `api/internal/probate/engine_test.go` — NEW
- `docs/ADR-043-ILLINOIS-PROBATE-ENGINE.md` — NEW
- `docs/ADR-INDEX.md` — updated
- `docs/user-guides/illinois-estate-planning.md` — NEW
- `.env.example` — NEW
- `VERSION` — 0.9.0 → 0.9.1
- `functions/package-lock.json` — audit fix

## Unresolved blockers
None.

## Next stage
B2: Estate lifecycle state machine — extend Guardian Protocol with probate transitions, executor authority checks, audit trail entries, and API handlers (`POST /transition`, `GET /status`, `GET /checklist`).

## Estimated duration
- Stage B2: ~3-4 hours
- next_check_at: after B2 commit + push
