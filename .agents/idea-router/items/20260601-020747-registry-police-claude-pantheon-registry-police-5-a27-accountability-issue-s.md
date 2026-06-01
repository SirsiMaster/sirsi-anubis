---
from: "registry-police"
to: "claude-pantheon"
title: "Registry police: 5 A27 accountability issue(s)"
status: closed
opened: 2026-06-01T02:07:47Z
closed: 2026-06-01T02:09:57Z
---

## Instructions

# Registry Police Alarm — 2026-06-01T02:07:47Z

A27 two-tier accountability check found issues:

- **3 unmappable agent session(s)** — running agents launched outside any known repo (cwd=$HOME). They have no agent identity and no inbox. Operator must register them with an explicit repo, or relaunch from the repo dir.
- **2 registered-but-not-looping thread(s)** — registered in CTR but no recent heartbeat (A27 violation).

Run `sirsi thread discover` and `sirsi thread list` to inspect. Police is read-only/advisory; no process was killed or steered.

## Result

# Claude-Pantheon Triage — Registry Police A27 Alarm

**Item:** `20260601-020747-registry-police-claude-pantheon-registry-police-5-a27-accountability-issue-s`
**Responder:** claude-pantheon (thr-77348bd5999beca6)
**Responded at:** 2026-06-01T02:08Z
**Decision:** partially actioned + false-positive bug fixed

## Finding 1 — "2 registered-but-not-looping thread(s)" → FALSE POSITIVE (fixed)

The alarm's stale-loop count was a bug in `police/registry-police.sh`, not a real
A27 violation. The detector read `t.get("last_heartbeat") / t.get("lastHeartbeat")`
off each row, but `sirsi thread list --json` emits rows shaped as
`{idle_seconds, stale, thread:{... last_seen_at ...}}`. There is **no**
`last_heartbeat` field at any level, so `if not hb: n+=1` counted *every*
registered thread as not-looping. At alarm time there were 2 registered threads →
it reported 2.

Live verification: all 3 currently-registered threads report `stale: false`
(idle 9s / 39s) — they are alive and looping.

**Fix applied:** the detector now trusts the CLI's own `stale` boolean
(`t.get("stale") is True`) instead of reinventing heartbeat math on a guessed
field. Re-running the fixed logic against live data yields `stale-loop = 0 / 3`.

## Finding 2 — "3 unmappable agent session(s)" → REAL, operator-scoped

`sirsi thread discover` confirms pids 2412, 3110, 3887 running with cwd at
`$HOME` / non-repo dirs (one under `~/Documents/Codex/...`). These have no repo
identity and no inbox. Police is read-only/advisory and correctly did not steer
them. Resolution requires **operator action** (register with an explicit repo or
relaunch from the repo dir) — it is not a claude-pantheon code task and crosses no
repo I own.

## Net

- 1 of 2 reported issue classes was spurious (police bug) — now fixed so the
  alarm stops firing false A27 violations every run.
- The unmappable-session finding stands and is referred to the operator.
- No threads were killed, reniced, or steered.

Refs: PANTHEON_RULES.md Rule A27, Rule A26; `.agents/idea-router/police/registry-police.sh`
