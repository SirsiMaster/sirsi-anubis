# ADR-024: One Watcher Per Surface — Router-Prescribed Heartbeat

## Status
**Proposed** — June 1, 2026 (pending codex review, router item `162920`)

## Context
A27 (Heartbeat Loop Mandate) says every registered thread must run a watcher
from `register` → `close`. It did not say **how many**, and the answer drifted to
"as many as got built." On 2026-06-01 a single `claude-home` thread was kept
alive by up to **three independent mechanisms**, each authored in a different
session blind to the others:

| Mechanism | Spawned by | Heartbeat | Wakes Claude conversation? |
| :--- | :--- | :--- | :--- |
| `watch-router` fs-watcher | `sirsi thread register` (auto) | yes (FSEvents) | **No** |
| Python caffeinator | `.claude/hooks/router_inbox_check.py` | 60s loop | No |
| `/loop` Monitor | the agent, manually | per-tick | **Yes** |

This is precisely the failure R4 (Capability-First) exists to prevent: a
primitive rebuilt from scratch because its existence was invisible. Worse, only
**one** of the three (the `/loop` Monitor) can actually wake a Claude
conversation — the other two keep the *record* warm but never deliver an inbox
item to the agent (the wake-asymmetry: file drops notify Codex's poller, not
Claude). So the redundant mechanisms cost CPU, Spotlight `mds_stores` pressure
(the same write-amplification implicated in the 2026-06-01 lockup), and registry
churn — while adding zero wake coverage.

The root defect is that **arming was guessed per surface**. `register`
unconditionally spawned its own fs-watcher; the hook spawned a caffeinator; the
agent armed `/loop` — nobody owned the decision of *what the correct single
watcher for this surface is*.

## Decision

### 1. One watcher per surface — invariant
A registered thread runs **exactly one** watcher. The correct watcher is a
function of the **surface**, not of which session got there first:

| Surface | Canonical watcher | Heartbeat | Watches inbox | Wakes agent |
| :--- | :--- | :--- | :--- | :--- |
| `claude` | `/loop` + file Monitor on `items/` | per-tick (≤60s) | yes | yes |
| `codex` | app heartbeat (`ctr-thread-wake`) | native poll | yes | yes |
| `gemini` `gemma` `qwen` | surface-native loop, else `sirsi router daemon` | loop/daemon | yes | yes |
| `menubar` `tui` `vscode` `jetbrains` `cursor` `macapp` | native runloop ping | ≥60s, bounded | only if it acts on items | n/a (resident) |
| `mcp` `api` `webhook` `worker` | `sirsi router daemon` (or resident launch agent) | daemon | yes | n/a |

### 2. The router prescribes the watcher — register becomes a handshake
`sirsi thread register` **stops auto-spawning the fs-watcher.** Instead it
returns the prescribed watcher for the surface. A new `internal/router`
watcher-spec table (sourced once, consulted always — the R4 living inventory in
code) maps surface → `{type, mechanism, arm_instruction, heartbeat_interval}`.

`sirsi thread register --surface claude --json` returns, alongside `thread_id`:

```json
{
  "thread_id": "thr-…",
  "watcher": {
    "type": "loop-monitor",
    "mechanism": "/loop + Monitor on .agents/idea-router/items/",
    "arm_instruction": "Arm /loop watching items/ for `to: <agent>`; heartbeat each tick.",
    "heartbeat_interval_s": 60
  }
}
```

The thread's job is no longer "invent a watcher" but "arm the one the router
named." The router owns the mapping; the surface owns the arming. Re-register is
idempotent on `(agent_id, pid)` (A27) and returns the same spec.

### 3. The agent is *told* to arm — at register time (R2 enforcement)
The register handshake is the enforcement point. The SessionStart hook calls
`register --json` and **injects `watcher.arm_instruction` into the agent's
context**, so a Claude session is instructed to arm `/loop` the moment it
starts — closing the gap where R2 ("monitor armed") relied on the agent
remembering. A `Stop`-hook gate (`exit 2` until the surface's watcher is
detected alive) is the backstop for surfaces that ignore the instruction; it is
**off by default** and gated by `SIRSI_SUPERVISOR` (see §4).

### 4. Default-on, human off-switch
The supervisor is on by default via **user-scope** `~/.claude/settings.json`
hooks (not project-scope — the 2026-06-01 miss was caused by the hook living in
`sirsi-pantheon/.claude/`, so home-dir sessions never fired it). The single
off-switch is the env gate `SIRSI_SUPERVISOR=0`, honored by both the hook and
the `Stop`-gate.

## Consequences
- A thread can no longer accumulate redundant heartbeats: there is one watcher,
  named by the router, per surface. The caffeinator and the auto-spawned
  fs-watcher are **retired** for `claude` (the `/loop` Monitor subsumes both).
- `register` is no longer a side-effecting spawner; it is a pure handshake that
  *returns* what to do. Easier to test, no orphaned watcher processes when a
  thread is adopted under a new PID.
- Spotlight/`mds_stores` pressure from N redundant heartbeat loops drops to one
  bounded ping per live thread (ties to the write-amplification lockup fix).
- The surface→watcher table is the R4 capability inventory **in code** —
  consulted on every register, so no future session can rebuild a watcher
  without first being handed the canonical one.

## Deferred / follow-up
- Migrate `router_inbox_check.py` to user-scope and strip its caffeinator
  (superseded by the register handshake + `/loop`).
- `sirsi thread suspend` (resumable state carrying memory+plans) — A27 lifecycle
  completion, tracked separately.
- Wire the `sirsi router daemon` CLI verb for non-interactive surfaces (code
  exists in `internal/router/daemon.go`, no verb).

## Verification (to perform on acceptance)
- `register --surface claude --json` returns a `watcher` block and spawns **no**
  fs-watcher process (`/tmp/sirsi-router-watch-*.pid` absent for claude).
- A claude session shows exactly one heartbeat source; `sirsi thread list`
  reports fresh `last_seen` from the `/loop` tick alone.
- `register --surface menubar` prescribes the runloop ping at ≥60s; no inbox
  loop unless the surface acts on items.
- With `SIRSI_SUPERVISOR=0`, no watcher is prescribed/armed and the Stop-gate is
  inert.

Refs: PANTHEON_RULES.md A27 (heartbeat loop), A26 (router relay), A24 (autonomy),
A11/A19 (Spotlight write-amplification context); companion to ADR-022 (CTR
OS-truth liveness). Resolves the three-heartbeat accretion observed 2026-06-01.
