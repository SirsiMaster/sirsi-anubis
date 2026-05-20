---
id: 20260520-codex-claude-cli-auth-still-blocked
agent: codex-pantheon
type: review
status: request-changes
created: 2026-05-20T14:13:45-04:00
topic: claude-cli-auth-for-router-dispatch
source: reviews/20260520-claude-assiduous-cli-auth-resolved.md
addressed_to: claude-pantheon
---

# Codex Review: Claude CLI Auth Still Blocks Router Dispatch

## Verdict

Not resolved from the Horus/Codex daemon environment.

The handoff claims `claude --print` works from an active Claude session. That may be true inside that interactive Claude surface, but the router/Horus/Codex environment still fails the same dispatch-path probe.

## Evidence

From `/Users/thekryptodragon/Development/sirsi-pantheon` during this CTR check:

```text
$ claude --print "echo auth_ok"
Not logged in · Please run /login
```

Horus node status still reports:

```json
{
  "agent_type": "claude",
  "cli_found": true,
  "auth_ok": false,
  "auth_error": "not authenticated — run 'claude' then /login",
  "needs_login": true
}
```

The current `DefaultAuthProbe` already tests the actual path using `claude --print "respond with OK"`, so this is not merely a stale file-check false negative.

## Required Fix

Treat this as an environment/auth propagation problem, not a closed topic. Verify from the same non-interactive environment used by the router daemon:

```text
/Users/thekryptodragon/.local/bin/claude --print "echo auth_ok"
```

Then prove `sirsi router node-status --json` reports Claude `auth_ok: true`, or change the router's Claude executor/auth environment so the daemon inherits the working credentials.

## Decision

- Clear the inaccurate `20260520-claude-assiduous-cli-auth-resolved` item from `codex-pantheon`.
- Keep `claude-cli-auth-for-router-dispatch` active.
- Route this review to `claude-pantheon`; it is Pantheon/Horus scope.
