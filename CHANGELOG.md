# Changelog — Sirsi Anubis
All notable changes to this project are documented in this file.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and [Semantic Versioning](https://semver.org/).

---

## [Unreleased]
### Planned
- Phase 1 (Jackal): RAM guard, interactive mode, profiles
- Phase 2 (Jackal+): Container/VM scanning, offline disk scan
- Phase 3 (Hapi): VRAM management, storage optimization

---

## [0.2.0-alpha] — 2026-03-25 (Ship Week Day 5)
### Added (Day 5: Neural Brain Downloader)
- **Brain module** (`internal/brain/`) — on-demand neural model management
- **`anubis install-brain`** — download CoreML/ONNX model to `~/.anubis/weights/`
  - Progress bar with bytes/total/percentage display
  - SHA-256 checksum verification post-download
  - Platform-aware model selection (prefers CoreML on Apple Silicon)
- **`anubis install-brain --update`** — check for and install latest model version
- **`anubis install-brain --remove`** — self-delete all weights and manifest
- **`anubis uninstall-brain`** — alias for `--remove`
- **Manifest-driven versioning** — remote `brain-manifest.json` + local `manifest.json`
- **Classifier interface** — pluggable backends (Stub, future ONNX, CoreML)
- **StubClassifier** — heuristic file classification (30+ file types, 9 categories)
  - Path-based detection: `node_modules/`, `__pycache__/`, `.cache/`
  - Extension-based: source, config, media, archives, data, ML models
  - Concurrent batch classification via goroutines
- **22 brain tests** — downloader + inference (manifest roundtrip, hash, batch, 35+ classification cases)
- **`--json` support** on all brain commands
- **Pro upsell footer** — tier messaging on brain commands

### Refs
- Canon: ANUBIS_RULES.md, docs/DEVELOPMENT_PLAN.md
- ADR: ADR-001
- Changelog: v0.2.0-alpha — Day 5 Neural Brain

## [0.1.0-alpha.2] — 2026-03-21
### Fixed (Session 2: Clean, Lint, Optimize)
- **CI pipeline** — fixed go.mod version mismatch (`go 1.26.1` → `go 1.22.0`)
- **golangci-lint** — added `.golangci.yml` config, replaced deprecated `exportloopref` with `copyloopvar`
- **errcheck** — fixed unchecked `cmd.Help()` return value
- **gofmt** — applied formatting to 4 source files with drift
- **Portfolio CI** — fixed FinalWishes (`go 1.25.0` → `go 1.24.0`), tenant-scaffold (missing `package-lock.json`)

### Added (Session 2: Tests + Documentation)
- **Unit tests** — `types_test.go` (FormatSize, ExpandPath, PlatformMatch), `safety_test.go` (all protection layers), `scanner_test.go` (extractBundleID, guessAppName, isSystemBundleID), `engine_test.go` (mock rules, category filtering, clean safety)
- **ADR-002** — Ka Ghost Detection algorithm (5-step process, 17 residual locations)
- **CONTRIBUTING.md** — contributor guide with scan rule examples and safety rules
- **SECURITY.md** — security policy, threat model, protected paths, data privacy

---

## [0.1.0-alpha.1] — 2026-03-20
### Added (Session 1: Ka Ghost Hunter)
- **Ka module** (`internal/ka/`) — Ghost detection engine scanning 17 macOS locations
- **22 new scan rules** — AI/ML (6), virtualization (4), IDEs (5), cloud (4), storage (3)
- **`anubis ka`** — Ghost hunting CLI command with `--clean`, `--dry-run`, `--target` flags
- **Launch Services scanning** — detects phantom app registrations in Spotlight
- **Bundle ID extraction** — heuristic parser for plist filenames and directory names
- **System filtering** — `com.apple.*` and known system services excluded from ghosts

---

## [0.1.0-alpha] — 2026-03-20
### Added (Phase 0: Project Genesis)
- **Project scaffolding** — Go 1.22+ module, directory structure for all 4 modules
- **ANUBIS_RULES.md v1.0.0** — Operational directive with 16 universal rules + 5 Anubis-specific safety rules
- **GEMINI.md + CLAUDE.md** — Auto-synced copies of ANUBIS_RULES.md
- **ADR-001** — Founding architecture decision (Go, cobra, agent-controller, module codenames)
- **ADR system** — Template + Index established (next available: ADR-002)
- **Architecture Design** — Module architecture, data flow, component interaction
- **Safety Design** — Protected paths, dry-run guarantees, trash-vs-delete policy
- **CI/CD** — GitHub Actions workflow: lint, test, build, binary size guard
- **Default scan rules config** — YAML-based rule definitions
- **LICENSE** — MIT (free and open source forever)
- **VERSION** — `0.1.0-alpha`

### Refs
- Canon: ANUBIS_RULES.md, docs/ARCHITECTURE_DESIGN.md, docs/SAFETY_DESIGN.md
- ADR: ADR-001 (Founding Architecture)
- Changelog: v0.1.0 — Project Genesis

---

## [0.0.1] — 2026-03-20
### Added
- Initial product concept ("Deep Cleanse") born from manual Parallels cleanup session
- Competitive analysis vs Mole (open-source Mac cleaner)
- Name selection: Sirsi Anubis (Egyptian god of judgment)
- Module codenames: Jackal, Scarab, Scales, Hapi
- 60+ scan rule categories across 7 domains identified
- Agent-controller architecture designed
- Network topology awareness (VLAN, subnet, relay) specified
