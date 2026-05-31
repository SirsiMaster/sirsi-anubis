# CLI Compatibility Matrix — v0.22 → v0.23

This matrix lists every user-visible `sirsi` verb and how v0.23 (v0.22 BubbleTea TUI implementation removed) behaves relative to v0.22 (that TUI present). Module-internal subcommands and debug flags are out of scope — see `sirsi <verb> --help` for those.

**Governing decisions:** ADR-018 (TUI Sunset — *Partially In Force / Amended By ADR-020*) and ADR-020 (Interactive Surface Reopened, closed Hybrid C: new Mole-grade TUI first cross-platform, Mac native later). The v0.22 TUI implementation was removed in v0.23; the CLI is unchanged. The replacement interactive surface is a new Mole-grade TUI (in design — no code lands before `docs/TUI_DESIGN_PROOF.md` clears codex review); native macOS SwiftUI follows as a later-phase polish-bar upgrade.

## Top-level verbs

| Verb | v0.22 | v0.23 | Notes |
| :--- | :--- | :--- | :--- |
| (no args) | Launched TUI gateway | **Prints help** | Intentional. The TUI launcher was the only behavior change. |
| `scan` | Unchanged | Unchanged | — |
| `clean` | Unchanged | Unchanged | `--dry-run` semantics unchanged. |
| `ghosts` | Unchanged | Unchanged | — |
| `duplicates` | Unchanged | Unchanged | — |
| `purge` | Unchanged | Unchanged | — |
| `analyze` | Unchanged | Unchanged | Disk explorer remains plain styled output (lipgloss tables); no interactive navigation. |
| `installer` | Unchanged | Unchanged | — |
| `diagnose` | Unchanged | Unchanged | — |
| `fix` | Unchanged | Unchanged | — |
| `network` | Unchanged | Unchanged | — |
| `monitor` | Unchanged | Unchanged | — |
| `status` | Unchanged | Unchanged | Was already non-TUI. Per ADR-020 / Hybrid C, "live dashboard" rendering returns first in the new Mole-grade TUI (cross-platform); Mac native SwiftUI carries the same flow in a later phase. |
| `audit` | Unchanged | Unchanged | — |
| `risk` | Unchanged | Unchanged | — |
| `hardware` | Unchanged | Unchanged | — |
| `diagram` | Unchanged | Unchanged | — |
| `quickstart` | Unchanged | Unchanged | Guided text flow, never used the TUI. |
| `setup` | Unchanged | Unchanged | — |
| `permissions` | Unchanged | Unchanged | — |
| `guides` | Unchanged | Unchanged | — |
| `version` | Unchanged | Unchanged | — |
| `anubis <verb>` | Unchanged | Unchanged | — |
| `isis <verb>` | Unchanged | Unchanged | — |
| `maat <verb>` | Unchanged | Unchanged | — |
| `ra <verb>` | Unchanged | Unchanged | — |
| `agent` | Unchanged | Unchanged | — |
| `thread` | Unchanged | Unchanged | — |
| `router` | Pull-model verbs only | Pull-model verbs only | Push-model removed in [Unreleased] (separate change, not TUI-related). |

## Removed flags

| Flag | Where | Reason |
| :--- | :--- | :--- |
| `--live` | `sirsi status` | Live dashboard was a TUI surface; replaced by `sirsi status` (snapshot) + planned Mac app. |
| TUI launch flag on root | `sirsi` (no args) | No-args now prints help (see top of matrix). |

## What changed vs. what only moved

- **Removed entirely (no replacement in v0.23):** the interactive TUI gateway (was `sirsi` no-args), live status dashboard.
- **Behavior preserved, surface narrowed:** every scan/clean/diagnose verb still runs and still produces styled `lipgloss` output. No verb lost a flag.
- **Coming back under ADR-020 / Hybrid C:** gateway navigation, live dashboards, scan/clean interactive flows ship first in the new Mole-grade TUI on macOS/Windows/Linux. Mac native SwiftUI carries the same flows in a later phase. Not promised for any v0.23.x. Phase-1 audits (`cmd/sirsi-menubar/`, `mobile/*.go`, `ios/Pantheon/`, Mole reference) become Mac-conditional records per `docs/PHASE1_RESCOPE_NOTE.md`.

## Scripting impact

None expected. All CLI verbs return the same exit codes and the same `--json` output. CI pipelines and scripts that called `sirsi scan --json`, `sirsi clean --dry-run`, etc. are unaffected.

## References

- ADR-018: Native macOS App + CLI Pivot (TUI Sunset) — *Partially In Force / Amended By ADR-020*
- ADR-020: Interactive Surface Reopened — closed Hybrid C (TUI first cross-platform, Mac native later)
- CHANGELOG.md `[Unreleased]` — v0.23 entry (now scoped as "v0.22 TUI implementation removed", not "TUI surface category abandoned")
- Phase-0 completion decision: `.agents/idea-router/decisions/20260521-claude-pantheon-tui-elimination-phase0-complete.md`
