# ADR-022: CTR Liveness Is OS Truth, Not Heartbeat Recency

## Status
**Accepted** — June 1, 2026

## Context
The CTR thread registry (`internal/router/threads.go`) tracks live agent
sessions so Horus and the menubar can show which conversations are alive on a
workstation. It classified a thread's liveness from two weak signals — the
recorded `status` string and heartbeat recency (`last_seen_at`) — never from
the live OS process table. Three coupled defects followed, observed after a
reboot that left the registry at **1050 records** for ~2 live sessions:

1. **Revival (B1)** — a late heartbeat could flip a `closed` record back to
   `active`, so a session that had ended reappeared as live with a fresh
   `last_seen_at` while still carrying `last_error: reaped`.
2. **Zombie-blindness (B2)** — the read-time reaper used `kill -0` (signal 0)
   to test liveness. A **defunct** process (zombie, state `Z`) answers `kill -0`
   successfully because it still occupies a slot in the table until its parent
   reaps it. Zombies were therefore never reaped and presented as `active`
   forever.
3. **Volume** — registration was non-idempotent: every heartbeat/discover tick
   for an already-terminal session minted a brand-new record. A tight
   register → reap → register loop produced hundreds of tombstones.

The tombstone churn had a second-order cost: each registry write re-triggered
macOS Spotlight indexing (`mds_stores`/`mdworker`), contributing to the
indexing storm that helped lock the machine (see incident 2026-05-31).

This is a direct extension of Rule A27 (Heartbeat Loop Mandate): registration
must mean "alive and watching," and the registry must agree with `ps`.

## Decision
**A registered thread's liveness is decided against the live OS process table,
and terminal states are irreversible.**

1. **Terminal states.** Introduce `reaped` (PID confirmed gone/defunct by OS
   truth) and `stale-heartbeat` (PID alive but quiet) statuses, plus
   `ThreadStatus.IsTerminal()` (`closed` ∪ `reaped`). `Heartbeat` refuses to
   write to any terminal record — resuming requires a new registration (new
   thread ID). A terminal record can never be resurrected by a late heartbeat.
2. **OS-truth liveness primitive.** A new `internal/router/liveness*.go`
   distinguishes **alive / gone / defunct** by reading `ps -o stat=` (a leading
   `Z` is defunct), build-tagged for unix/windows, exposed as an **injectable**
   function pointer behind a `sync.RWMutex` per Rules A16/A21. `kill -0` is
   demoted to a fallback existence check only.
3. **Reaper retires gone *and* defunct.** `ReapDeadThreads(routerRoot, host)`
   sets dead PIDs to `reaped`. It is **host-scoped**: a thread recorded on a
   different host is never reaped, because we cannot observe a remote process
   table. Threads with no PID are unverifiable and never reaped.
4. **Idempotent registration.** `RegisterThread` reuses the existing live
   record for the same `(agent_id, pid)` instead of minting a new one. One live
   session → exactly one thread.
5. **Surfacing + hygiene.** `sirsi thread list` reaps on read (the read IS the
   event), prints an OS-truth integrity warning, and marks reaped threads `💀`.
   `CollectNodeStatus` annotates each thread with `os_state` so Horus/menubar
   can never render a dead PID as live. A new `sirsi thread prune --older-than`
   deletes terminal tombstones (0 = all).

## Alternatives Considered
1. **Keep `kill -0` only**: simplest, but it is precisely the B2 bug — it cannot
   distinguish a live process from a zombie. Rejected.
2. **Parse `/proc` directly**: accurate on Linux but absent on macOS (the
   primary workstation surface) and more code than one `ps` call. Rejected for
   portability and Rule 0 (minimal code).
3. **Periodic reaper daemon**: a background sweeper polling the registry.
   Rejected — violates the LEAN ethos and Rule A27's event-driven preference;
   reaping on read needs no daemon and no polling.
4. **Treat dead PIDs as `closed`** (the prior reaper's behavior): conflates
   "operator ended it" with "OS says it died," so the integrity surface cannot
   tell intentional shutdown from a crash. Rejected in favor of a distinct
   `reaped` state.

## Consequences
- **Positive**: the registry agrees with `ps`; dead and zombie sessions cannot
  masquerade as active; the registry self-heals on every read; tombstone churn
  (and its Spotlight-indexing amplification) is bounded by `prune`; liveness is
  unit-testable without spawning real zombies (injectable prober).
- **Negative**: every read shells out one `ps` per local thread (cheap at this
  scale, but non-zero); a new terminal status widens the state space callers
  must handle.
- **Risk**: `ps -o stat=` output format differs across platforms — mitigated by
  build tags and the `Z`-prefix check plus a `kill -0` existence fallback. A
  mis-detected live process as gone would wrongly reap a session; mitigated by
  the host scope, the PID>0 guard, and the requirement that re-registration
  (cheap) restores it.

## References
- PANTHEON_RULES.md Rule A27 (Heartbeat Loop Mandate), A26 (Idea Router), A21
  (Concurrency-Safe Injectable Mocks), A16 (Side Effect Injection)
- ADR-017 (Ra/Horus CTR Hypervisor), ADR-021 (Deities Not Single-Repo)
- Commit `ca6e343`; router item `20260601-024355...execute-fix-ctr-false-active`
- `internal/router/{threads,liveness,nodestatus}.go`, `cmd/sirsi/threadcmd.go`
