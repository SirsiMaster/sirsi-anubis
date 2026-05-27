# Case Study: Day One With The Knowledge Substrate

> **We installed Understand-Anything on sirsi-pantheon at lunchtime. By dinner we had a 3,340-node semantic graph, a hypergraph vision document committed to the workspace canon, and a bidirectional sync wired between Thoth and the graph. Here's exactly what happened and what we learned.**

---

## The Problem

By May 2026, sirsi-pantheon had grown to:
- 367 Go files across 41 modules in `internal/`
- 8 binaries under `cmd/`
- 33 Swift files (iOS), 24 Kotlin files (Android), 17 TypeScript files (VS Code extensions)
- 282 markdown documents (ADRs, case studies, user guides, agent threads)
- 907 total files across 19 languages and 6 detected frameworks

When a new AI session asked "what imports `internal/mcp/tools.go`?", the answer required reading 367 Go files. That's ~80,000 lines of Go alone before the AI could form an opinion. Thoth's 100-line compressed memory helped with *why* questions but couldn't answer structural ones.

**The gap:** a semantic ground-truth layer derived directly from source. Not curated. Not Thoth-shaped. Just *what literally exists*.

---

## The Install

```bash
brew install pnpm          # Wasn't installed; pnpm 11 default-denies postinstall scripts
```

```
/plugin marketplace add Lum1104/Understand-Anything
/plugin install understand-anything
/reload-plugins
```

The plugin (`Lum1104/Understand-Anything` v2.7.5) ships as a Claude Code plugin with skills, agents, and a JavaScript/TypeScript workspace under `packages/{core,dashboard}`. First-time setup required:

1. `pnpm install --config.dangerouslyAllowAllBuilds=true` — pnpm 11 defaults to blocking postinstall scripts; tree-sitter parsers and esbuild both need them. Took ~6 seconds and compiled 12 native tree-sitter grammars.
2. `pnpm --filter @understand-anything/core build` — TypeScript compile of the shared library. ~30 seconds.

**Total install time: under 90 seconds wall clock.**

---

## The First Run

```
/understand /Users/thekryptodragon/Development/sirsi-pantheon
```

The skill runs a 7-phase pipeline:

| Phase | Wall time | What happened |
|---|---|---|
| 0 — Pre-flight | <1s | Plugin root resolved, git commit captured (`22ec913`) |
| 0.5 — `.understandignore` | <1s | Starter file generated from `.gitignore` + detected dirs. No exclusions activated (full polyglot pass) |
| 1 — SCAN | ~56s | 907 files indexed, 486 code files, 19 languages, 6 frameworks detected (Cobra, Lipgloss, SQLite, systray, Makefile, GitHub Actions), **2,277 internal import edges resolved by static analysis** |
| 1.5 — BATCH | <2s | 56 semantic batches via Louvain community detection on the import graph (sizes 3–32 files) |
| 2 — ANALYZE | ~12 minutes | **56 parallel LLM subagent dispatches** (sliding 5–10 concurrent). Each subagent reads its batch's files, emits GraphNodes (file, function, class) and GraphEdges (imports, contains, exports, calls, tested_by) |
| 3 — ASSEMBLE REVIEW | ~2 minutes | Merge step + path-convention `tested_by` linker. 14 duplicate IDs collapsed; 17 `step:` → `pipeline:` prefix normalizations. Cross-batch import edge verification: all 2,277 import edges from the static-analysis import map present in the assembled graph |
| 4 — ARCHITECTURE | ~1.5 minutes | 9 architectural layers identified, all 924 file-level nodes assigned exactly once |
| 5 — TOUR | ~2 minutes | 14-step pedagogical tour starting at `cmd/sirsi/main.go` |
| 6 — REVIEW | <2s | Inline deterministic validator. 0 issues, 252 orphan warnings (markdown docs with no edges — expected) |
| 7 — SAVE | <3s | `knowledge-graph.json` (2.9 MB), `meta.json`, `fingerprints.json` (907 baseline hashes) |

