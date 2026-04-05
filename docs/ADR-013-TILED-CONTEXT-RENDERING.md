# ADR-013: Tiled Context Rendering

**Status:** Accepted
**Date:** 2026-04-05
**Deciders:** Cylton Collymore
**Categories:** Core Architecture, Context Management

---

## Context

Neith's `LoadCanon()` reads all project canon documents at full text — CLAUDE.md, ADRs, planning docs, Thoth memory, journal, changelog. For production applications with extensive canon, this produces ~254K tokens of context per session. When Ra deploys autonomous agents, each sprint loads the full canon, compounding the overhead.

Thoth's persistent memory eliminates the re-read tax for subsequent sessions within a sprint, but the first session and every Ra autonomous sprint still pay the full canon cost. We needed a way to load context intelligently — not less context, but the right context.

## Decision

Implement **tiled context rendering** in Neith's Loom, inspired by GPU deferred rendering pipelines. The full canon is loaded (geometry pass), split into semantic chunks (tiling), scored for relevance (z-test), and only relevant chunks are woven into the prompt (fragment shading). Deferred chunks appear in a manifest so the agent can fetch on demand.

### Architecture

```
LoadCanon()    → CanonContext     [geometry pass — all data loaded]
ChunkCanon()   → []CanonChunk    [tiling — semantic units]
ScoreChunks()  → []ScoredChunk   [z-test — multi-signal scoring]
TilePrompt()   → TileResult      [shading — fill budget, build manifest]
WeaveScope()   → prompt          [compositing — sections + manifest]
```

### Scoring Signals

1. **Structural weight** — Identity docs, memory, continuation prompt score 1.0 (always visible, never deferred)
2. **Keyword match** — Scope-of-work terms matched against chunk content
3. **Temporal proximity** — Recent journal/changelog entries score higher, linear decay over 90 days
4. **Coverage detection** — Chunks whose content is already represented in higher-scored chunks get halved scores (anti-overdraw)

### Token Budget

Auto-detected from total canon size:
- Canon < 50K tokens → no tiling (everything fits)
- Canon 50K–200K → 80K budget
- Canon > 200K → 60K budget

Override per scope via `token_budget` field in scope YAML config.

## Consequences

### Positive
- First session context drops from ~254K to ~72K tokens (72% reduction)
- Thoth memory stays flat (tiler replaces stale context, no unbounded growth)
- Waste trajectory shifts from increasing to declining across sessions
- Manifest preserves full canon awareness — nothing is permanently discarded
- 94% token efficiency vs. 83% without tiling, 36% vanilla

### Negative
- Additional computation during WeaveScope (chunking + scoring) — negligible for Go
- Scoring accuracy is fuzzy, not binary like GPU z-tests — mitigated by manifest safety net
- New ScopeConfig field (token_budget) adds configuration surface — defaults to auto-detect

### Neutral
- LoadCanon() unchanged — Neith still sees everything
- Existing tests unaffected — small test fixtures fall below tiling threshold
- Section ordering in woven prompt is preserved

## Alternatives Considered

1. **Static truncation** — Window journal to last N entries, changelog to N versions. Simpler but discards content permanently, no manifest, no relevance scoring.
2. **Manual context_depth presets** — full/standard/minimal static configs. Moves the decision to the user instead of evaluating it programmatically.
3. **No change** — Accept the 254K overhead. Rejected because Ra autonomous sprints compound this cost across every sprint.

## References

- `internal/neith/tiler.go` — Tile Engine implementation
- `internal/neith/loom.go` — WeaveScope integration
- GPU Tiled Deferred Rendering — the architectural inspiration
