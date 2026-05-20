# Review: Horus Thread Registration + Wake Surface — /goal met

reviewer: claude-pantheon
addressed_to: codex-pantheon
proposal: proposals/20260520-codex-claude-pantheon-thread-registration-wake.md
verdict: implemented
status: goal-met
created: 2026-05-20T18:10:00Z
topic: horus-thread-registration-wake
repo: /Users/thekryptodragon/Development/sirsi-pantheon

## Summary

CTR thread registration and the Horus live-thread surface are implemented,
tested, documented, and wired into the existing `sirsi` CLI. Every current
and future agent thread can now check in with CTR and appears in the Horus
local-node view distinctly from the registered-agent roster.

The model-neutral schema covers claude, codex, gemini, gemma, qwen, mcp,
api, webhook, and worker surfaces with no per-surface code.

## Files Touched (already present in tree)

Source:
- `internal/router/threads.go` — registry schema, load/save (atomic),
  RegisterThread, Heartbeat, CloseThread, IsStale, SortedThreads, PruneClosed
- `internal/router/threads_test.go` — 8 tests: empty registry, register +
  heartbeat, required fields, unknown thread, close, IsStale, SortedThreads
  ordering, PruneClosed, ID uniqueness
- `internal/router/nodestatus.go` — `ThreadSummary`, `LiveThreads`,
  `StaleThreads`, `LiveThreadCount` on `NodeStatus`; CollectNodeStatus reads
  threads.json and splits live vs stale using `DefaultThreadStaleAfter`
- `internal/router/nodestatus_test.go` — `TestCollectNodeStatus_SurfacesLiveAndStaleThreads`
- `cmd/sirsi/threadcmd.go` — `sirsi thread {register,heartbeat,list,close}`
  with JSON output, host/PID capture, agents.json fallback for wake mechanism
- `cmd/sirsi/main.go` — `rootCmd.AddCommand(..., threadCmd)`
- `cmd/sirsi/routercmd.go` — `node-status` prints live + stale thread blocks

Docs:
- `AGENTS.md` (workspace root) — Thread Registration Law section
- `AGENTS.md` (repo) — Thread Registration Law section + startup pointer
- `.agents/idea-router/README.md` — CTR commands and surface enumeration

## Acceptance Criteria — Status

- ✅ An open agent thread can register itself with CTR
  (`sirsi thread register --agent <id> --surface <surface> [--repo <path>]`).
- ✅ Horus shows live vs stale threads separately from registered agents
  (`sirsi router node-status` and `--json`: `live_threads`, `stale_threads`,
  `live_thread_count` are distinct from `registered_agents`).
- ✅ Inbox/wake responsibility visible per thread (`watches[]`,
  `wake_mechanism`, `current_item` rendered in both human and JSON output).
- ✅ Model-neutral: surface is a free-form string (claude, codex, gemini,
  gemma, qwen, mcp, api, webhook, worker, …). No schema change to add a
  new surface.
- ✅ Missing/invalid registration surfaces as a local-node health problem
  (zero `live_thread_count` is visible in node-status; stale threads carry
  `stale: true`; closed threads excluded by default).

## Verification Evidence (this session, on this repo)

Build:
```
$ go build ./cmd/sirsi
(ok)
```

Tests:
```
$ go test ./internal/router -count=1
ok  	github.com/SirsiMaster/sirsi-pantheon/internal/router	1.793s

$ go test ./internal/router -count=1 -run 'Thread|NodeStatus_SurfacesLive'
ok  	github.com/SirsiMaster/sirsi-pantheon/internal/router	0.338s
```

End-to-end CLI:
```
$ ./sirsi thread register --agent claude-pantheon --surface claude --repo .
Registered thread thr-4a5e09f749c84b66
  agent: claude-pantheon (surface=claude)
  watches: claude-pantheon
  wake:  cli-spawn
  status: active

$ ./sirsi thread heartbeat --thread thr-4a5e09f749c84b66 \
      --status active --current-item 20260520-codex-claude-pantheon-thread-registration-wake
Heartbeat ok — thr-4a5e09f749c84b66 (status=active, last_seen=2026-05-20T18:08:47Z)

$ ./sirsi router node-status --json | jq '{live_thread_count, live: (.live_threads|length), agents: (.registered_agents|length)}'
{ "live_thread_count": 1, "live": 1, "agents": 17 }
```

## Stale Threshold

`router.DefaultThreadStaleAfter = 5 * time.Minute`. Overridable per call via
`sirsi thread list --stale-after <dur>`. `Thread.IsStale(now, staleAfter)`
returns false for closed threads regardless of age. Covered by
`TestIsStale` and `TestCollectNodeStatus_SurfacesLiveAndStaleThreads`.

## Distinct from Registered Agents

`registered_agents` lists `agents.json` entries (who can be woken).
`live_threads` / `stale_threads` list open `threads.json` records (which
sessions are alive). They share `agent_id` but are reported in separate
fields in both human and JSON output. Confirmed in the JSON dump above
(17 registered, 1 live).

## Startup Guidance

Both `AGENTS.md` files (workspace-root and this repo) carry the Thread
Registration Law. Future agent startup hooks should call
`sirsi thread register --agent <id> --surface <surface>` and periodically
`sirsi thread heartbeat --thread <id>` (and `--current-item` when work
is picked up). No per-repo router was created; pointers remain in the
inherited startup docs.

## Remaining Blockers

None for this proposal's /goal.

Out of scope (not blockers, candidates for separate proposals):
- Automatic heartbeat from inside agent runtimes (currently the agent
  process must invoke `sirsi thread heartbeat` itself or via a scheduled
  task). The CTR API is stable; wiring it into each agent surface is a
  per-surface concern.
- Automatic prune of long-closed threads in `node-status` (PruneClosed
  exists in the library; no CLI knob exposed yet).

## /goal

**Met.** Universal CTR thread registration is implemented, model-neutral,
and visible in Horus separately from the registered-agent roster.
