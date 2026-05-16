# Proposal: Router Runner + CLI Spinners

author: claude
status: needs-review
created: 2026-05-16

## /plan

### Stream A: Router Runner (`sirsi router` commands)

Own: `cmd/sirsi/router.go` (new), `internal/router/`

Implement:
1. `sirsi router status` — one-shot: show pending items for each agent, active topics, last read times
2. `sirsi router watch` — poll loop: print pending items every 10s, exit on Ctrl+C
3. `sirsi router inbox <agent>` — show and optionally ack inbox items

Files to create/modify:
- `cmd/sirsi/router.go` (new cobra command)
- Wire into `cmd/sirsi/main.go` init()

No changes to `internal/router/` — the package already has PollInbox, AckInbox, ReadState.

### Stream B: CLI Progress Spinners

Own: `cmd/sirsi/anubis.go`, `cmd/sirsi/main.go`, `internal/output/terminal.go`

Implement:
1. Add `Spinner(label string) func()` to output package — starts a spinner, returns a stop function
2. Wrap long CLI operations (scan, purge, installer, analyze, ghosts, diagnose) with spinner start/stop
3. Spinner suppressed when `--json` or `--quiet` or inside TUI

Files to modify:
- `internal/output/terminal.go` — add Spinner function
- `cmd/sirsi/anubis.go` — wrap scan, ghosts
- `cmd/sirsi/main.go` — wrap purge, installer, analyze

### /goal

Stream A complete when: `sirsi router status` prints inbox state, `sirsi router watch` polls and prints, tests pass.
Stream B complete when: `sirsi scan` shows a spinner during the scan phase in CLI mode, suppressed in JSON/TUI mode.

## Risks

- Router watch adds a long-running polling loop — must handle Ctrl+C cleanly
- Spinner must not interfere with JSON output or TUI rendering

## Tests / Verification

- Router: test that status command prints without error
- Spinner: verify suppression in JSON mode, verify stop clears the line
