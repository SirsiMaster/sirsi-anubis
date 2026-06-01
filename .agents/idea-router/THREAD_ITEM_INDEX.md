# Router Thread/Item Relationship Index

Generated: 2026-05-31T20:15:00-04:00 by codex-pantheon.
Updated: 2026-05-31T21:05:00-04:00 by codex-pantheon after new Claude writebacks.
Updated: 2026-05-31T21:18:00-04:00 by codex-pantheon after closing Codex reviews.

## Current Truth

- `sirsi thread list` reports no active registered threads.
- `threads.json` contains only closed/reaped Claude Pantheon thread registrations.
- `processes.json` is Pantheon's host process scout ledger. It records every visible PID from `sirsi thread scout`, including IDEs, terminals, agent CLIs, app helpers, system processes, and ordinary processes.
- Open router item ownership is therefore agent-level right now. Thread-level ownership starts when a Claude session registers and heartbeats `current_item`.
- `state.json.pending` has been reconciled to match open item frontmatter for the current Claude repo-agent inboxes.
- Terminal-visible Claude sessions may still be alive without appearing here when they are home-launched, unregistered, or unmappable. `sirsi thread discover` reported this exact condition: live sessions discovered, but unmappable because their cwd/env does not resolve to a registered repo agent.

## Ownership Rule

1. Router item owner agent = frontmatter `to`.
2. Owning thread = active `threads.json` entry where:
   - `agent_id` equals the item owner agent, and
   - `watches` includes that agent, and
   - `current_item` equals the item id once work begins.
3. If no such active thread exists, the item is `thread_unassigned` and must be claimed by the next registered thread for that agent.
4. Historical thread/lane claims remain useful provenance, but closed threads do not own new work.
5. Process visibility is broader than thread ownership: `processes.json` may know a PID exists even when `threads.json` correctly refuses to assign it router work.

## Process Scout

`sirsi thread scout` refreshes `.agents/idea-router/processes.json` with every host-visible PID that the current Pantheon process can see. The scout is read-only: it records PID, PPID, user, command, RSS, VSZ, CPU, host, first_seen, last_seen, status, and a coarse role.

Roles:

- `agent`: Claude, Codex, Gemini, Gemma, Qwen, and related helpers
- `ide`: VS Code, Cursor, Windsurf, Antigravity, Xcode, Zed
- `terminal`: Terminal, iTerm, Warp, shells, and terminal emulators
- `system`: kernel/launchd/window/server-style system processes
- `process`: everything else

Automatic sweep now runs:

- `sirsi thread discover --json` to register mappable repo-launched agent sessions
- `sirsi thread scout --json` to refresh host process awareness

Control remains guarded: Pantheon may observe broadly, but process intervention must still go through Guard/Throttle/Slay safety rules and explicit operator approval.

## Open Items By Agent

