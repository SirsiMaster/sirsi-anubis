# ADR-021: Deities Must Not Assume Single-Repo — Osiris Workstation-Scoping

| Field | Value |
| :--- | :--- |
| **Status** | Proposed — 2026-05-31 |
| **Date** | 2026-05-31 |
| **Author** | claude-pantheon (`thr-7452fa9c16e656c9`, lane: pantheon-runtime-restore) |
| **Reviewer** | codex-pantheon (review requested) |
| **Supersedes** | None |
| **Related** | [ADR-015](ADR-015-DEITY-HIERARCHY.md), [ADR-017](ADR-017-RA-HORUS-CTR-HYPERVISOR.md), [ADR-010](ADR-010-MENUBAR-APPLICATION.md), [ADR-006](ADR-006-SELF-AWARE-RESOURCE-GOVERNANCE.md) |

## Context

After an out-of-memory reboot, the menubar app (`Pantheon.app` / `sirsi-menubar`) was rebuilt from source (v0.22.0-beta) and relaunched under a LaunchAgent. Its Recent Activity feed showed every deity succeeding except one:

```
✅ isis — doctor completed
✅ anubis — anubis judge completed
✅ anubis — anubis weigh completed
❌ osiris — osiris assess failed
```

Investigation traced the failure to a single line:

- `cmd/sirsi-menubar/stats.go:84` sets `RepoDir: "."` — a **relative** path.
- A LaunchAgent-spawned process inherits launchd's working directory, which is `/` (confirmed via `lsof -a -p <pid> -d cwd`).
- Osiris assesses **uncommitted git work**. Running it against `/` (not a git repository) errors → the menubar logs `osiris assess failed`.
- Proof the deity itself is healthy: the identical Osiris assessment run from `~/Development/sirsi-pantheon` succeeds and returns `🔴 critical — 142 uncommitted files on main`.

The shallow read is "fix the path / pin a repo." That is wrong, and the user named why:

> *"Sirsi and its Pantheon components are NOT restricted to repo management by any means… if we are doing that, rather than recognizing that we have a design problem."*

## The Design Problem

`RepoDir: "."` is a **CLI-era assumption** — *"the repo I am currently standing in"* — baked into a **workstation-resident** surface. The menubar does not stand in a repo; it stands on the machine.

This contradicts the Pantheon hierarchy established in **ADR-015**: **Horus is the Local Workstation Lord — everything on ONE machine reports to Horus.** A deity that can see only one repo, and *fails* when it sees none, cannot serve a workstation-wide lord. Pinning Osiris to a single repo would not fix this — it would hardcode the wrong model deeper.

Osiris is the symptom. The disease is the latent single-repo assumption in any deity whose domain is actually machine-wide.

## Decision

**1. Establish the principle (canon).**
Any deity whose domain is workstation-scoped — risk of data loss (Osiris), hygiene (Anubis), quality (Ma'at), resource pressure (Isis/Guard) — MUST source its scope from **workstation discovery**, never from the process working directory. `RepoDir: "."` and equivalent cwd-relative defaults are an anti-pattern in any long-running, workstation-resident surface (menubar, daemon, dashboard).

**2. Make Osiris workstation-scoped.**
Osiris in the menubar (and dashboard) aggregates risk across **all active repositories on the machine**, surfacing the rollup and the worst offender, e.g. `🔴 142 in pantheon · 🟡 6 in assiduous`.

**3. Source the repo set from infrastructure that already exists (LEAN — no new discovery layer).**
The repo set is the union of:
- The **CTR thread registry** (`sirsi thread`) — every live agent session already registers its repo. `agents.json` enumerates all active repo cwds (assiduous, finalwishes, nexus, pantheon, porch-and-alley, homebrew-tools, …).
- The **`sirsi thread discover`** primitive (committed 2026-05-31, `10a97b7`) — already reconciles live sessions into the registry. This is the discovery mechanism; Osiris consumes it.
- Optional user-configured dev roots (menubar config), deduped against the above.

**4. Degrade gracefully.**
Zero repos, or a non-git directory, renders a benign state (`✅` / `n/a`) — **never `failed`**. "Nothing to assess" is not an error.

## Alternatives Considered

1. **Pin `RepoDir` to a chosen repo** (e.g. always `~/Development/sirsi-pantheon`): Rejected — hardcodes the single-repo model that is itself the bug; arbitrary on a multi-repo workstation; silently blind to risk in every other repo.
2. **Graceful-degradation only** (non-git dir → `✅`, no aggregation): Rejected as a *final* answer — it silences the symptom while leaving the workstation lord blind to real, machine-wide data-loss risk. Acceptable only as an interim "stop the bleeding" patch ahead of this ADR landing.
3. **Per-cwd assessment** (resolve `"."` to an absolute cwd): Rejected — that *is* the current behavior; a launchd cwd of `/` is never a meaningful repo.
4. **A new repo-discovery walker** (scan the filesystem for `.git`): Rejected as primary — duplicates the CTR registry, which already knows the *active* repos with far less I/O and better signal (active sessions, not every dormant clone on disk). Filesystem walk may serve as an opt-in fallback for users not running CTR threads.

## Consequences

- **Positive**: Osiris matches its mandate (machine-wide data-loss risk). The menubar stops showing a false failure. A general anti-pattern (`cwd`-relative scope in workstation surfaces) is named in canon, preventing the next deity from repeating it. Reuses CTR/`thread discover` rather than building new discovery — net LOC may *drop*.
- **Negative**: Osiris gains a dependency on the CTR registry; needs a defined behavior when CTR is empty (fall back to configured dev roots, else benign). Aggregation across many repos must stay cheap enough for the menubar's refresh tick (git status is fast, but N repos × tick needs a sane cap / cache).
- **Risk**: Scope creep into "rewrite every deity." Mitigation: this ADR *names the principle* but only *mandates the Osiris fix*; other deities are audited, not auto-refactored. Anubis/Ma'at are already workstation-wide and out of scope here.

## Scope Boundary

- **In scope**: the principle (canon), the Osiris workstation-scoping implementation, graceful no-repo handling.
- **Out of scope**: refactoring Anubis/Ma'at (already workstation-wide); filesystem-walk discovery (optional fallback, separate item); any menubar redesign.

## References

- `cmd/sirsi-menubar/stats.go:84` — the `RepoDir: "."` defect.
- Commit `10a97b7` — `feat(router): sirsi thread discover` — the discovery primitive Osiris will consume.
- [ADR-015](ADR-015-DEITY-HIERARCHY.md) — Horus as Local Workstation Lord (the mandate this restores).
- [ADR-017](ADR-017-RA-HORUS-CTR-HYPERVISOR.md) — CTR thread registry (the repo-set source).
- Evidence: menubar cwd `/` (`lsof`), `osiris risk` succeeds from a real repo (🔴 142 files), fails from `/`.

## /goal

ADR reviewed by codex-pantheon and the user. On acceptance: implement `osiris` workstation aggregation sourced from the CTR registry + graceful no-repo handling, with a CHANGELOG entry and tests, in the `claude-pantheon` lane. No code lands before this ADR is accepted.
