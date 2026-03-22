# ‍‍‍𓂀 Sirsi Anubis — Continuation Prompt
**Date:** March 22, 2026 (Saturday, 7:20 PM ET)
**Session:** Coverage Hardening + Case Studies + Dogfooding Metrics
**Repo:** `github.com/SirsiMaster/sirsi-anubis`
**Path:** `/Users/thekryptodragon/Development/sirsi-anubis`

---

## CRITICAL: Read Before Starting

1. **Run `/session-start`** — the Thoth workflow at `.agent/workflows/session-start.md`
2. **Read `.thoth/memory.yaml`** — compressed project state (~100 lines). This replaces reading source files.
3. **Read `.thoth/journal.md`** — timestamped reasoning (10 entries).
4. **Read `ANUBIS_RULES.md`** — the 12 non-negotiable safety rules.
5. **Scope**: Launch execution + case studies + production polish.
6. **Deadline: Friday March 28** — April investor demos require complete product.
7. **All code compiles and 453 tests pass** — do NOT break the build.
8. **ADR-003 is ACTIVE** — every release must update BUILD_LOG.md, build-log.html, CHANGELOG, Thoth.
9. **Case studies**: Every product claim needs a `docs/case-studies/` entry with measured data.

---

## 𓁟 Thoth — Session Management

Thoth is the project's persistent knowledge system. It eliminates re-reading source files AND tracks session health. Two responsibilities:

### 1. Project Memory (Read at start, update at end)
| Layer | File | When |
|:------|:-----|:-----|
| Memory | `.thoth/memory.yaml` | **ALWAYS first** — architecture, decisions, limitations |
| Journal | `.thoth/journal.md` | When WHY matters — 10 timestamped entries |
| Artifacts | `.thoth/artifacts/` | Deep dives — benchmarks, audits, **roi-metrics.md** |

### 2. Context Window Monitoring (Track throughout session)

Thoth tracks session health to prevent context exhaustion. After every sprint:

```
## 📊 Session Metrics — Sprint [N]
| Metric | Value |
|--------|-------|
| ⏱️ Session elapsed | Xh Ym |
| 💬 Conversation depth | Turn N |
| 📂 Files ingested | N files (~XK lines) |
| ✏️ Output generated | ~N lines code/text |
| 🔀 Commits this session | N |
| 📝 Files modified | N |

### Context Health
| Indicator | Status |
|-----------|--------|
| Estimated fill | ~XX% |
| Checkpoint signals | None / Detected |
| Degradation risk | Low / Medium / High |

### Recommendation
🟢 Continue | 🟡 Wrap within 2-3 tasks | 🔴 Wrap NOW
```

**Heuristic model:**
- Turns 1–5: ~10–20% filled. Green zone.
- Turns 5–15: ~20–60% filled. Peak productivity.
- Turns 15–25: ~60–85% filled. Watch for quality.
- Turns 25+: >85% filled. Wrap protocol.

**Checkpoint signals:** If the system truncates the conversation, you are at 85%+. Wrap immediately.

**Wrap protocol (when 🟡 or 🔴):**
1. Commit all work
2. Push to GitHub
3. Update CHANGELOG.md, BUILD_LOG.md (per ADR-003)
4. Update `.thoth/memory.yaml` and `.thoth/journal.md`
5. Generate new `docs/CONTINUATION-PROMPT.md`
6. Report final session metrics

**AG Monitor Pro** is also installed as a VS Code extension (`~/.antigravity/extensions/shivangtanwar.ag-monitor-pro-1.0.0`) for real token tracking. Run `AG Monitor: Export Usage Report` for precise data.

---

## What Exists Right Now (All Working)

### Binary
- **Version:** 0.3.0-alpha (tagged `v0.3.0-alpha`)
- **Size:** ~8 MB (macOS arm64), ~2 MB (agent)
- **Go:** 1.22+, Cobra CLI, lipgloss terminal UI
- **Tests:** 453 passing, 15 test suites, 0 lint warnings
- **GoReleaser:** Verified — 12 binaries across 6 platforms all compile

### 17 CLI Commands

| Command | Module | Description |
|:--------|:-------|:-----------|
| `anubis weigh` | jackal | Scan workstation (64 rules, 7 domains) |
| `anubis judge` | cleaner | Clean with trash-first safety |
| `anubis ka` | ka | Ghost app hunter |
| `anubis guard` | guard | RAM audit + process slayer |
| `anubis sight` | sight | Launch Services + Spotlight repair |
| `anubis profile` | profile | 4 scan profiles |
| `anubis seba` | mapper | Interactive infrastructure graph |
| `anubis hapi` | hapi | GPU detection, dedup, snapshots |
| `anubis scarab` | scarab | Network discovery + container audit |
| `anubis mirror` | mirror | File deduplication (CLI + GUI) |
| `anubis install-brain` | brain | Neural model downloader |
| `anubis uninstall-brain` | brain | Remove neural weights |
| `anubis mcp` | mcp | MCP server (5 tools, Thoth included) |
| `anubis scales enforce` | scales | Policy engine enforcement |
| `anubis book-of-the-dead` | (hidden) | System autopsy |
| `anubis initiate` | (cli) | macOS permission wizard |
| `anubis version` | updater | Version + update check |

### Module Test Coverage

**15 modules HAVE tests:**

