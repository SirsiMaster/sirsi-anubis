# Pantheon Interactive Surfaces — Rewrite/Update Master Plan (DRAFT for review)

| Field | Value |
| :--- | :--- |
| **Status** | DRAFT — routed to codex-pantheon + claude-home for vetting before ANY execution |
| **Author** | claude-pantheon (lane: `pantheon-mac-native-cli-pivot`) |
| **Date** | 2026-06-01 |
| **Governs** | [ADR-020](ADR-020-INTERACTIVE-SURFACE-REOPENED.md) (ladder + surfaces-are-router-threads), A27 |
| **Reviewers** | codex-pantheon (sequencing/architecture), claude-home (version-contract lane + drift) |

## 0. Thesis — one core, four surfaces (not four apps)

This is **convergence onto one contract**, not four rewrites. `internal/dashboard` (9 files) is the spine; every surface is a thin renderer + a router thread (A27). No surface forks core logic. Grounded snapshot (2026-06-01):

| Rung | Surface | Code today | State |
| :--- | :--- | :--- | :--- |
| 1 | **CLI** | 26 top-level verbs | shipped; binary `v0.21.0` vs repo `0.22.0-beta` (drift) |
| 2 | **menubar** | 1,063 LOC Go (`fyne.io/systray`) | Step 1 (router-registration) DONE `543e959`; actions still Terminal-spawns |
| 3 | **TUI** | 1,607 LOC scaffold (`internal/tui/`, 15 files) | design-proof cleared; NO launch path, NO dashboard wiring |
| 4 | **SwiftUI** | `ios/` 35 `.swift` files (iOS app) | no macOS app yet; iOS app exists as reuse source |

**The single most important precondition: freeze + complete the `internal/dashboard` contract first.** Every rung consumes it. If it churns, all four surfaces churn. Contract-first, surfaces-second.

## 1. Per-rung plan

### Rung 1 — CLI (foundation; maintenance + version contract)
- **Target:** stays the canonical surface; every other surface's actions must map to a real CLI verb (no orphan UI actions).
- **Work:** (a) **version contract** — *owned by claude-home, plan `143914`* (fixes the `v0.21.0` drift; `internal/version`, `sirsi doctor`, `self-update`). I do not duplicate this. (b) **verb-coverage audit** — enumerate every menubar/TUI action and confirm a backing verb + dashboard endpoint exists.
- **Owner:** claude-home (version) + claude-pantheon (audit). **Mostly done/maintenance.**

### Rung 2 — menubar (Step 2: functional, not a launcher)
- **Target:** every menu action produces an **in-surface result via the dashboard/result contract** — views open the running dashboard; operations execute and post a `notify` result — **no `spawnTUIWithCommand` Terminal launches**.
- **Work:** replace the ~17 `spawnTUIWithCommand(...)` cases in `cmd/sirsi-menubar/main.go` event loop with dashboard-contract calls. Codex constraints: no `internal/tui/`, no frequent heartbeat, no new long-running loop.
- **BLOCKER:** claude-home has **uncommitted version-contract edits in `cmd/sirsi-menubar/main.go`** (+ all `cmd/*/main.go`). Step 2 cannot start until that is committed (router item `20260601-153803`). 
- **Owner:** claude-pantheon. **Approved by codex (item `145331`).**

### Rung 3 — TUI (scaffold → first runnable cut)
- **Target:** wire the `internal/tui/` scaffold to the in-process dashboard contract → first reachable, Mole-grade cut → `v1.0-alpha.0`.
- **Work:** (a) launch path (`sirsi tui` or no-args opt-in — TBD, needs `docs/CLI_COMPATIBILITY.md` review); (b) bind the 3 proof screens (scan / Ra fleet / router inbox) to live dashboard data; (c) keyboard model + command palette from the design proof. **No resurrection of v0.22.**
- **Gate:** Mole-grade quality bar (`PHASE1_MOLE_INSPECTION.md`); codex review of the first cut.
- **Owner:** claude-pantheon. Dependency: dashboard contract frozen (§0).

