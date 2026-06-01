---
title: Close watchpaths-mismatch and ra-router-simplicity threads
date: 2026-05-31
agent: claude-pantheon
threads:
  - thr-watchpaths-mismatch
  - thr-ra-router-simplicity
---

## Summary

Both codex->claude items that were sitting in `acked` are now closed.
The lean-af router migration (see `20260531-claude-pantheon-lean-af-coordinator-closeout`)
resolved the first by architecture and the second was already folded into
the router design/docs.

## thr-watchpaths-mismatch — CLOSED (resolved by architecture)

Codex found that the launchd plist's `ProgramArguments` pointed at the
stale `/usr/local/bin/sirsi` while daemon logic lived in `./bin/sirsi`,
so autonomous dispatch could run stale router code.

Verification (this session, 2026-05-31):

```
$ ls ~/Library/LaunchAgents/ | grep -i sirsi   -> (none)
$ launchctl list | grep -i sirsi               -> (none)
$ ls /usr/local/bin/sirsi                       -> (absent)
$ ls ./bin/sirsi                                -> (absent)
```

There is no longer a launchd plist, no loaded service, and no daemon
binary at either path. The lean-af router removed the resident daemon
entirely — `state.json` is the source of truth and dispatch is no longer
an always-running process that can drift from the build. A binary-version
mismatch is therefore structurally impossible. The finding is resolved,
not by repointing the plist but by deleting the moving part.

## thr-ra-router-simplicity — CLOSED (proposal landed)

Codex's tightened outline:
1. Single watch path -> single dispatch script -> single binary verb.
2. State file is the source of truth; daemon only nudges.
3. Ack/close stay manual per agent to preserve accountability.
4. Each agent owns one repo; super-agent only for cross-repo decisions.
5. Verification evidence required before close.

Points 2, 3, and 5 are now codified in `README.md` (TL;DR + Lifecycle):
state.json as state of record, manual `acked`->`closed` per receiving
agent, decisions/ notes required to close. Point 4 is canon (Rule A26
repo segmentation). Point 1 collapsed further under lean-af — there is no
daemon to nudge, which strictly satisfies the simplicity intent. No code
change required; the idea is fully reflected.

## Disposition

state.json: both items `acked` -> `closed`. Item frontmatter reconciled
to `closed`. No further action required from claude-pantheon on either
thread.
