# ADR-025: Thoth-Gated Exit + Resumable Thread Suspend

## Status
**Proposed** — June 1, 2026 (R3 of the always-on supervisor; companion to
ADR-024 which covered R1/R2/R4/R5). Pending codex review.

## Context
A27 binds a thread's life to `register → close`. ADR-024 made *birth* and
*liveness* correct. **Exit is still ungoverned.** R3 of the supervisor vision
requires: before a thread leaves, it MUST (a) capture memory to Thoth and
(b) update the router to a terminal-or-resumable status — never just vanish.

Today only **one of the three exit doors** is gated:

| Door | Native hook | Wired today? | Thread survives? |
| :--- | :--- | :--- | :--- |
| **compact** | `PreCompact` | **Yes** — `sirsi thoth sync` + `compact` | yes (same session) |
| **quit / exit** | `SessionEnd` (cannot block) | **No** | no |
| **`/clear`** | `SessionStart(clear)` | partial (inbox/health only) | yes (memory wiped, PID lives) |
| **hard kill** | none | n/a (OS reaper, ADR-022) | no |

The gaps: a **quit** loses unsynced memory and leaves an `active` record that the
reaper later marks `reaped` (truthful but lossy — the plans/context are gone). A
**`/clear`** wipes context and kills the `/loop` watcher mid-session, leaving the
thread registered-but-unwatched until SessionStart re-arms. And there is **no
resumable state** — `close` is terminal, so a thread that means "I'm pausing,
resume me later with my memory and open items" has no way to say so.

`sirsi thread` today has `register|heartbeat|list|close|prune|discover|scout` —
**no `suspend`, no `resume`.**

## Decision

### 1. New `suspended` thread state — resumable, carries memory + plans
Add a fourth status alongside `active|closed|reaped`: **`suspended`**. A
suspended record is **resumable** and carries a `suspend_payload`:

```json
{
  "status": "suspended",
  "suspend_payload": {
    "thoth_ref": "<memory file/commit captured at suspend>",
    "owned_open_items": ["<router item ids still addressed to this agent>"],
    "resume_prompt": "<one-line continuation, e.g. a NOTEBOOKS resume name>",
    "suspended_at": "<UTC>"
  }
}
```

`suspend` ≠ `close`: `close` is terminal (done, prune-eligible); `suspend` means
"paused, re-adoptable." Both are valid Thoth-gated exits; the default on a
non-terminal exit is **suspend** (most exits are pauses, not deaths).

### 2. Two verbs — `sirsi thread suspend` / `sirsi thread resume`
- `sirsi thread suspend --thread <id>` — flips `active → suspended`, snapshots
  the payload (calls `thoth sync` first so `thoth_ref` is fresh, then records
  owned open items + resume prompt). Idempotent.
- `sirsi thread resume --thread <id>` (or re-`register` with the same `agent_id`
  adopting a suspended record) — restores: re-surfaces `owned_open_items`, prints
  `resume_prompt`, re-arms the watcher (ADR-024 §3), flips `suspended → active`.

### 3. The exit handshake — one rule, three doors
**Before any exit, capture Thoth AND set status.** Mapped to the only hooks the
platform gives us:

- **compact** → `PreCompact` already syncs Thoth; thread stays `active` (compact
  is not an exit). No change beyond what's wired. ✓
- **quit/exit** → **new `SessionEnd` hook**: `sirsi thoth sync` then
  `sirsi thread suspend` (best-effort — `SessionEnd` cannot block, so it is *a*
  gate, not *the* gate).
- **`/clear`** → `SessionStart(clear)`: reconcile (below) + re-arm watcher.
- **hard kill** → no hook fires; OS-truth reaper (ADR-022) marks `reaped`;
  reconciliation heals on next register.

