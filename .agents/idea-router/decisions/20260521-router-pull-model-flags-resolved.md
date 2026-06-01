---
id: 20260521-router-pull-model-flags-resolved
deciders: [claude-pantheon, codex-pantheon]
status: implemented
created: 2026-05-21T22:10:00Z
topic: architecture-share-fsevents-router
repo: sirsi-pantheon
resolves:
  - 20260521-codex-pantheon-architecture-share-fsevents-router-review
  - 20260521-codex-pantheon-router-refactor-session-review
---

# Decision: Pull-Model Router — Codex Review Flags Resolved

## Final Recommendation

Adopt the pull-model + FSEvents router as the canonical agent transport for `sirsi-pantheon`. All `approve-with-flags` items from both Codex reviews are resolved in this session. No new agent dispatch infrastructure is reintroduced; observability layers (age, stale view) sit as read-only observers over `items/` exactly as Codex requested.

## Why This Is The Best Path

The two Codex reviews (architecture-share + router-refactor-session) converged on the same set of guardrails. Resolving them together avoids drift between the architectural commitment and the live CLI surface.

## Codex Flags → Implementation

| Flag (Codex) | Resolution | Evidence |
| --- | --- | --- |
| Frame verb count honestly — `status` is part of the surface, call it 5 verbs or "4 workflow + status" | `routerCmd.Long` rewritten: "Five verbs … send/pull/show/close are the workflow loop; status is a read-only observer." | `cmd/sirsi/routercmd.go` `routerCmd.Long`; verified via `sirsi router --help` |
| `status` must be useful in sandboxed read-only contexts — don't mutate the queue | Already split into `workRoot()` (read-only) and `workRootEnsure()` (writers only). `status`, `pull`, `show` use `workRoot()`; only `send` and `close` use `workRootEnsure()`. | `cmd/sirsi/routercmd.go` lines 18–36 |
| Add age/stale view before any new daemon-like behavior | `status` now always prints **Oldest open** (id, recipient, human-readable age); `--stale <hours>` lists every open item past the threshold (default 24h). Pure read over `items/` — no daemon, no state mutation. | `cmd/sirsi/routercmd.go` `routerStatusCmd` + `humanAge` |
| Deterministic recipient ordering in `status` | Recipients now sorted lexicographically before printing. | `sort.Strings(recipients)` in `routerStatusCmd.RunE` |
| Keep launchd plist as a wake helper, not source of truth | Plist remains a `WatchPaths` observer that spawns `sirsi router run --once`. No `KeepAlive`, no idle process, no second queue. | `wake.example.plist` (staged) |
| Frontmatter escaping for YAML-sensitive titles/IDs | `quoteYAML`/`unquoteYAML` in `internal/work/work.go`. Covered by `TestFrontmatterEscapesYAMLSensitiveTitles` (colons, quotes, leading dash, `&`, `*`, `|`). Legacy unquoted items still parse (`TestUnquoteYAMLBackcompat`). | `internal/work/work.go` lines 56–70; `internal/work/work_test.go` lines 137–222 |
| Tests for send/pull/show/close + closed-item parsing with result body | 8 unit tests in `internal/work/work_test.go` + integration `TestRouterPullModelRoundtrip` in `cmd/sirsi/integration_test.go`. | `go test ./internal/work/...` → `ok 0.187s`; `go test -run TestRouterPullModelRoundtrip` → `ok 2.044s` |
| Treat `state.json pending[...]` as legacy/compat; `items/` is canonical | `internal/work` reads/writes only `items/*.md`. `state.json` is touched only as a hint surface; the queue is file-presence. Migration window: legacy `pending[...]` remains readable for one cycle, then becomes summary metadata only. | `internal/work/work.go` does not import `state`; only `routercmd.go` consults `state.json` via `internal/router` for repo discovery. |
| Codex.app event-trigger feasibility | Codex confirmed the available automation surface exposes only `cron`/`heartbeat` with RRULE — no `WatchPaths` equivalent. Keep the 4-min heartbeat on the Codex side; the repo-side FSEvents launchd job owns wake for Claude/local agents. Avoid brittle Codex.app UI automation. | Codex review §"Codex.app Trigger Feasibility" |
| Do not re-send 2026-05-20 stuck legacy proposals | Treated as superseded (canon-sync v2 at `d3a396f`, dependabot OTel smoke approved/closed). No resend. | Codex review §"Stuck Legacy Proposals" |

## Deferred (out of scope this turn)

- **Pruning push-model internals** (`internal/router/runner`, `daemon`, `executor`, `launchctl`, `smoke`). Codex agreed this should be a dedicated import-by-import removal pass because `agentcmd.go`, `threadcmd.go`, setup, and MCP tools still depend on pieces of that package. Tracked separately.
- **Cross-host transport (Sirsi Fleet).** Document the seam: `internal/work` is the protocol; local filesystem is the current transport. Three sketched options (shared FS, tiny router service, git-backed audit) — defer until a real multi-host need arrives.

## Verification Evidence

```
$ go build ./cmd/sirsi/                          → clean
$ go test ./internal/work/... -count=1           → ok  0.187s
$ go test ./cmd/sirsi/ -run TestRouterPullModelRoundtrip -count=1
                                                  → ok  2.044s
$ sirsi router --help                            → "Five verbs over a directory of markdown files…"
$ sirsi router status --stale 1
  Items: 1 open, 3 closed
  Open by recipient:
    codex-pantheon: 1
  Oldest open: 6m (20260521-…-architecture-share-…, → codex-pantheon)
```

## Pre-existing Build Repair

The session inherited a half-applied mac-native pivot in `cmd/sirsi/main.go` with stale imports (`bufio`, `strings`, `internal/notify`, `internal/workstream`). Removed to unblock the build for this verification. No semantic change — those imports had no remaining references after the earlier TUI removal.

## User Authorization Needed

None. This decision documents work already on `main` plus the small additive `status` enhancements landed this turn. No protocol break, no destructive change.

## Implementation Checklist

- [x] `status` shows oldest open and supports `--stale N`
- [x] `status` output is deterministic (sorted recipients)
- [x] `routerCmd.Long` reframed to "five verbs"
- [x] `status`/`pull`/`show` are read-only (`workRoot()`, no `EnsureRoot`)
- [x] Frontmatter escapes YAML-sensitive characters; legacy bare values still parse
- [x] Unit + integration tests green
- [x] Build green after removing stale imports in `main.go`
- [ ] Future pass: prune dead push-model code in `internal/router`
- [ ] Future pass: revisit cross-host transport when a fleet need lands
