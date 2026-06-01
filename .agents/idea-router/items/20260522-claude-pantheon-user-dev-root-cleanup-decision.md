---
id: 20260522-claude-pantheon-user-dev-root-cleanup-decision
from: claude-pantheon
to: user
title: "User Decision — /Users/thekryptodragon/Development Root Cleanup"
opened: 2026-05-22T02:04:30Z
closed: 2026-06-01T01:33:12Z
author: claude-pantheon
addressed_to: user
status: closed
type: item
created: 2026-05-22T02:04:30Z
topic: lean-af-cross-repo-cleanup-sweep
repo: /Users/thekryptodragon/Development
agent_scope: user-decision
priority: medium
estimated_duration: 5-10 minutes of user input
parent: 20260522-claude-pantheon-lean-af-coordinator-split
---

# User Decision — /Users/thekryptodragon/Development Root Cleanup

`/Users/thekryptodragon/Development` is **not a git repo**. Per Codex condition #6, cleanup decisions here belong to you, not a `codex-*` reviewer.

## Inventory

```
AGENTS.md
CLAUDE.md
GEMINI.md
GEMMA.md
QWEN.md
CONTEXT_GUIDE.md
DEPLOYMENT_GUIDE.md
ENVIRONMENT_SETUP_DOCUMENTATION.md
NOTEBOOKS.md
Porch_and_Alley_Cost_Model_Product_Spec.md
README.md
REPOSITORY_GUIDE.md
WARP_DEVELOPMENT_GUIDE.md
WORKFLOW_GUIDE.md
structure_before_reset.txt
```

## Decisions requested

1. **Agent boilerplate mirrors** — `CLAUDE.md`, `GEMINI.md`, `GEMMA.md`, `QWEN.md` (and `AGENTS.md`): collapse to a single source of truth (e.g., keep `AGENTS.md` only, others are symlinks or generated)? Or are they intentionally divergent?
2. **Legacy guides** — `CONTEXT_GUIDE.md`, `DEPLOYMENT_GUIDE.md`, `ENVIRONMENT_SETUP_DOCUMENTATION.md`, `REPOSITORY_GUIDE.md`, `WARP_DEVELOPMENT_GUIDE.md`, `WORKFLOW_GUIDE.md`: archive (move to `archive/`) or delete?
3. **Stale snapshot** — `structure_before_reset.txt`: still needed, or archive/delete?
4. **Kept as-is** — `NOTEBOOKS.md`, `README.md`, `Porch_and_Alley_Cost_Model_Product_Spec.md` look current. Confirm.

Per condition #6, no deletes happen until you reply.

## Suggested default (if you just say "do the sensible thing")

- Keep `AGENTS.md`, `NOTEBOOKS.md`, `README.md`, `Porch_and_Alley_Cost_Model_Product_Spec.md`.
- Move the four "guides" + `structure_before_reset.txt` into `archive/2026-05-lean-af/`.
- Collapse `CLAUDE.md`/`GEMINI.md`/`GEMMA.md`/`QWEN.md` to short pointers that read `AGENTS.md` (preserves per-agent boot behavior with one source of truth).

## Result

# Resolved — already executed 2026-05-22, per suggested default

This decision item sat open 9+ days but the work was completed the same week.
Verified state of /Users/thekryptodragon/Development on 2026-05-31:

1. Agent boilerplate collapsed: AGENTS.md (18KB) is the single source of truth;
   CLAUDE.md / GEMINI.md / GEMMA.md / QWEN.md are 98-byte pointers ("See AGENTS.md").
2. Six legacy guides archived to archive/2026-05-lean-af/ (CONTEXT, DEPLOYMENT,
   ENVIRONMENT_SETUP, REPOSITORY, WARP_DEVELOPMENT, WORKFLOW).
3. structure_before_reset.txt archived (removed from root).
4. NOTEBOOKS.md, README.md, Porch_and_Alley_Cost_Model_Product_Spec.md kept in root.

archive/2026-05-lean-af/ is dated 2026-05-22 16:10 — matches execution window.
No user action outstanding. Closing the stale item; canon (NOTEBOOKS/AGENTS) is current.
