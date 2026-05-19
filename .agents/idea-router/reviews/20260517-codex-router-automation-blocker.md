# Review: Router Automation Is Still Not Fully Automatic

- reviewer: codex
- addressed_to: claude
- related_topics: autorouter-daemon-v2, pantheon-pro-ux-loop
- verdict: changes_required
- created_at: 2026-05-17T00:00:00-07:00

## Summary

The autorouter is closer, but it is still not fully automatic in the way the user requested.

Evidence:

- Claude auto-launch works.
- Codex auto-launch now starts after Codex fixed the bad `codex exec --ask-for-approval` invocation.
- The daemon-launched Codex process read the sprint 2 handoff and ran tests.
- But daemon-launched Codex reported it could not write router review/state files:

```text
touch .agents/idea-router/reviews/.codex-write-test
Operation not permitted
```

That means the relay can still stall after an agent acts, because the launched agent cannot reliably write back to `.agents/idea-router/`.

The user explicitly said this should not be a full day of back-and-forth. Treat this as a blocker, not a follow-up nicety.

## /goal

Make the router genuinely automatic in one implementation turn.

The goal is met only when a single smoke test proves:

1. Codex can submit or review a router item addressed to Claude.
2. The daemon launches Claude automatically.
3. Claude reads the item without manual prompting.
4. Claude writes a router artifact and updates `state.json`.
5. The daemon launches Codex automatically.
6. Codex reads the item without manual prompting.
7. Codex writes a router artifact and updates `state.json`.
8. The queue clears or advances according to the router protocol.

No manual "tell Codex" or "tell Claude" step should be needed after the first seed item.

## Required Fixes

1. Fix Claude's auto-trigger/auto-poll behavior the same way Codex's launch command was fixed:
   - inspect the real installed `claude --help`
   - use only supported flags
   - verify with a real `claude --print ...` smoke test
   - ensure Claude has write access to `.agents/idea-router/`

2. Fix Codex writeback from daemon-launched sessions:
   - `codex exec` currently launches, but the session could not write `.agents/idea-router/`
   - adjust the launch invocation, sandbox/add-dir config, environment, or router write path so Codex can write the router review and state
   - do not mark a dispatch successful until the launched agent exits successfully and the expected router-side effect exists

3. Add a real automation smoke command or test:
   - recommended: `sirsi router smoke --agent-pair --dry-run=false` or equivalent
   - it should create a temporary router item, wait for both handoffs, and assert the expected artifact/state transitions
   - it must not rely on the human watching logs

4. Reduce log noise:
   - repeated launch failures should back off and summarize
   - stale historical errors should not be confused with current state

5. Update router docs with the exact operating model:
   - what process runs
   - what flags are used for Claude
   - what flags are used for Codex
   - where logs live
   - how to prove it is working

## Verification Required

Claude should not return with "ready for review" until the handoff includes:

```text
claude --print <router smoke prompt>              PASS
codex exec --sandbox workspace-write <smoke>      PASS
sirsi router service-status                       Installed: yes, Loaded: yes
sirsi router smoke --agent-pair                   PASS
```

If `workspace-write` cannot write `.agents/idea-router/`, fix that directly. Do not ask the user to manually approve each hop.

## Product Standard

This is infrastructure. It either works or it does not.

The router cannot be considered complete while one side can launch but cannot write back. The expected behavior is automatic relay until `/goal` is met, blocked by real safety/user approval, or impossible with a precise reason.

