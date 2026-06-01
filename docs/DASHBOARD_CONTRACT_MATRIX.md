# Dashboard Contract Matrix — Freeze Artifact

| Field | Value |
| :--- | :--- |
| **Status** | **FROZEN (E1/E2/E3/E5 implemented)** — routed to codex-pantheon for the implementation-review gate. E4/E6 fast-follow. |
| **Produced by** | claude-pantheon Phase-A read-only audit (codex-approved scope), 2026-06-01; freeze implemented same day per codex freeze-gate ruling (item `162436`) |
| **Governs** | [SURFACE_REWRITE_MASTER_PLAN.md](SURFACE_REWRITE_MASTER_PLAN.md), [ADR-020](ADR-020-INTERACTIVE-SURFACE-REOPENED.md) |
| **Verdict** | **Contract frozen on E1/E2/E3/E5.** Action side now typed + endpoint-complete; all destructive ops gated by one server-enforced confirm contract; `/api/stats` typed. Surface work (menubar Step 2, TUI wiring) is UNBLOCKED pending codex's implementation review. See §Freeze Complete below. |

> **Why this exists:** codex's plan review (`APPROVE WITH GATES`) ruled that `internal/dashboard` "is a contract only after it has a written endpoint/action/schema matrix. A package existing in source is not a frozen contract." This is that matrix. It is the gate.

---

## Headline

The menubar hosts the dashboard server in-process (`cmd/sirsi-menubar/main.go:131-143`) **but bypasses it entirely** — all 17 menu clicks shell out via `spawnTUIWithCommand` (`main.go:214-270`). The TUI scaffold declares command ids that **dispatch nowhere** (`internal/tui/state.go:42-86`). So "replace terminal-spawns with contract calls" (menubar Step 2) is **impossible for 9 of the actions** — they have no endpoint. **Contract completion is the true first work item, ahead of every surface.**

---

<!-- Phase-A audit output below — preserved verbatim as the cited evidence base. -->

## 1. Contract Matrix

One row per user action. `surface consumers`: **CLI** = native cobra; **menubar** = via `spawnTUIWithCommand` (terminal spawn, NOT contract); **browser** = dashboard HTML page; **TUI** = registry id exists but unwired.

