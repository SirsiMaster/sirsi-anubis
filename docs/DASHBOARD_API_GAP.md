# Dashboard API Gap

Map every current `ios/Pantheon/Services/PantheonBridge.swift` call to **exactly one** of:

- **1:1** — existing endpoint, shape matches.
- **Adapter** — existing endpoint exists, but envelope/shape needs a Swift-side adapter or a new compat endpoint. The choice is in `DASHBOARD_ENVELOPE_DECISION.md`; this column only flags that an adapter is required.
- **New endpoint** — no equivalent on the dashboard server. Must be added before the Mac bridge can call it.
- **CLI one-shot** — no in-process equivalent makes sense; the Mac app shells out to `sirsi <verb> --json` instead.

Ambiguous cases are marked **AMBIGUOUS** rather than papered over, per codex's step-3 condition.

## Gap table

| iOS bridge call | Underlying Go | Existing dashboard endpoint | Disposition |
| :--- | :--- | :--- | :--- |
| `anubisCategories()` | `mobile.AnubisCategories` → static list of 7 categories | none | **New endpoint** `GET /api/anubis/categories` (returns the static list verbatim) **OR** CLI one-shot `sirsi scan --list-categories --json`. The static-list case is so small that a new endpoint is the leaner choice. |
| `anubisScan(rootPath:, categories:)` | `mobile.AnubisScan` → `jackal.Engine.Scan` | `GET /api/findings` returns the **latest persisted** scan; it does not run a fresh scan on demand | **AMBIGUOUS**. Two paths: (a) `POST /api/scan` new endpoint that runs `jackal.Engine.Scan` with `rootPath`/`categories` and returns the new findings (the iOS bridge runs synchronously with a 5-minute timeout); (b) submit the scan via `POST /api/run?cmd=scan` and consume `/api/events`. Path (a) matches iOS semantics; path (b) is the long-term LEAN choice for any operation > a few seconds. **Recommend (a) for parity, with a note that the runner path is the canonical replacement.** |
| `kaHunt(includeSudo:)` | `mobile.KaHunt` → `ka.Scanner.Scan` | `GET /api/ghosts` runs the same ghost scan | **Adapter**. The dashboard endpoint already runs `ka.Scanner.Scan(ctx, false)`. It does not accept the `includeSudo` flag — the iOS call passes `false` by default anyway. Drop the parameter for v1; if `includeSudo=true` is ever needed, add a `?include_sudo=true` query param (additive). |
| `kaEnumerateApps()` | `mobile.KaEnumerateApps` → `ka.EnumerateInstalledApps` | none | **New endpoint** `GET /api/ka/apps`. Small wrapper; or CLI one-shot `sirsi ka apps --json`. **Recommend new endpoint** for parity with `/api/ghosts`. |
| `thothInit(projectRoot:)` | `mobile.ThothInit` → `thoth.Init` | none | **New endpoint** `POST /api/thoth/init` with body `{"project_root":"<str>"}`. |
| `thothSync(root:)` | `mobile.ThothSync` → `thoth.Sync` | none | **New endpoint** `POST /api/thoth/sync` with body `{"repo_root":"<str>"}`. |
| `thothCompact(root:)` | `mobile.ThothCompact` → `thoth.Compact` | none | **New endpoint** `POST /api/thoth/compact` with body `{"repo_root":"<str>"}`. |
| `thothDetectProject(root:)` | `mobile.ThothDetectProject` → `thoth.DetectProject` | none | **New endpoint** `GET /api/thoth/project?root=<str>`. |
| `sebaDetectHardware()` | `mobile.SebaDetectHardware` → `seba.DetectHardware` | partial — `/api/horus/report` includes hardware-adjacent fields but not the full `HardwareProfile` | **New endpoint** `GET /api/seba/hardware`. |
| `sebaDetectAccelerators()` | `mobile.SebaDetectAccelerators` → `seba.DetectAccelerators` | none | **New endpoint** `GET /api/seba/accelerators`. |
| `seshatIngest(options:)` | `mobile.SeshatIngest` → `seshat.Ingest` | none | **New endpoint** `POST /api/seshat/ingest` with the existing options JSON as the body. |
| `seshatListSources()` | `mobile.SeshatListSources` → `seshat.ListSources` | none | **New endpoint** `GET /api/seshat/sources`. |
| `seshatListTargets()` | `mobile.SeshatListTargets` → `seshat.ListTargets` | none | **New endpoint** `GET /api/seshat/targets`. |
| `seshatListKnowledgeItems()` | `mobile.SeshatListKnowledgeItems` → `seshat.ListKnowledgeItems` | none | **New endpoint** `GET /api/seshat/items`. |
| `brainClassify(filePath:)` | `mobile.BrainClassify` → `brain.Classify` | none | **New endpoint** `POST /api/brain/classify` with body `{"file_path":"<str>"}` **OR** CLI one-shot. Classify is sub-second; new endpoint is leaner. |
| `brainClassifyBatch(paths:, workers:)` | `mobile.BrainClassifyBatch` → `brain.ClassifyBatch` | none | **New endpoint** `POST /api/brain/classify-batch` with body `{"paths":[<str>,...],"workers":<int>}`. |
| `brainModelInfo()` | `mobile.BrainModelInfo` → `brain.ModelInfo` | none | **New endpoint** `GET /api/brain/model`. |
| `brainInstallModel(modelPath:)` | `mobile.BrainInstallModel` → `brain.InstallModel` | none | **New endpoint** `POST /api/brain/install-model` with body `{"model_path":"<str>"}`. Model install can take seconds-to-minutes — consider runner/events path for v2. |
| `rtkFilter(rawOutput:, config:)` | `mobile.RtkFilter` → `rtk.Filter` | none | **CLI one-shot** `sirsi rtk filter --json` is a more honest fit — RTK is invoked at command boundaries, not as a persistent app feature. Skip a dashboard endpoint until a real Mac UX needs it. |
| `rtkDefaultConfig()` | `mobile.RtkDefaultConfig` → `rtk.DefaultConfig` | none | **CLI one-shot** `sirsi rtk config --default --json`. Same reasoning as `rtkFilter`. |
| `steleReadRecent(count:)` | `mobile.SteleReadRecent` → reads stele file | `GET /api/stele?limit=<n>` | **Adapter**. The iOS call returns a typed Swift array; dashboard returns `[]stele.Entry`. Same shape, no envelope. Mac bridge decodes directly. Just need to confirm `count` ↔ `limit` parameter rename. |
| `steleStats()` | `mobile.SteleStats` → `stele.Stats` | none | **New endpoint** `GET /api/stele/stats`. |
| `steleVerify()` | `mobile.SteleVerify` → `stele.Verify` | none | **New endpoint** `GET /api/stele/verify`. |
| `vaultStore(...)` | `mobile.VaultStore` → `vault.Store` | none — `/api/vault/{search,stats,prune}` cover the other operations | **New endpoint** `POST /api/vault/store`. |
| `vaultGet(...)` | `mobile.VaultGet` → `vault.Get` | none | **New endpoint** `GET /api/vault/get?key=<str>`. |
| `vaultSearch(q:, limit:)` | `mobile.VaultSearch` → `vault.Search` | `GET /api/vault/search?q=<str>&limit=<n>` | **1:1** (shape matches, no envelope mismatch beyond the global one). |
| `vaultStats()` | `mobile.VaultStats` → `vault.Stats` | `GET /api/vault/stats` | **Adapter**. Note the 200-OK-with-in-body-error pattern when vault is unavailable; the iOS bridge must treat presence of `error` as a soft-state signal. |
| `vaultPrune(olderThan:)` | `mobile.VaultPrune` → `vault.Prune` | `POST /api/vault/prune?older_than=<dur>` | **1:1**. |
| `horusParseDir(root:)` | `mobile.HorusParseDir` → `horus.GoParser.ParseDir` | `GET /api/horus/scan?path=<str>` | **Adapter**. Same `SymbolGraph` shape; `root` ↔ `path` parameter rename. |
| `horusFileOutline(root:, filePath:)` | `mobile.HorusFileOutline` → `horus.FileOutline` | none | **New endpoint** `GET /api/horus/outline?path=<str>&file=<str>`. |
| `horusContextFor(root:, symbolName:)` | `mobile.HorusContextFor` → `horus.ContextFor` | partial — `/api/horus/query?name=<str>` returns matching symbols, but not the surrounding context | **New endpoint** `GET /api/horus/context?path=<str>&name=<str>`. |
| `horusMatchSymbols(root:, pattern:)` | `mobile.HorusMatchSymbols` → `horus.Query.MatchSymbols` | `GET /api/horus/query?path=<str>&filter=<str>` | **Adapter**. Same shape; `pattern` ↔ `filter` parameter rename. |

