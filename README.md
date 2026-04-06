# Sirsi Pantheon

**Infrastructure hygiene for your dev machine.** Finds waste that CleanMyMac misses, audits network security, and gives your AI tools persistent memory.

[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-C8A951?style=flat)](LICENSE)
[![Version](https://img.shields.io/badge/Version-0.15.0-1A1A5E?style=flat)](VERSION)
[![Tests](https://img.shields.io/badge/tests-1%2C663%20passing-brightgreen?style=flat)](.github/workflows/ci.yml)

```bash
brew tap SirsiMaster/tools && brew install sirsi-pantheon
```

---

## What It Does

### Scan for waste
```bash
pantheon scan                 # 58 rules across 7 domains — caches, build artifacts, orphaned files
pantheon scan --all           # Deep scan
pantheon scan --json          # Machine-readable output
```

### Hunt ghost apps
```bash
pantheon ghosts               # Find remnants of apps you already uninstalled
pantheon ghosts --sudo        # Include system directories
```

Ghost detection catches Launch Services phantoms, orphaned plists, leftover caches, and Spotlight ghosts that standard cleanup tools miss.

### Deduplicate files
```bash
pantheon dedup ~/Downloads ~/Documents
```

Three-phase scan: size grouping → partial hash (8 KB per file) → full hash. Opens a web UI with smart keep/delete recommendations.

### System health diagnostic
```bash
pantheon doctor               # RAM pressure, disk space, kernel panics, Jetsam events
pantheon doctor --json
```

### Network security audit
```bash
pantheon network              # DNS, WiFi, TLS, CA certs, VPN, firewall — read-only
pantheon network --fix        # Auto-apply encrypted DNS + firewall with safety rollback
pantheon network --rollback   # Restore DNS to pre-fix state
```

The `--fix` command uses a three-layer safety model: TCP probe before changing config, watchdog polling after, auto-revert within 6 seconds if resolution fails. [Case study →](docs/case-studies/isis-dns-safety-rollback.md)

### Hardware profiling
```bash
pantheon hardware             # CPU, GPU, RAM, Neural Engine, accelerators
pantheon hardware --json      # Full hardware profile
```

Detects Apple Silicon (ANE, Metal), NVIDIA (CUDA), AMD (ROCm), and Intel. Routes ML workloads to the fastest available accelerator.

### AI project memory
```bash
pantheon thoth init           # Create .thoth/ knowledge system in your project
pantheon thoth sync           # Sync from source + git history
pantheon mcp                  # Start MCP server for Claude, Cursor, Windsurf
```

Thoth gives AI coding sessions persistent memory via the [Model Context Protocol](https://modelcontextprotocol.io). Instead of re-explaining your project every session, the AI reads `.thoth/memory.yaml` and starts with full context.

### Code quality governance
```bash
pantheon quality              # Full governance audit (coverage, formatting, static analysis)
pantheon quality --skip-test  # Use cached coverage
```

Runs automatically on every `git push` via the pre-push gate. Three depth tiers: fast (10-30s default), standard (60-90s), deep (3-5 min pre-release).

### Knowledge ingestion
```bash
pantheon seshat ingest --source chrome       # Chrome bookmarks + history
pantheon seshat ingest --all-profiles        # All Chrome profiles
pantheon seshat export notebooklm            # Push to Google NotebookLM
```

Ingests from Chrome, Gemini, Claude, Apple Notes, and Google Workspace. Exports to NotebookLM, Thoth, and Gemini. All data stays local.

---

## Install

### Homebrew (macOS / Linux)
```bash
brew tap SirsiMaster/tools
brew install sirsi-pantheon
```

### From source
```bash
git clone https://github.com/SirsiMaster/sirsi-pantheon.git
cd sirsi-pantheon
go build -o pantheon ./cmd/pantheon/
```

### Binary
Download from [GitHub Releases](https://github.com/SirsiMaster/sirsi-pantheon/releases).

---

## All Commands

| Command | What It Does |
|:--------|:-------------|
| `pantheon scan` | Find infrastructure waste (58 rules, 7 domains) |
| `pantheon ghosts` | Detect remnants of uninstalled apps |
| `pantheon dedup [dirs]` | Find duplicate files with three-phase hashing |
| `pantheon doctor` | One-shot system health diagnostic |
| `pantheon network` | Network security audit (DNS, WiFi, TLS, firewall, VPN) |
| `pantheon hardware` | CPU, GPU, RAM, Neural Engine detection |
| `pantheon guard` | Real-time resource monitoring |
| `pantheon quality` | Code governance audit |
| `pantheon thoth init/sync` | AI project memory |
| `pantheon mcp` | MCP server for AI IDEs |
| `pantheon seshat ingest` | Knowledge ingestion from browsers and AI tools |
| `pantheon diagram` | Generate architecture diagrams (Mermaid/HTML) |
| `pantheon version` | Show version and module info |

Every command supports `--json`, `--quiet`, and `--verbose` flags.

---

## Editions

| Edition | Scope | Price |
|:--------|:------|:------|
| **Pantheon** | Single machine — all commands above | **Free forever** |
| **Pantheon Ra** | Fleet management — multi-repo orchestration, subnet scanning, compliance | Contact us |

The free edition has no feature gating, no telemetry, no time limits. MIT licensed.

---

## Security & Privacy

- **Zero telemetry.** No analytics, no tracking, no data leaves your machine. Non-negotiable.
- **Dry-run by default.** Every destructive operation requires explicit `--confirm`.
- **35 protected paths.** System directories, keychains, and SSH keys are hardcoded as undeletable.
- **Trash-first cleaning.** Removals go to Trash with a full decision log.
- **DNS safety model.** Network fixes probe before changing config, auto-revert on failure.

---

## Development

```bash
git clone https://github.com/SirsiMaster/sirsi-pantheon.git
cd sirsi-pantheon
git config core.hooksPath .githooks    # Enable pre-push quality gate
go test ./...                          # 1,663 tests across 27 packages
go build ./cmd/pantheon/
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## Built by Sirsi Technologies

[sirsi.ai](https://sirsi.ai) · [GitHub](https://github.com/SirsiMaster) · [Pantheon Hub](https://pantheon.sirsi.ai)

MIT License — free and open source forever.
