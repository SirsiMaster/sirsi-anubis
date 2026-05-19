# Decision: Net Goal-Weaving Architecture Doctrine

- author: codex
- status: active
- topic: portfolio-architecture
- created_at: 2026-05-18
- custodian: Net 𓁯, The Weaver

## Decision

Net 𓁯 is the portfolio goal-weaving standard. Net keeps every repo aligned to its full `/goal`, product surface target, phase plan, language choices, data architecture, and completion evidence.

Agents must not treat simplification as a reason to use weak primitives. The portfolio standard is simpler vendor surface area with robust architecture by design.

## Rules

- Prefer GCP-native, Firebase-native, or open-source tooling before paid third-party services.
- Prefer PostgreSQL/Cloud SQL/AlloyDB for canonical business truth: contracts, payments, subscriptions, ledgers, probate, estates, deals, audit trails, and reporting.
- Use Firestore for real-time UX state, notifications, collaboration, presence, and denormalized dashboard views.
- Use Cloud Storage for files, media, generated PDFs, evidence packets, exports, and user-owned artifacts.
- Use SQLite for local CLI/desktop state, router ledgers, caches, and offline indexes.
- Do not add a paid vendor without an ADR covering why GCP/open-source is insufficient, what data leaves Sirsi/client control, cost risk, exit path, and replacement plan.
- Do not reopen language choices without a measured blocker and written ADR.
- Agents must include `eta_for_review`, `next_check_at`, or `estimated_duration` in router handoffs.
- Agents should work independently toward the full `/goal` whenever their repo-scoped work can proceed without waiting.

## Surface Targets

- FinalWishes: web and mobile.
- Assiduous: web and mobile.
- Porch & Alley: web and mobile.
- Sirsi Nexus: native web, native mobile app, and native desktop app.
- Pantheon: local desktop and web.
- Homebrew tools: distribution only.

## Language Freeze

- Go: APIs, CLIs, daemons, routers, validators, Cloud Run services, and explicit domain engines.
- TypeScript/React: web UX, dashboards, admin surfaces, Firebase-facing UX.
- React Native/Expo: mobile UX and device flows unless native constraints require modules.
- Rust: native desktop, hardened local agents, media/edge runtimes, OS-adjacent secure storage, and high-performance local runtimes when justified.

This is a freeze against aesthetic rewrites, not against evidence-based change.
