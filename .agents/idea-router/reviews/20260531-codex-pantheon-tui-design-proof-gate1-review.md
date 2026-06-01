# Codex Review — TUI Design Proof Gate 1

## Decision

Gate 1 passes for codex-pantheon review. User sign-off is still required before any Go implementation lands.

Gate 2 may be an `internal/tui/` scaffold only: primitives, state model, command registry, renderer contract, fixture screens, and tests. No functional resurrection of the old v0.22 TUI and no operator-facing default launch path are approved by this review.

## Seven Gate Conditions

1. **Layout system — PASS**

   The five primitive model and max-depth binary split tree are a real structural answer to v0.22's view-stack/decorative-tab failure. Named layouts by view are the correct constraint.

2. **Density & typography — PASS WITH PRECONDITION**

   The min/target terminal sizes, no-wrap table policy, truncation, and color ladder are sufficient. Precondition for scaffold: the glyph-width policy must apply to all non-ASCII layout-bearing glyphs, not just Egyptian hieroglyphs. Box drawing, `◉`, arrows, checkmarks, timers, and ellipsis need measured-width handling or ASCII fallback.

3. **Keyboard model — PASS**

   Modeless operation plus `Ctrl-K` palette is right for an operator console. The reachability tiers are clear and testable.

4. **Error states — PASS**

   Blast-radius-altitude error surfaces and destructive-action confirmation align with Pantheon governance. Scaffold should include error fixtures, not only happy-path screens.

5. **Accessibility — PASS WITH PRECONDITION**

   `--no-altscreen`, non-color severity tokens, high contrast, reduced motion, and visible focus are enough for Gate 1. Precondition: scaffold must keep linear/no-altscreen rendering as a first-class renderer, not a later adapter.

6. **Sample screens — PASS**

   The three mocks cover scan, fleet deploy, and router inbox with clear density and alignment. They are good enough as proof-of-craft for the next scaffold gate.

7. **Different from v0.22 deltas — PASS**

   The proof addresses the known failure modes structurally: no decorative top-level tabs, generated hints from command registry, semantic color, registry-driven deity attribution, live command IDs, and safe glyph policy.

## Scaffold Preconditions

- Implement command/hint wiring as data, with tests proving no visible hint references an unregistered command.
- Include fixture-based renderer tests for the three proof screens at 80x24, 100-column, and 120x40.
- Include width/fallback tests for every non-ASCII glyph used in layout-bearing cells.
- Include `NO_COLOR`, reduced-motion, and no-altscreen fixtures in Gate 2.
- Keep `sirsi` no-args behavior unchanged until a separate default-surface decision lands.

## Risk Flag

The one v0.22-class risk still visible in the proof is overconfidence in terminal glyph behavior. The safe-hieroglyph decision is excellent, but the same discipline needs to cover all UI glyphs and box drawing before code.

## Status

Codex half of Gate 1 is approved. The user must still approve before Gate 2 starts.
