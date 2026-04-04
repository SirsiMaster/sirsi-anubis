# ADR-014: The Stele — Append-Only Hash-Chained Event Ledger

**Status:** Accepted
**Date:** 2026-04-03
**Deciders:** Cylton Collymore, Ra Session

---

## Context

The Pantheon governance loop requires multiple deities (Ra, Ma'at, Thoth, Seshat, Neith) and the Ra Command Center TUI to share state during multi-sprint autonomous agent deployments. The initial implementation used:

- Per-scope log files (human-readable, written by agents)
- `events.jsonl` (structured events, pushed by agents)
- `deployment.json` (metadata, written by Ra)
- PID files, exit files (process tracking)

This created scattered state, redundant I/O, polling overhead, and no verification that entries were authentic or untampered.

## Decision

Replace all scattered state files with a **single append-only hash-chained event ledger** called **The Stele**.

### What is a Stele?

In Egyptian archaeology, a stele (στήλη) is a stone slab inscribed with official decrees, laws, and records. Once carved, it cannot be altered. The Pantheon Stele is the digital equivalent: an append-only file where every deity inscribes its work, each entry chained to the previous by a cryptographic hash.

### Architecture

```
~/.config/ra/stele.jsonl     ← The single source of truth
```

**Write model**: Append-only. Each entry is a JSON line containing:
- `seq`: Monotonic sequence number
- `prev`: SHA-256 hash of the previous entry (hash chain)
- `deity`: Which deity wrote this entry (ra, maat, thoth, seshat, agent:<scope>)
- `type`: Event type (sprint_start, tool_use, governance, text, commit, etc.)
- `scope`: Which repo scope this relates to
- `data`: Event-specific payload
- `ts`: ISO-8601 timestamp
- `hash`: SHA-256 of the entire entry (excluding this field)

**Read model**: Memory-mapped. Every consumer `mmap`s the file and tracks its own read offset. On Apple Silicon unified memory, this is zero-copy — CPU, GPU, and ANE threads all see the same physical memory pages. No IPC, no sockets, no message passing.

### Why Not Webhooks/WebSockets/Polling?

| Approach | Problem |
|----------|---------|
| Webhooks | Requires a server, network stack, HTTP overhead |
| WebSockets | Bidirectional complexity, connection management |
| Polling | Wastes CPU cycles, stale data between intervals |
| Message queues | External dependency, serialization overhead |
| **Stele (mmap)** | Zero-copy reads, no server, no IPC, hardware-native |

On Apple Silicon with unified memory (M1/M2/M3/M4), an `mmap`'d file is backed by the same physical pages accessible to CPU, GPU (Metal), and ANE. A Metal compute shader or ANE inference thread can read the Stele directly without any data transfer. This is the optimal IPC mechanism for Pantheon's single-machine architecture.

### Hash Chain (Local Hashgraph)

Each entry's `prev` field contains the SHA-256 hash of the previous entry. This creates a verifiable chain:

```
Entry 0: { seq: 0, prev: "000...", deity: "ra", type: "deploy_start", ... }
Entry 1: { seq: 1, prev: sha256(entry_0), deity: "agent:assiduous", type: "sprint_start", ... }
Entry 2: { seq: 2, prev: sha256(entry_1), deity: "agent:assiduous", type: "tool_use", ... }
Entry 3: { seq: 3, prev: sha256(entry_2), deity: "maat", type: "governance", ... }
```

Any consumer can verify the chain by walking entries and checking hashes. If an entry is corrupted or injected, the chain breaks. This is not consensus (single writer per entry) but **integrity verification** — the Stele is trustworthy because it's self-proving.

### Who Writes, Who Reads

| Deity | Writes | Reads |
|-------|--------|-------|
| **Ra** | `deploy_start`, `deploy_end`, `kill` | Everything (orchestrator) |
| **Agent** | `sprint_start`, `sprint_end`, `tool_use`, `text`, `commit` | Own scope entries |
| **Ma'at** | `governance` (QA gate results) | Agent entries for the scope being checked |
| **Thoth** | `compact`, `journal_write`, `memory_update` | Everything (knowledge keeper) |
| **Seshat** | `scribe` (session summaries) | Agent + governance entries |
| **Neith** | `drift_check`, `scope_weave` | Everything (scope alignment) |
| **Command Center** | Nothing (read-only) | Everything (display) |

### Consumer Offsets

Each consumer tracks its own byte offset into the Stele:

```
~/.config/ra/offsets/
  command-center.offset    → 48392  (CC has read up to byte 48392)
  maat.offset              → 45000  (Ma'at has processed up to byte 45000)
  thoth.offset             → 48392  (Thoth is caught up)
```

On each read cycle, the consumer:
1. `mmap`s the Stele (or re-reads if the file grew)
2. Reads from its saved offset to EOF
3. Processes new entries
4. Updates its offset file

### File Rotation

When the Stele exceeds 10MB, Ra rotates it:
1. Rename `stele.jsonl` → `stele.YYYYMMDD-HHMMSS.jsonl`
2. Create new `stele.jsonl` with a genesis entry pointing to the archive
3. Reset all consumer offsets to 0

Archives are retained for Seshat to compile into long-term knowledge items.

## Consequences

**Positive:**
- Single source of truth for all Pantheon state
- Zero-copy reads on Apple Silicon (unified memory mmap)
- Verifiable integrity via hash chain
- No IPC overhead — no servers, sockets, or message passing
- Any future deity or tool can consume from the same Stele
- ANE/Metal threads can read directly from mapped memory

**Negative:**
- Append-only means the file grows (mitigated by rotation)
- Write contention if multiple agents append simultaneously (mitigated by OS-level atomic append on macOS with O_APPEND)
- Hash chain adds ~64 bytes per entry (negligible)

## Implementation

- **Package**: `internal/stele/`
- **Writer**: `stele.Append(deity, eventType, scope, data)` — appends a hash-chained entry
- **Reader**: `stele.NewReader(consumerName)` — returns an offset-tracking reader
- **Verifier**: `stele.Verify()` — walks the chain and reports breaks
- **Integration**: Replace events.jsonl, per-scope log writing, deployment.json state tracking
