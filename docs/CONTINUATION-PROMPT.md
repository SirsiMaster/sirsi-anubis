# ‍‍‍𓂀 Sirsi Anubis — Continuation Prompt
**Date:** March 23, 2026 (Sunday, 4:24 AM ET)
**Session:** Session 9 — Ma'at Foundation + Coverage Sprint
**Repo:** `github.com/SirsiMaster/sirsi-anubis`
**Path:** `/Users/thekryptodragon/Development/sirsi-anubis`
**CI Status:** ✅ Green (pre-push hook active since session 8)

---

## CRITICAL: Read Before Starting

1. **Run `/session-start`** — the Thoth workflow at `.agent/workflows/session-start.md`
2. **Read `.thoth/memory.yaml`** — compressed project state (~135 lines). This replaces reading source files.
3. **Read `.thoth/journal.md`** — timestamped reasoning (13 entries).
4. **Read `ANUBIS_RULES.md`** — the 15 non-negotiable safety rules (includes A14, A15).
5. **Deadline: Friday March 28** — April investor demos require complete product.
6. **All code compiles and 470 tests pass** — do NOT break the build.
7. **ADR-003 is ACTIVE** — every release must update BUILD_LOG.md, build-log.html, CHANGELOG, Thoth.
8. **Rule A14 (Statistics Integrity)** — every public number must be independently verifiable.
9. **Rule A15 (Session Definition)** — a session = one AI conversation between continuation prompts.
10. **Pre-push hook is ACTIVE** — `.githooks/pre-push` runs golangci-lint before every push. Do NOT skip it.

---

## 𓁟 Thoth — Session Management

Thoth is the project's persistent knowledge system. Two responsibilities:

### 1. Project Memory (Read at start, update at end)
| Layer | File | When |
|:------|:-----|:-----|
| Memory | `.thoth/memory.yaml` | **ALWAYS first** — architecture, decisions, limitations |
| Journal | `.thoth/journal.md` | When WHY matters — 13 timestamped entries |
| Artifacts | `.thoth/artifacts/` | Deep dives — benchmarks, audits, **roi-metrics.md** |

### 2. Context Window Monitoring (Track throughout session)

After every sprint, report session metrics per the template in `.thoth/memory.yaml`.
Heuristics: Turns 1-5 ~10-20%, Turns 5-15 ~20-60%, Turns 15-25 ~60-85%, Turns 25+ >85%.
If truncation detected, wrap immediately.

---

## What Exists Right Now (All Working)

### Core Modules (19 internal packages)
| Module | Package | Description |
|:-------|:--------|:------------|
| Jackal | `internal/jackal/` | 58 scan rules across 7 domains |
| Ka | `internal/ka/` | Ghost app detection (17 macOS locations) |
| Mirror | `internal/mirror/` | File dedup (27.3x faster via partial hashing) |
| Guard | `internal/guard/` | RAM audit + process slayer |
| Cleaner | `internal/cleaner/` | Trash-first deletion with decision log |
| Hapi | `internal/hapi/` | GPU detection, dedup engine, snapshots |
| Sight | `internal/sight/` | LaunchServices ghost repair |
| Scarab | `internal/scarab/` | Network discovery + fleet sweep |
| Brain | `internal/brain/` | Neural model downloader + classifier |
| MCP | `internal/mcp/` | Model Context Protocol server |
| Scales | `internal/scales/` | Policy engine + violation reporting |
| Profile | `internal/profile/` | Scan profiles (quick/full/custom) |
| Stealth | `internal/stealth/` | Ephemeral execution + post-run cleanup |
| Ignore | `internal/ignore/` | .anubisignore file support |
| Logging | `internal/logging/` | slog-based structured logging |
| Platform | `internal/platform/` | OS abstraction: Darwin, Linux, Mock |
| Mapper | `internal/mapper/` | Filesystem mapper (no tests) |
| Output | `internal/output/` | Terminal rendering (no tests) |
| Updater | `internal/updater/` | Version check + advisory system |

