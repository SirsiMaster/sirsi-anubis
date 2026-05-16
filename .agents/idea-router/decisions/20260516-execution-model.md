# Decision: Autonomous Execution Model for All SirsiMaster Workstreams

deciders: user (Cylton), codex, claude
status: approved-by-user
date: 2026-05-16

## Final Recommendation

All SirsiMaster workstreams follow this execution loop:

```
Claude writes /plan → submits to router → Codex reviews/edits →
User approves → Claude executes → submits sprint to router →
Codex reviews → Claude fixes → repeat until /goal met
```

## Roles (Non-Negotiable)

| Role | Agent | Responsibility |
|------|-------|----------------|
| **Doer** | Claude | Writes plans, writes code, runs tests, commits, pushes, submits sprints |
| **Reviewer** | Codex | Edits plans, reviews code, requests changes, approves, sharpens /goal |
| **Approver** | User | Approves finalized plans, approves questionable changes only |

The user is NOT the message bus. The router is the shared contract.

## Execution Protocol

1. Claude writes a `/plan` with explicit `/goal` completion condition
2. Claude submits `/plan` to router addressed to Codex
3. Codex reviews, edits, and finalizes the `/plan` (adds requirements, sharpens /goal)
4. User approves the finalized plan (or delegates standing approval)
5. Claude executes against the plan — implements, tests, commits, pushes
6. After each sprint/milestone, Claude submits work to router addressed to Codex
7. Codex reviews the sprint — approves or requests changes
8. Claude fixes and resubmits — relay continues autonomously
9. Loop until `/goal` is met, blocked by safety/user, or impossible with stated reason

## Autorouter Is Critical Path

The router-runner-v1 (automatic trigger) is the mechanism that makes this truly autonomous. Until it exists:
- Agents MUST check `state.json` and pending items at session start
- Agents MUST continue the relay without waiting to be told "check the router"
- The pending inbox IS the trigger

Once autorouter v1 exists:
- A submission by Claude automatically wakes Codex for review
- A submission by Codex automatically wakes Claude for implementation
- The user only intervenes for approval gates

## Scope

This applies to ALL SirsiMaster repositories and workstreams:
- sirsi-pantheon
- SirsiNexusApp
- FinalWishes
- Assiduous
- Any future repos

## Enforcement

Per Rule A26, Ma'at treats unmandated cross-repo edits, missing `/plan`, missing `/goal`, or unclosed router handoffs as governance failures.

## Why This Is The Best Path

- Eliminates the user as message bus
- Creates auditable decision trail in git
- Both agents have clear, non-overlapping responsibilities
- Work continues autonomously until done
- Autorouter removes the last manual step
