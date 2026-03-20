# Continuation Prompt — Sirsi Anubis
**Date:** March 20, 2026
**Version:** 0.1.0-alpha
**Context:** Session 1 wrap — Project genesis through Ka Ghost Hunter
**Priority:** Clean, lint, optimize, then continue Phase 1 development

---

## 0. Identity

You are working on **sirsi-anubis** — Sirsi Technologies' infrastructure hygiene CLI tool.

- **GitHub**: https://github.com/SirsiMaster/sirsi-anubis
- **Local Path**: `/Users/thekryptodragon/Development/sirsi-anubis`
- **Language**: Go 1.22+
- **CLI Binary**: `anubis` (5.2 MB)
- **Agent Binary**: `anubis-agent` (2.4 MB)
- **Rules File**: `ANUBIS_RULES.md` v1.1.0 (synced to `GEMINI.md` + `CLAUDE.md`)

**READ `ANUBIS_RULES.md` FIRST. It is the operational directive. It contains 16 universal rules + 13 Anubis-specific rules including safety protocols, commit traceability, feature documentation, and CI/CD gates.**

---

## 1. What Exists (Completed in Session 1)

### Administrative Scaffolding (P0 Complete)
| Document | Status | Lines |
|----------|--------|-------|
| `ANUBIS_RULES.md` v1.1.0 | ✅ | 308 |
| `GEMINI.md` + `CLAUDE.md` (synced copies) | ✅ | 308 each |
| `README.md` | ✅ | 147 |
| `CHANGELOG.md` (Keep a Changelog) | ✅ | 49 |
| `VERSION` → `0.1.0-alpha` | ✅ | 1 |
| `LICENSE` (MIT) | ✅ | 21 |
| `docs/ADR-001-FOUNDING-ARCHITECTURE.md` | ✅ | 81 |
| `docs/ADR-INDEX.md` | ✅ | 40 |
| `docs/ADR-TEMPLATE.md` | ✅ | 22 |
| `docs/ARCHITECTURE_DESIGN.md` | ✅ | 264 |
| `docs/SAFETY_DESIGN.md` | ✅ | 196 |
| `.github/workflows/ci.yml` | ✅ | 86 |
| `configs/default_rules.yaml` | ✅ | 282 |
| `.gitignore` | ✅ | 33 |

### Working Code
| Module | Package | Status | Description |
|--------|---------|--------|-------------|
| **Jackal** (Scan Engine) | `internal/jackal/` | ✅ Working | `ScanRule` interface, `Engine` with concurrent rule execution, result aggregation |
| **Ka** (Ghost Hunter) | `internal/ka/` | ✅ Working | Cross-references installed apps against 17 residual locations, queries Launch Services |
| **Cleaner** (Safety Module) | `internal/cleaner/` | ✅ Working | Hardcoded protected paths, trash-vs-delete, dry-run enforcement |
| **Output** (Terminal UI) | `internal/output/` | ✅ Working | Gold/Black lipgloss theme, styled findings, summary boxes |
| **34 Scan Rules** | `internal/jackal/rules/` | ✅ Working | 6 general, 4 virtualization, 5 dev, 6 AI/ML, 5 IDEs, 4 cloud, 3 storage, 1 Xcode |

### CLI Commands
| Command | Status | Description |
|---------|--------|-------------|
| `anubis weigh` | ✅ Working | Scan machine — found 67.8 GB across 31 findings on dev machine |
| `anubis judge --dry-run` | ✅ Working | Preview cleanup |
| `anubis judge --confirm` | ✅ Working | Clean with trash (default) or delete |
| `anubis ka` | ✅ Working | Ghost detection — found 130 ghosts (9.8 GB) on dev machine |
| `anubis ka --clean --dry-run` | ✅ Working | Preview ghost cleanup |
| `anubis ka --target "name"` | ✅ Working | Hunt specific ghost |
| `anubis version` | ✅ Working | Version display |
| `anubis guard` | ❌ Not built | RAM management |
| `anubis hapi` | ❌ Not built | VRAM/storage optimizer |
| `anubis scarab` | ❌ Not built | Fleet sweep |
| `anubis scales` | ❌ Not built | Policy engine |

### Git History (6 commits)
```
06f473f feat(ka): 𓂓 Ka module — Ghost Hunter + 22 new scan rules
cdaafc5 feat(jackal): Phase 1 MVP — working scanner + cleaner 🐺
16e5d91 feat(docs): port FinalWishes governance rules to ANUBIS_RULES v1.1.0
f0ef10f fix(core): add Go source files, fix .gitignore binary patterns
f30a7ef feat(core): Project genesis — Sirsi Anubis v0.1.0-alpha
```