### CLI Commands (17)
| Command | Module | Description |
|:--------|:-------|:------------|
| `anubis weigh` | jackal | Scan workstation (58 rules, 7 domains) |
| `anubis judge` | cleaner | Clean with trash-first safety |
| `anubis ka` | ka | Ghost app hunter |
| `anubis guard` | guard | RAM audit + process slayer |
| `anubis sight` | sight | Fix Spotlight ghost registrations |
| `anubis hapi` | hapi | GPU detect + VRAM status |
| `anubis scarab` | scarab | Network discovery |
| `anubis mirror` | mirror | File dedup (GUI or CLI) |
| `anubis seba` | - | Dependency graph visualization |
| `anubis book-of-the-dead` | - | Deep system autopsy |
| `anubis initiate` | - | macOS permission granting |
| `anubis install-brain` | brain | Download neural models |
| `anubis uninstall-brain` | brain | Remove neural models |
| `anubis mcp` | mcp | Start MCP server |
| `anubis scales` | scales | Enforce policies |
| `anubis profile` | profile | Manage scan profiles |
| `anubis version` | updater | Version + update check |

### Global Flags
- `--json` — JSON output
- `--quiet` — suppress non-error output
- `--verbose` / `-v` — enable debug logging (slog to stderr)
- `--stealth` — ephemeral mode

### Test Coverage
| Package | Tests | Coverage |
|:--------|------:|:---------|
| brain | 22 | Unit + integration |
| cleaner | 30 | 77% — safety-critical |
| guard | 12 | RAM + process |
| hapi | 20 | GPU, dedup, snapshots |
| ignore | 17 | Pattern matching |
| jackal/rules | 11 | Rule registry |
| ka | 28 | 42.7% — ghost detection |
| logging | 6 | Level modes |
| mcp | 5 | Server lifecycle |
| mirror | 12 | Dedup engine |
| platform | 11 | All implementations |
| profile | 16 | Scan profiles |
| scales | varies | Policy engine |
| scarab | 18 | Network + ARP parsing |
| sight | 4 | LaunchServices |
| stealth | 9 | Cleanup engine |
| **Total** | **470** | **17 suites** |

### Infrastructure
- CI: `.github/workflows/ci.yml` (lint + test + build)
- Release: `.github/workflows/release.yml` (goreleaser on v* tag push)
- **Pre-push hook**: `.githooks/pre-push` (gofmt + go vet + golangci-lint + go build)
- v0.3.0-alpha released on GitHub (6 binaries + checksums)
- Homebrew tap: `SirsiMaster/homebrew-tools` (repo exists, needs PAT)
- VS Code extension scaffold: `extensions/vscode/`
- ADRs: 001 (founding), 002 (Ka ghost detection), 003 (build-in-public)

### Case Studies (3 verified)
- `docs/case-studies/thoth-context-savings.md` — 98.7% context reduction
- `docs/case-studies/mirror-dedup-performance.md` — 27.3x faster, 98.8% less I/O
- `docs/case-studies/ka-ghost-detection.md` — 5-step algorithm, 17 locations

### Sirsi Pantheon (Repos)
| Repo | Deity | Role | Version |
|:-----|:------|:-----|:--------|
| `sirsi-anubis` | 𓂀 Anubis | Judgment — workstation hygiene | v0.3.0-alpha |
| `sirsi-thoth` | 𓁟 Thoth | Knowledge — persistent AI memory | v1.0.0 |
| `SirsiNexusApp` | ☀️ Ra | Portal — client platform | In development |
| *new* | 🪶 Ma'at | Truth — QA/QC governance agent | **Build this session** |

---

## What's Next

### Priority 1: 🪶 Ma'at — QA/QC Governance Agent

**Ma'at** is the Egyptian goddess of truth, justice, balance, and cosmic order. Her feather was the standard against which hearts were weighed. She is not a judge — she IS the standard.

**Ma'at's role in the Pantheon:** Every feature must be justified against canon before it exists. Ma'at weighs plans against execution and determines worthiness. She is the QA/QC deity — an embodied, empowered agent that constantly weighs, constantly checks, constantly assesses.

**This is the prototype for the future agent architecture.** In a later phase, all deities (Anubis, Thoth, Ka, Scarab, Guard) become autonomous agents. Ma'at goes first because she governs the others.

