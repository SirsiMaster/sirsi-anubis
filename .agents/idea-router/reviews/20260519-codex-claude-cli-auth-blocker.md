# Review: Claude CLI Auth Blocks Live Router Dispatch

reviewer: codex-pantheon
addressed_to: claude-pantheon
status: blocker
created: 2026-05-19T10:55:00-04:00
eta_for_review: 2026-05-19T11:30:00-04:00
next_check_at: 2026-05-19T11:30:00-04:00
topic: claude-cli-auth-for-router-dispatch

## /goal

Restore live router dispatch to Claude agents, or document the exact operator action required to authenticate the local Claude CLI so `sirsi router work --target claude-pantheon` can launch work without manual shuttling.

## Evidence

Codex attempted live dispatch:

```text
sirsi router work --target claude-pantheon
```

The router found runnable work but Claude CLI failed:

```text
Router work check: 1 runnable dispatches ready.
Dispatching to claude-pantheon for review 20260519-codex-ra-horus-canon-review-request-changes — Review: Ra/Horus CTR Hypervisor Canon Completion
  Dispatching to claude-pantheon (claude)...
Not logged in · Please run /login
  Failed: agent exited with code 1: exit status 1
  Warning: notification failed: dispatch to claude-pantheon failed: exit status 1
```

## Required Work

1. In an authenticated Claude/Pantheon environment, confirm whether the local `claude` CLI can be logged in non-interactively or must be fixed by the user running `/login`.
2. If this can be fixed by configuration, implement it in the Pantheon router/operator setup without storing secrets in git.
3. If user action is required, update Pantheon docs/operator output so `sirsi router work` or the new Horus local-node status surface clearly reports:
   - Claude CLI auth status
   - exact required action: `claude` → `/login`
   - which agent dispatches are blocked
4. Add this blocker into the Ra/Horus local-node implementation so Horus reports local auth failures as part of daemon/agent health.
5. Write back to `codex-pantheon` with:
   - what was fixed
   - whether live dispatch now works
   - exact verification command output
   - remaining manual action, if any

## Constraint

Do not ask the user to manually shuttle router items. The point of this workstream is to make the router surface the blocker and recover cleanly.
