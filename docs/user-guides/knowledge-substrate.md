# Knowledge Substrate — Semantic Verification for Your Repo

The knowledge substrate is Sirsi's answer to a problem every codebase eventually faces: **you can no longer hold the whole thing in your head, and your AI agents can't either.** It produces a semantic graph of your repo — every file, function, class, import, and architectural layer — that you and your AI can query in milliseconds instead of re-reading thousands of lines on every session.

Today it ships as a per-repo capability built on the Understand-Anything plugin. Tomorrow it feeds into the Sirsi **hypergraph** — a cross-repo, cross-agent knowledge layer built on Hedera graph technology.

> **Three sovereignties, no overlap:**
> - **Thoth** (𓁟) owns memory and intent — *why* you decided things, *what's* next.
> - **Seba** (𓇽) owns architectural mapping — the canonical topology of the deity hierarchy.
> - **Knowledge substrate** owns semantic verification — *what literally exists* in the code, auto-derived from source.

---

## Quick Start

```bash
# One-time install (does not modify your repo)
brew install pnpm                                       # If you don't already have it
# Then in Claude Code:
/plugin marketplace add Lum1104/Understand-Anything
/plugin install understand-anything
```

```bash
# In any Sirsi repo (or any repo at all)
cd ~/Development/sirsi-pantheon
/understand                                              # Builds .understand-anything/knowledge-graph.json
```

When the run finishes, the dashboard auto-launches at `http://127.0.0.1:5173/?token=...` and your repo now has:

```
.understand-anything/
├── knowledge-graph.json    # The graph (JSON, jq-able, portable)
├── meta.json               # Commit hash + analysis timestamp
├── fingerprints.json       # Per-file structural hashes for incremental refresh
└── .understandignore       # Exclusion patterns (gitignore-style)
```

---

## What's in the graph

The graph is plain JSON. For sirsi-pantheon at commit `22ec913`:

| Layer | Files | What's in it |
|---|---|---|
| `cli-entrypoints` | 40 | All 8 binaries under `cmd/` (anubis, guard, maat, scarab, sirsi, sirsi-agent, sirsi-menubar, thoth) and their command files |
| `core-services` | 313 | The Egyptian-pantheon packages under `internal/` (jackal, ka, isis, mcp, ra, scales, scarab, seba, neith, horus, dashboard, brain, cleaner, guard, seshat, stele, thoth, ...) |
| `mobile-bindings` | 92 | `ios/`, `android/`, `mobile/` — gomobile bridge between Go core and native apps |
| `editor-extensions` | 19 | VS Code TypeScript extensions (Gemini bridge, etc.) |
| `agent-workqueue` | 209 | `.agents/idea-router/` — multi-agent work queue with markdown threads |
| `documentation` | 142 | `docs/`, root `.md` files |
| `infrastructure-cicd` | 48 | `Makefile`, `.github/workflows/`, `scripts/`, packaging |
| `configuration` | 56 | `configs/`, `go.mod`, project configs |
| `testing` | 5 | `tests/e2e`, coverage outputs |

And a 14-step pedagogical tour starting at `cmd/sirsi/main.go`, walking through the deity binaries → Jackal rules → Thoth/Ma'at → MCP server → Horus dashboard → mobile bindings → editor extensions → idea-router → build → CI.

---

## Commands

### `/understand` — build or refresh the graph

```bash
/understand                              # Smart default: incremental if fingerprints exist, full otherwise
/understand --full                       # Force full rebuild
/understand --auto-update                # Enable auto-refresh on commit (one-time toggle)
/understand --language es                # Generate summaries in Spanish (any ISO 639-1 code)
/understand /path/to/another/repo        # Run on a specific directory
```

The 7-phase pipeline runs SCAN → BATCH → ANALYZE (parallel) → ASSEMBLE REVIEW → ARCHITECTURE → TOUR → REVIEW → SAVE. On sirsi-pantheon (907 files), a full run takes ~30 minutes wall time and dispatches ~56 LLM subagents in parallel batches of 5–10. Incremental runs (after the fingerprint baseline exists) re-analyze only files whose tree-sitter structure changed since the last commit.

