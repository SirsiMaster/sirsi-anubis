# Proposal: Ra/Horus CTR Hypervisor Canon + Product Surface Update

author: codex
addressed_to: claude-pantheon
status: needs-implementation
created: 2026-05-19T10:00:00-04:00
eta_for_review: 2026-05-19T12:30:00-04:00
next_check_at: 2026-05-19T12:30:00-04:00
estimated_duration: 90 minutes
repo: /Users/thekryptodragon/Development/sirsi-pantheon
agent_scope: repo-segmented

## /goal

Make the Ra/Horus CTR ownership model canonical, visible, and operational across Pantheon's code, product, governance, website, changelog, development log, ADRs, and case study surfaces.

Completion means the repo consistently says:

```text
Ra owns CTR / Idea Router as Sirsi-wide orchestration.
Horus owns each desktop's local runtime node and operator view.
Thoth preserves continuity.
Ma'at validates governance.
Net keeps portfolio goals aligned.
```

## Context

Codex already codified the core ownership split in:

- `/Users/thekryptodragon/Development/AGENTS.md`
- `AGENTS.md`
- `.agents/idea-router/README.md`
- `.agents/idea-router/DESIGN.md`

Those edits establish that CTR / Idea Router is a Ra-owned Pantheon feature, while Horus owns the local desktop node: daemon health, local agent/window visibility, local repo status, and local operator surface.

The user now wants this treated as a Pantheon innovation and propagated into canonical files, website surfaces, ADRs, changelog, development log, and case studies.

## /plan

1. Read the current Ra/Horus canon:
   - `/Users/thekryptodragon/Development/AGENTS.md`
   - `AGENTS.md`
   - `.agents/idea-router/README.md`
   - `.agents/idea-router/DESIGN.md`
   - `docs/ADR-015-DEITY-HIERARCHY.md`
   - `docs/PANTHEON_HIERARCHY.md`
   - `docs/DEITY_REGISTRY.md`

2. Add or update an ADR for the innovation.
   - Preferred: create a new next-numbered ADR, e.g. `docs/ADR-017-RA-HORUS-CTR-HYPERVISOR.md`, unless an existing ADR clearly owns this exact decision.
   - Update `docs/ADR-INDEX.md`.
   - The ADR must state the ownership boundary: Ra is Sirsi-wide orchestration; Horus is per-desktop runtime/operator surface.

3. Update canonical Pantheon docs.
   - `docs/ARCHITECTURE_DESIGN.md`
   - `docs/PANTHEON_HIERARCHY.md`
   - `docs/DEITY_REGISTRY.md`
   - any relevant user guide such as `docs/user-guides/ra.md`
   - preserve existing terminology unless it conflicts with the new canon.

4. Implement the missing code surface for the Ra/Horus split.
   - Audit existing router code in `cmd/sirsi/routercmd.go` and `internal/router/`.
   - Audit existing Ra/Horus desktop/TUI surfaces in `internal/output/`, `internal/ra/`, and `cmd/sirsi-menubar/`.
   - Preserve existing `sirsi router work`, daemon, and service-status behavior.
   - Add a concrete local-node status surface if it does not already exist. Preferred shape: a `sirsi router node-status` or `sirsi horus router-status` command that reports:
     - router home path
     - registered agents
     - pending work by agent
     - work-queue item statuses
     - launchd daemon installed/loaded/configured-binary health
     - recent dispatch failures, especially unauthenticated Claude CLI
   - Wire the same local-node status into the existing TUI/operator surface if feasible without destabilizing the current TUI refactor.
   - Add tests for any new router/Horus status aggregation code.
   - If a better existing code surface already covers this, document it and improve naming/copy so the Ra/Horus split is clear.

5. Update website/product surfaces.
   - `docs/index.html`
   - `docs/pantheon/ra.html`
   - any docs navigation/index pages that should expose the innovation.
   - Make the copy user-comprehensible: "Ra routes the work; Horus shows what is happening on this machine."

6. Update changelog and development log.
   - `CHANGELOG.md`
   - `docs/BUILD_LOG.md`
   - If the HTML build log mirrors the Markdown log, update `docs/build-log.html` only if local conventions show it is manually maintained.

7. Add or update case study coverage.
   - Add a focused case study under `docs/case-studies/` explaining the problem: multiple agents/windows, no unified control surface, manual shuttle fatigue, and CTR as Ra/Horus split.
   - Update `docs/CASE-STUDIES.md` and `docs/case-studies.html` if those are the repo's case study indexes.

8. Verify.
   - Run documentation grep checks proving the ownership split is present.
   - Run any available doc/site validation that does not require external services.
   - If no formal docs test exists, state that and provide the grep/file evidence.

9. Write back to the router.
   - Create a review/completion file in `.agents/idea-router/reviews/`.
   - Address it to `codex-pantheon`.
   - Include `verdict`, files changed, verification evidence, blocker status, and whether the `/goal` is met.
   - Include `eta_for_review` or `next_check_at`.

## Required Constraints

- Stay inside `/Users/thekryptodragon/Development/sirsi-pantheon`.
- Do not edit other repos.
- Do not fork the router into other repos.
- Do not remove existing canon unless replacing conflicting language with the Ra/Horus split.
- Keep the work commercially polished: no placeholder docs, no "TODO later" claims for required surfaces.
- This is not docs-only. If no code changes are needed, prove the existing code already provides the Ra/Horus local-node status surface and name the exact command.
- Preserve the existing dirty worktree. Do not revert unrelated changes, especially current TUI/router changes from other workstreams.

## Required Output

The router completion writeback must include:

- changed files
- concise summary of the innovation
- verification commands and outputs
- remaining risks, if any
- explicit `/goal met` or `/goal not met`
