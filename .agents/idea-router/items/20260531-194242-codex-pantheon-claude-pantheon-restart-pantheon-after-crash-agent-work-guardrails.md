---
from: "codex-pantheon"
to: "claude-pantheon"
title: "Restart Pantheon after crash: agent work guardrails"
status: closed
opened: 2026-05-31T19:42:42Z
closed: 2026-05-31T19:50:38Z
---

## Instructions

---
id: 20260531-codex-pantheon-claude-pantheon-restart-pantheon-after-crash-agent-guardrails
author: codex-pantheon
addressed_to: claude-pantheon
topic: restart-pantheon-after-crash-agent-guardrails
repo: /Users/thekryptodragon/Development/sirsi-pantheon
agent_scope: repo-segmented
eta_for_review: 2026-05-31T18:30:00-04:00
next_check_at: 2026-05-31T18:45:00-04:00
estimated_duration: 2 hours
---

# Restart Pantheon After Crash: Agent Work Guardrails

## Context

While Codex was assessing Sirsi Nexus, an agent-initiated analysis path ballooned to roughly 135 GB of application memory and crashed the working environment. This is exactly the class of incident Pantheon is supposed to prevent: AI/tooling work should be resource-governed before it can harm the operator's machine.

The user wants this routed as implementation work for the Pantheon Claude agent:

- Think through the task and governance.
- Build and test the Pantheon-side capability.
- Submit back to Codex for quality review.
- Assess the actual Pantheon code first to determine how much already exists in current tools.

## Governance

This work is governed by ADR-006 / Rule A16: Pantheon tools MUST NOT make a bad situation worse.

Additional incident-specific governance:

1. No unbounded recursive analysis over `$HOME`, `~/Development`, or `.codex/sessions`.
2. No whole-file JSONL transcript ingestion. Session logs must be treated as hazardous unless read through bounded ranges or filtered tools.
3. No Python implementation for this feature unless an ADR explains why Go cannot do it. The crash path involved a Python analysis script; the first implementation should stay inside Pantheon's Go runtime.
4. Any command wrapper must enforce output limits before returning text to an agent.
5. Any process intervention must be reversible or dry-run by default. Killing remains opt-in and protected by existing Guard/Slay safety rules.
6. Tests must use injectable providers/mocks. Do not require real process killing, real memory pressure, or real `/Applications` edits.

## Existing Pantheon Primitives To Assess And Reuse

Codex verified these files exist and are relevant:

- `internal/yield/yield.go`: CPU load governance with `Check`, `ShouldYield`, and `WarnIfHeavy`.
- `internal/guard/doctor.go`: one-shot diagnostic with RAM pressure, swap, top memory consumers, crash logs, and Sirsi process checks.
- `internal/guard/audit.go`: process grouping and RSS accounting.
- `internal/guard/watchdog.go`: sustained CPU watchdog with bounded alert channel and self-backoff.
- `internal/guard/throttle.go`: reversible renice-based pressure relief.
- `internal/guard/slayer.go`: dry-run/default-safe process termination groups.
- `internal/rtk/rtk.go` and `cmd/sirsi/rtk.go`: output stripping/dedup/truncation.
- `internal/vault/vault.go` and `cmd/sirsi/vault.go`: context sandbox for large output.
- `internal/horus/`: structural code graph/outlines so agents can inspect code without full-file ingestion.
- `extensions/vscode/src/crashpadMonitor.ts`: Crashpad monitoring for IDE crash leading indicators.
- `cmd/sirsi/threadcmd.go`: router thread watch/heartbeat infrastructure.

Do not duplicate these capabilities. First produce an implementation note in the commit/artifact that says what was reused, what was extended, and what remains unbuilt.

## /plan

Implement the smallest Pantheon feature that would have prevented or sharply contained the Sirsi Nexus crash.

Recommended shape:

1. Add an "agent work safety" package or command layer that composes:
   - system preflight from `yield.Check` and `guard.Doctor`
   - process pressure facts from `guard.Audit`
   - output budget filtering through `rtk.Filter`
   - optional vault storage for oversized output
2. Add a CLI surface, such as one of:
   - `sirsi agent preflight`
   - `sirsi agent safe-run -- <command...>`
   - `sirsi guard agent`

   Choose the name that best matches existing CLI conventions.
3. The command must return a machine-readable JSON mode for agents and a concise human mode for the operator.
4. Add policy detection for known hazardous agent work:
   - broad scans of `$HOME` or `~/Development`
   - direct `.codex/sessions/*.jsonl` reads without a narrow range/filter
   - commands likely to emit unbounded output
   - Python scripts used for repo-wide analysis without explicit budget flags
5. Add tests for:
   - healthy preflight permits work
   - high CPU/RAM/swap causes warn or block according to severity
   - hazardous command patterns are blocked or require explicit override
   - output over budget is truncated or vaulted
   - no real system processes are killed in tests
6. Update docs:
   - ADR or design note for "Agent Work Safety Governor"
   - README or CLI compatibility note if a new command is added
   - Thoth memory/journal after meaningful implementation

## /goal

Claude must submit back to Codex with:

1. A router artifact addressed to `codex-pantheon`.
2. Summary of existing Pantheon code reused vs new code built.
3. Files changed.
4. Test commands and results.
5. Explicit answer to: "Would this have prevented or contained the 135 GB crash? If not fully, what remains?"

Codex quality gate:

- Review behavior, tests, and governance fit.
- Confirm the feature is integrated enough to be usable by future Codex/Claude sessions.
- Confirm no new unbounded scanner or context-ingestion path was introduced.

## Notes For Claude

This is Pantheon work, not Sirsi Nexus work. Do not edit `SirsiNexusApp` for this task.

Keep implementation narrow and dogfoodable. A working preflight/safe-run primitive is better than a broad dashboard-only design.

## Result

User redirected Codex to fix and unify directly. Codex implemented internal/agentguard, wired sirsi agent preflight and safe-run, added docs/AGENT_WORK_SAFETY.md, updated README and Thoth, and verified with go test ./internal/agentguard, go test ./cmd/sirsi, go build ./cmd/sirsi, plus JSON CLI smoke checks. Claude implementation handoff no longer needed for this item; Codex retains quality ownership.
