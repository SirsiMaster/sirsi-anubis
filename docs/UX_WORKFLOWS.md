# UX Workflow Audit — Sirsi Pantheon v0.21.x

**Date:** May 14, 2026
**Auditor:** Claude (code review) + Codex (acceptance criteria)
**Method:** Source code trace of each CLI handler through to output

---

## Audit Results

| Command | Discover | Progress | Complete | Error | Empty | Narrow | Plain | Status |
|---------|----------|----------|----------|-------|-------|--------|-------|--------|
| `sirsi scan` | Pass | Partial | Pass | **Fail** | Pass | Pass | Warn | Swallows scan errors silently |
| `sirsi clean` | Pass | Pass | Pass | Warn | Pass | Pass | Warn | Some errors swallowed in partial clean |
| `sirsi purge` | Pass | Warn | Pass | Pass | Pass | Warn | Pass | No progress during scan; 30-char column |
| `sirsi installer` | Pass | Warn | Pass | Pass | Pass | Warn | Pass | No progress; 35-char column |
| `sirsi diagnose` | Pass | Warn | Pass | Pass | Pass | Pass | Pass | No spinner for health check |
| `sirsi analyze` | Pass | Warn | Warn | Pass | Pass | Warn | Pass | Suggests TUI instead of in-place drill |
| `sirsi audit` | Pass | Pass | Pass | Pass | Pass | Pass | Warn | Ma'at deity vocab in headers |
| `sirsi risk` | Pass | Warn | Pass | Pass | Warn | Pass | Warn | Osiris deity vocab in headers |
| `sirsi ghosts` | Pass | Warn | Pass | **Fail** | Pass | Pass | Warn | Swallows scan errors; KA deity header |

**Legend:** Pass = acceptable, Warn = works but improvable, Fail = blocks quality, Partial = sometimes works

---

## Critical Issues (Fix Required)

### 1. Silent error swallowing in scan and ghosts

`cmd/sirsi/anubis.go`:
- Line 153: `res, _ := engine.Scan(ctx, ...)` — ignores scan engine errors
- Line 536: `ghosts, _ := scanner.Scan(ctx, ...)` — ignores ghost scan errors

Users get no feedback when the scan engine fails. This violates Rule A1 (every destructive operation must be safe) and basic usability (don't leave users dangling).

**Fix:** Check and report errors. At minimum: `output.Warn("Scan error: %v", err)`.

### 2. Deity vocabulary in CLI headers

Multiple commands print deity names in their output headers:
- `"ANUBIS — Scan"` → should be `"Scan"` or `"Infrastructure Scan"`
- `"ANUBIS — The Sight (KA)"` → should be `"Ghost App Detection"`
- `"ANUBIS — Purge"` → should be `"Build Artifact Purge"`
- `"ANUBIS — Analyze"` → should be `"Disk Analyzer"`
- `"ANUBIS — Installer"` → should be `"Installer Cleanup"`

These headers are the first thing users see. Module attribution can appear in footers or `--verbose` output.

---

## Improvements (Non-Blocking)

### 3. No progress feedback for long operations

`scan`, `purge`, `installer`, `analyze`, `ghosts`, `risk`, and `diagnose` all block without visible progress. The TUI versions have spinners, but CLI-only runs show nothing until completion.

**Fix:** Add a simple spinner or `"Scanning..."` line for operations that take >1s.

### 4. Hardcoded column widths

`purge` (30 chars), `installer` (35 chars), `analyze` (30 chars) use fixed-width `printf` formatting. On narrow terminals (<80 cols), output will misalign or wrap.

**Fix:** Use `lipgloss` table rendering (already available) or detect terminal width.

### 5. `analyze` suggests TUI instead of showing results

After analysis, the only next-step suggestion is "Launch TUI for drill-down navigation." Users running CLI-only workflows get no actionable path forward.

**Fix:** Add `sirsi analyze --depth 2` or similar for CLI drill-down.

---

## What Passed Well

- All 9 commands are discoverable via `sirsi --help` and support `--help` individually
- All commands show completion results with timing
- Empty states are handled consistently ("No findings", "No installer files found", etc.)
- JSON output mode (`--json`) works on all commands that support it
- The Pantheon brand styling (gold/black) is consistent and readable
- The "What's Next" suggestion pattern is present on most commands

---

## Codex UX Review Contract

Per the router proposal, every future Claude implementation review must include:

```
## UX Workflow Review
- Entry point:
- Progress feedback:
- Completion state:
- Error/empty state:
- Cancellation/back navigation:
- Output visible on screen:
- Next action clear:
- Plain-language outcome:
- Internal/module names hidden or justified:
- User left dangling? yes/no:
```

This document will be updated as workflows are fixed.