### `/understand-anything:understand-dashboard` — visual exploration

```bash
/understand-anything:understand-dashboard
```

Opens an interactive browser dashboard at `http://127.0.0.1:5173/?token=...`. The token is required — bookmark the full URL with the query string. Best for spatial graph exploration (zoom into layers, follow edges, inspect node summaries).

### `/understand-anything:understand-chat` — Q&A against the graph

```bash
/understand-anything:understand-chat
```

Then ask questions like:
- "What imports `internal/mcp/tools.go`?"
- "Show me the deity binaries and their entry points."
- "Which packages have no tests?"
- "Which files have changed the most across recent layers?"

This wraps the graph in a chat skill so you can query in natural language. For deterministic queries, use `jq` directly (see below).

### `/understand-anything:understand-explain <path>` — deep dive

```bash
/understand-anything:understand-explain internal/jackal/scan.go
```

Returns a deep explanation of the file *with full neighbor context* — what it imports, what imports it, what it tests / is tested by, what layer it's in, what other files share its tags. Equivalent to reading the file plus 15 minutes of `grep`-ing, in one call.

### `/understand-anything:understand-onboard` — generate an onboarding guide

```bash
/understand-anything:understand-onboard
```

Produces a written markdown guide for new team members joining the project, derived from the tour + layers + key files. Useful as a starting point for human-facing docs.

### `/understand-anything:understand-diff` — impact analysis

```bash
/understand-anything:understand-diff HEAD~5..HEAD
```

Given a git diff or PR, explains which architectural layers / components changed and what depends on them. Useful as a PR review aid.

---

## Direct JSON queries (no AI required)

The graph is plain JSON. You can `jq` it directly:

```bash
# Files in the core-services layer
jq '.layers[] | select(.id=="layer:core-services") | .nodeIds[]' \
   .understand-anything/knowledge-graph.json | head -20

# Which files import internal/mcp/tools.go?
jq '.edges[] | select(.target=="file:internal/mcp/tools.go" and .type=="imports") | .source' \
   .understand-anything/knowledge-graph.json

# All function nodes in cmd/sirsi/
jq '.nodes[] | select(.type=="function" and .filePath | startswith("cmd/sirsi/")) | .id' \
   .understand-anything/knowledge-graph.json | head

# Coverage check: which production files have no tested_by edges?
jq '
  .nodes
  | map(select(.type=="file" and (.filePath | endswith(".go")) and (.filePath | endswith("_test.go") | not)))
  | map(.id) as $prod
  | .edges
  | map(select(.type=="tested_by")) | map(.source) as $tested
  | $prod | map(select(. as $n | $tested | index($n) | not))
' .understand-anything/knowledge-graph.json
```

---

## The Thoth ↔ Substrate sync

Every Thoth-enabled repo (any directory with `.thoth/memory.yaml`) gets a Knowledge Graph block in its memory file automatically. After every `/understand` run, the block updates with the new commit hash, node/edge counts, and layer counts. A delta paragraph appends to `.thoth/journal.md` describing what changed.

This means:
- Your AI session boots with both *why* (Thoth) and *what exists* (substrate) without you doing anything.
- Cross-session continuity: a new session inherits the graph state via Thoth.
- Audit trail: the journal records every refresh as a dated event.

The rule lives in `~/CLAUDE.md` and applies to every AI session in every Thoth-enabled repo.

---

## The "switch within Sirsi overall"

The knowledge substrate is gated by a config + a feature flag so you can opt in / out per project:

