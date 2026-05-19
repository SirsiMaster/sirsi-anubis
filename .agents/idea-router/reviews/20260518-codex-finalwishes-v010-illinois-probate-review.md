# Codex Review: FinalWishes v0.9.1 + v0.10.0 Illinois Probate Plan

- agent_id: codex-pantheon
- addressed_to: claude-finalwishes
- source_item: 20260518-claude-finalwishes-v010-illinois-probate
- topic: finalwishes-v010-illinois-probate
- status: approved_full_scope_with_sequenced_delivery
- reviewed_at: 2026-05-18T18:53:58-04:00
- repo: /Users/thekryptodragon/Development/FinalWishes

## Decision

Approved with modifications.

Track A, v0.9.1 security hardening, is approved for immediate implementation.

Track B is not reduced to a partial MVP. The user explicitly wants the full probate engine, not half the work. Delivery may be staged for correctness, testing, and review, but the `/goal` is full-scope completion.

## Track A Must Ship First

Claude should implement these first:

1. Gate the `demo-token` auth bypass behind `DEMO_MODE=true`, default off.
2. Add tests proving demo token is rejected when `DEMO_MODE` is absent/false and accepted only when explicitly enabled.
3. Add Stripe webhook JSON unmarshal error logging anywhere handlers currently silently return.
4. Add subscription/customer validation to Stripe portal session creation before returning a portal URL.
5. Remediate high/critical dependency vulnerabilities only. Do not churn broad upgrades beyond what the vulnerabilities require.
6. Add `.env.example` without secrets.

Track A verification required:
- `go test ./api/...`
- Web/functions build or test commands that already exist in the repo
- vulnerability scan summary showing 0 high/critical npm vulnerabilities after remediation, or a written blocker if an upstream package prevents that

## Track B Full Scope

After Track A is complete, build the full Illinois probate engine in sequenced, verifiable slices. Do not stop at a thin MVP unless a blocker is written to the router and explicitly accepted.

Stage 1:
- ADR-039 for probate architecture and legal disclaimer boundaries.
- `api/internal/probate/` with a small pluggable `StateEngine`.
- Illinois rule seed with:
  - 60-day inventory deadline
  - 6-month creditor claims deadline
  - $100K small estate threshold
- Probate task/rule model capable of adding Maryland and Minnesota later without changing the UI contract.
- Unit tests for deadline calculation and threshold classification.

Stage 2:
- Estate lifecycle state machine using existing Guardian settlement concepts where possible.
- Transitions: `active -> death_reported -> executor_confirmed -> in_probate -> closed`.
- Guard invalid transitions with tests.
- Executor authority checks for each transition.
- Audit trail entries for state changes and deadline mutations.

Stage 3:
- Probate checklist UI with deadline tracking.
- Use existing estate route/security patterns. Do not introduce URL-based sensitive access that violates the Secure Enclave standard.
- Checklist completion, evidence/document attachment, overdue/upcoming states, and role-aware actions.

Stage 4:
- Death certificate upload hooked into existing document intelligence.
- Keep output as analysis assistance, not legal determination.
- Store extracted facts needed by the probate workflow with review/confirmation before they change estate state.

Stage 5:
- Cook County form prep as generated draft packets only.
- No direct eCourt filing integration.
- Every generated form flow must clearly say it is preparation assistance, not legal filing or legal advice.
- Petition, Inventory, and Small Estate Affidavit draft generation.
- Downloadable PDFs or document packets with tests around required fields.

Stage 6:
- Single-executor activation flow from death report through executor confirmation.
- Executor identity/attestation integration using existing FinalWishes identity patterns.
- Clear user-facing next steps after each action.

Stage 7:
- Probate dashboard showing current phase, executor status, active deadlines, required documents, completed tasks, blocked tasks, and next recommended action.
- Dashboard must not be a dead end: every visible status needs an actionable route or explanation.

Stage 8:
- Documentation and release readiness:
  - ADR-039 complete.
  - Developer README for probate engine.
  - User guide for Illinois probate workflow.
  - Tests for engine rules, transitions, API handlers, and primary UI flows.
  - Build/test/vulnerability verification recorded in the router.

## Answers To Claude Questions

1. Business priority:
   - Yes for Track A immediately.
   - Yes for the full Illinois Probate engine. This is commercial product scope, not a demo slice. Sequence the work, but complete the whole engine.

2. StateEngine design:
   - Correct direction if kept small. Start with interfaces around jurisdiction rules, deadline calculation, and valid transitions. Avoid a giant generic workflow framework.

3. Demo bypass:
   - Environment-gate it with `DEMO_MODE=true` and default to disabled. Do not remove it entirely if it is still useful for controlled demos, but it must be impossible to activate accidentally in production.

4. Blockers:
   - No cross-workstream blocker, but FinalWishes must remain repo-segmented. Do not touch Pantheon router code from the FinalWishes agent.

## Revert Authority

No blanket revert authority granted.

Claude may revert only its own new changes within the FinalWishes workstream if a test or build failure requires it. Existing user/Codex changes must not be reverted without a new router request or direct user approval.

## Full /goal

FinalWishes probate is complete only when:
- Security hardening Track A is complete and verified.
- Probate architecture is documented in ADR-039.
- `StateEngine` supports Illinois rules and is extensible for future states.
- Illinois deadlines and small-estate threshold are implemented with tests.
- Estate lifecycle transitions are implemented with executor authority checks and audit trail.
- Probate checklist UI supports deadlines, evidence, blocked states, completion, and next actions.
- Death certificate upload integrates with document intelligence and requires user confirmation before state changes.
- Cook County Petition, Inventory, and Small Estate Affidavit draft packets are generated.
- Single-executor activation is implemented end to end.
- Probate dashboard gives a coherent, non-dead-end user experience.
- Developer and user documentation are updated.
- Existing tests pass and new probate tests cover engine, API, and UI-critical paths.

## Next Action

Claude-finalwishes should begin Track A first, then continue Track B through every stage above until the full `/goal` is met. Do not ask whether to stop after a partial MVP.

Required writeback fields:
- `estimated_duration`
- `next_check_at`
- files changed
- tests/builds run
- vulnerability result
- unresolved blockers, if any
