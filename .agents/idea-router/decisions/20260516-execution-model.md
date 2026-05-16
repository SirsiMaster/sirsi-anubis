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

1. **Either agent writes a `/plan`** with explicit `/goal` completion condition
2. **Both agents wrangle the /plan to the point of execution** — back-and-forth via router until the plan specifies the proper, elegant, long-term implementation. No shortcuts. No "let me just do this because it's easier."
3. User approves the finalized plan (or delegates standing approval)
4. Claude executes against the plan — implements, tests, commits, pushes
5. After each sprint/milestone, Claude submits work to router addressed to Codex
6. Codex reviews the sprint — approves or requests changes
7. Claude fixes and resubmits — relay continues autonomously
8. Loop until `/goal` is met, blocked by safety/user, or impossible with stated reason

## Non-Negotiable Rules

- **Measure twice, cut once.** Do not implement until the plan is properly wrangled.
- **No shortcuts.** Never negotiate toward a simpler implementation because it's faster. Build the proper solution or do not build at all.
- **No bypass.** Both agents must agree the plan is right before Claude writes a line of code.
- **Proper, elegant, long-term.** Every implementation must be the kind you'd ship to production and never revisit. Band-aids are governance failures.

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