| Module | Coverage | Notes |
|:-------|:---------|:------|
| jackal | 93% | Scan engine |
| cleaner | **77.2%** | Safety + deletion — near target |
| ka | **42.7%** | Ghost detection (improved this session) |
| guard | 42 tests | RAM audit |
| brain | has tests | Neural downloader |
| mcp | has tests | MCP server |
| mirror | has tests | File dedup |
| scales | has tests | Policy engine |
| **ignore** | ✅ 17 tests | .anubisignore (new this session) |
| **jackal/rules** | ✅ 11 tests | 64 rule registry (new this session) |
| **profile** | ✅ 16 tests | Scan profiles (new this session) |
| **stealth** | ✅ 9 tests | Ephemeral cleanup (new this session) |
| **hapi** | ✅ 20 tests | GPU detect, dedup, snapshots (new this session) |
| **scarab** | ✅ 12 tests | Network discovery (new this session) |
| **sight** | ✅ 9 tests | LaunchServices (new this session) |

**2 modules have ZERO tests (low priority — display-only):**

| Module | Priority | Why low |
|:-------|:---------|:--------|
| **mapper** | 🟢 Low | Graph generation (display) |
| **output** | 🟢 Low | Terminal rendering (display) |

### Infrastructure
- CI: `.github/workflows/ci.yml` (lint + test + build)
- Release: `.github/workflows/release.yml` (goreleaser on v* tag push)
- VS Code extension scaffold: `extensions/vscode/`
- ADRs: 001 (founding), 002 (Ka ghost detection), 003 (build-in-public)

### Sirsi Pantheon (Repos)
| Repo | Deity | Version |
|:-----|:------|:--------|
| `sirsi-anubis` | 𓂀 Anubis | v0.3.0-alpha |
| `sirsi-thoth` | 𓁟 Thoth | v1.0.0 |
| `SirsiNexusApp` | ☀️ Ra (coming) | In development |

Thoth is standalone at `github.com/SirsiMaster/sirsi-thoth`:
- `npx thoth-init` auto-detects language, scaffolds `.thoth/`, injects into Cursor/Windsurf/Claude/Gemini/Copilot IDE rules
- No MCP required — just rules files

### Build-in-Public (Live)
- `docs/build-log.html` — public HTML page (Swiss Neo-Deco)
- `docs/BUILD_LOG.md` — sprint chronicle in markdown
- SirsiNexus Portal cross-linked ↔ Anubis
- "Weigh. Judge. Purify." tagline
- "From Anubis to Ra" section for roadmap context

---

### Priority 1: Launch Execution

Everything is tested, built, and documented. Time to ship.

```
- GitHub Release v0.3.0-alpha (goreleaser already verified, 12 binaries)
- Product Hunt submission (copy in docs/LAUNCH_COPY.md)
- Hacker News Show HN (copy in docs/LAUNCH_COPY.md)
- Investor demo rehearsal (script in docs/INVESTOR_DEMO.md)
```

### Priority 2: Expand Case Studies

Case study system exists at `docs/case-studies/`. Thoth study is complete.
Two more studies need writing (data exists, narratives needed):

```
1. Mirror Dedup Performance  — 27.3x faster, 98.8% less I/O (real benchmark data)
2. Ka Ghost Detection       — 23 GB Parallels remnants found (real discovery)
3. Expand SirsiNexus stub   — $111/session savings (star investor number)
```

### Priority 3: Production Polish

```
- Structured logging (replace fmt.Printf with slog)
- Platform abstraction interface (enable testing moveToTrash etc.)
- Convert pitch deck stub to full HTML slide (ADR-024 compliant)
- Linux folder picker (zenity)
- VS Code extension completion
```

---

## Key Context

1. **"Weigh. Judge. Purify."** — canonical tagline (was "Purge", updated to "Purify")
2. **Sirsi Pantheon** — Egyptian-themed tools: Anubis, Thoth, Ka, Ra, Seba, Hapi, Scarab
3. **Thoth is independent** — standalone repo, works without Anubis or MCP
4. **ADR-003** — build-in-public is mandatory, enforced by session workflow Step 6
5. **Voice rule**: Never "the user wanted/suggested." Use direct verbs: built, fixed, refactored.
6. **Audience**: GUI for everyone (parents, students, hobbyists). CLI for devs/AI engineers.
7. **Anubis→Ra**: Anubis is standalone preview; Ra is the full module coming in SirsiNexus
8. **April investor demos** — product must be complete by March 28
9. **Case study system** — `docs/case-studies/` + `scripts/thoth-roi.sh` in all 4 repos
10. **Pitch deck stub** — `SirsiNexusApp/docs/pitch-deck/slide-pantheon-proof.md` needs HTML conversion

---

## Dev Machine Specs

- **CPU:** Apple M1 Max (10 cores)
- **GPU:** Apple M1 Max (32 cores, Metal 4)
- **Neural Engine:** ✅ Available
- **RAM:** 32 GB unified memory
- **Disk:** 926 GB

---

## Rules of Engagement

1. **Read `.thoth/memory.yaml` FIRST** — do not re-read source files the memory already covers.
2. **Build → Test → Commit → Push** after every feature.
3. **Never break the build** — `go build && go test ./... && go vet ./...` must pass.
4. **ADR-003 is enforced** — every release updates 7 artifacts.
5. **Check actual struct field names** before using them.
6. **Binary size budget:** controller < 15 MB, agent < 5 MB.
7. **Monitor context** — report session metrics after every sprint. Wrap at 🔴.
8. **Voice**: Direct verbs only. No "the user wanted."
9. **Thoth manages the session** — memory for context, monitoring for health. Both are mandatory.

---

## Start Command

```bash
cd /Users/thekryptodragon/Development/sirsi-anubis
cat .thoth/memory.yaml
go build ./cmd/anubis/ && go test ./... && echo "✓ Ready"
```

Then begin Priority 1: Launch execution (GitHub Release v0.3.0-alpha)