### Internal Modules (Updated Hierarchy)
| Module | Codename | Archetype | Role |
| :--- | :--- | :--- | :--- |
| Local Scanner | **Jackal** 🐺 | The Hunter | Patrols and cleans individual machines |
| Ghost Hunter | **Ka** 𓂓 | The Spirit | Detects dead app remnants and residual hauntings |
| Fleet Sweep | **Scarab** 🪲 | The Transformer | Rolls across VLANs, subnets, domains |
| Policy Engine | **Scales** ⚖️ | The Judgment | Weighs findings against defined policies |
| Resource Optimizer | **Hapi** 🌊 | The Flow | Controls VRAM, GPU memory, and storage flow |

---

## 2. Immediate Priority: Clean, Lint, Optimize

### 2A. Code Quality & Linting
- [ ] Add `.golangci.yml` configuration file with linter settings
- [ ] Run `golangci-lint run ./...` and fix all warnings/errors
- [ ] Run `go vet ./...` and fix any issues
- [ ] Ensure `gofmt` is applied to all files (Rule A6)
- [ ] Add missing error handling (some scan rules silently skip errors)
- [ ] Review `internal/ka/scanner.go` — the `isInstalled()` function does expensive string matching in a loop; consider building a trie or prefix map

### 2B. Test Coverage
- [ ] Add unit tests for `internal/jackal/types.go` — `FormatSize()`, `ExpandPath()`, `PlatformMatch()`
- [ ] Add unit tests for `internal/cleaner/safety.go` — `ValidatePath()` with protected paths
- [ ] Add unit tests for `internal/ka/scanner.go` — `extractBundleID()`, `guessAppName()`, `isSystemBundleID()`
- [ ] Add table-driven test for each scan rule (Rule A6 — every rule needs at least one test)
- [ ] Add integration test: `TestWeighCommand` with test fixtures
- [ ] Target: 60%+ coverage before Phase 1 release

