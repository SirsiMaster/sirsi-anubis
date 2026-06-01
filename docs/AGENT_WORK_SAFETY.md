# Agent Work Safety Governor

Pantheon now treats AI agent execution as a first-class local resource risk. The goal is simple: agent work should be bounded before it can degrade the operator's Mac.

This layer unifies existing Pantheon primitives instead of replacing them:

- `internal/yield`: CPU load preflight.
- `internal/guard`: RAM, swap, process, crash, and Sirsi process diagnostics.
- `internal/rtk`: output deduplication and truncation.
- `internal/vault`: context sandbox for future large-output storage.
- `internal/horus`: structural code inspection without full-file ingestion.

## Commands

```bash
sirsi agent preflight
sirsi agent preflight rg .codex/sessions/example.jsonl
sirsi agent preflight -- rg --files internal/agentguard
sirsi agent safe-run -- rg --files internal/agentguard
```

`preflight` returns a verdict:

- `allow`: no safety issue detected.
- `warn`: work may proceed, but Pantheon found elevated system or output risk.
- `block`: the command matches a high-risk agent pattern or the machine is in a critical state.

`safe-run` executes a command only after preflight. It captures stdout/stderr through a fixed output budget and applies RTK filtering before printing or returning JSON.

## Blocked Patterns

The first policy pass targets the crash pattern that motivated this feature:

- unbounded scans over `$HOME` or `~/Development`
- direct reads of `.codex/sessions/*.jsonl` with `cat`, `rg`, `grep`, or Python
- Python-based repo-wide or transcript-wide analysis without explicit budgets

These are not banned forever. They must be narrowed, routed through bounded Pantheon tools, or run with explicit override and output budgets.

## Why This Exists

The Sirsi Nexus crash reached roughly 135 GB of application memory while an agent was assessing code and transcripts. Pantheon already had Guard, Yield, RTK, Vault, and Horus, but they were separate tools. The agent safety governor is the missing front door that composes them before work begins.
