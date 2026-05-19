# Codex Plan: Pantheon TUI Controller Refactor

- agent_id: codex-pantheon
- addressed_to: claude-pantheon
- topic: tui-controller-refactor
- status: approved-for-implementation
- created_at: 2026-05-18T18:53:58-04:00
- repo: /Users/thekryptodragon/Development/sirsi-pantheon

## /goal

Refactor the Pantheon TUI so workflows are controller-owned, renderer functions are dumb, and no TUI file remains over roughly 500 lines.

Completion criteria:
- `internal/output/tui.go` is reduced to app wiring/model glue only.
- `internal/output/tui_render.go` is split so no render file is over roughly 500 lines.
- Scan, clean, analyze, status, command execution, and selection/prompt flows each have focused controller/state files.
- Process-global pending state is removed from scan/analyze/select flows.
- Destructive actions route through one safety gateway.
- Existing output/TUI tests pass.
- New or adjusted tests cover controller transitions for at least scan/select, analyze navigation, and destructive clean confirmation.

## Implementation Boundary

Claude owns implementation in `sirsi-pantheon` only.

No cross-repo edits. If another repo appears relevant, write a router item instead of touching it.

## Proposed File Layout

Keep package `output` for now to avoid import churn.

Target split:
- `internal/output/tui.go`: `LaunchTUI`, `TUIModel`, app init, high-level `Update`, `View` dispatch.
- `internal/output/tui_messages.go`: Bubble Tea message types.
- `internal/output/tui_actions.go`: `tabAction`, action registry, command aliasing, action runner interface.
- `internal/output/tui_runner.go`: command/native execution and streaming adapters.
- `internal/output/tui_select_controller.go`: select request state, selection key handling, selected-item execution.
- `internal/output/tui_scan_controller.go`: scan progress/result handling and scan follow-up suggestions.
- `internal/output/tui_clean_controller.go`: dry-run/confirm cleanup orchestration and safety gateway.
- `internal/output/tui_analyze_controller.go`: analyze result state, drill-down/back behavior.
- `internal/output/tui_prompt_controller.go`: freeform command prompt state and key handling.
- `internal/output/tui_status_controller.go`: active/vitals/notifications refresh behavior.
- `internal/output/tui_render_shell.go`: tab bar, layout shell, running/done shells, bottom hints.
- `internal/output/tui_render_status.go`: status dashboard rendering.
- `internal/output/tui_render_analyze.go`: analyzer rendering.
- `internal/output/tui_render_select.go`: selection/prompt rendering.

This is a suggested split, not sacred. The important invariant is controller ownership and small files.

## Controller Contract

Use small controller structs with explicit state and methods. Example shape:

```go
type ScanController struct {
    progress []string
    selected *selectRequest
}

func (c *ScanController) Start(action tabAction, runner ActionRunner) tea.Cmd
func (c *ScanController) HandleProgress(line string) (changed bool)
func (c *ScanController) HandleResult(result nativeResultMsg) (nextActions []suggest.Action, err error)
```

Keep controllers boring:
- They own workflow state.
- They return state changes, messages, and next actions.
- They do not render lipgloss views directly unless the file is explicitly a renderer.
- They do not read/write process globals.

## Required Safety Gateway

Introduce one cleanup/destructive gateway, even if it is initially small:

```go
type SafetyGateway interface {
    ConfirmClean(items []jackal.Finding, source string) error
}
```

All `clean --confirm`, selected cleanup, duplicate purge, and future destructive actions should pass through this gateway. The first implementation can preserve current behavior, but the call path must be centralized.

## Migration Steps

1. Baseline:
   - Run `go test ./internal/output -count=1`.
   - Record existing failures if any before editing.

2. Mechanical split:
   - Move message types, helpers, and render sections into separate files without changing behavior.
   - Keep tests passing after this step.

3. Action runner:
   - Introduce `ActionRunner`.
   - Wrap current native and external command execution behind it.
   - Keep existing command behavior intact.

4. Remove process-global pending state:
   - Replace `pendingSelectReq`, `pendingAnalyzeRes`, `scanProgressCh`, and related mutex/global state with fields owned by `TUIModel` or focused controllers.
   - Preserve async message delivery using explicit tea messages.

5. Workflow controllers:
   - Extract scan, clean, analyze, prompt, and status transitions.
   - Keep `TUIModel.Update` as a dispatcher.

6. Renderer split:
   - Split `tui_render.go` into focused render files.
   - Renderers receive state and return strings. No command execution and no hidden state mutation.

7. Tests:
   - Add table tests for select flow transitions.
   - Add analyze navigation tests for drill-down/back/exit.
   - Add a clean confirmation gateway test that proves destructive cleanup goes through the central gateway.
   - Run `go test ./internal/output -count=1`.
   - If router-adjacent output behavior changed, also run `go test ./cmd/sirsi ./internal/output -count=1`.

## Guardrails

- Do not change command names or user-facing workflows unless required to preserve existing behavior.
- Do not absorb the Pro UX remaining work into this refactor. Session persistence and CommandResult smoke tests are a separate Claude-owned item.
- Do not create a new package unless import boundaries make it clearly cleaner.
- Prefer several small, reviewable commits or router handoffs over one giant rewrite.

## ETA Contract

Claude must write back with:
- `estimated_duration`
- `next_check_at`
- files changed
- tests run
- any behavior that intentionally changed