## Tally

- **1:1 (no work):** 2 calls — `vaultSearch`, `vaultPrune`.
- **Adapter (parameter rename or in-body-error handling):** 6 calls — `kaHunt`, `steleReadRecent`, `vaultStats`, `horusParseDir`, `horusMatchSymbols`, plus the global envelope flatten (every call).
- **New endpoint required:** 19 calls — most of Thoth, Seba, Seshat, Brain, parts of Ka/Horus/Stele/Vault, optional Anubis categories.
- **CLI one-shot:** 2 calls — `rtkFilter`, `rtkDefaultConfig`.
- **AMBIGUOUS (decision needed):** 1 — `anubisScan` (new POST endpoint vs runner-events path).

## Implementation order suggestion (out of scope for this batch)

When the implementation batch begins, the rational order is:

1. Stats + ghosts + doctor + guard + horus/{scan,query,report} + vault/{search,stats,prune} — every endpoint already exists; only the envelope decision blocks them.
2. Stele endpoints — small additions to round out a working stele view.
3. Anubis (resolve AMBIGUOUS, ship as `POST /api/scan` for parity).
4. Thoth + Seba + Seshat + Brain — new endpoints, each thin.
5. RTK stays CLI-only.

This audit decides only the gap. Implementation order is a Phase-2 batch-two artifact.
