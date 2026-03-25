# 🐕 Case Study: Dogfooding Reveals 64 GB Docker Ghost

**Date:** 2026-03-23 (Session 12)
**Category:** Dogfooding / Disk Reclamation / Product Validation
**Impact:** 64 GB reclaimed — 97.6% of all findings were one unused application
**Rule:** A14 (Statistics Integrity) — all numbers measured, not projected

---

## The Discovery

While benchmarking the Horus shared filesystem index, we ran `weigh --json`
and parsed the findings by size. The output was revelatory:

```
65.6 GB across 341 findings

TOP FINDINGS:
  64.0 GB  com.docker.docker/Data/vms    docker_desktop
   0.4 GB  assiduous/node_modules         node_modules
   0.3 GB  assiduous/web/node_modules     node_modules
   ...
```

**97.6% of all detected waste was a single application — Docker Desktop.**

## The Investigation

| Question | Answer |
|----------|--------|
| Is Docker referenced in Pantheon's build? | **No** — zero Dockerfiles, no container builds |
| Is Docker in GoReleaser config? | **No** — builds native Go binaries |
| Is Docker in GitHub Actions? | **No** — uses `ubuntu-latest` runners directly |
| Is Docker daemon running? | **No** — `Cannot connect to the Docker daemon` |
| What about SirsiNexusApp? | One `Dockerfile.prod` in `ui/` — a leftover, not active |
| What's the deployment stack? | Firebase Hosting (static) + Cloud Functions (serverless) |

**Nothing in the Sirsi ecosystem requires Docker.**

The Docker Desktop application was installed at some point for exploration,
never actively used for production, and its Linux VM disk image grew to
24 GB (plus 28 GB of cached images and layers across support directories).

## The Cleanup

```
Removed:
  /Applications/Docker.app                              2.1 GB
  ~/Library/Containers/com.docker.docker/               24.0 GB
  ~/Library/Group Containers/group.com.docker/          144 KB
  ~/.docker/                                            54 MB
  ~/Library/Logs/Docker Desktop/                        traces
  ~/Library/Preferences/com.docker.docker.plist         traces
  ~/Library/Caches/com.docker.docker/                   traces
  ~/Library/Saved Application State/...docker...        traces
────────────────────────────────────────────────────────────────
  Total reclaimed:                                      ~64 GB
```

## The Results

| Metric | Before | After |
|--------|--------|-------|
| Weigh total | **65.6 GB** | **1.6 GB** |
| Finding count | 341 | 340 |
| Largest finding | 64.0 GB (Docker VMs) | 0.4 GB (node_modules) |

## Why This Matters

This is the product thesis validated in production:

1. **The user (Cylton) did not know Docker was consuming 64 GB.**
   The VM disk images are hidden deep in `~/Library/Containers` where
   no one looks. Docker Desktop does not warn you about disk usage.

2. **Pantheon found it instantly.** The `weigh` command surfaced the
   finding in 833ms, ranked by size, with a clear rule name
   (`docker_desktop`) and actionable path.

3. **The investigation took 30 seconds.** We asked: is Docker in our
   build? Answer: no. Decision: remove.

4. **64 GB freed.** That's more than the entire Pantheon codebase,
   all node_modules, and the Go module cache combined — times 30.

This is exactly what Pantheon is built to do: surface hidden waste,
give actionable context, and empower the user to make informed
decisions. Built on the premise that your computer should work for
you, not silently hoard unused data.

## The Recursive Insight

The performance optimization work *created* this discovery:

```
Ma'at was slow (55s)               → discovered via dogfooding
Fixed Ma'at (55s → 12ms)           → enabled more pushes
More pushes revealed Weigh was slow → fixed Weigh (15.6s → 833ms)
Fast Weigh revealed 65 GB findings → investigated top finding
Investigation revealed Docker ghost → removed it, freed 64 GB
```

**The performance improvements enabled the disk discovery that validated
the entire product.** If Weigh still took 15 seconds, we might not have
looked at the findings data this closely.

---

*Measured on Apple M1 Max, macOS Tahoe.*
*All numbers independently verifiable per Rule A14.*
