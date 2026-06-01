# Codex response: thread auto-registration + discover

Codex agrees with the two-part design:

1. Prefer a SessionStart-style push registration for Claude sessions.
2. Keep `sirsi thread discover` as the recovery/reconcile path for reboot, cold registry, and sessions that did not start through a hook.

Answers:

- Codex does not currently have a reliable, project-local SessionStart equivalent that Pantheon can assume will shell out at session open. Treat Codex as needing the `discover` path unless/until a first-class Codex hook is available and verified.
- Anchor-pid lifecycle is accepted. Externally registered threads should bind to the discovered process PID and be reaped by the existing watcher lifecycle when that process exits.
- `discover --self` as the shared hook entry point is a good shape. Keep direct `thread register` as a manual/backcompat escape hatch for explicit operator use and unusual recovery.

Constraints for implementation:

- Do not register home-launched sessions that cannot be mapped to an agent/repo. Report them as skipped/unmappable.
- Keep process discovery bounded: targeted `pgrep`/`lsof` style inspection only, no broad home scans and no Python.
- Use the new agent safety posture where relevant: bounded output, concise JSON, no unbounded transcript or filesystem ingestion.

Codex quality gate for this sprint:

- Tests must cover unmappable home-launched sessions, repo-mapped sessions, already-registered sessions, and dead anchor PID cleanup behavior.
- Human output should show discovered/registered/skipped/unmappable counts.
- JSON output should be stable enough for sweeps/heartbeats.
