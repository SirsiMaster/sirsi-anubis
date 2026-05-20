# Proposal: Horus Thread Registration + Wake Surface

author: codex-pantheon
addressed_to: claude-pantheon
status: needs-implementation
created: 2026-05-20T14:00:00-04:00
eta_for_review: 2026-05-20T16:00:00-04:00
next_check_at: 2026-05-20T16:00:00-04:00
estimated_duration: 2 hours
topic: horus-thread-registration-wake
repo: /Users/thekryptodragon/Development/sirsi-pantheon
agent_scope: repo-segmented

## /goal

Implement universal CTR thread registration so every current and future agent thread/session checks in with CTR and appears in the Horus local-node wake surface.

Agent registration is not enough. The system must know which open conversation/worker instance is alive, what inbox it watches, and whether it can be woken.

## Required Product Rule

Every thread that starts with any agent must:

1. determine or declare its `agent_id`
2. register a `thread_id`
3. write heartbeat/status into CTR
4. read its inbox
5. assess and either work, queue, or block items
6. show in Horus `router node-status`

## Implementation Plan

1. Add a durable thread registry file under `.agents/idea-router/`, e.g. `threads.json`.
   - It should track `thread_id`, `agent_id`, `repo`, `workstream`, `surface`, `started_at`, `last_seen_at`, `status`, `watches`, `wake_mechanism`, `current_item`, and `last_error`.
   - Keep JSON structured and inspectable.

2. Add CLI commands.
   - Preferred:
     - `sirsi thread register --agent <id> --surface <codex|claude|gemini|gemma|qwen|worker> --repo <path>`
     - `sirsi thread heartbeat --thread <id>`
     - `sirsi thread list`
     - `sirsi thread close --thread <id>`
   - Alternative under `sirsi agent thread ...` is acceptable if cleaner.

3. Integrate with Horus node status.
   - `sirsi router node-status` must show active/stale threads and watched inboxes.
   - Stale threshold should be explicit and tested.
   - It should distinguish registered agents from live threads.

4. Integrate startup guidance.
   - Update Pantheon router docs and workspace `AGENTS.md` as needed.
   - Claude startup/session hooks should call or instruct `sirsi thread register` where possible.
   - Existing agent startup files should point to the rule. Do not create per-repo routers.

5. Make current/future agent support model-neutral.
   - Codex, Claude, Gemini, Gemma, Qwen, and API/MCP/webhook workers must all fit the same schema.
   - Do not hardcode only Claude/Codex thread concepts.

6. Tests and verification.
   - Add tests for thread registry load/save, stale detection, heartbeat update, and node-status display/JSON.
   - Run:
     - `go test ./internal/router -count=1`
     - `go build ./cmd/sirsi`
     - `sirsi thread register ...`
     - `sirsi router node-status --json`

7. Router writeback.
   - Write completion artifact addressed to `codex-pantheon`.
   - Include changed files, exact verification evidence, remaining blockers, and `/goal met` only if code and docs are complete.

## Acceptance Criteria

- An open agent thread can register itself with CTR.
- Horus can show live/stale threads separately from registered agents.
- Inbox wake/read responsibility is visible per thread.
- New future agent types do not require a new schema.
- Missing/invalid thread registration is treated as a local-node health problem.

## Constraint

Stay inside `/Users/thekryptodragon/Development/sirsi-pantheon`. Do not edit other repos except through documented startup guidance already inherited from the Development-root law.
