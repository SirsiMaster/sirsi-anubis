# ADR Index — Sirsi Pantheon (Architecture Decision Records)

This index tracks **all** architectural decisions for the Sirsi Pantheon ecosystem.

**Total ADRs: 20** | **Next available: ADR-021**

---

## Master Registry

| ID | Title | Status | Date |
|----|-------|--------|------|
| [ADR-001](ADR-001-FOUNDING-ARCHITECTURE.md) | Founding Architecture — Go, cobra, agent-controller, module codenames | Accepted | 2026-03-20 |
| [ADR-002](ADR-002-KA-GHOST-DETECTION.md) | Ka Ghost Detection — 5-step algorithm, 17 residual locations, bundle ID matching | Accepted | 2026-03-20 |
| [ADR-003](ADR-003-BUILD-IN-PUBLIC.md) | Build-in-Public as Canonical Process — required release artifacts, transparency rules, dual-audience docs | Accepted | 2026-03-22 |
| [ADR-004](ADR-004-MAAT-QA-GOVERNANCE.md) | Ma'at QA/QC Governance Agent — observe/assess/weigh/report, feather weight scoring, agent prototype | Accepted | 2026-03-23 |
| [ADR-005](ADR-005-PANTHEON-UNIFICATION.md) | Pantheon Unified Platform — all deities as sub-systems, single brand, single install | Accepted | 2026-03-23 |
| [ADR-006](ADR-006-SELF-AWARE-RESOURCE-GOVERNANCE.md) | Self-Aware Resource Governance — Guard module + yield-based resource management | Accepted | 2026-03-23 |
| [ADR-007](ADR-007-UNIFIED-FINDINGS-PORTAL.md) | Unified Findings Portal — Horus as canonical aggregator for deity findings | Accepted | 2026-03-24 |
| [ADR-008](ADR-008-SHARED-FILESYSTEM-INDEX.md) | Shared Filesystem Index — Walk once, query everywhere via Horus manifest cache | Accepted | 2026-03-24 |
| [ADR-009](ADR-009-INJECTABLE-SYSTEM-PROVIDERS.md) | Injectable System Providers — standard interface injection for 99% coverage | Accepted | 2026-03-24 |
| [ADR-010](ADR-010-MENUBAR-APPLICATION.md) | Pantheon Menu Bar Application — native macOS status bar + Finder presence | Accepted | 2026-03-25 |
| [ADR-011](ADR-011-DEITY-ALIGNMENT.md) | Deity Alignment & Context Architecture — canonical scopes for all deities | Accepted | 2026-03-25 |
| [ADR-012](ADR-012-VSCODE-EXTENSION.md) | Pantheon VS Code Extension — always-on Guardian, status bar ankh, Thoth context | Accepted | 2026-03-25 |
| [ADR-013](ADR-013-TILED-CONTEXT-RENDERING.md) | Tiled Context Rendering — GPU-inspired relevance scoring, token budgets, deferred manifest | Accepted | 2026-04-05 |
| [ADR-014](ADR-014-STELE-LEDGER.md) | Stele Ledger — append-only hash-chained event log for all deity communications | Accepted | 2026-04-03 |
| [ADR-015](ADR-015-DEITY-HIERARCHY.md) | Deity Hierarchy — Horus as local workstation lord, Ra as fleet lord | Accepted | 2026-04-24 |
| [ADR-016](ADR-016-TUI-PRIMARY-INTERFACE.md) | TUI as Primary Interface — shared suggest engine, streaming, view stack, persistent state | **Superseded by ADR-018** | 2026-05-06 |
| [ADR-017](ADR-017-RA-HORUS-CTR-HYPERVISOR.md) | Ra/Horus CTR Hypervisor — multi-agent orchestration canon, ownership boundary | Accepted | 2026-05-19 |
| [ADR-018](ADR-018-NATIVE-MAC-APP.md) | Native macOS App + CLI as Pantheon's Interactive Surfaces — TUI sunset, standalone SwiftUI + menubar companion | **Partially In Force — Amended By ADR-020** | 2026-05-21 |
| [ADR-019](ADR-019-KNOWLEDGE-SUBSTRATE.md) | Knowledge Substrate — Thoth/Seba/Understand three-tool split, JSON-as-architectural-code, bidirectional sync, Hedera hypergraph direction | Accepted | 2026-05-26 |
| [ADR-020](ADR-020-INTERACTIVE-SURFACE-REOPENED.md) | Interactive Surface Reopened — Multi-Track Evaluation; closed Hybrid C (TUI first cross-platform, Mac native later) | Accepted (Hybrid C) | 2026-05-29 |

---

## Categories

### Core Architecture
- ADR-001: Founding Architecture
- ADR-005: Pantheon Unified Platform
- ADR-006: Self-Aware Resource Governance
- ADR-007: Unified Findings Portal
- ADR-010: Pantheon Menu Bar Application
- ADR-012: Pantheon VS Code Extension
- ADR-014: Stele Ledger
- ADR-015: Deity Hierarchy
- ADR-016: TUI as Primary Interface *(superseded by ADR-018)*
- ADR-017: Ra/Horus CTR Hypervisor
- ADR-018: Native macOS App + CLI *(v0.22 TUI sunset; partially in force — amended by ADR-020)*
- ADR-019: Knowledge Substrate
- ADR-020: Interactive Surface Reopened — closed Hybrid C (TUI first cross-platform, Mac native later)

### Ghost Detection & Indexing
- ADR-002: Ka Ghost Detection
- ADR-008: Shared Filesystem Index

### Quality & Governance
- ADR-004: Ma'at QA/QC Governance
- ADR-009: Injectable System Providers (Testing Architecture)

### Context Management
- ADR-013: Tiled Context Rendering

### Process
- ADR-003: Build-in-Public as Canonical Process

---

## ADR Numbering History

| Range | Status |
|:------|:-------|
| ADR-001 | Active — Founding Architecture |
| ADR-002 | Active — Ka Ghost Detection |
| ADR-003 | Active — Build-in-Public Process |
| ADR-004 | Active — Ma'at QA/QC Governance |
| ADR-005 | Active — Pantheon Unification |
| ADR-006 | Active — Resource Governance |
| ADR-007 | Active — Unified Findings Portal |
| ADR-008 | Active — Shared Filesystem Index |
| ADR-009 | Active — Injectable System Providers |
| ADR-010 | Active — Menu Bar Application |
| ADR-011 | Active — Deity Alignment |
| ADR-012 | Active — VS Code Extension |
| ADR-013 | Active — Tiled Context Rendering |
| ADR-014 | Active — Stele Ledger |
| ADR-015 | Active — Deity Hierarchy |
| ADR-016 | **Superseded** by ADR-018 — TUI as Primary Interface |
| ADR-017 | Active — Ra/Horus CTR Hypervisor |
| ADR-018 | **Partially In Force — Amended By ADR-020** — Native macOS App + CLI (v0.22 TUI sunset) |
| ADR-019 | Active — Knowledge Substrate |
| ADR-020 | Active — Interactive Surface Reopened (closed Hybrid C) |
| ADR-021+ | Next available |

> **Last updated:** May 29, 2026 — ADR-020 **accepted** as Hybrid C (TUI first cross-platform, Mac native later); ADR-018 status changed to *Partially In Force / Amended By ADR-020*. User authorization recorded via the ADR-020 closure (`docs/ADR-020-INTERACTIVE-SURFACE-REOPENED.md` § Closure).
