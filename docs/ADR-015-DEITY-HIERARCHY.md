# ADR-015: Deity Hierarchy — Horus as Local Lord, Ra as Fleet Lord

## Status
**Accepted** — 2026-04-24

## Context

The deity architecture was flat — every deity was a peer module with no hierarchy.
Horus was narrowly defined as "code graph" (AST parsing, symbol outlines) despite
being named "The All-Seeing." Ra orchestrated deployments but had no reporting
arm. There was no clear answer to "who owns the local workstation?"

The v0.17.0 dashboard session exposed this: we built a localhost dashboard that
collects scan findings, ghost reports, guard alerts, doctor diagnostics, vault
search, code graph data, and Ra deployment status. That dashboard IS Horus — it
sees everything on the local machine. We just called it "Pantheon Dashboard."

Additionally, the product tier split (Anubis Free vs Ra Enterprise) maps directly
to this hierarchy: Horus = free local intelligence, Ra = paid fleet intelligence.

## Decision

### Hierarchy

```
Ra 𓇶 (Fleet Lord — Enterprise)
 ├── Receives reports from all Horus instances
 ├── Fleet-wide Neith (cross-endpoint alignment)
 ├── Aggregated Thoth/Seshat (fleet memory/knowledge)
 └── Fleet dashboard (sees all nodes)

Horus 𓂀 (Local Workstation Lord — Free)
 ├── Anubis (scan engine) → findings
 ├── Ka (ghost hunter) → ghost reports
 ├── Isis (guard/watchdog) → health alerts
 ├── Ma'at (quality gate) → governance scores
 ├── Seba (hardware/infra) → platform profile
 ├── Code Graph → symbol analysis (one capability, not identity)
 ├── Vault → context sandbox
 ├── Local Neith → scope alignment
 ├── Local Thoth → machine memory
 └── Local Seshat → machine knowledge
     ↓
     Reports up to Ra via Stele + HTTP
```

### Deity Authority Changes

| Deity | Old Role | New Role |
|-------|----------|----------|
| **Horus** 𓂀 | Code Graph (narrow) | **Local Workstation Lord** — sees and reports everything on one machine |
| **Ra** 𓇶 | Agent Orchestrator | **Fleet Lord** — receives Horus reports, orchestrates across all endpoints |
| **Neith** 𓁯 | Scope Weaver (local) | **Universal Weaver** — local scope alignment + fleet-wide consistency |
| **Thoth** 𓁟 | Session Memory | **Local Memory** — per-machine, per-repo. Ra aggregates. |
| **Seshat** 𓁆 | Knowledge Bridge | **Local Knowledge** — per-machine ingestion. Ra aggregates. |
| Anubis 𓃣 | System Jackal | Unchanged — reports to Horus |
| Ka 𓂓 | Ghost Hunter | Unchanged — reports to Horus |
| Isis 𓁐 | Health & Remedy | Unchanged — reports to Horus |
| Ma'at 𓆄 | Quality Gate | Unchanged — reports to Horus |
| Seba 𓇽 | Infra & Hardware | Unchanged — reports to Horus |
| Osiris 𓁹 | State Keeper | Unchanged — reports to Horus |

### Concrete Changes

1. **Dashboard rename**: "Pantheon Dashboard" → "Horus — Local Workstation Monitor"
2. **Dashboard sidebar brand**: "☥ Pantheon" → "𓂀 Horus"
3. **Module table in CLAUDE.md**: Horus role updated to "Local Workstation Lord"
4. **TUI deity roster**: Horus role updated from "Code Graph" to "Workstation Lord"
5. **Horus Go package**: Keeps code graph capability. Gains `Report()` that aggregates
   all local deity outputs into a single `WorkstationReport` struct
6. **Ra**: Gains `/api/fleet/report` endpoint that ingests Horus reports
7. **Neith**: Gains `AlignFleet()` that compares weaves across Horus instances

### Transport (Phase 2 — Fleet)

Horus → Ra transport for fleet mode (not implemented now, design locked):
- Each Horus periodically POSTs a `WorkstationReport` to Ra's ingestion endpoint
- Report contains: scan summary, ghost count, guard health, Ma'at score, Neith drift
- Ra stores in a fleet-level SQLite (or PostgreSQL for scale)
- Ra dashboard renders fleet view from aggregated reports

## Alternatives Considered

1. **Keep flat architecture**: Every deity remains a peer. Dashboard stays "Pantheon."
   Rejected because: the product tier split (free/enterprise) maps to local/fleet,
   and the flat model has no clear ownership of "workstation state."

2. **Ra owns everything**: Ra is the single lord, Horus stays narrow.
   Rejected because: Ra is enterprise/paid. The free tier needs a strong local
   authority. Horus is that authority.

3. **New deity for workstation**: Create a new deity (e.g., "Ptah") as local lord.
   Rejected because: Horus literally means "The All-Seeing" — it's already the
   right name. Adding a new deity increases cognitive load.

## Consequences

- **Positive**: Clear ownership (Horus = local, Ra = fleet). Product tiers map to
  deity hierarchy. Dashboard has a real identity.
- **Positive**: The code graph capability isn't lost — it becomes one tab in Horus's
  workstation view, alongside scan results, guard alerts, etc.
- **Negative**: "Horus" meant "code graph" in all prior docs and commit messages.
  Renaming requires updating docs, comments, and mental models.
- **Risk**: Over-complicating the hierarchy before fleet mode exists. Mitigated by
  only implementing Phase 1 (rename + local report struct) now.

## References

- ADR-014: Stele Ledger (Horus→Ra transport mechanism)
- ADR-011: Deity Alignment (original deity definitions)
- Rule A25: Deity Registry & Attribution
- Strategic Assessment Q8: Free (Anubis) vs Paid (Ra) differentiation