### 2C. Repo Cleanup
- [ ] Remove the empty module directories that have no code yet: `internal/fleet/`, `internal/profile/`, `internal/sight/` (replaced by `ka`), `agent/`
- [ ] Rename `internal/guard/` to keep but mark as Phase 1 TODO
- [ ] Verify `cmd/anubis-agent/main.go` still compiles cleanly (it's a placeholder)
- [ ] Ensure `go mod tidy` is clean
- [ ] Verify `.gitignore` covers all build artifacts
- [ ] Add `CONTRIBUTING.md` (P1 doc) — adapted from SirsiNexusApp pattern
- [ ] Add `SECURITY.md` (P1 doc) — adapted for filesystem scanning tool

### 2D. Developmental Alignment
- [ ] Update `CHANGELOG.md` with all work completed in Session 1
- [ ] Verify `docs/ARCHITECTURE_DESIGN.md` reflects the Ka module addition
- [ ] Create `ADR-002-KA-GHOST-DETECTION.md` — document the ghost detection algorithm
- [ ] Update `docs/ADR-INDEX.md` to list ADR-002
- [ ] Sync `ANUBIS_RULES.md` → `GEMINI.md` + `CLAUDE.md` after any rules changes
- [ ] Verify all commit messages follow the Traceability Protocol (Rule A7)

---

## 3. Phase 1 Remaining Work (After Cleanup)

### 3A. Guard Module (RAM Management)
```
anubis guard                    # Audit RAM usage by process group
anubis guard --slay node        # Kill zombie Node.js processes
anubis guard --slay lsp         # Kill stale language servers
anubis guard --slay docker      # Kill orphaned Docker processes
anubis guard --budget 16GB      # Set RAM budget, alert when exceeded
```
- Package: `internal/guard/`
- Needs: process listing (`ps aux`), grouping by parent, memory calculation
- Safety: Never kill protected processes (kernel_task, WindowServer, etc.)

### 3B. Ka Enhancements
- [ ] Launch Services cleanup: `lsregister -kill -r -domain local -domain system -domain user` to rebuild
- [ ] Show Spotlight impact — which ghosts are polluting search results
- [ ] Add `--min-size` flag to filter small ghosts (<1 MB)
- [ ] Consolidate ghosts by vendor (e.g., group all Adobe ghosts together)
- [ ] Add LaunchAgent/LaunchDaemon detection for user-level agents

### 3C. Jackal Enhancements
- [ ] Add `--min-age` flag to `anubis weigh` (override per-rule minimums)
- [ ] Add `--exclude` flag for custom exclusions
- [ ] Add `profiles` support (`~/.config/anubis/profiles/`) for project-specific scan configs
- [ ] Interactive mode with `bubbletea` — select which findings to clean

### 3D. More Scan Rules
- [ ] npm/yarn/pnpm global caches (`~/.npm/`, `~/.yarn/`, `~/.pnpm-store/`)
- [ ] Next.js `.next` build directories
- [ ] Rust `~/.cargo/registry/` and `~/.cargo/git/`
- [ ] Java `~/.gradle/caches/` and `~/.m2/repository/`
- [ ] CocoaPods and Swift Package Manager caches

---

## 4. Phased Roadmap (Full)

| Phase | Codename | Status | Scope |
|-------|----------|--------|-------|
| **0** | **Genesis** | ✅ Done | Scaffolding, docs, ADR-001 |
| **1** | **Jackal** | 🔨 In Progress | Local CLI — scan, clean, guard, Ka, profiles |
| **2** | **Jackal+** | 📋 Planned | Container/VM scanning, AI/ML rules, offline disk scan |
| **3** | **Hapi** | 📋 Planned | VRAM management, storage optimization, resource flow balancing |
| **4** | **Scarab** | 📋 Planned | Agent-controller, VLAN/subnet discovery, fleet sweep |
| **5** | **Scarab+** | 📋 Planned | SAN/NAS/S3 scanning, storage backends |
| **6** | **Scales** | 📋 Planned | Policy engine, fleet-wide enforcement, reporting |
| **7** | **Temple** | 📋 Future | Web dashboard / native SwiftUI GUI |

---

## 5. Key Architecture Details

### ScanRule Interface (internal/jackal/types.go)
```go
type ScanRule interface {
    Name() string
    DisplayName() string
    Category() Category
    Description() string
    Platforms() []string
    Scan(ctx context.Context, opts ScanOptions) ([]Finding, error)
    Clean(ctx context.Context, findings []Finding, opts CleanOptions) (*CleanResult, error)
}
```

### Rule Registration (internal/jackal/rules/registry.go)
All rules are registered in `AllRules()`. Platform filtering happens via `Platforms()` method.
New rules: create a Go file in `internal/jackal/rules/`, implement `ScanRule`, add to `AllRules()`.

### Two Rule Types
1. `baseScanRule` — path-based scanning with glob expansion (caches, logs, prefs)
2. `findRule` — directory-name searching within project trees (node_modules, target, .terraform)

### Safety Module (internal/cleaner/safety.go)
- `ValidatePath()` is called before EVERY deletion
- Protected paths are hardcoded — `.git`, `.env`, `.ssh`, `/System/`, `/usr/`, keychains
- Cannot be overridden by flags, config, or user input

### Ka Ghost Detection (internal/ka/scanner.go)
- Step 1: Build installed app index (/Applications + Homebrew casks)
- Step 2: Scan 12 user Library directories for orphaned bundle IDs
- Step 3: Query Launch Services (lsregister) for phantom registrations
- Step 4: Filter Apple system components (com.apple.*)
- Step 5: Merge into Ghost structs grouped by bundle ID

---

## 6. Technology Stack

| Layer | Technology |
| :--- | :--- |
| **Language** | Go 1.22+ |
| **CLI** | cobra v1.10.2 |
| **Terminal UI** | lipgloss v1.1.0 |
| **Build** | goreleaser (planned) |
| **CI/CD** | GitHub Actions |
| **Distribution** | Homebrew tap + GitHub Releases (planned) |

---

## 7. Portfolio Context

| Repo | Type |
| :--- | :--- |
| **SirsiNexusApp** | Platform Monorepo — Core infrastructure |
| **FinalWishes** | Tenant App — Estate planning |
| **Assiduous** | Tenant App — Real estate |
| **sirsi-anubis** (this repo) | Infrastructure Tool — Hygiene CLI |
| **sirsi-rook** (reserved) | Database Tool |
| **sirsi-rogue** (reserved) | Cybersecurity Sweeper |

---

## 8. Operational Reminders

1. **READ `ANUBIS_RULES.md` BEFORE ANY WORK** — it is the canonical directive
2. **Commit Traceability (Rule A7)** — every commit needs Refs: and Changelog: footers
3. **Feature Documentation (Rule A8)** — user guide + dev README in same commit
4. **Safety First (Rule A1)** — never delete without dry-run available
5. **Build Verification** — run `go build ./cmd/anubis/` and `go test ./...` before committing
6. **Sync agent files** — after rules changes: `cp ANUBIS_RULES.md GEMINI.md && cp ANUBIS_RULES.md CLAUDE.md`
7. **Context Monitoring (Rule A9)** — report session health after each sprint
8. **Push Protocol** — `git status` → `git add` → `git commit` → `git push`
9. **Identity** — `SirsiMaster` account exclusively

---

## 9. Session 2 Recommended Sprint Order

```
Sprint 1: Clean & Lint (2A + 2C)
  → .golangci.yml, fix all lint issues, remove empty dirs, go mod tidy

Sprint 2: Test Foundation (2B)
  → Unit tests for safety, types, Ka extraction, rule tests

Sprint 3: Developmental Alignment (2D)
  → ADR-002, changelog update, CONTRIBUTING.md, SECURITY.md

Sprint 4: Guard Module (3A)
  → anubis guard — RAM audit, orphan process killing

Sprint 5: Ka Enhancements (3B)
  → Launch Services rebuild, vendor grouping, min-size filter

Sprint 6: Polish & Release Prep
  → goreleaser config, Homebrew formula, v0.1.0-alpha tag
```

---

> *𓂀 "Weigh. Judge. Purge." — The jackal hunts. The Ka sees the dead.*
