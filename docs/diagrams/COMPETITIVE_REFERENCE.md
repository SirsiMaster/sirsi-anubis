# Competitive Reference: siddsachar/Thoth Architecture Diagrams

## Context
In May 2026, an open-source project called "Thoth" (github.com/siddsachar/Thoth)
launched with 6 architecture diagrams that set a new standard for developer-facing
visual documentation. This file captures what they did well so we can match or
exceed it in Sirsi Pantheon's own architecture maps.

## What They Published (6 Diagrams)

1. **Thoth Core Agent Architecture** — Full system overview with 11 numbered subsystems
2. **Thoth Memory System Architecture** — Knowledge graph, storage, retrieval, dream cycle
3. **Thoth Background Workflow Architecture** — Task engine, scheduler, pipeline, delivery
4. **Thoth Multi-Channel Architecture** — Messaging gateway, identity mapping, formatting
5. **Thoth Designer Studio Architecture** — Content creation, page model, export
6. **Thoth Safety, Privacy & Control Architecture** — Trust boundaries, permissions, audit

## Visual Design Patterns (What Makes Them Good)

### Layout
- Dark background (#0A0A0A range), warm rose/gold accent lines
- Numbered subsystem boxes (circled numbers: 1, 2, 3...) for reading order
- Central "core" module in the middle, satellite subsystems around it
- Clear spatial hierarchy: inputs left/top, core center, outputs right/bottom
- Legend at bottom with arrow types and color meanings

### Typography
- Title in large serif font, centered at top
- Subtitle in italic, smaller, different color
- Subsystem headers in CAPS with accent color
- Body text in smaller sans-serif, high contrast against dark bg
- Monospace for technical values (paths, commands, sizes)

### Color Coding
- Gold/rose for primary accent (headings, borders, numbers)
- Green for success/safety states
- Red for danger/blocked states
- Muted rose for secondary information
- Each subsystem has a consistent border color

### Data Flows
- Solid arrows for primary data flow (labeled with data type)
- Dashed arrows for background/scheduled flows
- Dotted lines for trust/security boundaries
- Arrow labels describe WHAT flows, not just that something flows

### Information Density
- Each subsystem box contains 4-8 bullet points
- Bullets are icon + label + brief description (one line each)
- No prose paragraphs inside boxes — just structured lists
- Flow Summary at bottom: numbered steps 1-7 showing end-to-end

## What Sirsi Pantheon Has That They Don't

- **Named deity hierarchy** with Egyptian mythology (not generic numbered boxes)
- **Two-tier architecture** (Horus local / Ra fleet) — natural diagram split
- **Shipping product** with real metrics (81 rules, 833ms scan, 98% context reduction)
- **Case studies** backing every architectural claim
- **Patent portfolio** (6 provisionals) — IP moat they lack
- **MCP integration** as a first-class protocol (not just LangChain tools)

## Priority for Sirsi Diagrams

Match their quality on these first (highest impact):
1. Core Architecture (our Diagram 01) — this is the one people screenshot
2. Thoth Memory System (our Diagram 04) — direct name competitor, must be better
3. Local Workstation (our Diagram 05) — shows the Horus aggregation model

Then extend beyond them:
4. Fleet Architecture (our Diagram 06) — they have nothing like this
5. Token Intelligence (our Diagram 03) — unique RTK→Vault→Horus pipeline
6. Governance Cycle (our Diagram 07) — deity-based governance loop