**Total wall clock: ~30 minutes.**

The dashboard auto-launched at `http://127.0.0.1:5173/?token=...`. Vite + React + a static JSON loader. No backend, no telemetry, no cloud sync.

---

## The Numbers

The final graph:

| Metric | Value |
|---|---|
| **Total nodes** | 3,340 |
| File nodes | 496 |
| Function nodes | 2,108 |
| Class nodes | 308 |
| Document nodes | 353 |
| Config nodes | 51 |
| Pipeline nodes | 23 |
| Service nodes | 1 |
| **Total edges** | 6,947 |
| `contains` edges | 2,433 |
| `imports` edges | 2,279 |
| `exports` edges | 1,816 |
| `tested_by` edges | 128 |
| `related` edges | 126 |
| `depends_on` edges | 70 |
| `calls` edges | 48 |
| **Layers** | 9 |
| **Tour steps** | 14 |
| **Artifact size** | 2.9 MB JSON |
| **Orphan warnings** | 252 (docs/configs with no edges — expected) |

---

## What Surprised Us

### 1. Per-batch LLM dispatch is the real cost

Phase 2 (Analyze) was 80% of total wall time. Each batch is a fresh Sonnet call reading 3–32 source files and emitting structured JSON. Tuning the concurrency from 5 → 10 cut Phase 2 by ~40% with no quality loss.

The 5-concurrent guideline in the skill instructions is an artificial floor, not a ceiling. With background dispatches and notification-driven progression, 10–12 concurrent worked fine. The bottleneck was per-batch latency, not parallelism.

### 2. Polyglot ratios shape graph density