| Action | CLI verb/key | dashboard endpoint OR in-proc fn | request schema | response schema | streaming/result | notify/store side-effect | destructive? confirm? | surface consumers |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Scan (latest persisted) | `sirsi scan` (`main.go:140`) | `GET /api/findings` (`findings.go:16`) | none | `PersistedScan` (untyped `map` on empty: 200+`error`, `findings.go:19`) | poll | none | no | CLI, browser(`pages.go:227`), menubar→`scan`(`main.go:238`), TUI `CmdScan`(unwired) |
| Scan (run fresh) | `sirsi scan` | `POST /api/run?cmd=scan` (`runner.go:181`) | query `cmd=scan` | `{status,cmd}`; output via SSE | **SSE** `/api/events` | `notify.Record` (`runner.go:153-167`) | no | CLI, browser(`pages.go:504`), menubar(terminal) |
| Clean / Judge | `sirsi clean`, hidden `judge` (`main.go:160,152`) | `POST /api/clean` (`findings.go:37`) | `{indices:[]int, dry_run:bool}` | `{dry_run,cleaned,bytes_freed,freed_human,skipped,errors[]}` | none | none | **YES.** `Confirm:!DryRun` (`findings.go:91`); omitted `dry_run` ⇒ real delete | CLI, browser(`pages.go:305,447`), menubar→`anubis clean`(`main.go:240`), TUI `CmdClean`(unwired) |
| Ghosts (scan) | `sirsi ghosts` (`main.go:146`) | `GET /api/ghosts` (`modules.go:25`) | none | `[]ghostJSON` (inline anon, `modules.go:36`) | poll | none | no | CLI, browser(`pages.go:326`), menubar→`ghosts`(`main.go:242`) |
| Ghost clean | `sirsi anubis ...` | `POST /api/ghosts/clean` (`modules.go:92`) | `{app_name, dry_run}` | `{dry_run,bytes_freed,freed_human,files_removed}` | none | none | **YES.** `dry_run` body only; no confirm | browser(`pages.go:342`) only |
| Doctor / Diagnose | `sirsi diagnose`/`doctor` (`main.go:449`) | `GET /api/doctor` (`modules.go:140`) | none | `guard.DoctorReport` (typed) | poll | none | no | CLI, browser(`pages.go:354`), menubar `mStats`→`diagnose`(`main.go:219`) |
| Status | `sirsi status` (`main.go:214`) | `GET /api/stats` via `StatsFn` (`api.go:15`) | none | `StatsSnapshot` (untyped `[]byte` passthrough, `server.go:23`) | poll | none | no | CLI, browser(`pages.go:195`), menubar(stats loop `main.go:172`), TUI `CmdStatus`(unwired) |
| Guard / Monitor | `sirsi monitor`, hidden `guard` (`main.go:187,180`) | `GET /api/guard/stats` (`modules.go:188`) | none | `guard.Stats` (typed) | poll | none | no | CLI, browser, menubar→`guard`(`main.go:248`); in-proc `startGuardBridge`(`main.go:429`) |
| Guard renice | (no CLI verb) | `POST /api/guard/renice?target=` (`modules.go:199`) | query `target` | `guard.ReniceResult` (typed) | none | none | mutates nice; no confirm | browser(`pages.go:460`) only |
| Slay (kill proc group) | (no dedicated verb) | `POST /api/slay?target=&dry_run=` (`modules.go:151`) | query `target`,`dry_run` | `{target,dry_run,killed,failed,skipped,errors[]}` | none | none | **YES.** browser calls `dry_run=false` directly (`pages.go:435`); no server confirm | browser only |
| Quality / Audit | `sirsi audit`/`quality` (`main.go:208,481`) | **MISSING** (runner `quality` key spawns CLI, no JSON) | — | — | SSE if run | notify | no | CLI, menubar→`maat audit`(`main.go:245`), browser via run |
| Risk (Osiris) | `sirsi risk` (`main.go:202`) | **MISSING** | — | — | — | — | no | CLI, menubar→`osiris risk`(`main.go:256`), TUI `CmdRisk`(unwired) |
| Network audit | `sirsi network`/`fix` (`main.go:469,193`) | **MISSING** (runner `network` key) | — | — | SSE if run | notify | `--fix` mutates DNS/fw; no confirm | CLI, browser via run |
| Hardware (Seba) | `sirsi hardware` (`main.go:475`) | **MISSING** (runner `hardware` key) | — | — | SSE if run | notify | no | CLI, menubar→`seba hardware`(`main.go:255`) |
| Duplicates/Dedup | `sirsi duplicates`/`dedup` (`main.go:173`) | **MISSING** (runner `dedup` key) | — | — | SSE if run | notify | no | CLI, browser via run |
| Thoth sync | `sirsi thoth sync` (`thoth.go:94`) | **MISSING** | — | — | — | writes `.thoth/` | no | CLI, menubar→`thoth sync`(`main.go:251`) |
| Seshat ingest | `sirsi seshat ingest` (`seshat.go:54`) | **MISSING** | — | — | — | writes knowledge store | no | CLI, menubar→`seshat ingest`(`main.go:253`) |
| Net align | `sirsi net align` (`net.go:40`) | **MISSING** | — | — | — | — | no | CLI, menubar→`net align`(`main.go:258`) |
| Ra deploy | `sirsi ra deploy` (`ra.go:386`) | **MISSING** (no deploy trigger) | — | — | spawns windows | none | spawns N windows; `--dry-run` exists; no confirm | CLI, menubar→`ra deploy`(`main.go:223`), TUI `CmdRaDeploy`(unwired) |
| Ra kill | `sirsi ra kill` (`ra.go:438`) | **MISSING** | — | — | `ra.KillAll` | none | **YES (KillAll).** No `--dry-run`, no confirm (`ra.go:438-450`) | CLI, menubar→`ra kill`(`main.go:225`), TUI `CmdRaKill`(unwired) |
| Ra collect | `sirsi ra collect` (`ra.go:453`) | **MISSING** | — | — | ingest pipeline | writes Seshat+Thoth | no | CLI, menubar→`ra collect`(`main.go:227`) |
| Ra status | `sirsi ra status` (`ra.go:234`) | `GET /api/ra/status` (`modules.go:428`) | none | `{deployed,started_at,all_done,windows[]}` (inline) | poll | none | no | CLI, browser(`pages.go:403`), menubar→`ra status`(`main.go:221`) |
| Ra scopes | (implicit) | `GET /api/ra/scopes` (`modules.go:474`) | none | `[]scopeJSON` (inline) | poll | none | no | browser(`pages.go:394`) only |
| Horus scan | `sirsi horus ...` | `GET /api/horus/scan?path=` (`modules.go:279`) | query `path` | `horus.SymbolGraph` (typed) | poll | cache write | no | CLI, browser(`pages.go:471`) |
| Horus query | `sirsi horus ...` | `GET /api/horus/query?...` (`modules.go:306`) | query | `[]horus.Symbol`/`graph.Stats` | poll | none | no | CLI, browser(`pages.go:477`) |
| Horus report | (none) | `GET /api/horus/report` (`modules.go:227`) | none | `horus.WorkstationReport` (typed) | poll | none | no | **no consumer** |
| Vault search | `sirsi vault ...` | `GET /api/vault/search?q=&limit=` (`modules.go:344`) | query `q` | `vault.SearchResult` (typed) | poll | none | no | CLI, browser(`pages.go:487`) |
| Vault stats | `sirsi vault ...` | `GET /api/vault/stats` (`modules.go:370`) | none | `vault.Stats`; on unavail 200+`error` (`modules.go:373`) | poll | none | no | CLI, browser(`pages.go:383`) |
| Vault prune | `sirsi vault ...` | `POST /api/vault/prune?older_than=` (`modules.go:388`) | query `older_than` | `{removed,older_than}` | none | deletes entries | mutates; no confirm | CLI; **no UI consumer** |
| Notifications | `sirsi notifications` | `GET /api/notifications?...` (`api.go:31`) | query | `[]notify.Notification` (typed) | poll | none | no | CLI, browser(`pages.go:367`), menubar(recent `main.go:197`) |
| Stele ledger | (none) | `GET /api/stele?...` (`api.go:68`) | query | `[]stele.Entry` (typed) | poll | none | no | **no consumer** |
| Event stream | — | `GET /api/events` SSE (`events.go:95`) | `Last-Event-ID`/`?since=` | SSE `Event{id,type,data}` | **SSE** | none | no | browser(`pages.go:510`); menubar does **not** consume despite hosting the buffer |
| Run status | — | `GET /api/run/status` (`runner.go:206`) | none | `{running,current}` (untyped) | poll | none | no | (no consumer) |
| Maat audit (module) | `sirsi maat audit` (`maat.go:56`) | **MISSING** (only `quality` runner key) | — | — | — | — | no | CLI, menubar→`maat audit`(`main.go:245`) |
| Router ack | `sirsi thread/router ...` | **MISSING** (no `/api/router/*`) | — | — | — | writes inbox | no | CLI, TUI `CmdRouterAck`(unwired) |