#### Phase 1: Foundation (this session)
```
1. ADR-004: Ma'at Architecture Decision Record
   - Role: QA/QC governance agent for the Sirsi Pantheon
   - Scope: plan verification, code quality, pipeline, test governance, release QA
   - Agent model: observe → assess → weigh → report/act
   - Canon linkage: no feature ships without plan linked to ADR/rule/priority

2. internal/maat/maat.go — core types + Verdict system
   - Verdict: Pass / Warning / Fail (with Feather weight score 0-100)
   - Assessment: what was weighed, against what standard, the verdict
   - CanonLink: ties a feature to its justification (ADR, rule, priority)

3. internal/maat/pipeline.go — CI pipeline monitoring
   - Poll gh run list for failures
   - Parse failure logs (gh run view --log-failed)
   - Categorize: lint → auto-fixable, test → report, build → report, infra → retry
   - Auto-fix lint issues (gofmt, misspell, goimports) + commit

4. internal/maat/coverage.go — test coverage governance
   - Per-module coverage thresholds (safety-critical = 80%+)
   - Compares current coverage against declared thresholds
   - Reports gaps with actionable context

5. internal/maat/canon.go — plan verification
   - Validates that features link to canon (ADR, ANUBIS_RULES, continuation prompt)
   - Scans commit messages for canon references
   - Reports unlinked changes as warnings

6. cmd/anubis/maat.go — CLI command
   - anubis maat              — full assessment (pipeline + coverage + canon)
   - anubis maat --pipeline   — CI status only
   - anubis maat --coverage   — test coverage audit
   - anubis maat --canon      — plan verification
   - anubis maat --watch      — daemon mode (future Phase 2)

7. Tests for all maat packages
```

#### Phase 2: Agent Mode (future session)
```
- Ma'at runs as background agent (--watch / daemon)
- Observes file changes, correlates with plans
- Posts quality scores (not just pass/fail)
- Blocks releases that don't meet the feather standard
- Pattern for converting all deities to agents
```

### Priority 2: Remaining Test Coverage
```
- Cleaner: 77% → 90% (safety-critical, Ma'at will enforce this threshold)
- Ka: 42.7% → 60% (test Clean with real file cleanup)
- Scanner edge cases: permission errors, symlink loops
```

### Priority 3: Homebrew Tap
```
- Create a GitHub PAT with repo:write scope for SirsiMaster/homebrew-tools
- Add it as HOMEBREW_TAP_TOKEN secret in sirsi-anubis settings
- Uncomment the brews section in .goreleaser.yaml
- Test with a new tag push
```

### Priority 4: Launch Execution
```
- Product Hunt submission (copy in docs/LAUNCH_COPY.md)
- Hacker News Show HN (copy in docs/LAUNCH_COPY.md)
- Investor demo rehearsal (script in docs/INVESTOR_DEMO.md)
```

### Priority 5: Production Polish
```
- Convert pitch deck stub to full HTML slide
- VS Code extension completion
- npm publish thoth-init
```

---

## Key Context

1. **"Weigh. Judge. Purify."** — canonical tagline
2. **Sirsi Pantheon** — Egyptian-themed tools: Anubis, Thoth, Ma'at, Ka, Ra, Seba, Hapi, Scarab
3. **Ma'at is the QA/QC deity** — she weighs plans against execution, constantly assessing
4. **Agent architecture** — Ma'at is the first deity to become an agent. All others follow.
5. **Thoth is independent** — standalone repo, works without Anubis or MCP
6. **ADR-003** — build-in-public is mandatory
7. **Rule A14** — every public number must be verifiable. No projections as measurements.
8. **Rule A15** — a session = one AI conversation between continuation prompts.
9. **Voice rule**: Never "the user wanted/suggested." Use direct verbs.
10. **April investor demos** — product must be complete by March 28
11. **v0.3.0-alpha is LIVE** — GitHub Release with 6 binaries
12. **Pre-push hook is active** — `.githooks/pre-push` gates every push with lint checks

---

## Session 8 Completed Work (for context)

- ✅ Platform interface wired into cleaner + mirror (replaced runtime.GOOS → platform.Current())
- ✅ CI lint fixes — 8 errors across 5 files (gofmt, govet/unusedwrite, misspell)
- ✅ Pre-push hook installed (.githooks/pre-push)
- ✅ CI green after 5 consecutive failures
- ✅ All artifacts canonized (CHANGELOG, memory, journal, continuation prompt)

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
9. **Thoth manages the session** — memory for context, monitoring for health.
10. **Rule A14**: Include the command to reproduce any public number.
11. **Ma'at governs quality** — every feature must link to canon. No unjustified code.

---

## Start Command

```bash
cd /Users/thekryptodragon/Development/sirsi-anubis
cat .thoth/memory.yaml
go build ./cmd/anubis/ && go test ./... && echo "✓ Ready"
```

Then begin Priority 1: Build Ma'at — the QA/QC governance agent.
