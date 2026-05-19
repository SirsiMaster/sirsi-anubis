# Case Study: TUI Controller Refactor — 2,383 Lines to 14 Focused Files

**Date:** May 2026
**Category:** Architecture
**Module:** Output (TUI)

## Problem

`tui.go` was a 2,383-line monolith containing the entire TUI: model, view, update, key handling, command execution, streaming, rendering, and native command functions. `tui_render.go` was another 1,035 lines. Together: 3,418 lines in 2 files.

Six package-level global variables with mutex guards coordinated state between goroutines and the bubbletea event loop.

## Solution

### Mechanical Split

| File | Lines | Purpose |
|------|-------|---------|
| tui.go | 322 | Model, Init, Update, launchers |
| tui_messages.go | 112 | Type definitions |
| tui_actions.go | 128 | Tab definitions, command registry |
| tui_native.go | 546 | 20 native command functions |
| tui_keys.go | 303 | Key handling |
| tui_runner.go | 411 | Command execution, streaming |
| tui_view.go | 444 | View rendering |
| tui_view_status.go | 225 | Status dashboard |
| tui_render_*.go | 4 files, max 345 | Focused renderers |

### Global Elimination

All 6 process globals removed. Native functions now return `nativeResult` directly (was 4-tuple + mutex-guarded shared state). Progress channels passed via closures, not package globals.

### Safety Gateway

`SafetyGateway` interface centralizes all destructive action confirmation.

## Evidence

- 7 controller transition tests prove state machine behavior
- All existing tests pass (19s)
- Zero behavior changes — pure refactor
