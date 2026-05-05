# Pantheon Architecture Diagrams

Visual architecture maps for Sirsi Pantheon. Each `.mmd` file is a Mermaid source
that should be rendered as a high-quality SVG or PNG in the Pantheon brand style.

## Diagram Index

| File | Title | Scope |
|------|-------|-------|
| `01-core-architecture.mmd` | Pantheon Core Architecture | Full system overview — all deities, data flows, tiers |
| `02-deity-hierarchy.mmd` | Deity Hierarchy & Governance | Ra → Net → Code Gods + Machine Gods |
| `03-token-intelligence.mmd` | Token Intelligence Pipeline | RTK → Vault → Horus reduction flow |
| `04-thoth-memory.mmd` | Thoth Memory System | Three-layer persistent memory architecture |
| `05-local-workstation.mmd` | Horus — Local Workstation Lord | Everything on one machine, all modules reporting to Horus |
| `06-fleet-architecture.mmd` | Ra — Fleet Architecture | Multi-node fleet with Horus instances reporting to Ra |
| `07-governance-cycle.mmd` | Governance Cycle | Net → Execute → Ma'at → Isis → Thoth → Net loop |
| `08-scan-pipeline.mmd` | Scan & Remediation Pipeline | Anubis scan → findings → policy → judge → clean |

## Brand Style Guide (for rendering)

- **Background**: `#020C08` (deep emerald-black) or `#0A0A0A` (pure black)
- **Primary accent**: `#C8A951` (gold) — headings, borders, highlights
- **Secondary accent**: `#10B981` (emerald) — status indicators, success states
- **Danger/alert**: `#EF4444` (red)
- **Warning**: `#EAB308` (yellow)
- **Body text**: `#F0EDE5` (papyrus white)
- **Dim text**: `rgba(240,237,229,0.6)`
- **Card backgrounds**: `rgba(255,255,255,0.03)` with `1px solid rgba(255,255,255,0.06)`
- **Heading font**: Cinzel (serif, uppercase, tracking 0.08em)
- **Body font**: Inter (sans-serif, weight 300-500)
- **Mono font**: JetBrains Mono
- **Module boxes**: Rounded 16px, subtle gold or emerald top-border accent
- **Numbered subsections**: Gold circled numbers (like the competitor's diagrams)
- **Flow arrows**: Solid for primary data, dashed for background/scheduled, dotted for trust boundaries
- **Legend**: Bottom of each diagram with arrow types and color meanings

## Rendering Instructions

These Mermaid files are source material, not final output. The goal is to produce
diagrams at the quality level of the Thoth (siddsachar) project architecture maps:
polished, dark-background, color-coded subsystem boxes with labeled data flows.

**Recommended workflow:**
1. Render Mermaid → SVG as a structural starting point
2. Import into Figma or re-render with a custom dark theme
3. Apply Pantheon brand tokens from the style guide above
4. Export as SVG (for web) and PNG @2x (for docs/deck)

**Target locations for rendered diagrams:**
- `docs/diagrams/rendered/` — final SVGs and PNGs
- `docs/pantheon/` — embedded in the live website (sirsi.ai/pantheon)
- Pitch deck — selected diagrams for investor materials

## Competitive Reference

See `docs/diagrams/COMPETITIVE_REFERENCE.md` for analysis of the siddsachar/Thoth
project's architecture diagrams — what they did well and what to match or exceed.
