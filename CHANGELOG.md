# Changelog — Sirsi Pantheon
All notable changes to this project are documented in this file.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and [Semantic Versioning](https://semver.org/).

**Building in public** — see [docs/BUILD_LOG.md](docs/BUILD_LOG.md) for the full narrative.

---

## [Unreleased]

> Cuts as **v0.23** per Codex review `20260521-codex-pantheon-tui-elimination-phase0-review`. Interactive surface direction reopened on 2026-05-29 (ADR-020) and closed as **Hybrid C**: a new Mole-grade TUI ships first cross-platform; Mac native follows later as the polish-bar upgrade. The `v1.0-alpha.0` slot now belongs to the first installable TUI cut, not the first Mac app.

### Added
- **Menubar bounds its own registry growth — `pruneOwnStaleRecords`** (authored by reaped sibling thread `thr-82f130c223a1b3d7`; committed by claude-pantheon under claude-home authorization `033748` after the author thread was reaped mid-edit — orphaned-work adoption, A23 provenance preserved in the commit trailer). On each menubar launch, after registering, the resident surface reaps this host's dead PIDs (via the ADR-024 Amendment 1 composite `ReapDeadThreads`) then deletes its **own** terminal records (`menubarAgentID`-scoped, keeps the live thread) older than a 1h retention window — so a frequently-relaunched resident surface never accretes unbounded reaped rows in `threads.json` (the A27 write-amplification → Spotlight `mds_stores` class). Agent-scoped: never touches other agents' history. Best-effort: a prune failure never blocks launch. `go build ./cmd/sirsi-menubar` green. Refs: PANTHEON_RULES.md A27/A21, ADR-022/ADR-024 Amendment 1; Changelog: v0.23.
- **ADR-026 step 4a (surface chrome) — menubar serves `/api/node-status` in-process** (claude-pantheon, surface-chrome lane handed off by claude-home `032339`). Wires `NodeStatusFn` into the menubar's embedded `dashboard.Config` (`cmd/sirsi-menubar/main.go`) so the resident menubar's in-process dashboard serves the ADR-026 ops read-model (`GET /api/node-status` [+ `?view=summary`]) — the producer the menubar previously left nil (503). Reuses the menubar's own `resolveRouterRoot()` (which carries the launchd `cwd=/` fallback, ADR-021) and derives the repo root for `router.CollectNodeStatus`; unresolved root → error → 503 (the designed graceful degrade, same as a nil `StatsFn`). Read-only, no destructive surface. Non-racing: `main.go` only, separate from the concurrent `register.go` idempotency fix. `go build ./...` + `go vet ./cmd/sirsi-menubar` clean. **Remaining (step 4b/5):** the NSMenu `OpsSummary` row rendering (glyph from `WorstIcon` + "Open full dashboard" affordance) sequences after the menubar `register.go` work lands and needs GUI verification; the TUI 4th pane (step 5) is deferred — `internal/tui/` is Gate-2 scaffold only (no launch path / no data-consuming panes per `docs/TUI_DESIGN_PROOF.md`), so a "4th pane" needs the live-data view infra behind the next TUI gate. Refs: docs/ADR-026-HORUS-OPS-DASHBOARD.md, ADR-021 (workstation-scoping), ADR-024 §1 (resident surface); Changelog: v0.23.
- **ADR-026 implemented (steps 1–3) — Horus ops-dashboard read endpoint + `sirsi router node-status` verb** (claude-home owns; design APPROVED by claude-pantheon `022950`; lane boundary `235419`/`235652`). Closes the canon/implementation gap where Rule A27 references `router node-status` — a verb that did not exist — and where the dashboard contract matrix had *Router ack → MISSING (no `/api/router/*`)*. **The gap was exposure, not computation**: `router.CollectNodeStatus()` already aggregates the whole operator read-model (agents + wake-health, queue, dispatch failures, live/stale threads w/ `os_state` per ADR-022, daemon + binary-drift per ADR-023, agent CLI auth) into one `NodeStatus` — but it was trapped in Go. **Step 1 (contract):** `router.NodeStatus` gains `SchemaVersion` (frozen at `1.0.0`; bumps only on breaking change, never on additive) — claude-pantheon caveat from the verdict. **Step 2 (endpoint):** `internal/dashboard/nodestatus.go` adds `GET /api/node-status` serving the typed `router.NodeStatus`, plus `?view=summary` returning a bounded `OpsSummary` projection (top-N=12 agents by pending+live signal, `more_agents` overflow row for the NSMenu budget — claude-pantheon caveat #2). The summary is a **pure reduction** of the same NodeStatus (every field derived; nothing sourced independently — the "one read-model" invariant applied to reads, like the frozen action contract applies to writes). New `NodeStatusFn NodeStatusCollector` producer hook in `dashboard.Config`; nil-safe 503 graceful-degrade. **Read-only**: zero `ConfirmGuard` surface, no method-gating, no side effects. **Step 3 (CLI):** `sirsi router node-status [--json]` (`cmd/sirsi/routernodestatus.go`) — `--json` shape is byte-identical to the HTTP body, so the menubar + TUI + scripts all consume the same bytes. Smoke run against the live registry surfaced 7 phantom `pid=0/os=unknown` records — *the exact ADR-024 Amendment 1 finding (3) class the verb now makes visible*. 5 tests `internal/dashboard/nodestatus_test.go` (full-contract serve, summary derivation + drift flag + needs-login projection, bounded truncation, 503 unwired, 500 collector-error) `go test -race ./internal/dashboard` green; `go build ./...` ok. **Boundary held** (no surface chrome touched): menubar rows + TUI 4th pane stay with claude-pantheon. Routed for codex arch-verify next. Refs: docs/ADR-026-HORUS-OPS-DASHBOARD.md, docs/HORUS_OPS_READMODEL_R4_INVENTORY.md, docs/user-guides/router-node-status.md; PANTHEON_RULES.md A8/A10/A22/A27/A26; Changelog: v0.23.
- **ADR-024 Amendment 1 implemented — worker-lifecycle gate + (pid, start_time) reap-key** (claude-pantheon sole writer, CLAIM `20260602-024522`; design APPROVED by claude-home `025217`; implementation routed to codex for review-of-code). Fixes two CTR registration-hygiene defects behind the registry churn. **(3) Reap-key = composite identity, not bare PID** — the highest-confidence systemic bug: `PIDStateOf(pid, startedAt)` now keys on `(pid, start_time)`, returning a new `PIDRecycled` state when a live PID's start signature differs from the one recorded at registration (the OS recycled the number onto a different process). `defaultPIDStart` reads `ps -o lstart=` (one cheap shell-out, same shape as the `stat=` probe; `""` on Windows/legacy → bare-PID fallback, zero regression). `Thread.StartTime` is captured at register; `ReapDeadThreads` and `RegisterThread`'s idempotency fast-path both key on the composite, so a recycled PID is reaped (not resurrected) and a re-register on a recycled PID mints fresh. **Critically, a genuinely live thread whose start still matches is never reaped** — the exact false-positive that retired live sessions this session. **(2) Worker-lifecycle gate** — `sirsi thread register` refuses one-shot `--print`/`-p` workers (injectable `oneShotProbe` + `ephemeralWorkerSkip`; no-op, not error): an ephemeral worker is neither an interactive nor a resident surface, so it must not enroll a persistent thread (the dominant phantom-record source). Finding (1)/menubar excluded per ruling `023813` (surface-chrome lock-holder owns it). New tests: `internal/router/adr024_amend_test.go` (composite-state matrix, recycled-reaped, live-survives, composite fast-path) + `cmd/sirsi/adr024_amend_test.go` (selective gate proves interactive surfaces still register). `go test -race ./internal/router ./cmd/sirsi` green; `go build ./...` ok; `start_time` capture smoke-verified via real `ps -o lstart`. Refs: docs/ADR-024-ONE-WATCHER-PER-SURFACE.md § Amendment 1, ADR-022 (OS-truth liveness, corrected), PANTHEON_RULES.md A27/A26/A16/A21; Changelog: v0.23.
- **ADR-025 implemented — Thoth-gated exit + resumable suspend (R3)** (claude-pantheon owns, router item `203859`; queued after the ADR-024 lane; codex round-2 *approve* against `9e74be4`). Completes A27's lifecycle (`register → heartbeat → suspend ⇄ resume → close`) so a thread can no longer vanish lossy. Core (`414142f`) landed the `suspended` status (resumable-but-not-live: prune skips it, the reaper skips it, `Heartbeat` rejects it, `RegisterThread` bypasses the live fast-path) + `SuspendThread`/`ResumeThread`. This lane adds the rest: **CLI verbs** `sirsi thread suspend` (`--self`/`--thread`, syncs Thoth first for a fresh `thoth_ref`, snapshots owned open items + resume prompt, stops the fs-watcher) and `sirsi thread resume` (restores owned items, prints the resume prompt, returns the ADR-024 watcher spec for re-arming) — `cmd/sirsi/threadsuspend.go`. **`ReconcileExits`** (`internal/router/threads.go`) is the authoritative SessionStart gate (ADR-025 §4): a **stale active** record (soft-exit/`clear`) heals in place active→suspended after a retro sync; a **reaped** record (hard kill) is never revived — if memory is recoverable a new suspended **successor** is minted carrying `reaped_from`, else an `UNRECOVERABLE` warning surfaces (never silent). Idempotent via `hasSuccessorFor` + a 24h lookback; host- and agent-scoped (each surface heals its own lineage). Exposed as `sirsi thread reconcile [--agent]`. **Hooks** (user-scope `~/.claude/settings.json`, ADR-024 §4 default-on, `SIRSI_SUPERVISOR=0` opt-out): new **`SessionEnd`** → `thoth sync` + `thread suspend --self` (best-effort, visible error — `SessionEnd` cannot block); **`SessionStart`** gains `thread reconcile --agent` (the guaranteed gate). 9 `internal/router` ADR-025 tests (6 core + 3 reconcile: stale-in-place, reaped-successor-then-warn, host/agent scoping) `go test -race` green; verbs smoke-tested register→suspend→resume→close on a throwaway thread. Note (codex): re-`register` matching a suspended record currently mints a fresh thread (per the shipped core test) rather than auto-adopting via resume — explicit `resume` is the supported path; flagged for review. Refs: docs/ADR-025-THOTH-GATED-EXIT.md, PANTHEON_RULES.md A27/A26/A22; ADR-024/ADR-022/ADR-014; Changelog: v0.23.
- **ADR-026 proposed — Horus ops-dashboard, one typed read-model for the operator surface** (claude-home owns the ops-view content lane; routed to claude-pantheon for review per the ratified lane boundary `20260601-235419`/`235652`). Realizes ADR-015 ("the dashboard **is** Horus") as a real operator command-center, distinct from the `internal/horus` code-graph. **Key finding: the gap is exposure, not computation** — `router.CollectNodeStatus()` already aggregates the entire operator read-model into one `NodeStatus` (agents+wake-health, router queue, dispatch failures, live/stale threads with `os_state` OS-truth liveness per ADR-022, daemon+binary-drift per ADR-023, agent CLI auth), but it is trapped in Go: **not** in the frozen dashboard contract (matrix row *Router ack → MISSING, no `/api/router/*`*), **no** CLI verb (Rule A27 canon references `router node-status`, a verb that does not exist), **no** surface render. ADR-026 promotes `NodeStatus` to a frozen additive contract served at typed **`GET /api/node-status`** (+ `?view=summary` → `OpsSummary` for the menubar), adds the **`sirsi router node-status [--json]`** verb, and defines menubar/TUI read-only *projections* of the one read-model (no surface re-aggregates — the action-contract principle applied to reads). **Challenges the resume framing's `/api/horus` name** (Rule A23): `/api/horus/*` is already the code-graph namespace, so the ops-view is `/api/node-status` (one name → one meaning). Read-only — zero destructive surface, nothing to confirm-gate. Boundary held: claude-home defines the read contract, claude-pantheon owns surface chrome. Also adds `docs/HORUS_OPS_READMODEL_R4_INVENTORY.md` (the R4 capability inventory — per-surface watcher specs from `watcherspec.go` + the read-model source/exposure ledger). Design-phase, no code until codex/claude-pantheon review. Refs: docs/ADR-026-HORUS-OPS-DASHBOARD.md, PANTHEON_RULES.md A27/A26/A22/A23/A19; ADR-015/020/022/023/024; Changelog: v0.23.
- **ADR-024 implemented — one watcher per surface, router-prescribed** (claude-home assigned, router item `200410`; A24 autonomy; commits `7c4cda5`/`10c5e93`/`9288534`). Kills the three-heartbeat accretion (auto fs-watcher + caffeinator + `/loop` Monitor) where only one mechanism actually woke the agent while all three churned CPU + Spotlight `mds_stores`. **Decision 2:** `sirsi thread register` stops auto-spawning the fs-watcher — it's now a pure handshake that RETURNS the canonical watcher via a new `internal/router` spec table (`watcherspec.go`, the R4 capability inventory in code): surface → `{type, mechanism, arm_instruction, heartbeat_interval_s, watches_inbox, resident}`. claude ⇒ `loop-monitor`; menubar/tui/IDE/macapp ⇒ resident `native-runloop` (not inbox workers, preserves ADR-020); codex/gemini/daemon surfaces mapped too. **Decisions 3+4 (supervisor hook):** retires the per-claude caffeinator + fs-watcher; `register --json` exposes `watcher.arm_instruction`; the hook does **check-then-arm on every SessionStart/wakeup** (F1) keyed on **OS truth `pgrep -f thr-<thread_id>`** — never TaskList (F2: falsely empty ⇒ duplicate) and never the shared `DIR=` loop body (collides with other agents' loops on a shared host — claude-deck's correction `838ad66`). `SIRSI_SUPERVISOR` gates it: default `on` = arming injection, `enforce` adds the Stop-gate backstop, `0` suppresses managed arming + Stop-gate while `register` still returns the spec. **Decision 5 (one inbox):** the reader already scanned `items/` only; the F3 root cause was sender guidance — `notify.go` now directs replies to `items/` via `router close --result` / `router send --type`, and `work.Item` gains a `type:` field (proposal/review/decision) so those collapse onto one addressed item instead of separate `reviews/`+`decisions/` channels. All **7 codex acceptance tests** pass (register-no-spawn/menubar-resident/idempotent-spec, SIRSI_SUPERVISOR=0, F1, F2, F3) across Go + python; `go test -race` green. Remaining deploy step: migrate the hook to user-scope `~/.claude/` (Decision 4 default-on). Refs: docs/ADR-024-ONE-WATCHER-PER-SURFACE.md, PANTHEON_RULES.md A27/A26/A24/A11/A19; Changelog: v0.23.
- **Dashboard action contract frozen — E1/E2/E3/E5** (codex freeze-gate ruling, router item `162436`; single-writer lane = claude-pantheon; commits `8675796`/`f5b3084`/`edb8a74`). Converges all four surfaces (CLI/menubar/TUI/SwiftUI) onto one typed `internal/dashboard` contract instead of each inventing action semantics. **E3:** typed `ActionRequest`/`ActionResult`/`PreparedAction` (`contract.go`) + typed `StatsResponse` mirroring the menubar `StatsSnapshot` JSON tags, so `/api/stats` is no longer an opaque `[]byte` at the boundary (decode-through with honest passthrough fallback). **E2:** `ConfirmGuard` (`confirm.go`) — server-enforced, single-use, tokenized two-phase confirmation (SHA-256 action hash; rejects missing/unknown/expired/mismatched/reused tokens; default = dry-run/prepare; the token is the safety boundary, no client `confirm()`). A shared `requireConfirm()` gates all 5 destructive endpoints; **fixes the `/api/clean` defect where an omitted `dry_run` deleted for real** (Rule A1, PARAMOUNT). **E1:** canonical `ActionSpec` registry (`actions.go`) folding the legacy 8 actions with the 12 gap-list actions (audit, maat, risk, network/fix, thoth/sync, seshat/ingest, net/align, ra/deploy/kill/collect), reachable via `GET /api/actions` + typed `POST /api/run`. **E5:** `POST /api/run` accepts a typed `ActionRequest` (server-defined base args + opt-in caller positional args — no arbitrary injection), runner+SSE retained for streaming; legacy `?cmd=` kept but cannot fire destructive. E4/E6 + renice-exemption deferred as documented fast-follow. Surface work (menubar Step 2, TUI wiring) is unblocked pending codex's implementation review. `go test -race ./internal/dashboard`: green. Refs: docs/DASHBOARD_CONTRACT_MATRIX.md, ADR-020, PANTHEON_RULES.md A1/A4/A16/A27; Changelog: v0.23.
- **One build-version contract + local binary-drift detection** (ADR-023; claude-home owns, codex-pantheon verifies — router item `20260601-143914...plan-version-contract`). Fixes the **CTR deploy-drift class** behind ADR-022: the fix `ca6e343` reached `~/.local/bin/sirsi` but `/opt/homebrew/bin/{sirsi,sirsi-menubar}` stayed stale, so the menubar ran the OLD `internal/router` reaper silently. Root cause was distribution, not logic — five disagreeing `var version` literals (`v0.21.0`/`v0.20.0`/`v0.4.0-standalone`…), no single source, no drift visibility. New `internal/version/build.go` holds one stamped contract (`Version`/`Commit`/`Date` + `Info`/`Current`), with a `debug.ReadBuildInfo()` fallback so plain `go build` self-reports honestly (A23). All 7 `cmd/*/main.go` literals replaced; `.goreleaser.yaml` + `Makefile` ldflags unified onto `internal/version.*`. `sirsi version --json` and a new `sirsi-menubar version [--json]` emit the same `Info` shape. New `internal/selfupdate` (`DetectMethod`/`ScanHost`/`BuildReport`) discovers sibling binaries and probes each via `version --json` (200 ms, **no network**), classifying **D2 sibling drift** + **D3 PATH drift**; no `internal/router` import (no cycle). `sirsi doctor` appends a `binary-drift` finding that the SessionStart `health-line.sh` surfaces automatically. Proven on the real binary: stamped `sirsi`+`sirsi-menubar` both report `v0.22.0-beta`+commit; a staged stale sibling triggered D2 (severity 2) and rendered `health:🔴 … binary-drift`. Verified atomic `sirsi self-update` + cosign signing deferred (follow-up router items). `go test -race ./internal/version ./internal/selfupdate`: green. Refs: PANTHEON_RULES.md A13/A23/A7; Changelog: v0.23.

### Fixed
- **CTR false-active resurrection + zombie-blind reaper** (claude-home owns; codex-pantheon verifies — router item `20260601-024355...execute-fix-ctr-false-active`). Three coupled defects let dead threads show `active` and the registry balloon to 1050 records: **(B1)** a late heartbeat could revive a terminal record; **(B2)** the reaper used `kill -0`, which a defunct (zombie `Z`) process answers, so zombies were never reaped; **(volume)** non-idempotent registration minted a new record per tick. Fixes (`internal/router/`): terminal `reaped`/`stale-heartbeat` statuses + `IsTerminal()`; `Heartbeat` refuses to revive terminal records; OS-truth liveness primitive (`liveness*.go`, build-tagged, injectable per A16/A21) distinguishing alive/gone/**defunct** via `ps -o stat=`; `ReapDeadThreads` (host-scoped); idempotent `RegisterThread` (one live session → one thread); `thread list` integrity warning + `💀`; `CollectNodeStatus.os_state` so Horus can't show a dead PID live; new `sirsi thread prune` to clear tombstones (the write-churn feeding Spotlight `mds_stores`). Proven on the real binary (register→kill→reap); live registry swept 1050→0, other lanes' 19 pending preserved. `go test -race`: green. Refs: PANTHEON_RULES.md A27/A26/A21/A16; Changelog: v0.23.
- **A27 watcher re-arm idempotency keyed on thread_id, not the shared loop body** (ADR-024). The prescribed re-arm check `pgrep -f "<thread-specific signature>"` was being implemented against the shared `DIR=.agents/idea-router/items` string — which every Claude surface runs verbatim — so on a shared host the check matches *other agents'* live loops and falsely reports "already armed," leaving a thread registered-but-unwatched. Observed 2026-06-01: a fresh `claude-deck` session's check matched `claude-home`'s running loop (`thr-7a3f16…`). ADR-024 `arm_instruction` and §3 now require the signature to include the **thread_id** (`pgrep -f thr-<thread_id>`), the same `(agent_id, pid)` identity ADR-022 reaps on. Refs: docs/ADR-024-ONE-WATCHER-PER-SURFACE.md; PANTHEON_RULES.md A27; Changelog: v0.23.

### Removed
- **v0.22 BubbleTea TUI implementation removed** (ADR-018, 2026-05-21; status now *Partially In Force — Amended By ADR-020*). All `internal/output/tui*.go` files (~4,800 LOC, 20 files), the TUI gateway entry point, `sirsi status --live`, and the no-args TUI launcher are gone. Binary dropped 24.2 MB → 22.2 MB. The `charm.land/bubbletea/v2` dependency is removed from `go.mod`. **Scope clarification (2026-05-29):** this removed the *unreleasable v0.22 implementation*, not the TUI surface category. ADR-020 reopened the surface decision and closed Hybrid C; a new Mole-grade TUI is the next interactive deliverable, designed from scratch (no restoration of the deleted code as foundation). The "intentional and immediate" framing applies to the v0.22 deletion only.
- **`sirsi` with no args** no longer launches an interactive surface; it now prints help. This holds for v0.23 until the new TUI lands. Per-verb behavior and flags are unchanged — see `docs/CLI_COMPATIBILITY.md` for the full matrix.

### Reopened
- **Interactive surface decision reopened and re-closed as Hybrid C** (ADR-020, 2026-05-29). After user direction *"TUIs are the wave… if we can't build one, it calls into question our ability to build Sirsi overall,"* the surface category was put under multi-track evaluation. Closure: new Mole-grade TUI ships first on macOS/Windows/Linux; Mac native SwiftUI follows in a later phase. No `internal/tui/` Go code lands before a `docs/TUI_DESIGN_PROOF.md` clears codex review (per ADR-020 §"Why This TUI Will Be Different" Gate).

### Proposed
- **ADR-021 — Deities Must Not Assume Single-Repo** (proposed 2026-05-31, routed to codex-pantheon). The menubar's Osiris reported `osiris assess failed` because `cmd/sirsi-menubar/stats.go:84` defaults `RepoDir: "."`, and a LaunchAgent-spawned menubar runs with cwd=`/` (not a git repo). The fix is not "pin a repo" — it names a design principle: deities whose domain is workstation-scoped (Osiris risk, Anubis hygiene, Ma'at quality, Isis pressure) must source scope from **CTR workstation discovery** (`sirsi thread` registry + `sirsi thread discover`), never the process cwd. Osiris becomes a workstation-wide risk aggregator; non-git/zero-repo states degrade to benign, never `failed`. No code lands before the ADR is accepted. See `docs/ADR-021-DEITIES-NOT-SINGLE-REPO.md`.

### Added
- **Menubar registers as a resident CTR router surface — surface-ladder Step 1** (claude-pantheon, commit `543e959`; approved by codex-pantheon router item `20260601-055029` with constraints). The user's interactive-surface ladder is `CLI > menubar > TUI > SwiftUI` + IDE plugins, and "registration" means **every surface is a router-registered thread** (A26/A27), not just a renderer. The existing Go `fyne.io/systray` menubar (`cmd/sirsi-menubar/`) — already running the in-process `internal/dashboard` server, guard bridge, and a 4h jackal scan — now registers one thread (`agent=sirsi-menubar`, `surface=menubar`) bound to its **own PID** (`os.Getpid()`, not a spawned Terminal child), heartbeating on a **bounded 60s ticker deliberately decoupled from the stats tick** to avoid the registry write-amplification that fed `mds_stores`. New `cmd/sirsi-menubar/register.go`: `resolveRouterRoot` (env → walk-up → conventional checkout; best-effort, skips registration rather than blocking launch), idempotent register (reuses live `(agent_id,pid)`; cross-restart dead records retired by ADR-022 OS-truth reaping, which `thread list` applies on read), bounded `heartbeatLoop`, `closeMenubarThread`. An explicit SIGTERM/SIGINT handler in `onReady` closes the thread on graceful shutdown (logout/launchd stop); `kickstart -k` SIGKILLs so that path relies on reaping. Verified live: `sirsi thread list` shows exactly **one** active menubar record across repeated restarts; real-SIGTERM close confirmed (record → `closed`). Surfaced a real deploy-drift bug in passing (two `sirsi-menubar` binaries — `/opt/homebrew/bin` 18.4MB vs `~/.local/bin` 12.1MB; redeployed both), feeding claude-home's `sirsi doctor`/`self-update` plan (`143914`). Menu-action → dashboard-contract refactor is **Step 2** (not in this change). Canon amendments (A27 resident-surface wording, ADR-020 ladder) pending codex's call on ADR form.
- **`internal/tui/` scaffold — Phase-2 batch-2 Gate 2** (claude-pantheon, routed to codex-pantheon for review). Gate 1 cleared (codex approved 2026-05-31 `reviews/20260531-codex-pantheon-tui-design-proof-gate1-review.md` + user sign-off), so the design proof becomes code — **scaffold only**, per Codex's review scope: primitives, state model, command registry, renderer contract, fixture screens, tests. **No functional resurrection of v0.22 and no operator-facing launch path** — `cmd/sirsi` does not import the package and `sirsi` no-args still prints help (`docs/CLI_COMPATIBILITY.md` unchanged). Contents: the 5 layout primitives (`primitives.go`), the binary split-tree + 3 named layouts (`layout.go`), the `Command` registry with **data-driven status hints** (`command.go` — a hint is a projection of registered commands, so a dead key cannot render: §7 delta 2), the state/`Reduce` contract with the Rule A1 destructive-confirm guarantee (`state.go` — `clean`/`ra.kill` arm a confirm modal, never fire on one keystroke), the renderer contract with a **first-class linear/no-altscreen renderer** (`renderer.go`), the closed **glyph-width policy** covering *all* non-ASCII layout glyphs — box-drawing, `◉`, `▸`, `→`, `✓`, `⏱`, `…` — each with measured single-width + ASCII fallback (`glyph.go`, addressing Codex's Gate-1 precondition that safe-glyph discipline extend beyond hieroglyphs), the semantic color ladder with `NO_COLOR`→attribute-only (`color.go`), and the 3 proof screens as fixtures (`fixtures.go`). Tests satisfy every Codex scaffold precondition: no hint references an unregistered/unwired command; fixture renders at **80×24, 100-col, 120×40** stay within width and use only grid-safe glyphs; width/fallback asserted for every non-ASCII glyph; `NO_COLOR`/reduced-motion/no-altscreen fixtures. `go test -race`: green, **93.6%** coverage; `golangci-lint`: 0 issues. Supersedes the prior "(draft) awaiting review" status of `docs/TUI_DESIGN_PROOF.md`.
- **`sirsi thread discover` — live-session reconcile for CTR** (commit `10a97b7`, codex-pantheon **approved** 2026-05-31). After a reboot the thread registry goes cold: every prior PID is reaped and nothing re-registers, because registration was manual-only (`sirsi thread register`). `discover` queries running agent processes on this host (bounded `pgrep -x`/`lsof` — no broad scans, no Python), resolves each one's cwd to an agent in `agents.json`, and registers live, *mappable* sessions anchored to their PID so the existing watcher/reaper lifecycle owns them. **Honesty by construction (Rule A23):** home-launched sessions (no repo binding) are reported `unmappable`, never registered; a cwd matching two agents of one surface (the real `codex-homebrew` vs `codex-homebrew-tools` collision) is reported `ambiguous`, never guessed. Pure decision logic in `internal/router/discover.go` (`ReconcileDiscovery`, surface-scoped longest-ancestor match) with 9 unit tests per Rule A16 (no real processes); enumeration + apply + CLI in `cmd/sirsi/threaddiscover.go`; `--self` for SessionStart-hook use (Phase 2); stable snake_case `--json` for sweeps. **Phase 1.5:** wired into the hourly verification `sweep.sh` (alongside `thread scout`) so a cold registry self-heals for repo-launched sessions. This is the CTR-discovery primitive **ADR-021** names for workstation-scoped deities. Live at install: `discovered=6 registered=0 unmappable=5 skip=1` — confirming the premise (every session was home-launched) and the already-registered skip path.
- **`docs/TUI_DESIGN_PROOF.md` (draft)** — Phase-2 batch-2 **Gate 1** deliverable per ADR-020 §"Why This TUI Will Be Different." The first artifact of the Hybrid C TUI track is *not code* — it is a design proof that must clear codex-pantheon + user review before any `internal/tui/` package is created. Specifies the 5-primitive layout system, density/typography rules, the load-bearing **glyph budget** (Egyptian hieroglyphs forbidden in layout-bearing cells; deity identity via BMP-safe sigils + color + a startup font-capability probe), modeless keyboard model with command palette, error-state altitudes, accessibility (linear screen-reader mode, `NO_COLOR`, high-contrast, reduced-motion), three canonical ASCII mocks (scan / Ra fleet / router inbox), and the six structural "different from v0.22" deltas. Awaiting review.
- **Knowledge Substrate** — semantic verification layer via the Understand-Anything Claude Code plugin. First run on 2026-05-26 produced `.understand-anything/knowledge-graph.json` (3,340 nodes, 6,947 edges, 9 architectural layers, 14-step pedagogical tour). Codified as **ADR-019**.
  - User-facing: `docs/user-guides/knowledge-substrate.md`
  - Web page: `docs/pantheon/knowledge-substrate.html` (→ `sirsi.ai/pantheon/knowledge-substrate`)
  - Case study: `docs/case-studies/2026-05-26-knowledge-substrate-day-1.md`
  - Three-tool split codified: Thoth (memory) / Seba (architectural map) / Knowledge Substrate (semantic verification). No deity sovereignty changes — Seba's mapping authority unchanged.
  - Bidirectional sync: `.thoth/memory.yaml` gains a `## Knowledge Graph (Understand-Anything)` block + `sync_protocol`; rule in `~/CLAUDE.md` so every Thoth-enabled repo auto-updates after `/understand` runs.
  - Long-term direction: cross-repo, cross-agent hypergraph on **Hedera Consensus Service**. Workspace-canon builder vision at `~/Development/HYPERGRAPH_VISION.md`; pointer added to `~/Development/AGENTS.md` § Knowledge Substrate.
  - CLI surface spec'd in ADR-019 § 6 (`sirsi hypergraph status|refresh|chat|explain|diff|layers|tour|export`), gated by `configs/hypergraph.yaml` `enabled:` and a `hypergraph` build tag. Implementation pending codex-pantheon review.
- `docs/CLI_COMPATIBILITY.md` — concise per-verb compatibility matrix for the v0.22 → v0.23 transition.
- **FSEvents-driven wake** — `.agents/idea-router/wake.example.plist` is a deployable launchd template. Copy to `~/Library/LaunchAgents/`, fill in paths, `launchctl load`. launchd watches `state.json`, `items/`, `proposals/` for any change and fires ONE dispatch pass per change (`sirsi router run --once`). `ThrottleInterval=10` prevents refire loops. **Zero idle process** — no polling daemon, no cron, no heartbeat. Replaces the prior 1-second-interval polling daemon (86,400 reads/day → ~0 idle, dispatch latency dropped from ≤1s to milliseconds).
- `internal/work/work_test.go` — round-trip coverage for Send/Get/ListInbox/Close, YAML quoting edge cases, status transitions.
- `internal/work` YAML quoting: `Send` now writes frontmatter values as double-quoted strings (`from: "claude-pantheon"`), escaped properly so titles/agent-ids containing colons, `|`, `*`, `&`, newlines, etc. round-trip cleanly. `Get`/`ListInbox` parsing handles both quoted and bare forms.
- `routercmd.go` split: `workRoot()` (read-only, no mkdir side-effect) for `status/pull/show`, `workRootEnsure()` (mkdir items/) for `send/close`. Audits no longer materialize an items/ directory.

- **Pull-model work queue** — bare-minimum any-to-any routing between agent threads, no daemon required:
  - `sirsi router send --from <id> --to <id> --title <s> --instructions <text-or-@file>` — write one work item
  - `sirsi router pull <agent>` — list open items addressed to an agent
  - `sirsi router show <id>` — print full body + frontmatter
  - `sirsi router close <id> --result <text-or-@file>` — flip status to closed
  - `sirsi router status` — count open vs closed items, open by recipient
  - New `internal/work` package; items live as plain markdown under `.agents/idea-router/items/`
- `.claude/hooks/router_inbox_check.py` now also surfaces pull-model items (was legacy-only).
- `TestRouterPullModelRoundtrip` integration test.

### Removed
- **Legacy push-model router CLI verbs**: `watch`, `run`, `daemon`, `work` (--poll), `install-agent`, `uninstall-agent`, `service-status`, `node-status`, `smoke`, `submit-existing`, plus the legacy `inbox` verb. The pull model replaces all of them with one mental model (file in items/ + recipient pulls). The `internal/router` Go package is left intact (still imported by `agentcmd.go`, `threadcmd.go`, `setup.go`, `internal/mcp/tools.go` for thread/agent registry reads) — a follow-up can prune the now-dead dispatcher/runner/launchctl code from that package.
- Any running `sirsi router daemon` process (e.g., from the launch agent) keeps running on its loaded binary, but restarts will fail with "unknown command" — uninstall the launch agent with `launchctl unload ~/Library/LaunchAgents/com.sirsi.idea-router.*.plist` if you previously installed one.

### Fixed
- Path containment check in the (removed) `submit-existing` verb used `filepath.EvalSymlinks` so tempdir tests worked on macOS; same pattern carries forward to `workRoot()`.

---

## [0.23.0-beta] — 2026-05-19

### Claude Router Inbox Hooks

- Added repo-local Claude Code hooks for router inbox awareness at session start and user prompt submit.
- Added `.claude/hooks/router_inbox_check.py` to read the Idea Router state and stay silent unless the registered Claude agent has pending work.

### Ra/Horus CTR Hypervisor Canon Completion

#### Code Surface
- `sirsi router node-status` — Horus local-node status command showing router home, registered agents, pending work by agent, work-queue item statuses, daemon health, configured binary, and recent dispatch failures
- `internal/router/nodestatus.go` — `CollectNodeStatus()` aggregation with `LaunchctlChecker` injectable for testability
- `internal/router/nodestatus_test.go` — 5 tests covering basic fields, pending-by-agent, sorted agents, daemon-not-installed, and work-queue summary with failures
- `internal/router/executor_test.go` — added non-Claude/non-Codex webhook registration and API wake dispatch coverage for universal agent wake proof

#### Documentation
- Case study indexed: `docs/case-studies/ra-horus-ctr-hypervisor.md`
- Rule D6 in DEITY_REGISTRY.md updated with Horus per-desktop node split
- PANTHEON_HIERARCHY.md §VII CTR Hypervisor boundary table verified
- ADR-017 propagated to ARCHITECTURE_DESIGN.md §2.8

## [0.22.0-beta] — 2026-05-18

### Hardening Sprint Complete (Codex-reviewed, 30+ commits)

#### Safety
- All deletion through `cleaner.DeleteFileReversible()` — no silent permanent delete
- `SafetyGateway` interface centralizes all destructive actions
- Protected path validation on every cleanup operation

#### UX — Pro Command Loop
- `CommandResult` shared model: every command ends with summary, evidence, next actions
- CLI progress spinners on all long operations
- `sirsi permissions` + auto-detect missing Full Disk Access on first scan
- `sirsi setup` checks dependencies and macOS permissions
- Outcome-first vocabulary: zero deity names in user-facing output
- Session state persistence across TUI sessions

#### TUI Refactor
- tui.go: 2,383 → 322 lines (14 focused files)
- All 6 process globals eliminated — message-passing via nativeResult
- 7 controller transition tests

#### Router v3 — Multi-Agent Work Queue
- Agent registry with 8 portfolio agents
- Pluggable executor with writeback verification
- Work item status tracking with dispatch ledger
- Autorouter daemon with fsnotify + polling

#### Ma'at
- Tiered coverage thresholds: Tier A (80%), B (50%), C (30%)
- Real coverage measurement (inverted flag fixed)

### Added — Pro UX Loop Sprint 2 Closeout
- TUI session state persistence
- Updated README and UX workflow docs

## [0.19.0] — 2026-05-06

### Added — TUI & CLI UX Overhaul (7 commits, ~1,500 lines)

#### TUI Post-Run UX (all 10 deities)
- Contextual "What's Next" panel after every deity command completes — gold-highlighted commands with descriptions
- Contextual input placeholder per deity/subcommand (replaces generic "What next?")
- Error state remediation — pattern-matching guidance (permission denied, timeout, connection refused) with deity-specific fallbacks
- 25+ missing TUI routing rules added (Ra test/lint/nightly, Thoth brain/status, Ma'at scales/heal, Seshat list/export/adapters/mcp, Seba book/compute, Horus all 5 subcommands, Anubis apps)
- Horus added to intentKeywords (was missing entirely)
- Help panel expanded with all routable commands + findings drill-down docs

#### Findings Drill-Down
- `findings <category>` filters by category with full detail (path, remediation, fixability, breaking warnings)
- Bare category names as shortcuts after scans (type `dev` to drill into dev findings)
- 20 findings shown in overview (up from 15) with richer per-finding rendering

#### Live Elapsed Timer
- Running commands show "𓃣 Anubis running... (12s)" with per-second updates
- Pipe-based command runner for future streaming capability

#### True Line-by-Line Streaming
- Commands stream output to the TUI viewport line-by-line via channel-based architecture
- Replaces batch buffering — users see progress as it happens

#### Dynamic Deity State Indicators
- Left panel dots reflect run outcomes: ✓ green (succeeded), ✗ red (failed), ◆ amber (has data), ● gold (active), · grey (never run)
- Running deity shows spinner in its roster cell

#### View Stack with Back Navigation
- Esc pops to previous view (findings → findings dev → esc → findings → esc → roster)
- Hints show stack depth: "esc back (2)"

#### Tab-to-Cycle Suggestions
- After a command completes, tab cycles through suggested next commands in the input bar
- Clears on any typed input

#### Persistent TUI State
- Deity state (✓/✗/◆) saved to `~/.config/pantheon/tui-state.json` between sessions
- Roster reflects historical state on relaunch

#### CLI Output Parity
- `output.NextSteps()` function added to terminal.go
- All primary deity CLI commands show "What's Next" footer after completion
- Covers: anubis weigh/judge, maat audit/pulse, seba hardware/diagram/fleet, osiris assess, ra status/deploy, net align, seshat ingest

#### Context-Aware Quick Actions
- Suggestions rotate based on what's been run, what failed, and what has actionable data
- Number shortcuts (1/2/3) work throughout the session, not just first command

---

## [0.18.0] — 2026-05-05

### Changed — Version Alignment
- Synced version across all surfaces: VERSION file, main.go, README badge, CHANGELOG, sirsi.ai/pantheon terminal demo

---

## [0.17.1] — 2026-04-24

### Added — Horus Dashboard, Advisory Intelligence, Deity Hierarchy (32 commits, 8,290 lines)

#### Horus Dashboard (`internal/dashboard/`, SPA)
- Terminal-first single-page application at localhost:9119 with 29 API endpoints
- 8 interactive views: Home, Scan, Ghosts, Guard, Notifications, Horus, Vault, Ra
- SSE streaming for live command output via `/api/events`
- Command input bar: scan, clean, ghosts, doctor, guard, network, hardware, quality, dedup, kill, renice
- Auto-switches to findings view after scan completes with bulk "CLEAN ALL SAFE" action
- Renamed from "Pantheon Dashboard" to "Horus — Local Workstation Monitor" (ADR-015)

#### Advisory Intelligence (`internal/jackal/advisory.go`)
- Every finding has: Advisory (what it means), Remediation (what Sirsi does), CanFix (bool), Breaking (bool)
- 77 rules with specific advisories (e.g., "Rebuilds automatically on next use → Move to Trash")
- 628/628 findings fixable — zero unfixable gaps
- Demonstrated: cleaned 628 findings down to 4, reclaimed ~30 GB

#### Scan Pipeline Overhaul
- 81 scan rules (was 58) — added 22 Git, CI/CD, and repo hygiene rules
- Git rules: stale branches, merged branches, large .git, orphaned worktrees, untracked artifacts, rerere cache, reflog bloat
- CI/CD rules: GitHub Actions cache, act runner, build output, Next.js/.turbo caches, dangling Docker images, BuildKit
- Repo hygiene: .env secrets, stale lock files, dead symlinks, oversized repos, coverage reports, Python venvs
- Severity classification: safe (274), caution (352), warning (2) — not everything is auto-cleanable
- Git rules use proper git commands: `git branch -D`, `git gc --aggressive`, `git worktree prune`
- `sirsi scan --json` outputs full structured results
- `sirsi clean [all|safe]` — bulk cleanup from CLI
- `sirsi judge` — interactive review with confirm prompt, wired to engine.Clean
- Findings persisted to `~/.config/pantheon/findings/latest-scan.json`
- Scan inscribes `anubis_scan` to Stele with category breakdown
- Scan includes ghost detection (Ka) — ghost residuals folded into findings

#### Deity Hierarchy (ADR-015)
- Horus 𓂀 = Local Workstation Lord — sees and reports everything on one machine
- Ra 𓇶 = Fleet Lord — receives Horus reports, orchestrates across all endpoints
- WorkstationReport struct: aggregated local state (`/api/horus/report`)
- Neith = Universal Weaver (local + fleet alignment)
- Thoth/Seshat = Local Memory/Knowledge (per-machine, Ra aggregates)

#### Horus Phase 2 — Live File Watching (`internal/horus/watcher.go`)
- fsnotify-based watcher for Go source changes
- 500ms debounced rebuild — batches IDE auto-format + save
- Skips .git, node_modules, vendor, .next, dist, .turbo
- Cache-first reads in dashboard `/api/horus/scan`

#### Guard Enhancements
- Auto-renice: `WatchConfig.AutoRenice` (opt-in, Rule A1)
- `reniceByPID()` — renice +10 + Background QoS on sustained CPU hogs
- `/api/guard/renice` — manual LSP renice from dashboard
- `/api/slay` — process slayer with 6 targets from dashboard
- `/api/doctor` — full diagnostic from dashboard

#### VS Code Extension (`extensions/vscode/`)
- PantheonDiagnostics: maps findings to inline VS Code diagnostics
- PantheonCodeActionProvider: quick-fix code actions for fixable findings
- New commands: `refreshDiagnostics`, `cleanFinding`
- Severity mapping: safe→Hint, caution→Info, warning→Warning

#### Distribution
- Install script rewritten: `curl -fsSL https://sirsi.ai/install.sh | sh` (no Go toolchain required)
- Homebrew formula fixed: binary names pantheon→sirsi
- Demo GIF rendered via VHS (`assets/demo.gif`)

### Tests
- 48 scan rule tests (git.go + ci.go) — all Git/CI/repo hygiene rules covered
- 7 Stele tests — append, hash chain, offset tracking, concurrent, continuity
- 9 Ra module tests — RADir, monitor, PID/exit files, deployment meta
- 13 new dashboard API tests — doctor, slay, guard, vault, clean, run
- 4 Horus watcher tests — start/stop, file change detection, non-Go skip, vendor skip
- Ghost rule registered for Clean dispatch (was silently skipping 162 findings)

### Case Study
- `docs/case-studies/628-of-628-fixable.md` — full remediation report, every finding documented

---

## [0.17.0] — 2026-04-20

### Added — Token Optimization Subsumption (3 new packages, 10 new MCP tools)

Four external token-optimization tools (RTK, Code Review Graph, Context Mode, Claude Context) have been evaluated, selected, and fully subsumed into native Go packages inside sirsi-pantheon. Zero new external dependencies — everything built on Go stdlib + existing `modernc.org/sqlite`.

#### RTK — Output Filter (`internal/rtk/`, v1.0.0)
Subsumes the external [RTK (Rust Token Killer)](https://github.com/rtk-ai/rtk) tool.

- **Why it exists:** AI coding assistants stuff raw terminal output (build logs, test results, git diffs) directly into the context window. 60-90% of this output is ANSI escape codes, repeated lines, and blank runs — invisible waste that consumes tokens without adding value. RTK filters this at the source.
- **What it does:** Strips ANSI escape sequences via regex, deduplicates consecutive identical lines using an FNV-1a hashed circular ring buffer, collapses runs of blank lines, and truncates oversized output with configurable tail preservation (keeps the last N lines for context).
- **MCP tool:** `filter_output` — explicit filtering of raw text with reduction statistics.
- **CLI:** `sirsi rtk filter` (stdin→stdout), `sirsi rtk stats` (reduction report).
- **Stele event:** `rtk_filter` — records original/filtered bytes, ratio, duplicate count.
- **Files:** `rtk.go`, `ansi.go`, `dedup.go`, `rtk_test.go` (12 tests).

#### Vault — Context Sandbox + Code Search (`internal/vault/`, v1.0.0)
Subsumes the external [Context Mode](https://github.com/mksglu/context-mode) and [Claude Context](https://github.com/zilliztech/claude-context) tools.

- **Why it exists:** Large tool outputs (build logs, test suites, API responses) consume the entire context window when returned inline. Meanwhile, AI assistants read full source files when they only need a few relevant functions. Vault solves both problems: it sandboxes output into SQLite FTS5 (queryable later without context cost), and it indexes source code for BM25-ranked retrieval (returns relevant chunks, not full files).
- **Output sandbox:** Stores any text blob in a SQLite FTS5 virtual table with porter-stemmed unicode tokenization. Full-text search returns BM25-ranked snippets. Metadata table tracks token count and creation timestamp. WAL mode for concurrent reads.
- **Code index:** Splits source files into semantically meaningful chunks — Go files at function/type boundaries using `go/ast`, other languages (Python, TypeScript, Rust, etc.) using 50-line sliding windows with 25-line overlap. BM25-ranked search over 20+ file extensions. Skips `node_modules`, `.git`, `vendor`, `dist`, and files >500KB.
- **MCP tools:** `vault_store` (sandbox output), `vault_search` (FTS5 query), `vault_get` (retrieve by ID), `vault_stats` (statistics), `code_index` (build index), `code_search` (BM25 code search).
- **CLI:** `sirsi vault store/search/get/stats/prune/index/code-search`.
- **Stele events:** `vault_store`, `vault_search`, `vault_prune`, `vault_index`, `vault_code_search`.
- **Dependencies:** Uses existing `modernc.org/sqlite` (pure Go, CGO-free). FTS5 compiled in by default.
- **Files:** `vault.go`, `codeindex.go`, `chunker.go`, `vault_test.go` (9 tests).

#### Horus — Structural Code Graph (`internal/horus/`, v1.0.0)
Subsumes the external [Code Review Graph](https://github.com/tirth8205/code-review-graph) tool.

- **Why it exists:** AI coding assistants read entire source files to understand code structure. A 700-line Go file contains maybe 30 lines of declarations and signatures — the rest is function bodies the AI doesn't need for understanding. Horus extracts just the structure, achieving 8-49x token reduction while preserving every type, function signature, and doc comment.
- **What it does:** Parses Go source using `go/ast`, `go/parser`, and `go/token` from the standard library. Extracts packages, imports, types, structs, interfaces, functions, methods, constants, and variables with their signatures, doc comments, line ranges, and export status. Builds a `SymbolGraph` that can be queried for file outlines, symbol context, and filtered listings.
- **Key innovation — FileOutline:** Returns declarations-only view of any Go file. No function bodies. Tested on Pantheon's own `tools.go` (700+ lines → ~30 lines = 23x reduction).
- **Key innovation — ContextFor:** Returns minimal context for understanding any symbol: its declaration, doc comment, parent type (for methods), and sibling methods. Everything an AI needs to understand a function's role without reading the file.
- **MCP tools:** `code_symbols` (extract all symbols with filters), `code_outline` (compact file outline), `code_context` (minimal symbol context).
- **CLI:** `sirsi horus scan/outline/symbols/context/stats`.
- **Stele events:** `horus_scan`, `horus_query`.
- **Cache:** GOB-encoded graph cache at `~/.config/sirsi/horus/` with manual invalidation.
- **Phase 1 (shipped):** GoParser using Go stdlib. Phase 2 (planned): TreeSitterParser for multi-language support behind `//go:build treesitter` tag.
- **Files:** `horus.go`, `go_parser.go`, `query.go`, `cache.go`, `horus_test.go` (10 tests).

#### Integration Summary
- **10 new MCP tools** registered (total: 16). Tools list: `filter_output`, `vault_store`, `vault_search`, `vault_get`, `vault_stats`, `code_index`, `code_search`, `code_symbols`, `code_outline`, `code_context`.
- **8 new Stele event types** for full observability: `rtk_filter`, `vault_store`, `vault_search`, `vault_prune`, `vault_index`, `vault_code_search`, `horus_scan`, `horus_query`.
- **3 new CLI command groups:** `sirsi rtk`, `sirsi vault`, `sirsi horus`.
- **31 new tests** across all 3 packages (all passing).
- **Module count:** 30 → 33 internal packages.
- **Version registry:** RTK v1.0.0, Vault v1.0.0, Horus v1.0.0 added.
- **Deity count:** 9 → 12 operational modules.
- **Zero new external dependencies.** Built entirely on Go stdlib (`go/ast`, `go/parser`, `go/token`, `regexp`, `hash/fnv`, `encoding/gob`, `database/sql`) + existing `modernc.org/sqlite` v1.44.0.

### Added — Mobile Bridge (iOS + Android)
- **8 new gomobile bindings** — `mobile/rtk.go`, `mobile/vault.go`, `mobile/horus.go`, `mobile/brain.go` (14 exported functions total). All use standard JSON envelope.
- **iOS SwiftUI views** — RTKView (output filter), VaultView (FTS5 search + store), HorusView (code graph browser), BrainView (neural classification with file picker, batch analysis, classification legend). All with shimmer loading, error retry, DeityHeader.
- **iOS models + bridge** — 4 new model files, 14 new bridge methods in PantheonBridge.swift. Deep links for `sirsi://rtk`, `sirsi://vault`, `sirsi://horus`, `sirsi://brain`.
- **Android Compose screens** — RTKScreen, VaultScreen, HorusScreen, BrainScreen. All with Material 3 cards, coroutine-based bridge calls, proper error handling.
- **Android nav drawer** — Replaced bottom nav (5 items max) with ModalNavigationDrawer containing "Core" (Home, Anubis, Ka, Thoth, Seba) and "Advanced" (RTK, Vault, Horus, Brain) sections.
- **Mobile version** — `0.16.0-ios` → `0.17.0`. iOS `project.yml` marketing version → `0.17.0`.
- **xcframework + AAR rebuilt** — Both artifacts include all new bindings.

### Fixed — CI Pipeline (fully green)
- **Go version** — All 4 workflows (CI, iOS, Android, Release) upgraded from Go 1.24 → 1.25 to match `go.mod`.
- **golangci-lint** — Using `install-mode: goinstall` to compile from source with Go 1.25 (pre-built binary was compiled with 1.24).
- **CoreML Darwin constraint** — Renamed `coreml_bridge.{m,h}` → `coreml_bridge_darwin.{m,h}` (same fix pattern as metal_bridge). Unblocks Linux/Android cross-compilation.
- **Android NDK** — Added `-androidapi 21` flag (NDK min API requirement).
- **iOS PantheonWidgets** — Added `CoreML.framework` linker flag (resolves undefined MLModel symbols).
- **Lint cleanup** — Resolved all errcheck, gosimple, govet shadow, ineffassign, misspell, and staticcheck violations across vault, stele, seba, help, neith, mcp, ka, benchmark, workstream.
- **Seba SSH test** — Replaced `os.Setenv` + `t.Parallel()` race with `t.Setenv` (fixes CI-only failure).

### Test Coverage
- **RTK** — 98.7% (12 → 20+ tests with table-driven subtests)
- **Horus** — 97.0% (10 → 46 tests)
- **Vault** — 86.7% (9 → 44 tests, structural limit from untestable error paths)

### Verified
- **CI** — All 5 jobs green: Lint ✅, Test ✅, Build (ubuntu/macOS/Windows) ✅.
- **Go build** — `go build ./...` passes.
- **Tests** — All tests pass. Total: 2,000+.
- **MCP** — `tools/list` returns all 16 tools.
- **Horus self-test** — Parsed sirsi-pantheon itself: 169 files, 328 types, 15 interfaces, 36 packages.

---

## [0.16.1] — 2026-04-18

### Added
- **Android app scaffold** — Full Kotlin/Jetpack Compose app (`android/`, 27 files). Material 3 theme with Pantheon gold/black/lapis branding. Five deity screens (Anubis, Ka, Thoth, Seba, Seshat) with NavHost navigation. `PantheonBridge.kt` mirrors iOS `PantheonBridge.swift` JSON bridge pattern. Proguard rules, externalized strings, CI workflow.
- **Android platform implementation** — `internal/platform/android.go` with `//go:build android` tag. Sandbox-aware filesystem, restricted commands, Android-specific protected paths.
- **Android CI workflow** — `.github/workflows/android.yml` (two-job: build AAR via gomobile + build APK via Gradle).
- **Android architecture doc** — `docs/ARCHITECTURE_ANDROID.md` following Neith's Triad (data flow, implementation order, decision points).
- **Makefile `android-aar` target** — `gomobile bind -target=android` builds `bin/android/pantheon.aar`.

### Changed
- **iOS version bump** — `mobile/mobile.go` version `0.15.0-ios` → `0.16.0-ios`. Marketing version in `project.yml` updated to `0.16.0`. `PantheonCore.xcframework` rebuilt.
- **go.mod** — Added `golang.org/x/mobile/bind` dependency (required for framework build). Go directive bumped to 1.25.0.

### Verified
- **Homebrew tap** — All 6 formulas at `v0.16.0-alpha` in `SirsiMaster/homebrew-tools`. `brew install sirsi-pantheon` serves current release. No action needed.
- **Go build** — `go build ./...` passes. All 1,895+ tests pass (`go test -short ./...`).

---

## [0.15.0] — 2026-04-06

### Added
- **Osiris CLI**: `sirsi osiris assess` (full checkpoint report with 5-level risk scoring) and `sirsi osiris status` (one-line summary for scripts/menubar). TUI intent routing, suggestions, and help guide all wired.
- **Seba hardware commands**: `seba hardware` (GPU/CPU/ANE/RAM dashboard), `seba profile` (saves JSON to ~/.config/pantheon/), `seba compute` (ANE tokenization with real latency measurement).
- **Net CLI registered**: `sirsi net status` and `sirsi net align` now functional. Previously the command existed but was never added to the root command.

### Changed
- **Hapi folded into Seba** (10 → 9 deities): All hardware profiling now under Seba v2.0.0. Removes a facade layer — Hapi was already just wrappers around Seba's detection code.
- **Ma'at pre-push hook**: Now skips deleted package directories (was failing on `internal/horus/` after removal).
- **Version synced to v0.15.0** across main.go, VERSION file, and CHANGELOG.

### Removed
- `cmd/pantheon/hapi.go` — CLI commands moved to `seba.go`.
- `internal/horus/` — 4 files deleted. MCP diagnostic replaced with file stat fallback.
- `docs/pantheon/hapi.html` — Stale deity page.
- Hapi from: version registry, TUI roster, intent keywords, suggestions, help guides, index.html, README, DEITY_REGISTRY.

### Fixed
- **Isis card (index.html)**: Developer metadata incorrectly showed `internal/maat/` package — now shows `internal/guard/`.
- **Net card (index.html)**: Commands updated from `neith audit, adr` to `net status, net align`.
- **neith.go → net.go**: Renamed CLI file and all internal references to match the Net deity name.

---

## [0.14.0] — 2026-04-05

### Added
- **Deity Consolidation (15 → 10)** — Sekhmet→Isis (health+remediation), Ka→Anubis (ghost hunting is hygiene), Khepri→Seba (fleet+infra mapping), Hathor→Anubis (dedup is hygiene), Horus removed (empty stub). Neith renamed to Net. Every deity now has a clear, distinct function with zero overlap.
- **Isis DNS Safety: Three-Layer Protection** — Pre-check gate (TCP probe before changing DNS), post-fix watchdog (polls resolution 3x over 6s, auto-reverts on failure), manual rollback (`sirsi isis network --rollback`). Fixes critical bug where `--fix` bricked internet on restricted networks. See case study: `docs/case-studies/isis-dns-safety-rollback.md`.
- **TUI `help` command** — Full in-TUI reference panel listing all commands, deities, and navigation keys.
- **TUI intent→subcommand inference** — Natural language like "check my dns" now dispatches to `isis network`, not bare `isis`. Maps keyword clusters to the most likely CLI args.
- **Narrow terminal fallback** — TUI gracefully degrades to stacked layout when terminal is <70 columns.

### Fixed
- **DNS auto-rollback failure (Rule A1 violation)** — `dnsReachable()` replaced nslookup (depends on DNS) with raw TCP connect to port 53 (transport-level, no DNS dependency). Fix path restructured: probe BEFORE changing config, not after.
- **Network keyword routing** — "network" now correctly routes to Isis (security) vs Seba (topology) based on multi-keyword scoring instead of always hitting the wrong deity.
- **`TestExtractAgeDays` timezone bug** — Date comparison used UTC midnight vs local time, causing off-by-one on timezone boundaries. Fixed to compare date strings.
- **`TestSmoke_Version` hardcoded version** — Updated to check for brand name instead of specific version string.

### Changed
- **Isis v2.0.0** — Absorbs all Sekhmet functionality: `doctor`, `guard`, `network`, `heal`. CLI: `sirsi isis network`, `sirsi doctor`.
- **Anubis v1.1.0** — Absorbs Ka (ghost hunting) and Hathor (file dedup). `sirsi anubis ka` and `sirsi dedup` both work.
- **Seba v1.2.0** — Absorbs Khepri (fleet discovery, container audit). `sirsi seba fleet` works.
- **Net v1.1.0** — Formerly Neith. Scope weaver for Ra task definition.

### Removed
- Sekhmet, Ka, Khepri, Hathor, Horus from deity roster and version display.
- All backwards-compatible aliases — clean codebase, no legacy bloat.

---

## [0.13.0] — 2026-04-05

### Added
- **TUI Inline Predictions** — Fish-shell-style ghost text suggestions. Static command tree covers all deities, subcommands, and flags.
- **Suggestion Engine** (`internal/output/suggestions.go`) — Context-aware completions: history → command tree → deity names → intent keywords.
- **Network Security Audit** (`sirsi isis network`) — Six-check posture audit: DNS, WiFi, TLS 1.3, CA certs, VPN, firewall. ~130ms.
- **`--fix` flag** for `isis network` — Auto-applies safe remediations (encrypted DNS, firewall enable).

### Fixed
- **Deity roster grid overflow** — Manual measure-then-pad approach for Egyptian hieroglyphs with unpredictable terminal widths.

### Changed
- **TUI hints** — `→ accept · ↑ history · help · ctrl+c quit`.
- **TUI key bindings** — Right-arrow accepts ghost text, Up/Down navigate command history.

---

## [0.12.0] — 2026-04-05

### Added
- **Pantheon TUI** — `sirsi` (no args) launches a persistent interactive session. Deity roster in a 3×5 column grid with active highlighting. Universal input bar accepts both natural-language requests ("find ghost processes") and direct CLI commands ("ka hunt ~/Dev"). Commands execute inside the TUI with output in a split-pane viewport. Input bar re-enables on completion. User stays in Pantheon until they quit.
- **Intent-based routing** — Natural-language input is scored against deity keyword maps and routed to the best-matching deity command.
- **Split-pane layout** — On first command, the view splits: left pane (deity roster + status), right pane (scrollable command output). Esc returns to full roster.
- **Active deity detection** — Reads Stele events and PID files to highlight deities with recent activity (gold dot indicator).

### Changed
- **`sirsi` entry point** — Bare `sirsi` now launches the TUI instead of printing help. All subcommands (`sirsi ka hunt`, `sirsi maat audit`, etc.) continue to work unchanged for scripting and CI.

---

## [0.11.0] — 2026-04-05

### Added
- **Neith Tiled Context Rendering (ADR-013)** — GPU-inspired context pipeline: chunks canon into semantic units, scores by keyword match/recency/structural weight/coverage, fills token budget with highest-scored tiles, defers the rest to a manifest. Reduces session 1 context from ~254K to ~72K tokens (72% reduction).
- **`ChunkCanon()`** — Splits CanonContext into addressable semantic chunks (journal entries, changelog versions, individual ADRs, planning docs).
- **`ScoreChunks()`** — Multi-signal visibility test: structural weight (always-visible HUD), keyword match, temporal proximity (90-day decay), coverage detection (anti-overdraw).
- **`TilePrompt()`** — Greedy budget-filling algorithm. Always-visible chunks bypass budget. Deferred chunks go to manifest.
- **`FormatManifest()`** — Generates deferred context table so agents know what exists and where to find it. Groups journal entries, caps at 20 rows.
- **`AutoTokenBudget()`** — Auto-detects budget from canon size: <50K = no tiling, 50K-200K = 80K budget, >200K = 60K budget.
- **`token_budget` field on ScopeConfig** — Per-scope override for token budget. Defaults to auto-detect.
- **Thoth auto-pruning** — Compact now defaults to MaxKeep=20 journal entries when no explicit pruning config set. Prevents unbounded journal growth.

### Changed
- **`WeaveScope()`** — Now runs the full tiling pipeline: chunk → score → tile → render + manifest. Section ordering preserved. Small canons (<50K tokens) skip tiling entirely.
- **Stele inscription** — Neith weave events now include `approx_tokens`, `tiled`, `rendered`, `total_chunks` metadata.
- **DEITY_REGISTRY** — Neith's domain updated to include "tiled context rendering."

### Documentation
- **ADR-013** — Tiled Context Rendering architecture decision record.
- **Case Study 020** — Full token economics analysis with three-tier comparison (vanilla/Pantheon/Pantheon+Tiling). Available as Markdown and standalone HTML.

---

## [0.10.0] — 2026-04-04

### Added
- **Stele Universal Event Bus** — All Pantheon deities now inscribe events to the Stele (`~/.config/ra/stele.jsonl`). Append-only, hash-chained, SHA-256 integrity. Promotes ADR-014 from Ra-only to ecosystem-wide.
- **`stele.Inscribe()` Convenience API** — Global singleton ledger with lazy initialization. Any deity can write events with one call, no lifecycle management.
- **30+ New Stele Event Types** — `thoth_sync`, `thoth_compact`, `maat_weigh`, `maat_pulse`, `seshat_ingest`, `neith_weave`, `neith_drift`, `ka_hunt`, `ka_clean`, `guard_start`, `seba_render`, `hapi_detect`, and more.
- **Ra ProtectGlyph `𓂀`** — Eye of Horus sentinel stamped into Terminal.app window titles. Windows bearing `𓂀` are immune to `KillAll`. Replaces fragile front-window heuristics that killed the Claude Code session.
- **`ProtectFrontWindow()`** — Stamps the user's Claude Code terminal before deploy.
- **Command Center Global Activity Feed** — Dashboard now displays deity-level events (Thoth sync, Ma'at weigh, etc.) in a live activity feed below scope cards.

### Changed
- **Command Center version** — Updated to v0.10.0.
- **Module versions bumped** — Thoth 1.1.0, Ma'at 1.1.0, Seshat 2.1.0, Hapi 1.1.0, Seba 1.1.0, Sekhmet 1.1.0, Neith 1.1.0, Ra 1.1.0. Stele 1.0.0 registered.
- **`buildTerminalScript`** — Spawned windows now `; exit` on completion (auto-close) and use `set custom title` inside `tell front window` block for reliable title assignment.
- **`KillAll`** — Single `𓂀` glyph check replaces TTY/name exclusion chains.

### Fixed
- **Session crash on `KillAll`** — Broad `osascript` window matching killed the Claude Code terminal. Now protected by ProtectGlyph.

---

## [0.9.0-rc1] — 2026-04-03

### Added
- **Ka v1.1.0 — Multi-Layer Ghost Matching** — Four-layer matching cascade (exact bundle ID, prefix/family, normalized name substring, nested directory scanning) eliminates 91 false positives. WhatsApp, Adobe Acrobat, and CleanMyMac no longer flagged as ghosts. Ghost residual size reduced from 6.2 GB to 165.2 MB. Case study: `docs/case-studies/ka-ghost-matching-v1.1.md`.
- **Module Versioning System** — `internal/version/modules.go` tracks per-deity module versions. `sirsi version` now displays all 15 module versions in a two-column layout.
- **Seshat v2.0 — Universal Knowledge Grafting** — 5 source adapters (Gemini, Claude, Chrome, Apple Notes, Google Workspace) + 3 target adapters (Thoth, GEMINI.md, NotebookLM). Secrets filter with regex-based detection and redaction.
- **Seshat Chrome Profile Support** — `--profile` flag for per-profile ingestion, `--all-profiles` for multi-profile sweep, `sirsi seshat profiles chrome` to list all profiles, `sirsi seshat open chrome --profile <name>` to launch Chrome with a specific profile.
- **Seshat NotebookLM Export** — `sirsi seshat notebooklm` exports KIs as Markdown and opens NotebookLM in the browser for drag-and-drop upload.
- **Neith Module** — Plan alignment engine with keyword-based log assessment, full tapestry validation (all 5 deity checks), drift detection, and CLI (`sirsi neith status`, `sirsi neith align`).
- **Ka Cross-Platform Ghost Detection** — `GhostProvider` interface with platform-specific implementations. macOS (full), Linux (XDG + dpkg + .desktop files), Windows (stub).
- **5 New MCP Tools** — `thoth_sync`, `maat_audit`, `anubis_weigh`, `judge_cleanup` (dry-run only), `pantheon_status`. Total: 11 tools, 4 resources.
- **Thoth /compact Integration** — `sirsi thoth compact -s "summary"` persists session decisions before context compression.
- **Sirsi Orchestrator** — Python orchestrator using claude-code-sdk to dispatch parallel Claude sessions across all Sirsi repositories. Commands: health, test, lint, task, broadcast, nightly.
- **Rich CLI Help System** — `sirsi help <deity>` with lipgloss-styled terminal guides for 12 deities. `--docs` flag opens web docs in browser. `--list` shows all available guides.
- **Per-Deity Binary Builds** — goreleaser now produces standalone binaries: `pantheon-anubis`, `pantheon-maat`, `pantheon-thoth`, `pantheon-scarab`, `pantheon-guard`. Each installable via `brew install SirsiMaster/tools/pantheon-<deity>`.
- **Getting Started Guide** — Full 7-step HTML walkthrough at sirsi.ai/pantheon/getting-started.
- **Deity Pages** — New HTML pages for Seshat, Isis, and Neith. All 15 deity pages now have how-to guides, FAQ sections, and platform support badges.
- **Sirsi Branding** — SVG logo assets (dark, light, icon), "by Sirsi Technologies" throughout all pages and README.

### Changed
- **Hapi → Seba Consolidation** — Hardware detection moved to `internal/seba/`. Hapi retains backward-compatible wrappers.
- **FAQ Expanded** — General, deity-specific, and troubleshooting sections with 15+ Q&As.
- **Platform Support Matrix** — Every deity page and the registry index now show macOS/Linux/Windows compatibility.

### Fixed
- All packages pass tests on macOS and Ubuntu CI
- Zero golangci-lint errors
- Smoke test updated for v0.9.0-rc1 version string

### Not Included (deferred)
- **Ra** — Web portal / hypervisor orchestration (not started)
- **Windows Ka** — Stub only; real implementation deferred
- **Flatpak/Snap/RPM** — Linux package managers beyond dpkg deferred

---

## [0.8.0-beta] — 2026-03-31 (The Honest Measurement)

### What This Release Is
v0.8.0-beta is the first credible public release of Pantheon. All metrics are verified by `go test -cover ./...` — no hardcoded numbers, no projections presented as measurements. The previous v1.0.0-rc1 claim was premature and has been corrected.

### Added
- **Thoth Knowledge System** — Go port of sirsi-thoth folded into Pantheon. `sirsi thoth init --yes <dir>` scaffolds .thoth/ project memory. Detects Go, TypeScript, Next.js, Rust, Python projects.
- **Ma'at Streaming Progress** — `maat audit` now shows per-package test results as they stream in, with color-coded verdicts. No more 2-minute silent waits.
- **`--skip-test` Flag** — `maat audit --skip-test` uses cached coverage for instant governance results without running the full test suite.
- **Ma'at Dynamic Module Discovery** — `DefaultThresholds()` now scans `internal/*/` dynamically instead of using a hardcoded registry. All 27 modules are now measured (was missing 10).
- **E2E Smoke Tests** — 9-test bash suite (`scripts/smoke.sh`) + 10-test Go E2E suite (`tests/e2e/smoke_test.go`) testing the compiled binary against the real filesystem.
- **Jackal Rules Coverage** — 93.1% coverage on scan rules (was 64.5%). 50+ new tests covering all rule constructors, Scan/Clean operations, Horus manifest branches, findRule depth/matchFile logic.

### Fixed
- **False Coverage Reports** — Ma'at was reporting 0% for 10 modules (thoth=83%, seshat=85%, neith=100%, etc.) due to hardcoded module registry. Fixed with dynamic discovery.
- **CI Pipeline** — Go 1.22 -> 1.24, golangci-lint v4 -> v6, 40+ lint errors resolved across 19 files.
- **Version Honesty** — Corrected v1.0.0-rc1 -> v0.8.0-beta. The v1.0.0-rc1 label was premature — it will be earned after 30-day dogfooding.

### Changed
- Version: `0.7.0-alpha` -> `0.8.0-beta`
- Go: 1.22 -> 1.24 across all CI workflows
- golangci-lint: v4 -> v6

### Verified Metrics (March 31, 2026)
| Metric | Value | Command |
|--------|-------|---------|
| Tests Passing | 1,500+ | `go test -short ./...` |
| Packages | 28/28 green | `go test ./...` |
| Weighted Coverage | ~85% | `go test -cover ./...` |
| Lint Errors | 0 | `golangci-lint run ./...` |
| Binary Size | ~12 MB | `go build ./cmd/sirsi/` |
| Scan Rules | 64 | `internal/jackal/rules/` |
| Internal Modules | 27 | `ls internal/` |
| E2E Smoke Tests | 9+10 | `scripts/smoke.sh` + `go test ./tests/e2e/` |

### What's NOT in This Release
- Ra (web portal) — not started
- Neith (orchestration) — stub only
- Windows/Linux ghost detection — macOS-first
- Cross-platform GUI — CLI only for now

### What's Next (v1.0.0-rc1 — earned, not declared)
- 30-day dogfooding on production machines
- Cross-platform testing (Linux, Windows)
- Neith orchestration implementation
- MCP plugin for Claude Code (desktop/IDE/CLI)

---

### Session 37 (2026-03-29) — The Great Pantheon Consolidation
- **Deity-First Architecture** — Successfully consolidated 12 fragmented command scripts into 6 Master Deity Pillars, achieving the "One Install. All Deities." standard.
  - **Anubis 𓃣**: Unified Hygiene, Ka Ghost Hunter, Mirror Dedup, and Guard Watchdog.
  - **Ma'at 𓁐**: Unified Scales Governance and Isis Autonomous Remediation.
  - **Thoth 𓁟**: Unified Knowledge Sync and Permanent Brain Ledger.
  - **Hapi 𓈗**: Unified Hardware Detection and Sekhmet ANE Acceleration.
  - **Seba 𓇼**: Unified Infrastructure Mapping, Project Book, and Scarab Fleet Discovery.
  - **Seshat 𓁆**: Unified Gemini Bridge, Brain Library, and MCP Context Server.
- **Universal Glyph Standard** — Purged all generic emojis (🏛️, 🌊, ⬥) and geometric symbols (⬥, ◇, ◆) across the entire platform. 
  - **CLI/TUI**: All headers, status indicators, and dashboards now use High-Fidelity Ancient Egyptian Hieroglyphs.
  - **Registry**: Remastered `docs/index.html` with click-to-flip cards reflecting the unified 6-pillar hierarchy.
- **Safety Restoration** — Restored the **⚠️ Universal Warning** signal throughout the platform to ensure absolute safety and recognition for destructive operations.
- **Monumental Baseline (𓉴)** — Promoted the **Great Pyramid (𓉴)** as the primary architectural anchor for the Pantheon platform and Web Registry, replacing legacy generic identifiers.
- **Hieroglyphic Menu** — Published the `glyph_menu.md` (Artifact) featuring over 25 categorized hieroglyphs for Master Pillar selection variety.
- **Hardening & Verification** — Resolved all compilation regressions, import collisions (fmt, os, InfoStyle), and unit test mismatches.
- **Stats**: 36 files modified, consolidated 13 legacy scripts, 100% build-readiness.

### Planned
- P1: npm publish thoth-init
- P2: Isis Phase 2 (test scaffold generation, errcheck auto-fix)
- P3: Thoth test coverage (internal/thoth/ at 0%)
- P4: Homebrew Formula update for marketing launch.

### Session 35 (2026-03-28) — Isis Phase 1 (The Healer) + Thoth CLI
- **Thoth CLI** (`cmd/pantheon/thoth.go`) — `sirsi thoth sync` wired to CLI.
  - Two-phase auto-sync: Phase 1 updates memory.yaml identity fields from source analysis. Phase 2 appends journal.md entries from git history.
  - `findRepoRoot()` walks up from cwd to locate `.thoth/` directory.
  - Flags: `--since`, `--dry-run`, `--memory-only`, `--journal-only`.
  - Self-dogfooded: the sync command updated its own memory.yaml in this session.
- **Isis Remediation Engine** (`internal/isis/`, 6 files, 24 tests) — Phase 1 of the Ma'at→Isis healing cycle.
  - `isis.go`: `Healer` struct, `Strategy` interface, `Heal()` orchestrator with dispatch, `Report` formatter.
  - `lint.go`: `LintStrategy` — runs `goimports` + `gofmt` with injectable `RunCmd` (Rule A21).
  - `vet.go`: `VetStrategy` — runs `go vet`, parses findings. Reports (no auto-fix — requires human judgment).
  - `coverage.go`: `CoverageStrategy` — uses `go/parser` AST analysis to find exported functions without tests.
  - `canon.go`: `CanonStrategy` — detects memory.yaml/journal drift and triggers `thoth.Sync()`.
  - `bridge.go`: `FromMaatReport()` converts Ma'at `Assessment` verdicts to Isis `Finding` structs.
- **Isis CLI** (`cmd/pantheon/isis.go`) — `sirsi isis heal`.
  - Dry-run by default (Rule A1 — safety first). `--fix` to apply changes.
  - Cache-based Ma'at weighing (~3ms) by default. `--full-weigh` for live `go test` (~5min).
  - Strategy filters: `--lint-only`, `--vet-only`, `--coverage-only`, `--canon-only`.
- **Distribution** — `tools/thoth-init/README.md` for npm publish. Local `npx thoth-init -y` verified.
- **Stats**: 14 files changed, +1,765 lines, 843+ tests, 27 modules, 42 commands.
- **Seshat VS Code Extension** (`extensions/gemini-bridge/`) — Full TypeScript extension for Gemini Bridge.
  - 7 source files: `extension.ts`, `commands.ts`, `dashboard.ts`, `knowledgeProvider.ts`, `chromeProfilesProvider.ts`, `syncStatusProvider.ts`, `paths.ts`.
  - **Activity Bar**: Dedicated sidebar with 3 tree views — Knowledge Items, Chrome Profiles, Sync Status.
  - **Dashboard Webview**: Gold-on-black Egyptian aesthetic with KI inventory, conversation count, bridge direction visualizer, and sync actions.
  - **Chrome Profile Discovery**: Reads Chrome's `Local State` to enumerate all profiles; highlights configurable default (`SirsiMaster`).
  - **6 Commands**: `seshat.listKnowledge`, `seshat.exportKI`, `seshat.syncToGemini`, `seshat.listProfiles`, `seshat.listConversations`, `seshat.showDashboard`.
  - **4 Config settings**: `seshat.defaultProfile`, `seshat.autoSync`, `seshat.pantheonBinaryPath`, `seshat.antigravityDir`.
  - VSIX packaged: `seshat-gemini-bridge-0.1.0.vsix` (430 KB, 12 files).
  - Publisher: `SirsiMaster`. License: MPL-2.0.
- **Neith's Triad Retrofit** — `ARCHITECTURE_DESIGN.md` upgraded from v1.0.0 to v2.0.0:
  - §7: **Data Flow Architecture** — Full Mermaid diagram mapping all CLI entry points, Code Gods, Machine Gods, Safety Layer, Output Layer, and Seshat's 6 external system directions.
  - §8: **Recommended Implementation Order** — Gantt chart of 7 build phases from Jackal through Distribution.
  - §9: **Key Decision Points** — 10-row decision matrix covering binary architecture, concurrency, policy language, safety model, UI framework, fleet transport, context format, deity coupling, distribution, and bridge direction.
  - Document now fully compliant with Rule A22.
- **Firebase Deploy** — 17 files deployed to `sirsi.ai/pantheon` with all 11 deity click-to-flip cards live.

### Session 29 (2026-03-27) — CI Green Sprint + Thoth Journal Sync + Rule A21
- **CI Remediation (P0)** — Resolved 22 lint errors across 16 files:
  - `errcheck`: 4 suppressed `fmt.Sscanf` returns in `stats.go`
  - `unused`: 3 wired/removed dead functions in menubar
  - `goimports`: 1 formatting fix in `sekhmet.go`
  - `shadow`: 6 renamed inner `err` vars in 5 test files + `publish.go`
  - `unusedwrite`: 8 added struct field assertions in 4 test files
- **Windows CI Fix** — Added `shell: bash` to test step (PowerShell splits `-coverprofile=coverage.out`).
- **Data Race Fix** — `AlertRing` mutex + `sampleTopCPUFn` accessor pattern (`getSampleFn()`/`setSampleFn()`).
  - Root cause: `defer func() { sampleTopCPUFn = old }()` raced with watchdog goroutines on locked OS thread.
  - Fix: `sync.RWMutex`-protected accessors. All 8 bridge tests pass with `-race -count=1`.
- **Rule A21 Canonized** — Concurrency-Safe Injectable Mocks. Ma'at governs: all package-level function pointers used for test injection MUST use mutex-protected accessors.
- **Thoth Journal Sync (P1)** — `internal/thoth/journal.go` (230 lines): auto-generates journal entries from git history.
  - `thoth sync` now runs Phase 1 (memory.yaml) + Phase 2 (journal.md from `git log`).
  - Supports `--since` and `--dry-run` flags. Closes the ghost transcript gap permanently.
- **Firebase Deploy (P2)** — 17 files deployed to `sirsi.ai/pantheon`.
- **gh CLI Upgrade (P3)** — `gh` 2.87.3 → 2.89.0.


### Session 28 (2026-03-27) — Ghost Transcripts Recovery + CI Remediation
- **Case Study 014** — "The Ghost Transcripts": discovered Antigravity IDE never writes `overview.txt` — 90+ conversations with zero transcripts.
- **Forensic Recovery** — Reconstructed journal entries 022-024 from git diffs, case studies, build log, and memory.yaml.
- **CI Remediation** — Fixed 3 CI failure categories: Windows `CGO_ENABLED` syntax, `coverprofile` parsing, 20+ lint errors.
- **Lint Hardening** — Fixed unused `version` vars (5 standalone binaries), unused struct fields (`lastSnapshot`, `autoEnabled`), misspelling (`cancelled`→`canceled`).
- **Binary Hygiene** — Removed tracked `thoth` binary from git, added to `.gitignore`.
- **Test Hardening** — Added `-short` flag to CI test runner to skip live syscall tests (30s timeout prevention).

## [0.7.0-alpha] — 2026-03-27 (Ecosystem Hardening — Sekhmet Phase)
### Added
- **Singleton Enforcement** — Implemented Unix domain socket locking (`platform.TryLock`) across all primary entry points (Menubar, Guard, MCP) to prevent process redundancy.
- **Hapi-Brain Bridge** — Created `internal/brain/hapi_bridge.go` for hardware-aware inference backend selection (CoreML vs ONNX).
- **Hardened Watchdog** — Sekhmet watchdog now enforces a 1.5GB memory governance threshold and tracks process prioritization.
- **MCP hardware tool** — Added `detect_hardware` tool to the MCP server for real-time accelerator and resource detection.

### Fixed
- **Triple Ankh Redundancy** — Resolved the issue of multiple pantheon-menubar instances running simultaneously.
- **MCP Standardization** — Refactored MCP server startup to utilize the standard `mcp.NewServer()` implementation with singleton hardening.
- **LaunchAgent Audit** — Synchronized `ai.sirsi.pantheon.plist` with the hardened singleton architecture.

### Session 25 (2026-03-27) — Sekhmet Phase II (ANE Tokenization)
- **HAPI Tokenization** — Extended the `Accelerator` interface with native `Tokenize` support.
- **Hardware Backends** — Implemented specialized tokenization for AppleANE, Metal, CUDA, and ROCm.
- **FastTokenize** — Built a deterministic BPE-style pure Go tokenizer as the universal CPU fallback.
- **Sekhmet CLI** — Integrated `sirsi sekhmet --tokenize` for direct hardware-accelerated testing.
- **Global Flags** — Centralized CLI flags in `cmd/pantheon/globals.go` to support modular command files.
- **Canon Sync** — Updated Thoth, BUILD_LOG, FAQ, and the Deity Registry.

### Session 24 (2026-03-27) — Pantheon v0.7.0-alpha Deployment
- **VSIX Packaging** — Built and sideloaded `sirsi-pantheon-0.7.0.vsix` for verify-before-publish testing.
- **OpenVSX Publish** — Published `SirsiMaster.sirsi-pantheon@0.7.0` to Open VSX using the SirsiMaster account (Rule A20).
- **Crashpad Verification** — Manually verified the Crashpad Monitor's threshold detection by clearing 34 stale dumps.
- **Site Deployment** — Deployed updated Deity Registry and Build Log (Sessions 23/24) to `sirsi.ai/pantheon`.
- **Status Sync** — Updated all public-facing stats: 21K+ lines of Go, 90.1% coverage, 11 deities, 12 ADRs.
- **Version**: 0.7.0-alpha.

### Session 23 (2026-03-26) — Crash Forensics + Crashpad Monitor
- **Crash Forensics** — Investigated IDE crash that required 2 reinstalls + 2 restarts.
  - 34 pending crash dumps in `Crashpad/pending/` — dating back weeks.
  - Root cause: Session 22 manifest patches created un-realizable Extension Host state.
  - Chain: V8 OOM (`electron.v8-oom.is_heap_oom`) → macOS Jetsam (`libMemoryResourceException`) → cascade.
  - V8 GC efficiency dropped to `mu = 0.132` (normal: >0.9) before heap exhaustion.
  - Crash dumps 2 & 3 confirmed as `libMemoryResourceException` — kernel memory pressure kills.
- **Rule A19 Hardened to ABSOLUTE PROHIBITION** — No `.app` bundle modifications ever.
  - Previous exception ("manifest-only patches are safe with re-signing") proven wrong.
  - Semantic integrity matters more than code signing — valid JSON can crash the Extension Host.
  - Case Study 011: `docs/case-studies/session-23-extension-host-crash-forensics.md`.
- **Crashpad Monitor** (`extensions/vscode/src/crashpadMonitor.ts`, 370+ lines) — **NOVEL FEATURE**.
  - Auto-detects Crashpad directory for Antigravity, VS Code, Cursor, Windsurf.
  - Polls `pending/*.dmp` every 5 minutes with rolling trend detection (3-reading window).
  - Extension Host crash identification via first-8KB string extraction from `.dmp` files.
  - Trend classification: `stable` / `growing` / `critical` with threshold-based alerts.
  - Status bar indicator: hidden when stable, 🟡 at 5+ dumps, 🔴 at 15+ dumps.
  - Premium webview report with timeline, forensics reference, and cleanup recommendations.
  - One-time session warning when trend shifts from stable.
  - New command: `pantheon.crashpadReport` (10 total commands, 7 modules).
  - Case Study 012: `docs/case-studies/session-23-crashpad-monitor.md`.
- **Canon Updated** — Journal Entry 020-021, build-log.html, PANTHEON_RULES.md, CLAUDE.md, GEMINI.md.
- **Version**: 0.7.0-alpha. Extension: 10 commands, 7 modules.

### Session 22 (2026-03-26) — Thoth Accountability Engine + Extension Triage
- **Thoth Accountability Engine** (`extensions/vscode/src/thothAccountability.ts`, 645 lines)
  - Cold-start benchmark: walks workspace source, compares against memory.yaml.
  - First measurement: ~1.5M source chars → ~19K memory.yaml = **371K tokens saved** per activation.
  - Dollar savings: configurable pricing tier (Opus $15/M, Sonnet $3/M, Haiku $0.25/M). Default: **$1.11/session**.
  - Freshness meter: compares memory.yaml mtime against most recent source edit. FRESH/STALE/OUTDATED status.
  - Coverage governance: cross-references `internal/` directories against memory.yaml mentions.
  - Context budget: memory.yaml token count as % of 200K context window (<5%).
  - Lifetime counter: persists total tokens, dollars, and sessions across restarts via `globalStorageUri`.
  - Premium webview report in Royal Neo-Deco design language (gold/lapis/obsidian).
  - Status bar: `$(bookmark)` with live savings display next to main PANTHEON ankh.
- **Extension Commands** — `pantheon.thothAccountability` command with 5-option QuickPick menu.
  - Integrated into `pantheon.showMetrics` system dashboard.
  - New configuration: `pantheon.thoth.accountability`, `pantheon.thoth.pricingModel`.
- **Extension Triage** — diagnosed and fixed 4 simultaneous extension issues:
  1. **AG Monitor Pro** (1988ms profile): disabled — `js-tiktoken` WASM init + `chokidar` watcher.
  2. **Pantheon 0.5.0** cascade unresponsive: sideloaded v0.6.0 with Accountability Engine.
  3. **Git extension** missing `title` properties: patched 2 Antigravity-added commands.
  4. **Antigravity extension** missing command declarations: patched 3 undeclared commands.
- **Gatekeeper Recovery** — modifying `.app` bundle broke macOS code signature.
  - Fix: `xattr -cr` + `codesign --force --deep --sign -` (ad-hoc re-signing).
  - Documented as case study with procedure for future `.app` manifest patches.
- **Version**: 0.6.0-alpha. Extension VSIX: 49.47 KB (13 files, 6 modules, 8 commands).

### Session 21 (2026-03-26) — Extension Live Testing + Memory GC
- **Guardian Rewrite** — Native `renice(1)` + `taskpolicy(1)` implementation. No CLI binary dependency for renice.
  - Discovers LSP processes via `ps`, applies nice +10 and Background QoS directly.
  - Skips already-deprioritized processes. Excludes host LSP (language_server_macos_arm) from warnings.
- **Memory Pressure GC** — Tracks per-process RSS across poll cycles.
  - When a third-party LSP exceeds 500 MB for 3+ consecutive checks, triggers VS Code's built-in LSP restart.
  - Maps process names to restart commands (gopls → `go.languageserver.restart`, tsserver → `typescript.restartTsServer`, etc.).
- **Codicon Status Bar** — Replaced invisible hieroglyph with `$(eye) PANTHEON` codicons. Loading spinner on init. Warning icon on pressure.
- **Warning Threshold** — Split total/third-party RAM tracking. Warning triggers on >1 GB third-party LSPs (host LSP at 4-6 GB is normal).
- **CLI Fix** — Commands now use correct Pantheon CLI flags (`weigh --dev --json`, `guard --json`).
- **Live Testing** — Verified end-to-end: all 3 LSPs reniced to nice 10 after 30s delay. Extension Host ~199 MB RSS.
- **Sideloaded** — Installed in both Antigravity and VS Code via VSIX.

### Session 20 (2026-03-25) — The Deployment Sprint
- **Firebase Hosting** — Deployed Deity Registry to `sirsi.ai/pantheon` via Firebase Hosting (15 HTML pages).
  - Created Firebase site `sirsi-pantheon` in project `sirsi-nexus-live`.
  - Configured hosting target with clean URLs and 1-hour cache.
- **Custom Domain** — Wired `sirsi.ai/pantheon` via Firebase Hosting API + GoDaddy CNAME.
  - Firebase: `HOST_ACTIVE`, `OWNERSHIP_ACTIVE`. SSL auto-provisioning.
- **Flip Cards** — Rebuilt Deity Registry index with click-to-flip 3D cards.
  - Front: user-facing (name, description, key metrics).
  - Back: developer info (package path, coverage, test count, commands, deps, ADR).
  - 3 action buttons per card: Full Page, Download (releases), Source (GitHub internal/ link).
- **Deity Page Fixes** — Updated all 12 deity pages:
  - URL display: subdomain → path format (`sirsi.ai/pantheon/anubis`).
  - Nav links: relative → absolute for Firebase deployment.
- **Canon Cleanup** — VERSION bump to `0.5.0-alpha`, extension icon created, CHANGELOG + Thoth updated.

### Session 19 (2026-03-25) — The Pantheon Extension
- **VS Code Extension** (`extensions/vscode/`) — Full TypeScript extension replacing JS scaffold (ADR-012).
  - `extension.ts`: Entry point — starts Guardian, status bar, Thoth on activation.
  - `guardian.ts`: Always-on renice (30s delay, 60s re-apply loop). Spawns `sirsi guard --renice lsp --json`.
  - `statusBar.ts`: Ankh (𓃣) icon with live RAM/CPU metrics. Polls `ps` directly (sub-50ms). Color-coded states.
  - `commands.ts`: 7 Command Palette entries (Scan, Guard, Renice, Ka, Thoth, Metrics, Settings).
  - `thothProvider.ts`: Context compression from `.thoth/memory.yaml` with file watching.
- **ADR-012**: Pantheon VS Code Extension architecture decision accepted.
- **ADR Index**: Updated to 12 ADRs (001–012).
- **Status**: Extension compiles (0 TypeScript errors), Go backend builds, 819+ tests passing.

### Session 16b (2026-03-24) — The Interface Injection Sprint
- **Coverage Breakthrough** — Weighted average pushed to **90.1%** (Rule A16 established).
- **Injectable Providers** — Established standard interface injection for signals and `exec.Command` (ADR-009).
- **Guard Module (89→91%)** — Full deterministic audit of process termination paths (root-failures, dry-runs).
- **Ma'at Module (80→88%)** — Deterministic CI pipeline assessment with injectable gh-cli runners.
- **Sight Module (78→93%)** — Rebuilt `Fix` and `ReindexSpotlight` with mockable system commands.
- **Antigravity CLI Wiring** — `sirsi guard --watch` now starts the full IPC bridge + AlertRing.
- **MCP Live Alerts** — Bridged watchdog alerts into MCP resources via `anubis://watchdog-alerts`.
- **Canon Realignment** — `ANUBIS_RULES.md` → `PANTHEON_RULES.md` (v2.0.0). ADR index updated.

## [0.4.0-alpha] — 2026-03-23 (Launch Execution + Modular Deities)

### Added
- **Homebrew Tap Integration** — Automated formula updates via `HOMEBREW_TAP_TOKEN`; `brew tap SirsiMaster/tools && brew install sirsi-pantheon`
- **ADR-007 Unified Findings Portal** — Canonical architecture for cross-deity finding aggregation
- **ADR-006 Self-Aware Resource Governance** — Guard module + yield-based resource management
- **Yield Module** (`internal/yield/`) — Cooperative resource yielding for process management
- **Horus Designation** — Assigned as the Unified Findings Portal deity
- **Horus Module** (`internal/horus/`) — Shared filesystem index, parallel walks, manifest cache (ADR-008)
- **Modular Deities (v2.1.0)** — ADR-005 updated with independent deployment standards
- **Ra (Hypervisor)** — v0.1.0-alpha overseer added to Pantheon architecture
- **Seba Rebrand** — `internal/mapper/` → `internal/seba/` (high-performance mapping)
- **Cross-Agent Referral Logic** — Initial implementation of inter-deity remediation referrals
- **Independent Deployment** — Support for standalone deity installation (e.g., `npx thoth-init`)

### Performance (Dogfooding-Driven)
- **Ma'at Diff-Based Coverage** — 55s → 15ms (3,667× speedup); only tests changed packages
- **Horus Shared Filesystem Index** — Walk once, all deities query; Weigh 15.6s → 7.2s (2.2×)
- **Jackal WalkDir Migration** — `filepath.Walk` → `filepath.WalkDir` (avoids stat per file)
- **Combined dirSizeAndCount** — Single walk replaces two separate walks per directory finding
- **Pre-push Gate** — Total gate time 65s → 5s (13× faster)
- **Feather Weight** — 69/100 → 81/100 over session

### Changed
- **Pantheon Unification** — Standardized GEMINI.md, CLAUDE.md, and Portfolio Standard across all 5 repos
- **Ma'at Governance** — Integrated pipeline monitoring, diff-based coverage default, `--full` flag
- **Improved Logging** — Wired Go 1.21 `slog` into `ka` and `cleaner` cores for better diagnostics
- **Release Pipeline** — GoReleaser brews section enabled with `HOMEBREW_TAP_TOKEN` cross-repo secret
- **Weigh CLI** — Horus integration, `--fresh` flag for forcing index rebuild

### Fixed
- **Missing Imports** — Resolved `undefined: logging` error in `internal/cleaner/safety.go`
- **Domain Purge** — Replaced all instances of `sirsinexus.dev` with `sirsi.ai` in SirsiNexusApp
- **MCP Versioning** — Corrected version reporting to match release tags
- **gofmt** — Fixed formatting in `yield_test.go`
- **.gitignore Collision** — Unanchored `sirsi` pattern was ignoring `cmd/sirsi/seba.go`


---

## [0.3.0-alpha] — 2026-03-21/22 (Ship Week — Mirror + Audit + Thoth)

### Added
- **Mirror module** (`internal/mirror/`) — file deduplication engine
  - Three-phase scan: size grouping → partial hash (first+last 4KB) → full SHA-256
  - 8-worker parallel hashing with semaphore-bounded I/O
  - Smart keep/delete recommendations: protected > shallow > oldest > largest
  - Media type classification: photos, music, video, documents (30+ extensions)
  - Flags: `--photos`, `--music`, `--min-size`, `--max-size`, `--protect`
  - JSON output via `--json` for pipeline integration
- **Mirror GUI** (`internal/mirror/server.go`) — local web UI
  - Native macOS Finder folder picker via `/api/pick-folder`
  - Filesystem browser API via `/api/browse`
  - Graceful SIGINT/SIGTERM shutdown
  - Filter chips, advanced options, results view with keep/remove badges
  - Egyptian dark theme, Inter font, gold accents
- **𓁟 Thoth knowledge system** — persistent AI memory
  - Three-layer architecture: memory.yaml → journal.md → artifacts/
  - `thoth_read_memory` MCP tool for AI IDEs
  - Standalone CLI: `tools/thoth-init/` (auto-detects language, counts lines)
  - Installed across 4 Sirsi codebases (428,000+ lines)
  - 98% context reduction benchmarked on real projects
- **Decision log** (`internal/cleaner/decisions.go`)
  - Per-file action recording: path, size, SHA-256, reason, timestamp
  - Persists to `~/.config/anubis/mirror/decisions/`
  - Trash-first policy on macOS (reversible, "Put Back" works)
- **Performance documentation** (`docs/MIRROR_PERFORMANCE.md`)
  - Real benchmark data: 27.3x faster, 98.8% less disk I/O
  - Algorithm explanation, scaling properties, safety claims
- **Build log** (`docs/BUILD_LOG.md`) — public build-in-public chronicle
- **12 mirror tests** + existing suite = 303 total

### Changed
- **Seba graph** — complete kinetic rewrite (self-contained Canvas renderer)
- **Guard optimization** — pre-lowercased orphanPatterns keys in hot loop
- **README** — added Mirror benchmarks, Thoth section, updated architecture
- **GoReleaser** — fixed brews vs homebrew_casks, removed stale .goreleaser.yml

### Fixed
- **GUI folder picker** — was returning browser-relative paths → native macOS Finder dialog
- **moveToTrash** — silently ignored filepath.Abs error (could trash wrong file)
- **Drop zone text** — said "Drop folders here" but D&D can't work → now says "Select folders"
- **Dead code removed** — symlink check, unused groupID, FollowLinks field
- **Lint fixes** — errcheck, capitalized errors, unnecessary Sprintf
- **GoReleaser CI** — deprecated format, stale config file

### Stats
- 17 CLI commands, 58 scan rules, 19 internal modules
- 470 tests across 17 packages, all passing (with `-race`)
- ~17,000 lines of Go
- Lint clean (golangci-lint + staticcheck)
- Test coverage range: 93% (jackal) to 0% (2 untested modules: mapper, output)
- 6 bugs found and fixed in audit cycle, 7 modules test-covered in test sprint

### Session 7 (2026-03-22)
- **Statistics audit** — corrected 5 categories of inflated claims across 12 files
  - Scan rules: 64→58 (verified). Tests: ~395→470 (verified).
  - Removed fabricated cross-repo savings and "3M tokens in 11 sessions" claim.
- **Structured logging** (`internal/logging/`) — Go 1.21+ slog to stderr
  - `--verbose` (debug), `--quiet` (error-only), `--json` (structured) modes
  - Instrumented mirror and ka scanners with debug points
- **Platform abstraction** (`internal/platform/`) — cross-platform interface
  - Darwin, Linux, Mock implementations
  - MoveToTrash, ProtectedPrefixes, PickFolder, OpenBrowser, SupportsTrash
  - Mock enables testing platform-specific code without system calls
- **Case studies** — 3 verified studies in `docs/case-studies/`
  - Thoth Context Savings, Mirror Dedup Performance, Ka Ghost Detection
- **CI fixes** — platform skip guards for macOS-only tests, homebrew tap disabled
- **Rules canonized** — A14 (Statistics Integrity), A15 (Session Definition)
- **GitHub Release** — v0.3.0-alpha published with 6 binaries
- **`SirsiMaster/homebrew-tools`** repo created (pending PAT setup)

### Session 8 (2026-03-23)
- **Platform interface wired** into cleaner and mirror modules (Priority 1 complete)
  - Replaced 3 `runtime.GOOS` checks in `cleaner/safety.go` with `platform.Current()`
  - Replaced `moveToTrash()` and `protectedPrefixes` map with platform interface calls
  - Replaced `OpenBrowser()` switch and `handlePickFolder` osascript in `mirror/server.go`
  - Tests use `platform.Set(&Mock{})` for cross-platform testing
- **CI lint fixes** — resolved 8 lint errors that broke 5 consecutive CI runs
  - `gofmt`: alignment in `ignore_test.go`, `registry_test.go`
  - `govet/unusedwrite`: struct assertions in `scarab_test.go`, `sight_test.go`
  - `misspell`: "cancelled" → "canceled" in platform package
- **Pre-push hook** (`.githooks/pre-push`) — mirrors CI checks locally
  - Runs gofmt + go vet + golangci-lint + go build before every push
  - Prevents lint issues from ever reaching the pipeline
- **Maat proposed** — pipeline purifier module (CI monitoring + auto-remediation)


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

### Added (Day 6: MCP Server + IDE Integrations)
- **MCP module** (`internal/mcp/`) — Model Context Protocol server
  - JSON-RPC 2.0 over stdio, MCP spec 2025-03-26 compliant
  - `initialize` handshake with capability negotiation
  - `tools/list` and `tools/call` for tool invocation
  - `resources/list` and `resources/read` for resource access
  - `ping` and method-not-found handling
- **`anubis mcp`** — start MCP server for AI IDE integration
- **4 MCP Tools:**
  - `scan_workspace` — run Jackal scan engine on a directory
  - `ghost_report` — run Ka ghost detection
  - `health_check` — system health summary with grade
  - `classify_files` — brain-powered semantic file classification
- **3 MCP Resources:**
  - `anubis://health-status` — system health as JSON
  - `anubis://capabilities` — modules, commands, and scan rules
  - `anubis://brain-status` — neural brain installation status
- **VS Code extension scaffold** (`extensions/vscode/`)
  - Extension manifest with Eye of Horus sidebar concept
  - 4 commands: scan workspace, ghost report, health check, install brain
  - Status bar icon, activity bar sidebar, configuration options
- **Workspace config** — `.anubis/config.yaml` template for per-project settings
- **14 MCP tests** — server lifecycle, tool calls, resource reads, error handling
- **IDE config examples** — Claude Code, Cursor, Windsurf setup instructions

### Refs
- Canon: ANUBIS_RULES.md, docs/DEVELOPMENT_PLAN.md
- ADR: ADR-001
- Changelog: v0.2.0-alpha — Day 6 MCP Server

### Added (Day 7: Scales Policy Engine + Agent Hardening)
- **Scales module** (`internal/scales/`) — YAML policy engine
  - Policy parser with validation (operators, severities, metrics)
  - Threshold normalization (KB/MB/GB/TB → bytes)
  - Built-in default workstation hygiene policy
- **Policy enforcement** (`internal/scales/enforce.go`)
  - Evaluates scan results against configurable thresholds
  - Generates pass/warn/fail verdicts with remediation suggestions
  - Collects metrics from Jackal (waste) and Ka (ghosts)
- **`anubis scales enforce`** — run policies against current state
  - Custom policy files via `-f` flag
  - JSON output support
  - Eye of Horus/Ra upsell for fleet enforcement
- **`anubis scales validate`** — validate policy YAML syntax
- **`anubis scales verdicts`** — show enforcement results
- **Agent hardening** (`cmd/anubis-agent/`)
  - Fixed command set: scan, report, clean, version (Rule A3)
  - All output JSON via AgentResponse envelope
  - Clean requires `--confirm` flag (Rule A1)
  - Health grading: EXCELLENT/GOOD/FAIR/NEEDS_ATTENTION
- **Example policy file** — workstation + CI/CD templates
- **13 scales tests** — parsing, validation, normalization, enforcement, verdicts

### Refs
- Canon: ANUBIS_RULES.md, docs/DEVELOPMENT_PLAN.md
- ADR: ADR-001
- Changelog: v0.2.0-alpha — Day 7 Scales + Agent

### Changed (Day 8: Polish + Release)
- **README.md** — complete rewrite with all 17 commands, 11 modules, MCP guide
- **VERSION** — bumped to `0.2.0-alpha`
- **Binary audit:**
  - `anubis`: 7.7 MB (< 15 MB budget ✅)
  - `anubis-agent`: 2.1 MB (< 5 MB budget ✅)
- **Test suite:** 72 tests, 7 packages, all passing
- **Code size:** 12,277 lines of Go across 15 internal modules
- **gofmt + go vet** — clean

---

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
