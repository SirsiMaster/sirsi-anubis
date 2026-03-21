# 𓂀 Sirsi Anubis — Canonical Development Plan
**Version:** 1.0.0
**Date:** March 21, 2026
**Status:** LOCKED — Canonical Reference

> This is the master roadmap. All sprint plans, session prompts, and feature
> work MUST trace back to this document. Changes require ADR review.

---

## Product Tiers

| Tier | Codename | License | Distribution | Scope |
|:-----|:---------|:--------|:-------------|:------|
| **Anubis** | Open Source | MIT | Homebrew, GitHub Releases, `go install`, Docker | Single workstation — scan, clean, ghost hunt, RAM guard |
| **Eye of Horus** | Subnet Edition | Licensed | Upgrade from Anubis | Local subnet scanning — VLAN sweep, LAN discovery, agent deployment |
| **Ra** | Enterprise Edition | Sirsi-only | Bundled with Sirsi Platform | Fleet-scale — multi-site, SAN/NAS, policy enforcement, reporting dashboard |

### Tier Boundaries

```
┌─────────────────────────────────────────────────────────┐
│                    RA (Enterprise)                       │
│   Fleet policies, multi-site, SAN/NAS/S3, dashboards   │
│ ┌─────────────────────────────────────────────────────┐ │
│ │              EYE OF HORUS (Subnet)                  │ │
│ │   VLAN sweep, LAN agents, local fleet orchestration │ │
│ │ ┌─────────────────────────────────────────────────┐ │ │
│ │ │               ANUBIS (Open Source)               │ │ │
│ │ │   weigh, judge, ka, guard, sight, hapi          │ │ │
│ │ │   34→60+ rules, CLI, JSON output, profiles      │ │ │
│ │ └─────────────────────────────────────────────────┘ │ │
│ └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

---

## Platform Matrix

### Operating Systems

| OS | Status | Binary | Notes |
|:---|:-------|:-------|:------|
| macOS (arm64) | 🔨 Active dev | `anubis-darwin-arm64` | Primary dev platform |
| macOS (amd64) | 🔨 CI builds | `anubis-darwin-amd64` | Intel Mac support |
| Linux (amd64) | Phase 2 | `anubis-linux-amd64` | Ubuntu/Debian/Fedora |
| Linux (arm64) | Phase 2 | `anubis-linux-arm64` | Raspberry Pi, ARM servers |
| Windows (amd64) | Phase 3 | `anubis-windows-amd64.exe` | WSL2 + native |
| Windows (arm64) | Phase 4 | `anubis-windows-arm64.exe` | Surface Pro, Snapdragon |

### GPU / Accelerator Detection

| Accelerator | Detection Method | Hapi Module | Phase |
|:------------|:----------------|:------------|:------|
| Apple Metal / MLX | `system_profiler SPDisplaysDataType`, Metal API | `hapi/metal.go` | 3 |
| NVIDIA CUDA | `nvidia-smi`, NVML library | `hapi/cuda.go` | 3 |
| AMD ROCm | `rocm-smi`, `/sys/class/drm/` | `hapi/rocm.go` | 4 |
| Intel oneAPI | `xpu-smi`, SYCL runtime | `hapi/intel.go` | 5 |

### Framework / Model Detection

| Category | Detection | Scan Rule | Phase |
|:---------|:----------|:----------|:------|
| PyTorch | `~/.cache/torch`, `torch.cuda.is_available()` | `pytorch_cache` | ✅ Done |
| TensorFlow | `~/.cache/tensorflow`, `tf.config.list_physical_devices` | `tensorflow_cache` | ✅ Done |
| HuggingFace | `~/.cache/huggingface`, `transformers` cache | `huggingface_cache` | ✅ Done |
| Ollama | `~/.ollama/models` | `ollama_models` | ✅ Done |
| Apple MLX | `~/.cache/mlx`, CoreML models | `mlx_cache` | ✅ Done |
| ONNX Runtime | `~/.cache/onnxruntime` | `onnx_cache` | 2 |
| vLLM | `~/.cache/vllm` | `vllm_cache` | 2 |
| JAX/Flax | `~/.cache/jax` | `jax_cache` | 2 |
| Stable Diffusion | `~/.cache/stable-diffusion`, ComfyUI models | `sd_models` | 2 |
| LangChain | `~/.langchain`, vector store caches | `langchain_cache` | 3 |

### IDE / Dev Tool Detection

| Tool | Detection | Scan Rule | Phase |
|:-----|:----------|:----------|:------|
| VS Code | `~/Library/Application Support/Code/` | `vscode_caches` | ✅ Done |
| JetBrains | `~/Library/Caches/JetBrains` | `jetbrains_caches` | ✅ Done |
| Xcode | `~/Library/Developer/Xcode/DerivedData` | `xcode_derived_data` | ✅ Done |
| Android Studio | `~/.android/avd`, `~/.android/cache` | `android_studio` | ✅ Done |
| Claude Code | `~/.claude/logs` | `claude_code` | ✅ Done |
| Gemini CLI | `~/.gemini/cache` | `gemini_cli` | ✅ Done |
| Cursor | `~/Library/Application Support/Cursor/` | `cursor_caches` | 2 |
| Windsurf | `~/Library/Application Support/Windsurf/` | `windsurf_caches` | 2 |
| Zed | `~/Library/Application Support/Zed/` | `zed_caches` | 2 |
| Neovim | `~/.local/share/nvim`, `~/.cache/nvim` | `neovim_caches` | 2 |
| Eclipse | `~/eclipse/`, `.metadata/` dirs | `eclipse_caches` | 3 |

---

## Phase Roadmap

### Phase 1: Jackal MVP — Local Workstation 🐺
**Status:** 🔨 In Progress | **Target:** v0.2.0-alpha

#### Sprint 1.0 — Foundation ✅ DONE
- [x] Project scaffolding (cobra CLI, lipgloss UI, Go 1.22)
- [x] ScanRule interface + Engine (concurrent scan orchestration)
- [x] Safety module (hardcoded protected paths, dry-run enforcement)
- [x] `anubis weigh` — scan with category flags
- [x] `anubis judge` — clean with `--dry-run`/`--confirm`/`--trash`
- [x] 12 initial scan rules (general Mac)

#### Sprint 1.1 — Ka Ghost Hunter ✅ DONE
- [x] Ka module (`internal/ka/`) — 5-step ghost detection algorithm
- [x] `anubis ka` command — ghost scan, clean, target filter
- [x] 22 new scan rules (AI, virtualization, IDEs, cloud, storage)
- [x] Launch Services (lsregister) scanning
- [x] Bundle ID extraction + system component filtering

#### Sprint 1.2 — CI + Quality ✅ DONE
- [x] CI pipeline green (go.mod fix, golangci-lint config, gofmt)
- [x] Unit tests (65+ cases, jackal 93%, cleaner 49%, ka 19.5%)
- [x] ADR-002, CONTRIBUTING.md, SECURITY.md, CHANGELOG
- [x] Portfolio CI fix (FinalWishes, tenant-scaffold)

#### Sprint 1.3 — Guard Module 🛡️
- [ ] `internal/guard/audit.go` — process enumeration + memory grouping
- [ ] `internal/guard/slayer.go` — orphan process detection + termination
  - Node.js orphans (`node --max-old-space-size`)
  - LSP servers (TypeScript, Rust Analyzer, gopls, pyright)
  - Docker Desktop background processes
  - Electron helper renderers
- [ ] `internal/guard/protector.go` — memory budget recommendations
- [ ] `cmd/anubis/guard.go` — `anubis guard`, `anubis guard --slay <target>`
- [ ] Safety: SIGTERM first, SIGKILL after 5s timeout, never kill root/system

#### Sprint 1.4 — Scan Rule Expansion (34 → 60+)
- [ ] Java/Gradle: `~/.gradle/caches`, `~/.m2/repository`
- [ ] Homebrew: `~/Library/Caches/Homebrew`, old formula versions
- [ ] npm/yarn/pnpm: global caches (`~/.npm/_cacache`, `~/.yarn/cache`)
- [ ] CocoaPods: `~/Library/Caches/CocoaPods`
- [ ] Swift Package Manager: `~/Library/Developer/Xcode/SPMRepositories`
- [ ] Composer (PHP): `~/.composer/cache`
- [ ] Ruby gems: `~/.gem/ruby/*/cache`
- [ ] Cursor IDE, Windsurf, Zed, Neovim caches
- [ ] ONNX, vLLM, JAX, Stable Diffusion caches
- [ ] iCloud Drive cache: `~/Library/Mobile Documents`
- [ ] nginx logs: `/var/log/nginx/`, `/usr/local/var/log/nginx/`
- [ ] Time Machine local snapshots
- [ ] Codex CLI: `~/.codex/`
- [ ] Spotlight index rebuild option

#### Sprint 1.5 — Sight Module 👁️
- [ ] `internal/sight/launchservices.go` — ghost app detection (extract from Ka)
- [ ] `internal/sight/rebuild.go` — Launch Services database rebuild
- [ ] `cmd/anubis/sight.go` — `anubis sight`, `anubis sight --fix`
- [ ] Spotlight re-index trigger for cleaned ghost apps

#### Sprint 1.6 — Profiles + Config
- [ ] `internal/profile/` — developer profile system
- [ ] `~/.config/anubis/config.yaml` — user preferences
- [ ] `~/.config/anubis/profiles/*.yaml` — named scan profiles
- [ ] `anubis profile create`, `anubis profile list`, `anubis profile use`
- [ ] Default profiles: `general`, `developer`, `ai-engineer`, `devops`

#### Sprint 1.7 — Distribution + Polish
- [ ] goreleaser setup (multi-platform binaries, checksums)
- [ ] Homebrew tap: `SirsiMaster/homebrew-tools`
- [ ] `go install` support (publish module)
- [ ] Binary size audit (target: anubis < 10 MB, agent < 5 MB)
- [ ] Shell completions (bash, zsh, fish, PowerShell)
- [ ] `docs/SCAN_RULE_GUIDE.md` — contributor guide for new rules
- [ ] Linux rule implementations (linuxRules() in registry.go)
- [ ] Man page generation

**Phase 1 Exit Criteria:**
- 60+ scan rules across all 7 categories
- `weigh`, `judge`, `ka`, `guard`, `sight` all working
- macOS + Linux support
- Homebrew + goreleaser distribution
- Test coverage > 70% on core modules
- README accurately describes shipped features

---

### Phase 2: Jackal+ — Deep Scanning 🐺+
**Target:** v0.3.0-beta

- [ ] Container scanning — scan inside Docker containers (`docker exec`)
- [ ] VM guest scanning — scan inside Parallels/UTM/VMware guests
- [ ] Offline disk scanning — scan mounted external drives
- [ ] Windows rule implementations
- [ ] Interactive TUI mode (bubbletea) — real-time scan progress
- [ ] Scan scheduling (cron-based, launchd)
- [ ] Markdown/HTML report generation (`output/markdown.go`, `output/html.go`)
- [ ] Disk image scanning (DMG, VHD, VMDK mounted volumes)

---

### Phase 3: Hapi — Resource Optimizer 🌊
**Target:** v0.4.0-beta

- [ ] `internal/hapi/metal.go` — Apple Metal/MLX unified memory audit
- [ ] `internal/hapi/cuda.go` — NVIDIA CUDA VRAM management (`nvidia-smi`)
- [ ] `internal/hapi/vram.go` — GPU memory audit and optimization
- [ ] `internal/hapi/snapshots.go` — APFS snapshot pruning
- [ ] `internal/hapi/dedup.go` — duplicate file detection (hash-based)
- [ ] `internal/hapi/compress.go` — compression analysis and recommendations
- [ ] `internal/hapi/tier.go` — hot/warm/cold storage tiering
- [ ] `cmd/anubis/hapi.go` — `anubis hapi`, `anubis hapi --gpu`, `anubis hapi --dedup`
- [ ] Platform detection: Metal vs CUDA vs ROCm vs CPU-only

---

### Phase 4: Scarab — Fleet Sweep (Eye of Horus) 🪲
**Target:** v0.5.0 | **License:** Eye of Horus (upgrade from OSS)

- [ ] `internal/scarab/discovery.go` — subnet/VLAN host discovery
- [ ] `internal/scarab/sweep.go` — parallel fleet scanning
- [ ] `internal/scarab/container.go` — Docker/Kubernetes container scanning
- [ ] `internal/scarab/vm.go` — VM guest agent deployment
- [ ] `internal/scarab/transport/ssh.go` — SSH transport
- [ ] `internal/scarab/transport/grpc.go` — gRPC transport
- [ ] `cmd/anubis/scarab.go` — `anubis scarab discover`, `anubis scarab sweep`
- [ ] Agent auto-deployment — push `anubis-agent` to targets
- [ ] `--confirm-network` safety flag (Rule A4)
- [ ] Agent binary: static, zero-dep, fixed command set (Rule A3)

---

### Phase 5: Scarab+ — Storage Backends 🪲+
**Target:** v0.6.0 | **License:** Eye of Horus / Ra

- [ ] `internal/scarab/storage.go` — NFS/SMB/iSCSI scanning
- [ ] S3-compatible storage scanning (AWS S3, MinIO, Backblaze)
- [ ] APFS/ZFS pool analysis
- [ ] SAN/NAS orphan detection
- [ ] Storage cost estimation (cloud pricing integration)

---

### Phase 6: Scales — Policy Engine (Ra) ⚖️
**Target:** v1.0.0 | **License:** Ra (Enterprise, Sirsi-only)

- [ ] `internal/scales/policy.go` — YAML policy parser
- [ ] `internal/scales/enforce.go` — fleet-wide enforcement
- [ ] `internal/scales/verdicts.go` — verdict reporting
- [ ] `cmd/anubis/scales.go` — `anubis scales enforce`, `anubis scales validate`
- [ ] Slack/Teams/webhook notifications
- [ ] Compliance reports (SOC2, CIS benchmarks)
- [ ] Integration with Sirsi Nexus platform API

---

### Phase 7: Temple — GUI 🏛️
**Target:** v2.0.0 | **License:** All tiers (OSS + Eye + Ra)

- [ ] SwiftUI native macOS app (wraps CLI engine)
- [ ] Web dashboard (Sirsi Nexus embedded)
- [ ] Real-time fleet monitoring
- [ ] Scan scheduling + history
- [ ] Visual ghost map / disk treemap

---

## Binary Size Budget

| Binary | Current | Target | Strategy |
|:-------|:--------|:-------|:---------|
| `anubis` | 5.2 MB | < 15 MB | All modules compiled in, UPX optional |
| `anubis-agent` | 2.4 MB | < 5 MB | Scan-only, no UI, no cobra, `CGO_ENABLED=0` |

**Build flags:**
```bash
# Controller (full)
go build -ldflags="-s -w" -o anubis ./cmd/anubis/

# Agent (minimal)
CGO_ENABLED=0 go build -ldflags="-s -w" -o anubis-agent ./cmd/anubis-agent/
```

---

## Scan Rule Target Count

| Category | Current | Phase 1.4 Target | Phase 2+ Target |
|:---------|:--------|:-----------------|:----------------|
| General Mac | 6 | 10 | 12 |
| Virtualization | 4 | 5 | 6 |
| Dev Frameworks | 5 | 14 | 18 |
| AI/ML | 6 | 11 | 15 |
| IDEs & AI Tools | 6 | 12 | 14 |
| Cloud & Infra | 4 | 7 | 10 |
| Cloud Storage | 3 | 5 | 6 |
| **Total** | **34** | **64** | **81** |

---

## Integration with Sirsi Ecosystem

| Sirsi Product | Anubis Integration |
|:--------------|:-------------------|
| **Sirsi Nexus** | Ra tier: fleet API, dashboard, policy management |
| **Sirsi-Sign** | Shared security infrastructure, agent auth certificates |
| **Sirsi Rook** | Database artifact scanning (stale backups, orphaned schemas) |
| **Sirsi Rogue** | Security sweep integration (CVE scanning, exposed secrets) |

---

## Decision Log

| Decision | Rationale | ADR |
|:---------|:----------|:----|
| Go 1.22+ | Static binary, cross-platform, contributor-friendly | ADR-001 |
| Agent-controller model | Fleet scalability without SSH key sprawl | ADR-001 |
| MIT open source | Community adoption, Anubis is preview/marketing for Sirsi | ADR-001 |
| Ka ghost detection | No competitor does cross-referenced ghost hunting | ADR-002 |
| 3-tier licensing | Open source → subnet → enterprise growth path | This document |
| Device-aware scanning | Platform detection adapts rules (Metal vs CUDA vs ROCm) | This document |

---

> **This plan is LOCKED.** Sprint plans reference specific sections.
> All feature work traces to a phase + sprint number here.
> Changes to scope or timeline require ADR review.