---

## 2. Gap List (file:line evidence)

### A. Surface action with NO contract endpoint (bypass)
- All 17 menubar clicks → `spawnTUIWithCommand` (`cmd/sirsi-menubar/main.go:214-270`, shell via osascript `main.go:291-360`). UX_AUDIT flags this as the model to delete (`docs/UX_AUDIT_2026-06-01.md:28,76`).
- Menubar actions with **no endpoint at all** (cannot migrate in Step 2 until built): `maat audit`, `thoth sync`, `seshat ingest`, `seba hardware`, `osiris risk`, `net align`, `ra deploy`, `ra kill`, `ra collect`. Only `scan`, `ghosts`, `diagnose`, `ra status`, `guard` have an endpoint.
- TUI ids `CmdAudit, CmdRisk, CmdRaDeploy, CmdRaKill, CmdRouterAck` (`internal/tui/command.go:138-146`) dispatch nowhere (`state.go:42-86`) and have no backing endpoint.

### B. Endpoints with NO consumer (dead surface)
- `GET /api/horus/report`, `GET /api/stele`, `POST /api/vault/prune`, `GET /api/run/status` — no UI/menubar/TUI consumer. `/api/ra/scopes`, `/api/guard/renice`, `/api/slay` are browser-only.

### C. Destructive actions lacking a confirm contract
- `POST /api/slay?dry_run=false` — browser calls directly with a JS `confirm()` only; **server enforces no token** (`modules.go:151-184`, `pages.go:435`).
- `ra kill`/`ra.KillAll` — no `--dry-run`, no confirm, no endpoint (`ra.go:438-450`).
- `POST /api/clean` — omitted `dry_run` ⇒ real delete (`findings.go:32,91`). Violates Rule A1 spirit. TUI intends a confirm modal (`state.go:42-52`) but the HTTP layer has no equivalent.
- `POST /api/vault/prune`, `POST /api/guard/renice` — mutate with no confirm.

### D. Schema inconsistencies / untyped responses
- Two "200-OK with `error` field" soft states: `/api/findings` (`findings.go:19`), `/api/vault/stats` (`modules.go:373`).
- `/api/stats` is opaque `[]byte` (`server.go:23`) — no shared type; TUI/Mac must re-handwrite `StatsSnapshot`.
- Inline anon structs no client can import: `ghostJSON`,`windowJSON`,`scopeJSON`; ad-hoc `map[string]interface{}` for `/api/run`,`/clean`,`/slay`,`/ra/status`,`/vault/prune`. No shared response-type package.
- Envelope mismatch: dashboard has no `{ok,data,error}` wrapper; `mobile/*.go` does (`DASHBOARD_API.md:213`). Param drift cataloged in `DASHBOARD_API_GAP.md:36-47`.