| Owner agent | Thread owner now | Repo | Open item ids |
| :--- | :--- | :--- | :--- |
| `claude-assiduous` | `thread_unassigned` | `/Users/thekryptodragon/Development/assiduous` | `20260522-024136-claude-pantheon-claude-assiduous-sirsi-router-ack-is-live-migration-helper-for-legacy-pending`; `20260522-024530-claude-pantheon-claude-assiduous-please-ack-adoption-of-caffeinate-contract-sirsi-router-ack-`; `20260522-claude-pantheon-route-assiduous-impl`; `20260526-163510-claude-pantheon-claude-assiduous-lean-af-cleanup-assiduous-3-pid-files-ignore-rules` |
| `claude-finalwishes` | `thread_unassigned` | `/Users/thekryptodragon/Development/FinalWishes` | `20260522-024136-claude-pantheon-claude-finalwishes-sirsi-router-ack-is-live-migration-helper-for-legacy-pending`; `20260522-024530-claude-pantheon-claude-finalwishes-please-ack-adoption-of-caffeinate-contract-sirsi-router-ack-`; `20260526-163510-claude-pantheon-claude-finalwishes-lean-af-cleanup-finalwishes-narrow-preserve-rag-legal-dirty-` |
| `claude-homebrew-tools` | `thread_unassigned` | `/Users/thekryptodragon/Development/homebrew-tools` | `20260522-024136-claude-pantheon-claude-homebrew-tools-sirsi-router-ack-is-live-migration-helper-for-legacy-pending`; `20260522-024530-claude-pantheon-claude-homebrew-tools-please-ack-adoption-of-caffeinate-contract-sirsi-router-ack-`; `20260522-claude-pantheon-route-homebrew-tools-impl`; `20260526-163510-claude-pantheon-claude-homebrew-tools-lean-af-cleanup-homebrew-tools-ds-store-ignore` |
| `claude-nexus` | `thread_unassigned` | `/Users/thekryptodragon/Development/SirsiNexusApp` | `20260522-024136-claude-pantheon-claude-nexus-sirsi-router-ack-is-live-migration-helper-for-legacy-pending`; `20260522-024530-claude-pantheon-claude-nexus-please-ack-adoption-of-caffeinate-contract-sirsi-router-ack-`; `20260522-claude-pantheon-route-nexus-impl`; `20260526-163510-claude-pantheon-claude-nexus-lean-af-cleanup-sirsinexusapp-codex-approved-ready-to-implem` |
| `claude-porch-and-alley` | `thread_unassigned` | `/Users/thekryptodragon/Development/porch-and-alley` | `20260522-024136-claude-pantheon-claude-porch-and-alley-sirsi-router-ack-is-live-migration-helper-for-legacy-pending`; `20260522-024530-claude-pantheon-claude-porch-and-alley-please-ack-adoption-of-caffeinate-contract-sirsi-router-ack-`; `20260522-claude-pantheon-route-porch-and-alley-impl`; `20260526-163510-claude-pantheon-claude-porch-and-alley-lean-af-cleanup-porch-and-alley-tsbuildinfo-ignore-rules` |
| `user` | `not_applicable` | `/Users/thekryptodragon/Development` | `20260522-claude-pantheon-user-dev-root-cleanup-decision` |

## Historical Pantheon Thread Provenance

| Thread id | Agent | Status | Lane / current item | Provenance |
| :--- | :--- | :--- | :--- | :--- |
| `thr-659f4c6e12bb2f32` | `claude-pantheon` | closed | Lane B, `pantheon-mac-native-cli-pivot`; `current_item=20260522-codex-pantheon-active-thread-coordination-locks` | Original Mac-native Lane B owner. |
| `thr-4990a8df4cbd1468` | `claude-pantheon` | closed / reaped | Lane B successor, `pantheon-mac-native-cli-pivot` | Identified as Pantheon development thread for Phase-0, Phase-1 audits, Phase-2 batch-1 docs, and ADR-018/TUI reopening handoff. |
| `thr-1ca491d095768e1a` | `claude-pantheon` | closed | `pantheon-interactive-surface-decision` | ADR-020 interactive surface reopening package. |
| `thr-7452fa9c16e656c9` | `claude-pantheon` | closed | Lane C, LEAN AF coordinator-only | Closed stale Pantheon coordinator items and TUI misroute/correction acknowledgements. |
| `thr-6c49114858a272f2` | `claude-pantheon` | closed | Lane C, LEAN AF coordinator status refresh | Wrote the 2026-05-31 coordinator refresh decision. |
| `thr-f582c02ec658042a` | `claude-pantheon` | closed | `current_item=lean-af-fanout-complete` | Earlier LEAN AF fanout/coordinator thread. |
| `thr-a441bbff379e62a9` | `claude-pantheon` | closed | `current_item=build-hook-heartbeat-ack-keepalive-for-all-future-agent-threads` | Build-hook heartbeat / non-Lane-B thread. |

## Current Gaps

- No open item is owned by a live registered Claude thread yet.
- `sirsi thread discover` / SessionStart registration remains the needed mechanism for converting agent-level ownership into thread-level ownership.
- When a repo Claude starts, it should heartbeat the exact item id it is working through `sirsi thread heartbeat --thread <thread_id> --current-item <item-id>`.
- `codex-pantheon` has no open review requests as of 2026-05-31T21:18:00-04:00. The dispatch guard, ADR-020 canon v2, and `sirsi thread discover` Phase 1 items were reviewed and closed.
