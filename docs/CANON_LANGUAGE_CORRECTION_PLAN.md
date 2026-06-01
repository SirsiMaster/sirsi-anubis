# Canon Language Correction Plan — Failed Implementation Removed ≠ Future TUI Abandoned

**Governing ADR:** ADR-020
**Status:** Plan — no edits applied yet. Codex review then implementation.

## Distinction To Preserve Everywhere

| Wrong wording | Right wording |
| :--- | :--- |
| "Interactive TUI eliminated" | "v0.22 BubbleTea TUI implementation removed (unreleasable); TUI as a surface remains under evaluation per ADR-020." |
| "TUI sunset" | "v0.22 TUI sunset; surface category under reopened evaluation." |
| "TUI eliminated v0.23 per ADR-018" | "v0.22 TUI implementation removed in v0.23 per ADR-018; surface decision reopened per ADR-020." |
| "no TUI" (as future-tense claim) | "no v0.22 TUI" or just drop the claim — let ADR-020 speak. |
| "Pantheon's interactive surface moves from TUI to native SwiftUI" | "v0.22 TUI removed as foundation; native macOS SwiftUI is one track under multi-track evaluation per ADR-020." |

The factual deletion stays in the record (the code was removed, the binary shrank). The strategic claim that Sirsi abandons the TUI surface category is corrected — that was a step ADR-018 took beyond evidence and is rescinded by ADR-020.

## File-By-File Edit Plan

### `CHANGELOG.md` (the `[Unreleased]` block)

Current wording uses "Interactive TUI eliminated" as the headline `### Removed` bullet, with reasoning that frames the elimination as intentional surface-category abandonment.

**Edit:**
- Keep the factual deletion claim: ~4,800 LOC removed, `charm.land/bubbletea/v2` dropped, binary shrank 24.2 MB → 22.2 MB.
- Reframe the headline: `### Removed — v0.22 TUI implementation` (not `### Removed — Interactive TUI`).
- Add a `### Reopened` block (Keep-a-Changelog allows it): "Surface decision reopened per ADR-020 (2026-05-29). v0.23 ships without the v0.22 TUI; the future surface direction is under evaluation."
- Remove the line "**This was intentional and immediate** — a broken or brand-damaging surface should not remain reachable behind a flag…" Keep the spirit (the deletion was correct) but drop the implication that the surface category was abandoned.

### `docs/CLI_COMPATIBILITY.md`

- Update the §"What changed vs. what only moved" "Removed entirely" line from "the interactive TUI gateway" to "the v0.22 BubbleTea TUI gateway."
- Update the closing reference to ADR-018 to also cite ADR-020.

### `docs/ADR-018-NATIVE-MAC-APP.md`

- Add an `## Amended By` section pointing at ADR-020.
- Update the `Status` from `Accepted` to `Partially In Force / Amended By ADR-020`.
- Keep the body unchanged — it's a historical record of the 2026-05-21 decision. ADR-020 is the corrected interpretation.

### `docs/ADR-INDEX.md`

- Update ADR-018's row: status changes from "Accepted" to "Partially In Force — Amended By ADR-020."
- Add ADR-020 row: "Interactive Surface Reopened / Multi-Track Evaluation — Proposed — 2026-05-29."
- Update the header total: 19 → 20 ADRs.

### `docs/ADR-001-FOUNDING-ARCHITECTURE.md`

Current reads: `excellent CLI ecosystem (cobra, lipgloss; bubbletea was used through v0.22 and removed in v0.23 per ADR-018), contributor-friendly`.

**Edit:** append ADR-020 to the citation: `…removed in v0.23 per ADR-018; surface decision reopened per ADR-020`.

### `docs/diagrams/05-local-workstation.mmd`

Current node label: `CLI["sirsi CLI<br/>Cobra · lipgloss (TUI removed v0.23)"]`.

**Edit:** `CLI["sirsi CLI<br/>Cobra · lipgloss (v0.22 TUI removed v0.23; surface TBD per ADR-020)"]`. Lighter touch: drop the parenthetical entirely and let ADR-020 carry it.

### `PANTHEON_RULES.md` (canonical source for AGENTS/CLAUDE/GEMINI)

Current tech-stack row:
> `| Interactive Surface | Native macOS SwiftUI app (planned) + CLI on all platforms | TUI eliminated v0.23 per ADR-018. Mac native app is the interactive surface; Windows/Linux are CLI-only. |`

**Edit:**
> `| Interactive Surface | Under multi-track evaluation per ADR-020 | v0.22 BubbleTea TUI removed in v0.23 per ADR-018. Surface direction (TUI / Mac native / hybrid) under evaluation; Mac native is one track among several. CLI ships on all platforms regardless. |`

### `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`

These mirror `PANTHEON_RULES.md`'s tech-stack row. Same edit, applied to each. If sync tooling is in place, the canonical edit to `PANTHEON_RULES.md` propagates; if not, three separate matching edits.

### `README.md`

If any user-facing surface claim references "TUI eliminated" or "Mac-native-only interactive surface," same kind of correction. (Verify scope at edit time — README content drift may require a fuller pass.)

## Files Explicitly Out of Scope For This Plan

- `internal/output/tui*.go` — already deleted; not restored by canon-language correction.
- `docs/case-studies/tui-controller-refactor.md`, `tui-predictions-sekhmet-network.md` — historical case studies of work that did happen. Leave intact.
- `docs/seba.html` — auto-generated graph from before the rename; not load-bearing for surface claims.
- `.agents/idea-router/` — the audit trail of decisions stays intact; the corrections live in canon, not in router history.
- Phase-1 audit docs — re-scoped per `docs/PHASE1_RESCOPE_NOTE.md` with header notes, not body rewrites.

## Edit Sequencing Recommendation

When codex acks this plan and the user picks a surface track:

1. `docs/ADR-020-INTERACTIVE-SURFACE-REOPENED.md` flips from Proposed → Accepted.
2. ADR-018 status updates (Partially In Force / Amended).
3. ADR-INDEX update.
4. `PANTHEON_RULES.md` tech-stack row (canonical).
5. CHANGELOG `### Reopened` block.
6. CLI_COMPATIBILITY, ADR-001 citation, diagram, README — small touches.
7. AGENTS/CLAUDE/GEMINI propagation from PANTHEON_RULES.md.

All in one canon-correction commit, separately from any code work that follows the surface pick.

## Files Touched In This Plan

8 canon/doc files. ~50 LOC of edits total. No code.

## /goal

Codex ack of the file list, the edit strings, and the sequencing recommendation. On ack — and after the user picks a surface track — apply as a single canon-correction commit.
