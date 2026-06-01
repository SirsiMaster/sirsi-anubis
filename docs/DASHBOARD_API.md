# Dashboard API

**Source:** `internal/dashboard/`
**Base URL (CLI/browser default):** `http://127.0.0.1:9119`
**Base URL (Mac app, planned):** `unix:///Users/<user>/Library/Application Support/ai.sirsi.pantheon/dashboard.sock`
**Status:** Contract documentation, captured from the current implementation.

## Conventions

- **Content-Type:** `application/json` for all `/api/*` endpoints, except `/api/events` (`text/event-stream`).
- **Success body:** the resource's JSON shape, written directly by `writeJSON(w, v)` in `internal/dashboard/server.go:200`. **No envelope.** Callers receive the resource as the root JSON object/array.
- **Error body:** `{"error": "<message>"}` written by `writeError(w, msg, code)`. The HTTP status code is the canonical error signal; the message is a human-readable string.
- **Method:** every endpoint is `GET` unless explicitly noted.
- **Default port:** 9119 (`internal/dashboard/colors.go:23`).

## HTML routes (out of scope for the Mac bridge)

These render HTML for the browser dashboard. The Mac app will not consume them.

| Path | Notes |
| :--- | :--- |
| `/` | Overview page |
| `/scan` | Findings page |
| `/ghosts` | Ghosts page |
| `/guard` | Guard page |
| `/notifications` | Notifications page |
| `/horus` | Horus page |
| `/vault` | Vault page |

## JSON API routes

### `GET /api/stats`

- **Source:** `apiStats` (api.go:15)
- **Request:** no params.
- **Response:** the JSON returned by the host's `Config.StatsFn` callback. In the menubar process this is `StatsSnapshot` (cmd/sirsi-menubar/stats.go:31). Contains `total_ram`, `used_ram`, `ram_percent`, `ram_pressure`, `uncommitted_files`, `time_since_commit`, `git_branch`, `osiris_risk`, `primary_accelerator`, `active_deities[]`, `ra_deployed`, `ra_scopes[]`, `disk_waste_estimate`, `timestamp`, `collected_in`.
- **503:** `{"error":"stats not available"}` if no `StatsFn`.
- **500:** `{"error":"stats collection failed"}` on host-side error.
- **Streaming:** poll. No SSE.

### `GET /api/notifications`

- **Source:** `apiNotifications` (api.go:31)
- **Query:** `limit` (default 50), `source` (filter by source), `severity` (filter by severity).
- **Response:** `[]notify.Notification` — always an array, `[]` when empty.
- **503/500:** standard error shape.
- **Streaming:** poll.

### `GET /api/stele`

- **Source:** `apiStele` (api.go:68)
- **Query:** `type` (filter), `limit` (default 100).
- **Response:** `[]stele.Entry`, newest-first. `[]` if the stele file is absent.
- **500:** `{"error":"stele read failed"}` on read error.
- **Streaming:** poll.

### `GET /api/events`

- **Source:** `apiEvents` (events.go:95)
- **Headers:** standard SSE — `text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`.
- **Query / Headers:** `Last-Event-ID` (header) or `?since=N` (query) to resume.
- **Response:** SSE stream of `Event` objects from `EventBuffer`. Polls the buffer every 1s and flushes new events.
- **503:** `{"error":"event stream not available"}` if no `EventBuffer`.
- **500:** `{"error":"streaming not supported"}` if the writer is not a `http.Flusher`.
- **Streaming:** **SSE**, long-lived.

### `POST /api/run`

- **Source:** `apiRun` (runner.go:181)
- **Method:** **POST required**, else 405 `{"error":"POST required"}`.
- **Query:** `cmd` (the runner command to start).
- **Response:** `{"status":"started","cmd":"<key>"}`.
- **503/409/400:** standard errors.
- **Streaming:** none on this endpoint; consume `/api/events` for runner output.

### `GET /api/run/status`

- **Source:** `apiRunStatus` (runner.go:206)
- **Response when idle:** `{"running": false}`. **When running:** `{"running": true, ...}` plus runner-internal fields.
- **Streaming:** poll.

### `GET /api/findings`

- **Source:** `apiFindings` (findings.go:16)
- **Response:** the full `PersistedScan` JSON (the latest scan), encoded directly with `json.NewEncoder`. Includes `findings[]`, `total_size`, etc.
- **No persisted scan:** `{"findings":[],"error":"No scan results. Run a scan first."}` — note this is **200 OK with an `error` field in the body**, not an HTTP error. Documented as-is.
- **Streaming:** poll.

### `POST /api/clean`

- **Source:** `apiClean` (findings.go:37)
- **Method:** **POST required**, else 405.
- **Request body:** `{"indices":[<int>,...], "dry_run":<bool>}`.
- **Response:** `{"dry_run":<bool>, "cleaned":<int>, "bytes_freed":<int>, "freed_human":"<str>", "skipped":<int>, "errors":[<str>,...]}`.
- **400:** invalid JSON / no findings selected / invalid index.
- **404:** `{"error":"no scan results available"}`.
- **500:** clean engine failure.
- **Streaming:** none.

### `GET /api/ghosts`

- **Source:** `apiGhosts` (modules.go:25)
- **Response:** `[]ghostJSON` with fields `app_name`, `bundle_id`, `total_size`, `total_files`, `size_human`, `in_launch_services`, `residuals[].{path,type,size_bytes,file_count}`. `[]` if none.
- **500:** scan failure.
- **Streaming:** poll. The scan itself runs synchronously; first response can take seconds.

