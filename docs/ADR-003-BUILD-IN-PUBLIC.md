# ADR-003: Build-in-Public as Canonical Process

## Status
**Accepted** — 2026-03-22

## Context
Most developer tools compete on marketing polish — landing pages, feature matrices, testimonials. The actual development process stays hidden. Bugs are quietly patched. Performance regressions are never disclosed. Users evaluate products based on branding, not engineering quality.

Anubis competes against established tools (CleanMyMac, DaisyDisk, OmniDiskSweeper) that have years of marketing investment. Matching that spend is not feasible. Instead, the development process itself becomes the differentiator — transparency as a competitive advantage.

## Decision
The build-review-revise-release cycle is a formal, canonical part of Anubis development. Every sprint must update the public record. This is not optional documentation — it is a release artifact, equal in importance to the binary and the changelog.

### Required artifacts per release
1. **VERSION** — bumped with every tagged release
2. **CHANGELOG.md** — technical changes, migration notes, breaking changes
3. **docs/BUILD_LOG.md** — sprint narrative with real numbers: what broke, what was fixed, benchmark data, test coverage, honest disclaimers
4. **docs/build-log.html** — public-facing HTML version (Swiss Neo-Deco design) accessible to non-developers
5. **.thoth/memory.yaml** — updated project state for AI session continuity
6. **.thoth/journal.md** — timestamped reasoning entry (the "why" behind decisions)
7. **ADR** — new Architecture Decision Record if the change involves a structural decision

### Process rules
- **Include mistakes**: Bugs found, CI failures, incorrect assumptions — all stay in the record. No scrubbing.
- **Include real benchmarks**: Measured data with verification commands. No synthetic tests or cherry-picked numbers.
- **Include honest gaps**: What isn't ready, what coverage is missing, what platforms are incomplete.
- **Use direct verbs**: "Built. Refactored. Fixed. Added." Never "the user wanted" or "the user suggested."
- **Dual-audience writing**: BUILD_LOG.md for developers who read GitHub. build-log.html for everyone else.
- **Cross-link**: Both SirsiNexus Portal and Anubis pages link to each other with clear messaging about the Anubis→Ra relationship.

### Voice and tone
The build log reads like engineering notes from a team you'd want to hire — confident in results, honest about gaps, precise with data. It does not read like marketing copy, feature announcements, or user stories.

## Alternatives Considered
1. **Private development, public releases only** — Standard approach. Rejected because it provides no differentiation against funded competitors and offers no way for evaluators to verify claims.
2. **Blog-style devlogs** — Common in indie gamedev. Rejected because blogs lose structure over time, aren't co-located with the code, and don't include verifiable data.
3. **Automated release notes only** — Tools like release-drafter. Rejected because auto-generated notes lack narrative, context, and honesty about what went wrong.

## Consequences
- **Positive**: Builds trust with technical evaluators who can verify claims. Creates a moat that marketing-only competitors cannot replicate. Forces engineering discipline (can't hide broken things). Gives future team members full context on every decision.
- **Positive**: The HTML build log serves as a landing page for non-technical stakeholders (investors, partners, prospective users) who don't read GitHub.
- **Negative**: Requires discipline every sprint to update multiple artifacts. Adds ~15 minutes per release cycle.
- **Risk**: Competitors could study the build log to understand our approach. Accepted — the execution advantage outweighs the information leak.

## References
- [BUILD_LOG.md](BUILD_LOG.md) — Sprint chronicle
- [build-log.html](build-log.html) — Public HTML version
- [ADR-001](ADR-001-FOUNDING-ARCHITECTURE.md) — Founding architecture
- [Thoth Specification](THOTH.md) — Knowledge system that enables session continuity
