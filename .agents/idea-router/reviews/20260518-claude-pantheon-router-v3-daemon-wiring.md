# Review: Router v3 Daemon Wiring

- agent_id: claude-pantheon
- repo: /Users/thekryptodragon/Development/sirsi-pantheon
- topic: router-v3-multi-agent-queue
- addressed_to: codex-pantheon
- verdict: implemented
- date: 2026-05-18

## /plan

Wire v3 executor into the live daemon/runner dispatch path so that
`sirsi router run` and `sirsi router daemon` dispatch by registered
agent_id with writeback verification, not legacy NotifyAgent.

## /goal

The daemon dispatches work by agent_id, launches the registered command,
and records status transitions. Live CLI commands use the v3 executor.

## What Changed

### runner.go
- RunnerOptions.Executor field: when set, Tick() uses Executor.Dispatch()
  instead of legacy NotifyFunc
- PendingDispatches: reads BOTH dynamic Pending map AND legacy fields
  with deduplication (no double dispatch)
- Calls MigratePending on state for v3 inbox resolution

### router.go
- ValidateAgent(agentID): validates against registry OR legacy whitelist
- PollInbox/AckInbox: use ValidateAgent instead of ValidateAuthor
- Registered agent IDs (claude-pantheon, codex-nexus, etc.) now accepted

### daemon.go
- DaemonOptions.Executor field passed through to Runner

### routercmd.go
- `sirsi router run` and `sirsi router daemon` create v3 Executor from
  registry + work queue and pass it to Runner/Daemon
- Target validation accepts registered agent IDs, not just codex/claude/all
- `path/filepath` import added for routerRoot construction

## Verification

```
go build ./cmd/sirsi/: pass
go test ./internal/router/ -count=1: pass (all tests including 9 runner tests)
sirsi router inbox claude-pantheon: finds pending item
sirsi router run --once --dry-run --target claude-pantheon: dispatches correctly
```

## Tests

All 9 runner tests pass with the new dispatch path. Legacy tests continue
to work because PendingDispatches reads both dynamic and legacy fields.

## Failures/Blockers

None. The flaky TestDaemonFSNotifyDispatchesStateChange did not fail this run.

## Next Action for Codex

Review this wiring and verify:
1. `sirsi router run --once --dry-run` uses v3 executor (not legacy NotifyAgent)
2. Registered agent IDs work throughout the CLI
3. Close router-v3-multi-agent-queue if satisfied
