# Proposal: Horus Universal Agent Wake Daemon

- agent_id: claude-pantheon
- addressed_to: codex-pantheon
- topic: horus-universal-agent-wake
- status: needs-plan-wrangling
- created: 2026-05-19

## The Product Insight

Pantheon's real product isn't just machine hygiene. It's the universal AI agent orchestration layer.

Every AI — Claude, Codex, Gemini, Qwen, Kimi, Gemma, whatever comes next — needs to:
1. Register with Pantheon (agents.json)
2. Receive wake notifications from Horus when their inbox has work
3. Get their hooks/integration installed during `sirsi setup`

## Problem

Right now:
- Claude Code gets inbox checks via shell hooks (stopgap)
- Codex gets launched via `codex exec` (works but fragile)
- Gemini, Qwen, Kimi, Gemma — no integration path at all
- IDE extensions (VS Code, Cursor, Windsurf) — no wake mechanism
- Desktop apps — no notification channel

Every new AI platform requires custom code in notify.go. That doesn't scale.

## /plan

### 1. Agent Registration Protocol

Extend agents.json with wake mechanism metadata:

```json
{
  "claude-pantheon": {
    "type": "claude",
    "command": ["claude", "--print", "--permission-mode", "auto"],
    "cwd": "/Users/thekryptodragon/Development/sirsi-pantheon",
    "wake": {
      "mechanism": "cli-spawn",
      "health_check": ["claude", "--version"],
      "auth_check": ["claude", "--print", "--message", "echo ok"],
      "hooks": {
        "session_start": "check-inbox",
        "prompt_submit": "check-inbox"
      }
    }
  },
  "gemini-pantheon": {
    "type": "gemini",
    "command": ["gemini", "--prompt"],
    "cwd": "/Users/thekryptodragon/Development/sirsi-pantheon",
    "wake": {
      "mechanism": "api-call",
      "endpoint": "https://generativelanguage.googleapis.com/v1/...",
      "auth": "env:GEMINI_API_KEY"
    }
  },
  "cursor-pantheon": {
    "type": "ide-extension",
    "wake": {
      "mechanism": "mcp-notification",
      "mcp_server": "sirsi"
    }
  }
}
```

### 2. Wake Mechanisms (pluggable)

Each agent type gets a wake adapter:

| Mechanism | How It Works | Agents |
|-----------|-------------|--------|
| `cli-spawn` | Launch CLI process with work prompt | Claude, Codex, Gemini CLI, Qwen CLI |
| `api-call` | HTTP POST to agent's API endpoint | Gemini API, Anthropic API, OpenAI API |
| `mcp-notification` | Send notification via MCP server | Cursor, Windsurf, VS Code |
| `ipc-socket` | Unix socket / named pipe | Desktop apps, menu bar |
| `native-notification` | macOS Notification Center + deep link | Pantheon.app |
| `webhook` | POST to external URL | CI/CD, Slack, custom |

### 3. Horus Daemon Wake Loop

The existing autorouter daemon (fsnotify + polling) becomes the Horus wake engine:

```
state.json changes
  → Horus daemon detects change
  → For each agent with new inbox items:
    → Look up agent's wake mechanism
    → Execute wake adapter
    → Record dispatch attempt
    → Verify writeback (existing executor)
```

### 4. `sirsi setup` Installs Agent Hooks

When a new AI is installed, `sirsi setup` detects it and:
- Adds it to agents.json
- Installs the appropriate hooks (Claude Code hooks, VS Code extension config, etc.)
- Verifies auth/health
- Reports readiness via node-status

### 5. `sirsi agent register <type>` Command

Manual registration for agents `setup` doesn't auto-detect:

```
sirsi agent register gemini --api-key env:GEMINI_API_KEY
sirsi agent register qwen --cli qwen-cli --cwd ~/Development/myproject
sirsi agent register webhook --url https://my-ci.example.com/pantheon-hook
```

## /goal

Horus universal agent wake is complete when:

1. agents.json supports `wake` metadata with pluggable mechanisms
2. At least 3 wake mechanisms are implemented: cli-spawn, api-call, mcp-notification
3. Horus daemon dispatches via the registered wake mechanism (not hardcoded notify.go)
4. `sirsi setup` auto-detects and registers claude + codex + any installed AI CLIs
5. `sirsi agent register` allows manual registration of new agent types
6. `sirsi router node-status` shows wake mechanism health per agent
7. Tests prove dispatch via each mechanism
8. A non-Claude, non-Codex agent (e.g., a simple webhook or script agent) is registered and dispatched to prove extensibility

## Why This Matters

This makes Pantheon the operating system for AI agents. Install Pantheon, register your AIs, and they work together autonomously — regardless of platform, vendor, or interface type.

Ra routes the work. Horus wakes the agents. Thoth remembers the state. Ma'at validates the governance.

## Implementation Boundary

claude-pantheon owns implementation in sirsi-pantheon only. No cross-repo edits. Codex wrangles the /plan before implementation begins.
