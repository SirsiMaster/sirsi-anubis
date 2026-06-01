# ADR-023: One Build-Version Contract + Local Drift Detection

## Status
**Accepted** — June 1, 2026

## Context
The CTR fix `ca6e343` shipped to `~/.local/bin/sirsi`, but `/opt/homebrew/bin/sirsi`
and `/opt/homebrew/bin/sirsi-menubar` stayed pre-fix. The menubar compiles
`internal/router`, so it kept running the OLD reaper while source was fixed —
codex caught it as VERIFY PARTIAL. This was not a one-off: it is a
**distribution-class bug**. There was no single source of truth for the binary
version and no way to *see* that siblings had drifted apart.

Evidence at the time of writing — five disagreeing version literals, zero
compiler enforcement:

| Binary | Declared |
| :--- | :--- |
| `VERSION` file | `0.22.0-beta` |
| `cmd/sirsi` | `v0.21.0` |
| `cmd/sirsi-menubar` | `v0.20.0` |
| `cmd/{anubis,maat,guard,thoth,scarab}` | `v0.4.0-standalone` |

Each `cmd/*/main.go` declared its own `var version`, and ldflags targeted
`main.version` per-binary (deities got only `version`, never commit/date). The
literals were the fallback whenever ldflags were absent, so they drifted freely.

## Decision
1. **One contract.** New `internal/version/build.go` holds the stamped triple
   (`Version`/`Commit`/`Date`) plus `Info` and `Current(binary)`. Every
   `cmd/*/main.go` sources its version from this package — no hand-edited
   literals remain. `Current()` falls back to `debug.ReadBuildInfo()` (vcs
   revision/modified) so a plain `go build` self-reports honestly (Rule A23).
2. **Unified ldflags.** `.goreleaser.yaml` and the `Makefile` stamp
   `internal/version.{Version,Commit,Date}` for **all** binaries, replacing the
   per-binary `main.version` targets.
3. **Uniform JSON.** `sirsi version --json` and a new `sirsi-menubar version
   [--json]` both emit the `Info` shape, so any sibling can be machine-probed.
4. **Local drift detection.** New `internal/selfupdate` package
   (`DetectMethod`, `ScanHost`, `BuildReport`, `DriftReport`) discovers sibling
   sirsi binaries on the host and probes each via `version --json` (200 ms
   timeout, **no network**). It classifies:
   - **D2 sibling drift** — `sirsi` fresh but `sirsi-menubar` (or another
     sibling) stale. *The tonight bug.*
   - **D3 path drift** — the `sirsi` resolved on `$PATH` differs from the
     running executable.
   (D1 self-vs-upstream and verified atomic `sirsi self-update` are deferred —
   see Deferred below.)
5. **Health-surface tie-in.** `sirsi doctor`/`diagnose` appends a `binary-drift`
   finding (`SeverityWarn` when drift, `SeverityOK` when in sync). The existing
   SessionStart `health-line.sh` surfaces it automatically — verified rendering
   `health:🔴 … binary-drift`.

`internal/selfupdate` deliberately does **not** import `internal/router`, so the
health surface can consume it without an import cycle.

## Consequences
- A stale sibling is now **visible** the moment `sirsi doctor` runs — the exact
  failure mode that shipped silently on 2026-06-01 cannot hide again.
- Release builds report one consistent version across every binary; unstamped
  dev builds say `dev` + VCS commit instead of lying with a frozen literal.
- Two integration tests that asserted the frozen `0.21.0` literal were updated
  to assert the version banner renders (the literal is the anti-pattern).

## Deferred (follow-up router items)
- **D1 CheckUpstream** — cached GitHub `/latest`, offline-tolerant.
- **`sirsi self-update`** — verified, atomic, install-method-aware, all-siblings
  replace (Homebrew delegates to `brew upgrade`; go-run refuses; raw does a
  sha256+signature-verified atomic rename of every sibling in the dir).
- **cosign signing** in goreleaser so `Apply` has something to verify.

## Verification
- `go test -race ./internal/version ./internal/selfupdate` — green.
- Stamped build → `sirsi` and `sirsi-menubar` both report `v0.22.0-beta` + commit.
- Staged a `v0.20.0` sibling → `sirsi doctor` flagged D2 (severity 2); health
  line rendered the `binary-drift` token.

Refs: PANTHEON_RULES.md A13 (versioning), A23 (truth), A7 (traceability);
supersedes the scattered `var version` literals. Companion to ADR-022 (CTR).
