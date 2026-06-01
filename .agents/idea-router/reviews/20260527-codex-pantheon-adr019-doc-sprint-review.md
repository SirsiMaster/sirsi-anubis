# Codex Review — ADR-019 Knowledge Substrate Doc Sprint

**Reviewer:** codex-pantheon  
**Date:** 2026-05-27  
**Item:** `20260527-180524-claude-pantheon-codex-pantheon-doc-sprint-executed-adr-019-4-new-docs-5-canon-edits-review-`  
**Decision:** Approved with required canon corrections before commit/release publication

## Summary

The doc sprint is directionally approved. ADR-019, the user guide, public web page, case study, changelog, architecture section, and index update correctly establish the Knowledge Substrate as a local semantic-verification feeder for the future Sirsi hypergraph. The three-tool split is the right shape:

- Thoth owns curated memory, intent, and journal continuity.
- Seba owns deity-canonical topology and architectural mapping.
- Understand-Anything / `.understand-anything/knowledge-graph.json` owns auto-derived semantic verification from source.

This review also incorporates the prior Codex completion review for the Understand-Anything / hypergraph handoff: `.understand-anything/` should be committed for Pantheon for now because the artifact is small enough, useful to agents immediately, and is a portable JSON feeder. Revisit only if churn or size becomes material.

## Required Corrections

1. **Update ADR-019 review status and decider metadata.**  
   ADR-019 currently says `codex-pantheon (review pending — router item 20260527-174419...)`. That is stale. Reference the completed Codex review artifact `reviews/20260527-codex-pantheon-understand-hypergraph-completion-review.md`, and this doc-sprint review, so the ADR no longer records a pending Codex decision after approval.

2. **Remove the stale `.understand-anything/` open question.**  
   ADR-019 Alternatives §5 still says commit-vs-gitignore is deferred. It is no longer deferred for Pantheon. Canonical policy: commit `.understand-anything/` in `sirsi-pantheon` for now; generic docs may still describe team tradeoffs, but the ADR should state Pantheon's current decision clearly.

3. **Normalize hypergraph gating language.**  
   ADR-019 §6 says implementation is gated by `--feature=hypergraph`, then later says Go `--tags hypergraph`. Pantheon is Go; use build tag language consistently: config-time switch `configs/hypergraph.yaml` and compile-time `hypergraph` build tag. Avoid `--feature=hypergraph` unless it is a real command or build flag.

4. **Treat `configs/hypergraph.yaml` as a local config policy, not an accidentally committed mutable config.**  
   The docs can specify the schema, but the implementation batch should decide whether the committed file is `configs/hypergraph.yaml.example` while real `configs/hypergraph.yaml` remains local/gitignored. If the actual YAML is committed, make that explicit and ensure it contains no machine-local state or future credentials.

5. **Keep `sirsi hypergraph` clearly marked as pending.**  
   The user guide and public page mostly do this, but README’s command table currently lists `sirsi hypergraph status` like a live command. Add a parenthetical such as “planned / pending implementation” or move it to a future surface until the CLI exists. Canon must not teach users to run a command that is not present.

6. **Correct ADR-INDEX stale metadata.**  
   `docs/ADR-INDEX.md` now has ADR-019 in the registry, but numbering history still says `ADR-019+ | Next available`, ADR-010 is still marked Proposed in history while the registry marks it Accepted, and the footer still says May 21 / ADR-018. Update it to ADR-020 next available and a May 27, 2026 ADR-019 note.

7. **Soften “TUI sunset/eliminated” language across canon.**  
   This is the most important product-canon correction. The current inherited TUI implementation was removed from the Mac v1 path because it was broken and brand-damaging, but the broader TUI ambition is not dead. User direction on 2026-05-27: TUIs are strategically important; inability to build one would call Sirsi’s broader execution credibility into question. Therefore:
   - ADR-INDEX should not summarize ADR-018 as plain “TUI sunset.” Use wording like “Native macOS App + CLI for Mac v1; inherited TUI implementation superseded.”
   - CHANGELOG may retain the factual removal of `internal/output/tui*.go`, but should avoid implying “no TUI ever.” Add an explicit note that this removes the failed inherited implementation, not the strategic possibility of a future Pantheon Operator TUI / command console.
   - README and public pages should not continue advertising the old no-args TUI if that surface was removed; describe the current CLI/native app path and reserve future TUI language for planned work.

8. **Case study commit reference is premature.**  
   `docs/case-studies/2026-05-26-knowledge-substrate-day-1.md` ends with `Commit reference: v0.22.0-beta — ADR-019 accepted...` but this sprint was explicitly uncommitted at handoff. Replace with “Documentation sprint reference” / “Planned release target” until an actual commit or tag exists.

9. **Use precise Hedera wording.**  
   User-facing docs should prefer “Hedera Consensus Service (HCS) as the event substrate” over “Hedera graph technology.” The hypergraph is Sirsi’s graph; Hedera is the ordered durable event log candidate.

10. **Keep command examples tied to verified plugin capabilities.**  
    The user guide documents `/understand --auto-update` and `/understand --language es`. If those flags are verified from the plugin, fine. If not, either verify before commit or mark them as plugin-supported examples only after confirmation. Canon should not invent plugin CLI flags.

## Non-Blocking Notes

- `internal/seba/` exists in the repo, so the Seba path in ADR-019 and architecture docs is acceptable.
- The architecture section’s three-sovereignty split is strong and should stay.
- The public page’s feature table correctly marks `sirsi hypergraph` as spec’d / implementation pending; keep that pattern consistent elsewhere.
- The README currently still describes the old TUI as live in install, quick start, and “How It Works.” That is outside the knowledge-substrate sprint but must be reconciled before any release cut that includes the TUI removal changelog.

## Completion Decision

Codex approves the ADR-019 / Knowledge Substrate direction and the documentation set after the above corrections. The next doc batch should be a small canon cleanup pass before Swift or hypergraph CLI implementation begins:

1. ADR-019 metadata + commit policy + build-tag/config language.
2. ADR-INDEX numbering/date/TUI-summary cleanup.
3. README/changelog current-surface reconciliation, including future-TUI preservation language.
4. Case-study commit-reference fix.
5. One verification pass on plugin command flags.

After that, Phase-2 can continue with the already-approved dashboard API/gap/envelope docs and then implementation planning.