### `POST /api/ghosts/clean`

- **Source:** `apiGhostClean` (modules.go:92)
- **Method:** POST required.
- **Request body:** `{"app_name":"<str>", "dry_run":<bool>}`.
- **Response:** `{"dry_run":<bool>, "bytes_freed":<int>, "freed_human":"<str>", "files_removed":<int>}`.
- **400/404/500:** standard.
- **Streaming:** none.

### `GET /api/doctor`

- **Source:** `apiDoctor` (modules.go:140)
- **Response:** `guard.DoctorReport` JSON (includes a `Score` field).
- **500:** `{"error":"doctor: <err>"}`.
- **Streaming:** poll. Synchronous run.

### `POST /api/slay`

- **Source:** `apiSlay` (modules.go:151)
- **Method:** POST required.
- **Query:** `target` (`guard.SlayTarget`), `dry_run` (default `true`; `dry_run=false` actually kills).
- **Response:** `{"target":"<str>","dry_run":<bool>,"killed":<int>,"failed":<int>,"skipped":<int>,"errors":[<str>,...]}`.
- **400/500:** standard.
- **Streaming:** none.

### `GET /api/guard/stats`

- **Source:** `apiGuardStats` (modules.go:188)
- **Response:** `guard.Stats` JSON (includes `PressureLevel`).
- **500:** standard.
- **Streaming:** poll.

### `POST /api/guard/renice`

- **Source:** `apiRenice` (modules.go:199)
- **Method:** POST required.
- **Query:** `target` (`lsp` default, or `all`).
- **Response:** `guard.ReniceResult` JSON.
- **400/500:** standard.
- **Streaming:** none.

### `GET /api/horus/report`

- **Source:** `apiWorkstationReport` (modules.go:227)
- **Response:** `horus.WorkstationReport` JSON — `timestamp`, `hostname`, `os`, `arch`, scan summary, `health_score`, `ram_pressure`, `ram_percent`, `git_branch`, `uncommitted_files`.
- **Streaming:** poll.

### `GET /api/horus/scan`

- **Source:** `apiHorusScan` (modules.go:279)
- **Query:** `path` (default `.`).
- **Response:** `horus.SymbolGraph` JSON (cached after first parse).
- **500:** `{"error":"horus scan: <err>"}`.
- **Streaming:** poll. Synchronous.

### `GET /api/horus/query`

- **Source:** `apiHorusQuery` (modules.go:306)
- **Query:** `path` (default `.`); one of `name`, `kind`, `filter`. If none provided, returns `graph.Stats`.
- **Response:** `[]horus.Symbol` for `name`/`kind`/`filter`; `graph.Stats` otherwise.
- **500:** parse failure.
- **Streaming:** poll.

### `GET /api/vault/search`

- **Source:** `apiVaultSearch` (modules.go:344)
- **Query:** `q` (required), `limit` (default 20).
- **Response:** `vault.SearchResult` JSON.
- **400:** `{"error":"missing q parameter"}`.
- **503:** `{"error":"vault not available"}`.
- **500:** search failure.
- **Streaming:** poll.

### `GET /api/vault/stats`

- **Source:** `apiVaultStats` (modules.go:370)
- **Response:** `vault.Stats` JSON. **Note:** if vault is not available, returns 200 OK with `{"totalEntries":0,"error":"vault not available"}` (in-body error, not HTTP error).
- **Streaming:** poll.

### `POST /api/vault/prune`

- **Source:** `apiVaultPrune` (modules.go:388)
- **Method:** POST required.
- **Query:** `older_than` Go duration (default `720h` = 30 days).
- **Response:** `{"removed":<int>,"older_than":"<dur>"}`.
- **400/500/503:** standard.
- **Streaming:** none.

### `GET /api/ra/status`

- **Source:** `apiRaStatus` (modules.go:428)
- **Response:** `{"deployed":<bool>,"started_at":"<rfc3339>","all_done":<bool>,"windows":[{"name":"<str>","pid":<int>,"state":"<str>","exit_code":<int>,"log_tail":"<str>","duration":"<str>"}]}`. When no deployment: `{"deployed":false,"windows":[]}` (200 OK).
- **Streaming:** poll.

### `GET /api/ra/scopes`

- **Source:** `apiRaScopes` (modules.go:474)
- **Response:** `[]scopeJSON` with `name`, `display_name`, `repo_path`, `priority`, `deadline`, `sprints`. `[]` when no scopes loaded (also returned 200 OK on loader failure).
- **Streaming:** poll.

## Envelope summary (the headline contract finding)

- **Success:** the resource is the root of the response body. **There is no `{ok, data, error}` wrapper.**
- **Error:** `{"error": "<message>"}` paired with an HTTP non-2xx status.
- **Two endpoints intentionally embed an `error` field in a 200 OK body** (`/api/findings`, `/api/vault/stats`) when the underlying resource is empty/unavailable. Callers must treat the presence of `error` as a soft-state signal, not a transport failure.

This shape **differs from `mobile/*.go`**, which uses a `Response{ok, data, error}` envelope. See `docs/DASHBOARD_ENVELOPE_DECISION.md`.

## Auth / transport

- **Auth:** none. Single-user, loopback-only.
- **Singleton:** `Server.Start()` acquires a process lock; only one dashboard runs at a time.
- **CORS:** not currently set. Mac app will use a unix socket, sidestepping browser-origin concerns; CLI/browser use is same-origin.
- **Versioning:** none. There is no `/api/v1/` prefix today. If the envelope decision lands as new endpoints, that batch may introduce `/api/v2/` — separate decision.