367 Go files produced 2,108 function nodes. 282 Markdown files produced 0 function nodes. The call/import graph is heavily Go-weighted — which matches reality (Go is Pantheon's core) but means architectural-layer queries dominated by markdown look sparse on edges.

The 252 orphan warnings are all in the documentation layer. Not a defect; documentation is leaf data, not a participant in the import graph.

### 3. Swift / Kotlin coverage is file-level only

Tree-sitter Swift and tree-sitter-kotlin grammars aren't bundled in Understand-Anything v2.7.5. Per-function / per-class extraction is missing for iOS (33 files) and Android (24 files). The graph still captures their file relationships and architectural-layer assignment, but function-level call graphs for those languages will need an upstream parser add.

We logged this as a verification gap, not a blocker.

### 4. The graph caught duplicates and normalization gaps the merge step had to fix

The merge script (`merge-batch-graphs.py`) reported:
- 14 duplicate node IDs across batches (kept last occurrence)
- 17 nodes used `step:Makefile:build` instead of `pipeline:Makefile:build` (normalized)
- 20 `tested_by` edges added via path-convention pairing (test↔production by filename)
- 10 `imports` edges recovered from the static import map that LLM analysis missed

Without these post-processing passes, the graph would have shipped with broken layer assignments and missing test coverage edges.

---

## What We Did Next (Same Day)

### 1. Three-tool clarification

The graph forced a question: how does this relate to Thoth (memory) and Seba (deity-owned architectural mapping)?

Resolution, codified the same evening:

- **Thoth** owns *why* (intent, decisions, sprint state)
- **Seba** owns *deity-canonical topology* (architectural mapping sovereignty per ADR-015 / Rule A25)
- **Knowledge substrate** owns *what literally exists* (auto-derived semantic verification)

No overlap. Three sovereignties, three artifact types, three audiences.

### 2. Bidirectional Thoth ↔ Substrate sync

Added a `## Knowledge Graph (Understand-Anything)` section to `.thoth/memory.yaml` with the current stats, query commands, and an explicit `sync_protocol` block. Appended a journal entry for 2026-05-26. Added a sync rule to `~/CLAUDE.md` so future AI sessions in any Thoth-enabled repo update both files automatically after every `/understand` run.

### 3. Hypergraph vision document

User framing (verbatim): *"the reason this works is that there will be an agent knowledge graph called the hypergraph in the future based on Hedera graph technology... understand just maps my repo."*

This reframed the entire local layer as **a feeder for a future cross-repo, cross-agent knowledge layer** built on Hedera Consensus Service. The per-repo `.thoth/` and `.understand-anything/` artifacts are local truth-anchors that will project into a unified hypergraph.

Captured at workspace root in `~/Development/HYPERGRAPH_VISION.md`. Pointer added to `~/Development/AGENTS.md` § Knowledge Substrate so every future builder agent discovers it at session boot.

### 4. JSON-as-architectural-code preference codified

User framing: *"i also like the reliance on json which is machine readable architectural code."*

Saved as `feedback_json_arch_code.md` in agent memory. Rule: every cross-tool / cross-agent artifact defaults to JSON. Anti-patterns: unsafe language-native serializations, proprietary binary formats, YAML anchors that break JSON round-trip, "we'll convert later."

This is now the design constraint that shaped ADR-019 § 3 and the future hypergraph schema.

---

## Measured Impact

Too early for hard numbers — first run was today. The qualitative claims to verify in subsequent sessions:

| Hypothesis | How we'll measure |
|---|---|
| AI session boot reads less source | Compare tokens-consumed-before-first-edit across the next 10 sessions vs. prior baseline |
| Structural questions get faster answers | Time-to-answer on "what imports X" / "what's in layer Y" before vs. after |
| Cross-session continuity improves | Count of "where were we?" re-onboarding turns at session start |
| Onboarding new collaborators speeds up | Time-to-first-merged-PR for next contributor |

Will report in a follow-up case study after one week and one month of use.

---

## Lessons For Other Sirsi Repos

When indexing the next repo (likely SirsiNexusApp):

1. **Install pnpm first.** Don't make the first AI session debug a missing prerequisite.
2. **Start with `/understand` on `HEAD`, not your working tree.** Uncommitted changes are not reflected. Commit or stash first.
3. **For polyglot repos, accept Swift/Kotlin will be file-level.** Don't gate the run on full coverage — that's an upstream-parser problem.
4. **Decide commit policy before the first run.** `.understand-anything/` is 1–3 MB JSON. Commit for solo dev; gitignore + CI rebuild for teams.
5. **Bump concurrency from 5 → 10 for Phase 2** if your machine has the headroom. Significant wall-time savings.
6. **Add the Knowledge Graph block to `.thoth/memory.yaml` the same session.** Don't let the artifact drift from the memory.

---

## What's Different Now

Before today, Sirsi had two formal knowledge artifacts per repo (Thoth memory + journal) and one informal one (tribal knowledge in agent memories + router items). Both required human or AI curation to stay current.

After today, sirsi-pantheon has three: memory, journal, **and** an auto-derived semantic graph. The third one regenerates on demand and never goes stale relative to the code.

More importantly, we now have a *direction* — a workspace-canon hypergraph vision document that says "this local layer is the feeder, not the destination" and a substrate (Hedera Consensus Service) named as the future event log.

When Sirsi Nexus eventually builds the hypergraph projector, the local artifacts produced today will be ingestable without translation. That's the bet, and it's now documented.

---

## Refs

- [ADR-019 — Knowledge Substrate](../ADR-019-KNOWLEDGE-SUBSTRATE.md)
- [Knowledge Substrate user guide](../user-guides/knowledge-substrate.md)
- `~/Development/HYPERGRAPH_VISION.md` — builder vision (workspace-canon)
- `~/Development/AGENTS.md` § Knowledge Substrate
- `~/.claude/projects/.../memory/project_hypergraph_vision.md` — claude-side memory
- `~/.claude/projects/.../memory/feedback_json_arch_code.md` — JSON preference
- [Understand-Anything plugin (upstream)](https://github.com/Lum1104/Understand-Anything)
- [Sirsi Pantheon GitHub](https://github.com/SirsiMaster/sirsi-pantheon)

---

**Commit reference:** v0.22.0-beta — ADR-019 accepted, knowledge-substrate docs landed.
