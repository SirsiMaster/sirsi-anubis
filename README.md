# 🏛️ Sirsi Pantheon

**Unified DevOps Intelligence Platform — One Install, All Deities.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-C8A951?style=flat)](LICENSE)
[![Version](https://img.shields.io/badge/Version-0.4.0--alpha-1A1A5E?style=flat)](VERSION)
[![Tests](https://img.shields.io/badge/tests-522%20passing-brightgreen?style=flat)](.github/workflows/ci.yml)
[![MCP](https://img.shields.io/badge/MCP-2025--03--26-purple?style=flat)](https://modelcontextprotocol.io)
[![Build in Public](https://img.shields.io/badge/building-in%20public-C8A951?style=flat)](docs/BUILD_LOG.md)

> *"One Install. All Deities."*

**Sirsi Pantheon** is a unified DevOps intelligence platform that brings together every deity in the Sirsi ecosystem into a single, lightweight binary. Install once, get infrastructure hygiene, QA/QC governance, persistent AI knowledge, and more.

### The Deities

| Deity | Domain | Description |
|:------|:-------|:------------|
| 𓂀 **Anubis** | Infrastructure Hygiene | Scan, judge, and purge waste — the foundational module |
| 🪶 **Ma'at** | QA/QC Governance | Coverage, canon verification, pipeline monitoring |
| 𓁟 **Thoth** | Persistent Knowledge | AI session memory that saves 98.7% of context |

Anubis is the **foundational module** — the reason Pantheon exists. It's not being replaced, it's being elevated. Every scan rule, every safety protection, every feature you see here started with Anubis. Pantheon is what happens when one deity proves the architecture works and the others join in.

---

## 🏗️ Why the Rebrand: Binary Size Tells the Story

We analyzed exactly what's inside the binary and discovered something remarkable:

| Component | Size |
|:----------|-----:|
| Go runtime + stdlib | ~10 MB |
| **All 20 Sirsi modules combined** | **~186 KB** |
| Total binary (stripped) | 8.3 MB |

**All application code — 20 modules, 18 commands, 13,813 lines — compiles to just 186 KB.** That's 1.5% of the binary. The other 98.5% is Go's runtime, which is paid once regardless of module count.

Adding a new deity module costs **2-30 KB**. We could double the module count and the binary would grow by ~100 KB. Pantheon isn't a monolith — it's a composable platform that happens to ship as a single file.

**The rename reflects reality:** the binary already contained every deity. Calling it "Anubis" was limiting what it actually is. Pantheon gives every deity equal standing under one roof while honoring Anubis as the foundation.

### Size Per Deity Module (compiled)

| Module | Compiled Size | Role |
|:-------|-------------:|:-----|
| mcp | 31.1 KB | AI IDE integration |
| jackal | 26.6 KB | Scan engine (58 rules) |
| mirror | 19.3 KB | File dedup (27x faster) |
| maat | 18.9 KB | QA/QC governance |
| brain | 15.6 KB | Neural classifier |
| guard | 8.8 KB | RAM audit |
| ka | 7.3 KB | Ghost app detection |
| *...12 more* | *2-10 KB each* | |
| **Total** | **~186 KB** | **1.5% of binary** |

---

## ⚡ Quick Start

### Install
```bash
# From source
go install github.com/SirsiMaster/sirsi-pantheon/cmd/pantheon@latest

# Or clone and build
git clone https://github.com/SirsiMaster/sirsi-pantheon.git
cd sirsi-pantheon && go build -o pantheon ./cmd/pantheon/
```

### Scan Your Machine
```bash
pantheon weigh                   # Full scan — discover all waste
pantheon weigh --dev             # Developer frameworks only
pantheon weigh --ai              # AI/ML caches only
pantheon weigh --json            # Machine-readable output
```

### Clean What Was Found
```bash
pantheon judge --dry-run         # Preview cleanup
pantheon judge --confirm         # Execute cleanup
```

### Hunt Ghost Apps
```bash
pantheon ka                      # Find remnants of uninstalled apps
pantheon ka --target "Parallels" # Hunt specific ghost
pantheon ka --clean --dry-run    # Preview ghost cleanup
pantheon ka --clean --confirm    # Release the spirits
```

### Find Duplicate Files
```bash
pantheon mirror ~/Downloads ~/Desktop  # Scan directories for duplicates
pantheon mirror --photos --min-size 1MB # Large photo duplicates only
pantheon mirror --json > report.json   # Export results
pantheon mirror                        # Launch GUI (browser-based)
```

### Run QA/QC Governance
```bash
pantheon maat                    # Full governance assessment
pantheon maat --coverage         # Test coverage governance
pantheon maat --canon            # Canon document verification
pantheon maat --pipeline         # CI pipeline monitoring
```

---

## 📋 All Commands

| Command | Deity | Description |
|:--------|:------|:-----------|
| `pantheon weigh` | 𓂀 Anubis | Scan workstation for infrastructure waste |
| `pantheon judge` | 𓂀 Anubis | Clean artifacts found by weigh |
| `pantheon ka` | 𓂀 Anubis | Hunt ghost apps — find spirits of the dead |
| `pantheon guard` | 𓂀 Anubis | RAM audit, zombie process management |
| `pantheon sight` | 𓂀 Anubis | Launch Services / Spotlight repair |
| `pantheon profile` | 𓂀 Anubis | Machine profiling and system info |
| `pantheon seba` | 𓂀 Anubis | Dependency graph mapper |
| `pantheon hapi` | 𓂀 Anubis | Resource optimizer (GPU, dedup, snapshots) |
| `pantheon scarab` | 𓂀 Anubis | Network discovery + container audit |
| `pantheon install-brain` | 𓂀 Anubis | Download neural classification model |
| `pantheon uninstall-brain` | 𓂀 Anubis | Remove neural weights |
| `pantheon mirror` | 𓂀 Anubis | Duplicate file scanner (GUI + CLI) |
| `pantheon scales enforce` | 𓂀 Anubis | Run hygiene policy enforcement |
| `pantheon book-of-the-dead` | 𓂀 Anubis | Deep system autopsy |
| `pantheon initiate` | 𓂀 Anubis | Grant macOS permissions |
| `pantheon mcp` | 𓁟 Thoth | Start MCP server for AI IDE integration |
| `pantheon maat` | 🪶 Ma'at | QA/QC governance assessment |
| `pantheon version` | — | Show version and deity roster |

### Global Flags
```bash
--json      # JSON output for scripting
--quiet     # Suppress non-essential output
--stealth   # Ephemeral mode — delete all Pantheon data after execution
```

---

## 🏛 Architecture

Pantheon is built on modules named after Egyptian mythology. Every deity maintains its identity while sharing a unified runtime:

| Module | Deity | Codename | Role | Status |
|:-------|:------|:---------|:-----|:-------|
| 🐺 **Jackal** | Anubis | The Hunter | Scan engine — 58 rules across 7 domains | ✅ |
| 𓂓 **Ka** | Anubis | The Spirit | Ghost app detection — 17 macOS locations | ✅ |
| 🪞 **Mirror** | Anubis | The Reflection | File dedup — 27x faster than naive hashing | ✅ |
| 🛡️ **Guard** | Anubis | The Guardian | RAM audit, zombie process management | ✅ |
| 👁️ **Sight** | Anubis | The Sight | Launch Services + Spotlight repair | ✅ |
| 🌊 **Hapi** | Anubis | The Flow | GPU detection, dedup, APFS snapshots | ✅ |
| 🪲 **Scarab** | Anubis | The Transformer | Network discovery + container audit | ✅ |
| 🧠 **Brain** | Anubis | Neural | On-demand model downloader + classifier | ✅ |
| 🔌 **MCP** | Thoth | Context | MCP server for AI IDE integration | ✅ |
| ⚖️ **Scales** | Anubis | The Judgment | YAML policy engine + enforcement | ✅ |
| 🪶 **Ma'at** | Ma'at | Governance | Coverage, canon, pipeline assessments | ✅ |

### Two Binaries

| Binary | Size | Stripped | Purpose |
|:-------|:-----|:--------|:--------|
| `pantheon` | 12 MB | **8.3 MB** | Full CLI — all deities, Mirror GUI |
| `pantheon-agent` | 3.2 MB | **2.1 MB** | Lightweight fleet agent (JSON only) |

---

## 📦 Scan Domains (58 Rules)

| Domain | Examples |
|:-------|:--------|
| 🖥️ **General Mac** | Caches, logs, crash reports, browser junk, downloads |
| 🐳 **Virtualization** | Parallels, Docker, VMware, UTM, VirtualBox |
| 📦 **Dev Frameworks** | Node/npm, Next.js, Rust/Cargo, Go, Python/conda, Java/Gradle |
| 🤖 **AI/ML** | Apple MLX, Metal, NVIDIA CUDA, HuggingFace, Ollama, PyTorch |
| 🛠️ **IDEs & AI Tools** | Xcode, VS Code, JetBrains, Claude Code, Codex, Gemini CLI |
| ☁️ **Cloud/Infra** | Docker, Kubernetes, nginx, Terraform, gcloud, Firebase |
| 📱 **Cloud Storage** | OneDrive, Google Drive, iCloud, Dropbox |

---

## 🧠 Neural Brain

Pantheon includes an on-demand neural classification engine:

```bash
pantheon install-brain             # Download CoreML/ONNX model
pantheon install-brain --update    # Check for latest version
pantheon install-brain --remove    # Self-delete weights
```

The brain classifies files into 9 categories: **junk**, **project**, **config**, **model**, **data**, **media**, **archive**, **essential**, **unknown**. Currently ships with a heuristic classifier; neural model backends (ONNX Runtime, CoreML) are in development.

---

## 🔌 MCP Server — AI IDE Integration

Pantheon doubles as a context sanitizer for AI coding assistants:

```bash
pantheon mcp    # Start MCP server (stdio)
```

### Configure Claude Code
```json
// ~/.claude/claude_desktop_config.json
{
  "mcpServers": {
    "pantheon": {
      "command": "pantheon",
      "args": ["mcp"]
    }
  }
}
```

### Configure Cursor / Windsurf
```json
// .cursor/mcp.json
{
  "mcpServers": {
    "pantheon": {
      "command": "pantheon",
      "args": ["mcp"]
    }
  }
}
```

### MCP Tools
| Tool | Description |
|:-----|:-----------|
| `scan_workspace` | Scan a directory for waste |
| `ghost_report` | Hunt dead app remnants |
| `health_check` | System health summary with grade |
| `classify_files` | Semantic file classification |
| `thoth_read_memory` | 𓁟 Load project context without reading source files |

---

## 𓁟 Thoth — Persistent Knowledge System

Thoth gives AI assistants **persistent memory across sessions**. Instead of re-reading thousands of lines of source code every time, the AI reads a ~100-line memory file for instant context.

```
.thoth/
├── memory.yaml      # Layer 1: WHAT — compressed project state
├── journal.md       # Layer 2: WHY — timestamped reasoning
└── artifacts/       # Layer 3: DEEP — benchmarks, audits, reviews
```

### Measured Impact (Dogfooding on This Repo)

| Metric | Without Thoth | With Thoth | Savings |
|:-------|:-------------|:-----------|:--------|
| Lines read at startup | 22,958 | 297 | **98.7% fewer** |
| Tokens consumed | 275,496 | 3,564 | **271,932 saved** |
| Context window used | 137.7% (⚠️ doesn't fit) | 1.7% | **136% preserved** |
| Cost per session (Opus 4) | $4.13 | $0.05 | **$4.08 saved** |

> We built Thoth because our own AI sessions were failing — the codebase was too large to fit in context. The before/after is measurable. [Read the case study →](docs/case-studies/thoth-context-savings.md)

---

## 🪶 Ma'at — QA/QC Governance

Ma'at ensures every change meets quality standards:

```bash
pantheon maat                    # Full governance assessment
pantheon maat --coverage         # Test coverage thresholds
pantheon maat --canon            # Canon document verification
pantheon maat --pipeline         # CI pipeline health
pantheon maat --json             # Machine-readable output
```

57 tests, 3 governance domains, per-module threshold enforcement.

---

## ⚖️ Policy Enforcement

Define infrastructure hygiene policies in YAML:

```yaml
api_version: v1
policies:
  - name: workstation-hygiene
    rules:
      - id: waste-limit
        metric: total_size
        operator: gt
        threshold: 20
        unit: GB
        severity: fail
        remediation: Run 'pantheon judge --confirm'
```

```bash
pantheon scales enforce                      # Run default policies
pantheon scales enforce -f custom-policy.yaml # Custom policies
pantheon scales validate -f policy.yaml      # Syntax check
```

---

## 🪞 Mirror — File Deduplication

Mirror finds duplicate files across any directory using a **three-phase scan**:

1. **Size grouping** — instant elimination of unique file sizes
2. **Partial hash** — SHA-256 of first 4KB + last 4KB (8KB per file)
3. **Full hash** — complete SHA-256 only for files that pass both phases

### Why This Matters

| Metric | Naive approach | Pantheon Mirror |
|:-------|:--------------|:----------------|
| 56 candidate files (97.8 MB) | Reads all 97.8 MB | Reads 448 KB partial, then only matched files |
| Disk I/O | 97.8 MB | **< 2 MB** |
| Time | 84 ms | **3 ms** |
| Speedup | 1x | **27.3x** |
| I/O reduction | — | **98.8%** |

*Benchmarked on real ~/Downloads directory, March 2026.*

---

## 🛡️ Product Tiers

| Tier | Scope | Price |
|:-----|:------|:------|
| **Pantheon Free** | Single workstation, all commands, Mirror GUI + CLI | Free forever |
| **Pantheon Pro** | Neural brain, importance ranking, semantic search | $9/mo |
| **Eye of Horus** | Subnet sweep (< 100 nodes) | $29/mo |
| **Ra** | Enterprise fleet, SAN/NAS, compliance | Contact |

---

## 🔒 Security & Privacy

- **Rule A11: No Telemetry** — zero analytics, tracking, or data collection
- **Rule A1: Safety First** — all destructive ops require `--confirm` or `--dry-run`
- **Rule A3: Fixed Agent Commands** — agent has no shell access
- **Trash-first cleaning** — every removal goes to Trash with full decision log
- **29 protected paths** — system dirs, user content dirs, keychains, and SSH keys are hardcoded as undeletable
- **`--stealth` mode** — Pantheon comes, judges, and vanishes (zero footprint)
- All scanning is local — no data leaves your machine

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Adding scan rules is straightforward — implement the `ScanRule` interface in `internal/jackal/rules/`.

---

## 📄 License

MIT License — free and open source forever. See [LICENSE](LICENSE).

---

## 🏢 Sirsi Technologies

Sirsi Pantheon is the DevOps intelligence platform from [Sirsi Technologies](https://github.com/SirsiMaster).

| Product | Role |
|:--------|:-----|
| **Sirsi Pantheon** | Unified DevOps Intelligence Platform |
| **Sirsi Nexus** | AI infrastructure platform |

---

*🏛️ One install. All deities. Nothing escapes the Weighing.*