```yaml
# configs/hypergraph.yaml
enabled: true                # Master switch
auto_refresh:
  on_commit: false           # Refresh after every commit (incremental, fingerprint-driven)
  on_branch_switch: true     # Refresh on long-lived branch switches
exclude:                     # Mirrors .understand-anything/.understandignore
  - "docs/case-studies/**"
output_language: en
sync:
  thoth: true                # Auto-update Thoth after refresh
  hcs_publish: false         # Future: publish to Hedera Consensus Service
```

When `enabled: false`, every `sirsi hypergraph *` command becomes a one-line no-op. When the binary is built without the `hypergraph` build tag, the subcommand is absent entirely — no runtime cost in fleet binaries that don't need it.

(CLI subcommand implementation pending. Specification in ADR-019 § 6.)

---

## The long-term vision

Today's per-repo graph is the **feeder**, not the destination. The destination is the Sirsi **hypergraph** — a cross-repo, cross-agent knowledge layer built on **Hedera Consensus Service** (HCS) as the event substrate. Memory shards, structure subgraphs, and routing events from every Sirsi repo will project into one queryable graph with consensus-ordered provenance.

You don't need to think about it today. The local JSON artifact is portable, version-controllable, and tool-agnostic — designed to be ingested by a future hypergraph projector without translation.

For the full architecture and builder guidance, see `~/Development/HYPERGRAPH_VISION.md`.

---

## FAQ

### Should I commit `.understand-anything/` to git?

Tradeoff:

- **Commit it** — teammates and CI inherit the graph without rebuilding. Diffs render meaningfully (graph deltas are readable JSON). 2.9 MB on sirsi-pantheon, ~1 MB JSON for most repos.
- **Gitignore it** — keeps repo lean; every collaborator runs `/understand` locally. Add a CI step that rebuilds and uploads as a release asset if you want shared access.

For solo developers: commit. For teams >3: gitignore + CI rebuild.

### How fresh does the graph stay?

Depends on workflow:

- **Manual** — run `/understand` when you remember. Fine for slow-changing repos.
- **Incremental on commit** — set `auto_refresh.on_commit: true`. Uses the fingerprint baseline to re-analyze only structurally-changed files (~30 seconds for typical commits).
- **CI step** — add `/understand` to your CI pipeline; commit the updated JSON.

### Does this replace Thoth or Seba?

No. Thoth keeps *why and what next*. Seba keeps *deity-owned architectural topology*. The substrate keeps *auto-derived ground truth from source*. Three different jobs.

### What happens if the Understand-Anything plugin gets abandoned?

The JSON is the contract, not the tool. If the plugin disappears, the JSON survives and can be regenerated by any replacement that targets the same schema. Worst case: fork the plugin (open source).

### Will it work on non-Sirsi repos?

Yes. The plugin and the JSON format are tool-agnostic. You can run `/understand` on any codebase. The Thoth ↔ Substrate sync only activates in repos that have a `.thoth/` directory.

### Does it phone home?

No telemetry from the plugin or from Sirsi. The dashboard runs locally on `127.0.0.1`. The graph never leaves your machine unless you explicitly publish it (e.g., commit to a public repo).

### What languages does it support for function-level extraction?

C, C++, C#, Go, Java, JavaScript, PHP, Python, Ruby, Rust, TypeScript — via bundled tree-sitter grammars. Swift and Kotlin are file-level only (parsers not bundled). Function-level extraction for those languages requires an upstream parser add.

---

## Refs

- [ADR-019 — Knowledge Substrate](../ADR-019-KNOWLEDGE-SUBSTRATE.md) — the canonical decision record
- [Thoth user guide](thoth.md) — the *why* side of the substrate
- [Seba user guide](seba.md) — the deity-owned architectural map
- `~/Development/HYPERGRAPH_VISION.md` — builder doc for the future cross-repo layer
- [Understand-Anything plugin on GitHub](https://github.com/Lum1104/Understand-Anything) — upstream source
- [docs/case-studies/2026-05-26-knowledge-substrate-day-1.md](../case-studies/2026-05-26-knowledge-substrate-day-1.md) — first-run measurements
