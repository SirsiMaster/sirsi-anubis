# Embedding Boundary — Sirsi Pantheon Internal APIs

**Version:** 1.0.0
**Date:** May 14, 2026
**Status:** Active
**Refs:** ADR-005 (Pantheon Unification), ADR-009 (Injectable Providers), ADR-015 (Deity Hierarchy)

---

## Purpose

This document defines which Pantheon internal APIs are stable enough to embed into external products (primarily Sirsi Nexus), which should be consumed via platform boundaries (gRPC, REST), and which must not be embedded.

The `mobile/` package already demonstrates the extraction pattern: internal APIs wrapped with JSON-RPC envelopes for cross-language consumption.

---

## Embedding Tiers

### Tier 1 — Stable APIs (Same-Module or Vendored Access)

These packages have clean exported APIs, no TUI coupling, and stable type signatures. Since they live under `internal/`, they cannot be directly imported by external Go modules. Access options:

1. **Same-module adapter** (recommended): Create a `nexus/` package inside this repo that wraps internal APIs, following the `mobile/` pattern
2. **Vendoring**: Copy into Nexus's vendor tree until public packages are extracted
3. **Public extraction** (post v1.0): Move stable APIs to `pkg/scan`, `pkg/clean`, etc.

The `mobile/` package already demonstrates option 1 — wrapping internal APIs with JSON envelopes for cross-boundary consumption.

| Package | Key Exports | Use Case |
|---------|-------------|----------|
| `internal/jackal` | `NewEngine()`, `Engine.Scan()`, `ScanResult`, `Finding` | File scanning engine (81 rules, 7 domains) |
| `internal/ka` | `NewScanner()`, `Scanner.Scan()`, `Ghost`, `Residual` | Ghost app detection |
| `internal/cleaner` | `ValidatePath()`, `DeleteFile()`, `DeleteFileReversible()` | Safe deletion with protected paths |
| `internal/vault` | `Open()`, `Store.Add()`, `Store.Search()` | SQLite FTS5 context sandbox |
| `internal/scales` | `Enforce()`, `Policy`, `Verdict` | Policy evaluation engine |
| `internal/horus` | `NewGraph()`, `ParseDir()`, `Symbol`, `SymbolGraph` | Structural code graph (AST symbols) |
| `internal/rtk` | `New()`, `Filter.Apply()` | Output filtering (ANSI strip, dedup, truncate) |

**Stability guarantee:** Type signatures will not break within a minor version. New fields may be added to result structs. These packages are `internal/` — use the adapter or vendoring pattern above until public extraction.

### Tier 2 — Stable with Adapters

These packages have useful APIs but may need thin wrappers for your context.

| Package | Key Exports | Adapter Notes |
|---------|-------------|---------------|
| `internal/scarab` | `AuditContainers()`, `Container` | Container discovery; wrap with your own fleet context |
| `internal/seshat` | `ParseCommitLog()`, `CanonAssessor` | Git integration; adapt canon references to your ADR format |
| `internal/seba` | `DetectAccelerators()`, `HardwareMap` | Hardware detection only; diagram generation is not stable |

### Tier 3 — Consume via Platform Boundary

These packages require long-running processes or OS-specific integration. Consume via gRPC agent, REST API, or CLI subprocess — not direct import.

| Package | Reason | Recommended Boundary |
|---------|--------|---------------------|
| `internal/guard` | Daemon supervision, process lifecycle | gRPC agent or SSH |
| `internal/dashboard` | TUI/browser rendering | REST API for data, custom rendering |
| `internal/maat` | Runs `go test -cover` subprocess | CLI subprocess (`sirsi audit --json`) |

### Tier 4 — Do Not Embed

These packages are Pantheon-specific infrastructure. Re-implement for your platform.

| Package | Reason |
|---------|--------|
| `internal/output` | Lipgloss brand styling, TUI event loop |
| `internal/router` | Filesystem protocol for Codex/Claude collaboration |
| `internal/mcp` | MCP server with Pantheon-specific tools |
| `internal/neith` | AI prompt assembly (CLAUDE.md context injection) |
| `internal/notify` | OS-specific notification (macOS, Linux) |
| `internal/updater` | GitHub release checking |

---

## Extraction Pattern

The `mobile/` package demonstrates how to extract internal APIs:

```go
// mobile/anubis.go — wraps jackal for iOS/Android

func AnubisScan(rootPath string) string {
    engine := jackal.DefaultEngine()
    engine.RegisterAll(rules.AllRules()...)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    result, err := engine.Scan(ctx, jackal.ScanOptions{Root: rootPath})
    return marshal(result, err) // → JSON envelope: {ok, data, error}
}
```

**For Nexus:** Create a `nexus/` package inside this repo following the same pattern, but using native Go types instead of JSON (no serialization overhead for same-module access). This is the recommended extraction path until v1.0 public packages are ready.

---

## Safety Invariants

When embedding any Tier 1/2 package, these safety guarantees carry over:

1. **`cleaner.ValidatePath()`** — 29 protected paths are hardcoded and cannot be overridden
2. **`cleaner.DeleteFileReversible()`** — Errors when platform has no trash (never silently permanent-deletes)
3. **`jackal.Engine.Scan()`** — Read-only; never modifies the filesystem during scan phase
4. **`ka.Scanner.Scan()`** — Read-only ghost detection
5. **All scan operations** respect `context.Context` for timeout/cancellation

---

## Non-Goals

- Pantheon will not provide a stable Go module versioning guarantee (e.g., `go get`) until v1.0.0
- Internal package paths (`internal/`) enforce Go's access restriction — external modules cannot import them directly
- **Nexus embedding path**: Create `nexus/` adapter package inside this repo (same-module access bypasses `internal/` restriction)
- **Alternative**: Consume via CLI subprocess (`sirsi scan --json`) or MCP tools for cross-repo integration
- Public packages (`pkg/`) will be extracted after v1.0 stabilization

---

## Decision Record

| Question | Options | Decision |
|----------|---------|----------|
| How should Nexus consume scan results? | (a) Same-module adapter, (b) CLI subprocess, (c) gRPC | (a) `nexus/` adapter for Tier 1; (b) CLI for Tier 3 |
| Should we publish internal packages as public modules? | (a) Yes, (b) No, (c) After v1.0 | (c) After v1.0 — too early to commit to API stability |
| How should Nexus handle deletion safety? | (a) Own implementation, (b) Same-module adapter | (b) `nexus/` adapter wrapping cleaner — safety rules must be canonical |
| How do external products (non-Nexus) integrate? | (a) Fork, (b) CLI/JSON, (c) MCP | (b) CLI subprocess with `--json` or (c) MCP tools |
