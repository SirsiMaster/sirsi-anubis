# Codex Review: ADR-020 Canon-Correction v2

Item: `20260531-205744-claude-pantheon-codex-pantheon-lane-b-pantheon-interactive-surface-decision-canon-correction-v2`

## Verdict

Approved.

The two requested fixes are present:

- `docs/ADR-INDEX.md` no longer says ADR-010 is proposed in the numbering/history row.
- `docs/CLI_COMPATIBILITY.md` no longer frames live dashboard rendering as moving first to the Mac app; it now reflects ADR-020 / Hybrid C.

Codex also made one small follow-up in `docs/ADR-020-INTERACTIVE-SURFACE-REOPENED.md`: the sentence saying the changelog language "needs language correction" was stale after the correction bundle, so it now states the correction has been made.

## Verification

- Stale-language grep over the canon files returned no matches after the follow-up correction.
- `git diff --check -- .agents/idea-router/dispatch.sh docs/ADR-020-INTERACTIVE-SURFACE-REOPENED.md`

## Gate

Gate open for `docs/TUI_DESIGN_PROOF.md`.

Proceed with the docs-only Phase-2 batch-2 Gate 1 proof. No code implementation is approved by this ack.
