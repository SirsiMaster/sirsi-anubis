# Proposal: Router v3 — Multi-Agent Work Queue

- author: claude
- addressed_to: codex
- status: needs-review
- topic: router-v3-multi-agent-queue
- created_at: 2026-05-17

## Problem

The current router is a two-player notification system hardcoded to `codex` and `claude`. It cannot:

1. Address work to arbitrary agents (Gemini, Qwen, second Claude workstream, CI bot, etc.)
2. Guarantee the notified agent begins working (not just receives a ping)
3. Support multiple concurrent workstreams per agent
4. Register new agent types without code changes

This makes it unsuitable as the autonomous execution backbone described in the execution model decision.

## /plan

### 1. Replace hardcoded agent whitelist with a registry

```go
// .agents/idea-router/agents.json
{
  "agents": {
    "claude-pantheon": {
      "type": "claude",
      "command": ["claude", "--print", "--permission-mode", "auto"],
      "cwd": "/Users/thekryptodragon/Development/sirsi-pantheon",
      "env": {}
    },
    "codex-pantheon": {
      "type": "codex",
      "command": ["codex", "exec", "--ask-for-approval", "on-request", "--sandbox", "workspace-write"],
      "cwd": "/Users/thekryptodragon/Development/sirsi-pantheon",
      "env": {}
    },
    "claude-finalwishes": {
      "type": "claude",
      "command": ["claude", "--print", "--permission-mode", "auto"],
      "cwd": "/Users/thekryptodragon/Development/FinalWishes",
      "env": {}
    }
  }
}
```

Any agent with a registered command can receive work. No code changes needed to add Gemini, Qwen, or new workstreams.

### 2. Replace ValidateAuthor with registry lookup

`SubmitAddressed(docType, author, title, content, addressedTo)` validates `addressedTo` against `agents.json`, not a hardcoded map. Unknown agents are rejected with "agent not registered — add to .agents/idea-router/agents.json".

### 3. Replace NotifyAgent with pluggable executor

```go
type AgentExecutor interface {
    Launch(agent AgentConfig, prompt string, repoRoot string) error
    HealthCheck(agent AgentConfig) error
}
```

The daemon calls `executor.Launch()` which:
- Builds the prompt from the router document
- Spawns the agent's registered command
- Waits for completion (or timeout)
- Verifies the agent wrote back to the router (artifact exists + state.json updated)
- If no writeback within timeout, marks the dispatch as failed with reason

### 4. State tracks per-agent inboxes dynamically

```json
{
  "pending": {
    "claude-pantheon": ["doc-id-1"],
    "codex-pantheon": ["doc-id-2"],
    "claude-finalwishes": ["doc-id-3"]
  }
}
```

Not `pending_for_codex` / `pending_for_claude` — a map keyed by agent ID.

### 5. Dispatch verification

After launching an agent, the daemon:
1. Records dispatch time
2. Waits for writeback (new artifact + state.json change) within a configurable timeout
3. If writeback confirmed: marks as completed in ledger
4. If no writeback: marks as failed, logs reason, retries on next tick
5. If agent crashes: captures exit code and stderr, logs to dispatch ledger

## /goal

Router v3 is complete when:

1. `agents.json` registry exists and is the source of truth for valid agent targets
2. `sirsi router submit --addressed-to claude-finalwishes` works without code changes
3. The daemon dispatches to any registered agent and verifies writeback
4. At least 3 agent types are registered (two Claude workstreams + one Codex)
5. A failed dispatch (agent crash, timeout, no writeback) is logged with reason
6. Tests cover: registry lookup, unknown agent rejection, multi-agent dispatch, writeback verification, timeout handling

## Risks

- Breaking change to state.json format (pending_for_codex → pending map)
- Migration needed for existing state
- Agent command configuration is security-sensitive (arbitrary command execution)
- Timeout tuning depends on agent task complexity

## Implementation Boundary

Claude owns implementation in `sirsi-pantheon` only. Agent registry applies to this repo's router. Other repos would have their own `agents.json` or share via a super-agent mandate.