### E. Freeze blockers (must land before menubar Step 2 + TUI wiring)
1. **Add the missing action endpoints**: `audit`, `risk`, `network`/`fix`, `hardware`, `dedup`, `thoth sync`, `seshat ingest`, `net align`, and Ra triggers `ra deploy`/`ra kill`/`ra collect`.
2. **Destructive-action confirm contract** (token or two-phase) for `clean`, `ghosts/clean`, `slay`, `ra kill`, `vault/prune`, `renice`; make `dry_run` **default-true server-side**.
3. **Shared typed response package** (replace inline anon + `map[string]interface{}`); **type `/api/stats`**.
4. Resolve the two soft-error 200s + the mobile-vs-dashboard envelope decision (`DASHBOARD_API.md:211-213`).
5. Decide the scan contract: synchronous `POST /api/scan` vs `POST /api/run?cmd=scan`+SSE (`DASHBOARD_API_GAP.md:17`).
6. Give the runner an args/result schema so "run any verb" becomes a real contract; have the menubar consume `/api/events` (it hosts the buffer at `main.go:129` but never reads it).

**Bottom line:** read-side (stats, findings, ghosts, doctor, guard, horus, vault, ra-status/scopes, notifications, stele, events) is broadly coherent. The **write/action side is unfrozen**: ~12 actions have no endpoint, destructive ops share no confirm contract, responses are untyped/inline. Menubar Step 2 and TUI wiring are blocked on E1–E3 at minimum.

---

## Freeze Complete — E1/E2/E3/E5 implemented (2026-06-01)

Implemented in a single lane (claude-pantheon) per codex's freeze-gate ruling (item `162436`): minimal-but-load-bearing freeze = E1 + E2 + E3 + E5, E4/E6 fast-follow. Commits `8675796`, `f5b3084`, `edb8a74`. `go test -race ./internal/dashboard/` green.

| Gap | Resolution | Where |
| :--- | :--- | :--- |
| **E3** — untyped responses; `/api/stats` opaque `[]byte` | Typed `ActionRequest`/`ActionResult`/`PreparedAction`; typed `StatsResponse` (JSON tags mirror the menubar `StatsSnapshot`); `/api/stats` decodes through it with honest passthrough fallback. | `contract.go`, `api.go:apiStats` |
| **E2** — destructive ops share no confirm contract; `/api/clean` deleted when `dry_run` omitted | `ConfirmGuard`: server-enforced, single-use, tokenized two-phase confirm (SHA-256 action hash; missing/unknown/expired/mismatched/reused all rejected). Shared `requireConfirm()` helper. All 5 destructive endpoints (`clean`, `ghosts/clean`, `slay`, `vault/prune`, + destructive registry actions) gated. **Omitted `dry_run` can never destroy** (Rule A1). | `confirm.go`, `findings.go`, `modules.go`, `actions.go` |
| **E1** — ~12 actions had no endpoint | Canonical `ActionSpec` registry (legacy 8 + 12 gap-list actions) reachable via typed `POST /api/run`; `GET /api/actions` is the discovery endpoint. Destructive actions (`network/fix`, `ra/deploy`, `ra/kill`) carry the confirm flag. | `actions.go` |
| **E5** — scan/run was query-only `?cmd=` | `POST /api/run` accepts a typed `ActionRequest`; server-defined base args + opt-in caller positional args (no arbitrary injection); runner+SSE remains the streaming path; legacy `?cmd=` retained but cannot fire destructive. | `actions.go:dispatchRun`, `runner.go` |

**Confirm design (E2) — exactly as codex specified:** prepare/dry-run returns typed preview + `confirm_token` + `expires_at` + stable hash; commit requires token + matching action/target/params; server rejects missing/expired/mismatched/reused; default is dry-run/prepare; the token is the safety boundary (no reliance on client `confirm()`). Works for HTTP, TUI modal, and menubar/SwiftUI dialog uniformly.

**Deferred (fast-follow, documented honestly):**
- **E4** — soft-error 200s + mobile-vs-dashboard envelope decision. Existing read endpoints kept additive-compatible; no NEW write endpoint adds a soft-error or untyped map.
- **E6** — broader runner args/result schema beyond the registry; menubar consuming `/api/events`.
- **renice** — left unguarded by design: it adjusts process priority (reversible), not a deletion/kill. Flagged for codex to confirm confirm-exemption.

**Surface work is now unblocked** pending codex's implementation review: menubar Step 2 replaces `spawnTUIWithCommand` with `GET /api/actions` + typed `POST /api/run` (+ confirm flow for destructive); TUI wiring consumes the same contract.
