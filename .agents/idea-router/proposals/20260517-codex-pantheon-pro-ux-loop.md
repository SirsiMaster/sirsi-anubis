# Proposal: Pantheon Pro UX Loop

- id: 20260517-codex-pantheon-pro-ux-loop
- author: codex
- addressed_to: claude
- status: needs-implementation
- topic: pantheon-pro-ux-loop
- created_at: 2026-05-17T00:00:00-07:00

## Context

Pantheon has substantial engine capability, but the product experience is not yet Pro-ready. The user reports that commands feel like one-off dead ends: no persistent input loop, weak or missing output, unclear next actions, and no coherent CLI/TUI/GUI flow that helps the user understand what happened and what to do next.

This is not an enterprise/Ra workstream. Ra belongs inside Sirsi as a later full-commitment product. This workstream is strictly Pantheon Pro: local developer workstation hygiene, diagnostics, cleanup, and repo readiness.

## /goal

Pantheon becomes a guided local Pro product loop instead of a set of disconnected commands.

For this first implementation slice, `/goal` is met when:

- The core command experience has a shared user-facing result contract.
- At least `scan`, `clean`, `ghosts`, `duplicates`, `diagnose`, `network`, `status`, `risk`, and `audit` have a path to render:
  - what started
  - visible progress or phase text
  - what happened
  - evidence/counts
  - warnings/errors
  - 2-4 recommended next actions
- One-shot CLI commands still work for scripts, but user-facing mode never ends with silence or a blank dead end.
- Interactive `sirsi` has a clear return path to a prompt/input element with the latest state and next actions.
- Deity/module vocabulary is hidden from normal user-facing output unless `--verbose` is active.
- Tests or smoke checks prove the first upgraded commands emit summaries and next actions.

## /plan

1. Audit the existing command flow for the Pro command set:
   - `sirsi`
   - `sirsi scan`
   - `sirsi clean`
   - `sirsi ghosts`
   - `sirsi duplicates`
   - `sirsi diagnose`
   - `sirsi network`
   - `sirsi monitor`
   - `sirsi status`
   - `sirsi risk`
   - `sirsi audit`
2. Add or reuse a shared result model, for example:

```go
type CommandResult struct {
    Command     string
    Summary     string
    Evidence    []Evidence
    Warnings    []string
    Errors      []UserError
    NextActions []NextAction
}

type NextAction struct {
    Label       string
    Command     string
    Description string
    Risk        string
}
```

3. Add a renderer used by CLI/TUI-compatible flows:
   - plain text by default
   - JSON-safe shape where relevant
   - no deity/internal labels in normal mode
4. Start with the highest-value commands:
   - `scan`
   - `clean`
   - `ghosts`
   - `diagnose`
5. Fix known UX blockers from `docs/UX_WORKFLOWS.md`:
   - no silent error swallowing in `scan`
   - no silent error swallowing in `ghosts`
   - progress feedback for long operations
   - completion summaries
   - useful next actions
6. Make interactive `sirsi` return to an input prompt after command completion, preserving:
   - last command result
   - current state summary
   - next-action suggestions
7. Update docs only after behavior exists:
   - README quick start
   - `docs/UX_WORKFLOWS.md`
   - any command help that now has next-action behavior
8. Submit back to Codex with:
   - exact files changed
   - exact commands/tests run
   - screenshots or transcript snippets of upgraded command output
   - explicit residual gaps

## A-Z Product Direction

A. Freeze scope. No new modules, deities, fleet/Ra, dashboards, or enterprise features until Pro UX is coherent.

B. Product promise: Pantheon Pro keeps a developer machine clean, healthy, and ready to ship.

C. Core journey: open Pantheon, see state, scan, understand findings, preview cleanup, clean safely, see result, choose next action, return to input/dashboard.

D. Pro command set: `sirsi`, `scan`, `clean`, `ghosts`, `duplicates`, `diagnose`, `network`, `monitor`, `status`, `risk`, `audit`.

E. UX contract: start, progress, result, evidence, next actions, return path, error/empty state.

F. One-shot CLI may exit, but must end with useful next commands.

G. Persistent CLI shell must show current state, last action, and next-action suggestions.

H. TUI must have persistent input, findings, details, history, and next-action panel.

I. Menu bar must show meaningful state and open the right context, not a disconnected terminal.

J. Normal language only. Hide deity/module words except in verbose/internal contexts.

K. Fix silent failures first.

L. Add progress feedback to long operations.

M. Use structured command results instead of ad hoc text-only returns.

N. Make suggestions mandatory.

O. Add session state: last scan, last cleanup, ignored findings, current warnings, pending recommendations.

P. Reconcile version/docs drift before paid release.

Q. Hide or defer weak/stub surfaces.

R. Add UX smoke tests.

S. Stabilize full test suite and split slow/flaky suites.

T. Dogfood on real machines for two weeks and fix every dead end.

U. Product dashboard should summarize machine state, waste, ghosts, duplicates, health, security, repo risk, recent actions, and recommended next step.

V. Cleanup must build trust: preview, risk labels, protected paths, Trash-first, undo/recovery, decision log.

W. Package Pro beta only after signed install, first run, uninstall, logs, and no dev-only output work.

X. Pricing gate: no charge until the core journey, TUI loop, CLI next actions, and menu bar state work.

Y. Launch as Pantheon Pro Beta: developer machine hygiene, cleanup, diagnostics, and local repo intelligence.

Z. Move fleet/enterprise to Ra later inside Sirsi.

## Implementation Boundary

Claude owns implementation inside `sirsi-pantheon` only. Do not edit other repos. Do not expand feature scope. This is a product-loop hardening workstream, not a module expansion workstream.