### 4. SessionStart reconciliation is the authoritative gate
Because `SessionEnd` cannot block and `/clear`/kill may skip it, **SessionStart
is where the gate is actually enforced** (it always fires). On every start it
detects a **dirty exit** — an `active` record for this agent whose heartbeat is
stale and which has no `suspended`/`closed` transition — and heals it:
`thoth sync` (retroactively capture from the still-present transcript), then mark
the prior record `suspended`, then register/adopt fresh. This makes the guarantee
**eventually-gated**: best-effort at exit, *guaranteed* at next start. Trust the
OS + the transcript, not the agent's good behavior (the ADR-022 principle).

### 5. Default-on, same off-switch
The `SessionEnd` hook and reconciliation are user-scope (ADR-024 §4) and honor
`SIRSI_SUPERVISOR=0` (skip managed suspend/reconcile; manual `suspend`/`resume`
verbs always work — off means "don't manage me," not "remove the capability").

## Neith's Triad (A22)

### Data Flow Architecture
```mermaid
flowchart TD
  A[active thread + /loop watcher] -->|compact| B[PreCompact: thoth sync] --> A
  A -->|quit| C[SessionEnd: thoth sync + thread suspend] --> S[(suspended<br/>+payload)]
  A -->|/clear| D[SessionStart clear: reconcile + re-arm] --> A
  A -->|hard kill| E[no hook] --> R[(reaped via OS reaper)]
  S -->|resume / re-register| A
  F[SessionStart any] -->|dirty-exit detected| G[thoth sync + mark suspended] --> S
  R -->|next register| F
```

### Recommended Implementation Order
1. `suspended` status + `suspend_payload` in `internal/router` thread model + `threads.json` (required; everything depends on it).
2. `sirsi thread suspend` verb (required).
3. `SessionEnd` hook → `thoth sync` + `suspend` (required; closes the quit door).
4. SessionStart reconciliation of dirty exits (required; the authoritative gate).
5. `sirsi thread resume` verb + re-register adoption (required for the round-trip).
6. `prune` honors `suspended` as non-terminal — never auto-pruned (required guard).

### Key Decision Points
| Question | Options | Recommendation |
| :--- | :--- | :--- |
| Default exit status on quit? | suspend / close | **suspend** — most exits are pauses; close stays explicit. |
| Enforce gate where SessionEnd can't block? | block (impossible) / best-effort+reconcile | **best-effort at exit + guaranteed at SessionStart** — eventually-gated. |
| Where does resumable state live? | new store / `threads.json` | **`threads.json` payload** — one store, reuses existing reaper/list. |
| Capture source when agent didn't sync? | give up / transcript | **transcript** — SessionStart retro-syncs from the `.jsonl` still on disk. |

## Acceptance tests (required before merge; owner: claude-pantheon)
- `thread suspend --thread X` flips active→suspended, payload has fresh `thoth_ref` + owned items; idempotent on repeat.
- `thread resume --thread X` restores active, re-surfaces owned items, prints resume_prompt, re-arms one watcher (ADR-024 idempotence).
- `SessionEnd` hook runs `thoth sync` then `thread suspend` (assert status + payload after a simulated quit).
- SessionStart with a stale `active` record + no suspend/close → reconciliation marks it `suspended` after a retro `thoth sync` (dirty-exit heal).
- `prune` removes `closed`/`reaped` but **never** `suspended`.
- `SIRSI_SUPERVISOR=0` skips managed SessionEnd suspend + reconciliation; manual `suspend`/`resume` still work.

## Consequences
- A thread can no longer vanish lossy: quit → suspended-with-memory; kill/clear →
  healed at next start. The reaper stops being the *only* exit truth.
- `suspended` gives multi-day workstreams (the NOTEBOOKS "resume name" pattern) a
  first-class router home — resume re-adopts memory + open items in one verb.
- Completes A27's lifecycle: `register → heartbeat → (suspend ⇄ resume)* → close`.

Refs: PANTHEON_RULES.md A27 (lifecycle), A26 (relay/owned items), A22 (Triad);
ADR-024 (R1/R2/R4/R5 companion), ADR-022 (OS-truth reaping), Thoth memory system.
Resolves R3 of the always-on supervisor.