### Rung 4 — SwiftUI (native macOS app; absorbs menubar)
- **Target:** native macOS app + `MenuBarExtra` that **absorbs rung-2 menubar** (built once in spirit). Reads dashboard contract via `PantheonBridge.swift` over a gomobile `xcframework` (the iOS app's pattern — `ios/` is the reuse source).
- **Work:** Phase-1 reuse audits already exist (`PHASE1_IOS_REUSE_AUDIT.md` etc.) as deferred records. Activate them; stand up `macos/` target; port the menubar's *behavior* (defined in rung 2) to SwiftUI.
- **Gate:** activates **only after** TUI clears its v1 quality bar.
- **Owner:** TBD (likely a dedicated Swift lane; see §3).

### (Future) IDE plugins
- Editor-embedded surface parallel to menubar; registers as `vscode`/`jetbrains`/`cursor` router threads; consumes dashboard contract. Out of scope for this plan's execution window; tracked for after rung 4 stabilizes.

## 2. Sequencing (dependency-ordered)

```
0. claude-home COMMITS version-contract tree            (unblocks rung-2 main.go) ← critical path
1. FREEZE internal/dashboard contract  (gap audit)      (unblocks rungs 2,3,4)
2. Rung 2: menubar Step 2  (codex-approved)             ‖ can parallel rung 3 (different files)
3. Rung 3: TUI scaffold → first runnable cut            ‖ can parallel rung 2
4. Rung 4: SwiftUI macOS app (absorbs menubar)          (after rung 3 v1 quality gate)
```

Steps 2 and 3 touch **disjoint packages** (`cmd/sirsi-menubar/` vs `internal/tui/`), so per A26 they can run as **parallel repo-segmented lanes** once step 1 lands.

## 3. Agent / subagent proposal (FOR YOUR REVIEW — nothing spawned yet)

Per the directive: I present agent use here for codex + claude-home to vet **before** I spawn anything.

- **Phase A — `dashboard-contract-audit` (1 agent, read-only):** enumerate every menubar/TUI/CLI action and map → dashboard endpoint; output the gap list that defines the contract freeze (§0/step 1). Low risk, read-only. **Recommend approve.**
- **Phase B — parallel rung implementation (2 agents, worktree-isolated):** once contract frozen + claude-home committed: one agent on rung-2 menubar Step 2, one on rung-3 TUI wiring. Disjoint files (A26-clean). Worktree isolation prevents working-tree collisions like the one we just hit. **Recommend approve *conditionally* (only after step 0+1).**
- **Phase C — SwiftUI lane:** a dedicated Swift agent for rung 4. **Defer** — propose separately when rung 3 clears.
- **NOT proposed:** any agent that edits `cmd/*/main.go` while claude-home's version-contract work is uncommitted. Hard no until committed.

I will use the **Workflow** harness only if you (the user) explicitly opt in; otherwise individual **Agent** tasks, lane-segmented.

## 4. Review asks

**codex-pantheon:** (1) Endorse contract-freeze-first sequencing? (2) OK to run Phase-A audit agent now (read-only)? (3) Any objection to parallel rung-2/rung-3 lanes post-unblock?

**claude-home:** (1) When will the version-contract tree (`cmd/*/main.go`, `internal/version`, `.goreleaser.yaml`, Makefile) be committed? That's the critical-path blocker for rung 2. (2) Confirm ownership split on `143914` (you 1-2, me review+3-5?). (3) Should each surface surface its own `version --json` (you've already added it to menubar) — i.e., is the version contract part of the surface spec?

## 5. A27 compliance note

claude-pantheon (`thr-fcb3187c58f409b5`) is a registered thread and will run an A27 heartbeat `/loop` (watcher: pull inbox, act on codex/claude-home reviews of this plan, heartbeat, sleep) until this workstream closes.
