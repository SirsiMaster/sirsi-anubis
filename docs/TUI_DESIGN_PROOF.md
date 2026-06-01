# TUI Design Proof — Pantheon Operator Console

| Field | Value |
| :--- | :--- |
| **Status** | **Gate 1 cleared** — codex-pantheon approved 2026-05-31 (`reviews/20260531-codex-pantheon-tui-design-proof-gate1-review.md`); user signed off for Gate 2. **Gate 2 scaffold landed** at `internal/tui/` (scaffold only — no launch path; see CHANGELOG [Unreleased]). |
| **Author** | claude-pantheon (lane: `pantheon-interactive-surface-decision`) |
| **Gate** | Phase-2 batch-2 Gate 1 — design proof. **No `internal/tui/` code lands until this clears.** |
| **Governs** | [ADR-020](ADR-020-INTERACTIVE-SURFACE-REOPENED.md) (Hybrid C), [ADR-018](ADR-018-NATIVE-MAC-APP.md), [ADR-016](ADR-016-TUI-PRIMARY-INTERFACE.md) |
| **Quality bar** | Mole-grade (`docs/PHASE1_MOLE_INSPECTION.md`) — typography, hierarchy, animation, density |
| **Brand** | Gold `#C8A951` · Black `#0F0F0F` · Deep Lapis `#1A1A5E` · *"Weigh. Judge. Purge."* |

> **Premise.** This proof exists because the v0.22 BubbleTea TUI was *"utterly unreleasable"* (ADR-018). The deleted code is **not** the foundation (ADR-020 §"What Stays True"). This is a new design. The proof is the gate: if it does not credibly clear the Mole-grade bar **in its medium**, the TUI track does not start. There is no second chance to discover "unreleasable" after Go is written.

The console name is **Horus** 𓂀 — the local-workstation all-seeing surface (ADR-015 deity hierarchy). The TUI is one rendering of Horus; the browser dashboard is another. Both consume the same in-process `internal/dashboard` contract (`docs/DASHBOARD_API.md`); the TUI needs **no IPC layer** (ADR-020 Closure §2).

---

## 1. Layout System

### 1.1 Primitive set

The screen composes from exactly **five** primitives. Anything not expressible in these is out of scope for v1 — constraint is the point.

| Primitive | Role | Example |
| :--- | :--- | :--- |
| **Frame** | The fixed outer chrome: title bar (row 0), status bar (last row), and a single content region between them. Never scrolls. | App shell |
| **Pane** | A bordered, independently-focusable rectangle inside the content region. Panes tile via a **binary split tree** (horizontal or vertical splits only). Max depth 2 in v1. | Findings list + detail |
| **Table** | A virtualized, column-aligned data grid inside a pane. Owns its own scroll, selection, and sort. | Scan results |
| **Palette** | A modal overlay (centered, dimmed backdrop) for the command palette and any pick-list. Exactly one palette open at a time. | `Ctrl-K` command palette |
| **Toast** | A transient, non-modal banner anchored above the status bar. Auto-dismiss. Never steals focus. | "Scan complete · 14 findings" |

### 1.2 Why this set, not v0.22's

v0.22 used a **view-stack** model (`internal/output/tui*.go`, ~4,800 LOC): full-screen views pushed/popped, with decorative tabs implying navigation that the stack didn't actually provide. Two failures fell out of that:

1. **Tabs were decorative** — they suggested lateral movement the view-stack could not do (§7, delta 1).
2. **Density collapsed** — full-screen-per-view wastes a 120×40 terminal; an operator wants the findings table *and* the selected finding's detail visible at once.

The **binary-split-tree pane model** fixes both: navigation is spatial (panes are where they are; you move focus between them), and density is a function of how you split, not a fixed one-view-per-screen. Tabs, where they appear, are **inside a single pane** and are genuinely navigable (each tab swaps that pane's table; focus and selection are real). The split tree caps at depth 2 so the layout is always describable in one sentence and never becomes the unreadable nesting that made v0.22 feel arbitrary.

### 1.3 The three layouts

v1 ships **three** named layouts; the operator never hand-tiles:

```
LAYOUT A — "Survey"          LAYOUT B — "Inspect"         LAYOUT C — "Stream"
(single table)              (master + detail, 60/40)     (table + live log, 70/30 vert)
┌──────────────┐            ┌─────────┬────────┐         ┌──────────────┐
│              │            │         │        │         │   table      │
│   table      │            │  table  │ detail │         ├──────────────┤
│              │            │         │        │         │   log pane   │
└──────────────┘            └─────────┴────────┘         └──────────────┘
scan, router inbox          finding → remediation        ra deploy, monitor
```

Layout is chosen by the **view**, not the user — each view (Scan, Ra, Inbox, …) declares which layout it wants. The operator switches *views*, not *tilings*.

---

## 2. Density & Typography

### 2.1 Target terminals and font bar

The console targets **modern GPU-accelerated terminals** with truecolor and ligature-capable monospace fonts:

| Terminal | Min version | Truecolor | Notes |
| :--- | :--- | :--- | :--- |
| Ghostty, WezTerm, Kitty, Alacritty | current | ✓ | Primary targets — full fidelity |
| iTerm2, macOS Terminal.app | iTerm2 3.4+, Terminal 2.12+ | iTerm ✓ / Terminal 256-only | Terminal.app degrades to 256-color ramp (§2.4) |
| tmux/screen passthrough | tmux 3.3+ | ✓ with `Tc` | Documented `.tmux.conf` requirement |

**Font bar (the user's "Mole-grade" invocation, applied to type):** the reference faces are **JetBrains Mono**, **SF Mono**, and **Iosevka** (term/fixed). The console does **not** bundle or assume a Nerd Font, because requiring a patched font is itself a v0.22-class failure mode. Icon glyphs come from a **capability-probed set** (§2.3), not a hard dependency.

### 2.2 Cell budget

The console assumes a **minimum 80×24** and is designed for **120×40**. Below 80×24 it shows a single centered message ("Horus needs ≥ 80×24") rather than rendering broken. Region budgets at 120×40:

| Region | Rows | Cols | Rule |
| :--- | :--- | :--- | :--- |
| Title bar | 1 | full | App name (gold), active view (white), context breadcrumb (dim) |
| Content | 36 | full | Pane tree; min pane interior 24×6 |
| Status bar | 1 | full | Key hints (dim), mode (lapis chip), counts (gold) |
| Toast | 1 (transient) | ≤ 60, right-anchored | Floats above status bar |
| Gutter between split panes | — | 1 | Single `│` / `─`, never double |

Table rows are **single-line, no wrapping** — a value that overflows its column is **truncated with `…`** and revealed in full only in the detail pane (Layout B) or on selection. Wrapping inside tables was a v0.22 density killer; this proof forbids it (consistent with the repo rule "lines must fit on one line").

### 2.3 Glyph budget — the load-bearing decision

**This is the section v0.22 failed hardest on, so it is the most explicit.**

The Egyptian deity glyphs (𓂀 Horus, 𓆄 Ma'at, 𓁯 Net, 𓂓 Ka, 𓇶 Ra, 𓁢 …) live in the **Egyptian Hieroglyphs Unicode block (U+13000–U+1342F)**. **Almost no terminal monospace font renders this block.** In JetBrains Mono, SF Mono, and Iosevka they are **tofu** (`□`) — and tofu for an *East-Asian-width-ambiguous* codepoint **breaks cell alignment**, which cascades into exactly the "glyph rendering failure" that helped sink v0.22.

**Rule G1 — Hieroglyphs are never layout-bearing in the TUI.** No table cell, border, or fixed-width region may contain a U+13000+ codepoint. Deity *identity* in the TUI is carried by **(color + name + a safe sigil)**, never by the hieroglyph.

**Rule G2 — Safe sigil set.** Each deity gets a 1-cell ASCII/BMP sigil that renders everywhere:

| Deity | Canon glyph (docs/CLI only) | TUI safe sigil | Color |
| :--- | :--- | :--- | :--- |
| Horus (console/dashboard) | 𓂀 | `◉` | Gold |
| Jackal (scan) | 🐺 | `J` on lapis chip | Lapis |
| Ka (ghost hunt) | 𓂓 | `*` | Dim white |
| Scales (policy) | ⚖️ | `=` | Gold |
| Hapi (resources) | `~` | `~` | Cyan |
| Ma'at (QA gate) | 𓆄 | `+` | Green |
| Ra (fleet) | 𓇶 | `^` | Gold |

**Rule G3 — Capability probe, then optional flair.** At startup Horus renders the hieroglyph `𓂀` to an off-screen cell and measures the reported cursor advance. If the terminal reports the expected single/however-wide advance **and** the codepoint is not substituted, the hieroglyph may appear **only in non-layout positions** (the splash, the About modal) as flair. Default and fallback are the safe sigils. The probe result is cached per `$TERM`+font in `~/.config/sirsi/tui.json`.

This means: **the brand survives, but never at the cost of a broken grid.** Box-drawing uses the standard light set (`│ ─ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼`), which *is* universally present, never the heavy/double set (alignment + degrade-to-ASCII both cleaner).

### 2.4 Color semantics (truecolor → 256 → 16 ladder)

Color is **semantic**, never decorative — every color means one thing:

| Token | Truecolor | Meaning | 256 fallback | 16 fallback |
| :--- | :--- | :--- | :--- | :--- |
| `brand` | `#C8A951` gold | Identity, headers, selected counts | 179 | yellow |
| `accent` | `#1A1A5E` lapis | Mode chips, active focus border | 17 | blue |
| `ok` | green | Pass, safe, healthy | 35 | green |
| `warn` | amber | Needs attention, reclaimable | 214 | yellow |
| `danger` | red | Destructive, error, protected-path block | 196 | red |
| `dim` | gray | Chrome, hints, truncation | 244 | default |

`NO_COLOR` env → the 16-set collapses further to **attribute-only** (bold/reverse/underline). Color is never the *sole* carrier of meaning (§5 accessibility).

---

## 3. Keyboard Model

### 3.1 Modeless, with a command palette

The console is **modeless by default** (not vim-modal) — operators are not text-editing, they are navigating and dispatching. There is no "insert mode" to get trapped in (a top-3 complaint that kills TUI adoption). The **command palette** (`Ctrl-K`) is the universal escape hatch: every action is reachable by name, fuzzy-searched, so discoverability never depends on memorizing keys.

### 3.2 Reachability tiers

| Keystrokes | What's reachable | Examples |
| :--- | :--- | :--- |
| **0 (always visible)** | Status-bar hints show the 4–6 context actions for the focused pane | `↑↓ move · enter inspect · c clean · / filter` |
| **1** | Global navigation + pane focus + primary verb of the view | `Tab` cycle panes · `1–6` jump to view · `enter` drill in · `esc` back/close |
| **2** | Anything, via palette | `Ctrl-K` then type ("clean", "deploy ra", "open inbox item") |

**Single-key global bindings (the entire reserved set — small on purpose):**

```
Ctrl-K  command palette        Tab / S-Tab  focus next / prev pane
1..6    jump to view           enter        inspect / activate selection
/       filter focused table   esc          dismiss modal / pop detail / back
?       help overlay           q            quit (confirms if work in flight)
g / G   top / bottom of table  r            refresh focused view
```

### 3.3 Conflict resolution with the terminal emulator

`Ctrl-K` is the only potentially-contested binding (some shells bind kill-line, but inside a fullscreen alt-screen app the emulator delivers it to us). Reserved emulator chords are **never** rebound: `Ctrl-C` (interrupt → graceful quit with confirm), `Ctrl-Z` (suspend → honored), `Ctrl-D`, `Ctrl-L` (redraw → honored as redraw), `Ctrl-+/-` (font zoom → emulator keeps). Mouse is **optional enhancement** (click-to-focus pane, wheel-scroll table); the console is **100% keyboard-complete** with mouse disabled.

### 3.4 Dispatch mechanism (the v0.22 "keys did nothing" fix)

In v0.22, key bindings were declared but **dispatched nowhere**. Here, every binding resolves through a single **`Command` registry**: a keypress or palette selection produces a `Command{ID, Args}` that is routed to the focused view's reducer, which returns the next state. **Every key that is shown in the status bar is, by construction, wired** — the status-bar hints are *generated from the focused view's registered commands*, so a hint cannot exist for an unwired key. This is the structural guarantee that §7 delta 2 demands.

---

## 4. Error States

Errors surface at the **altitude of their blast radius** — never one global error modal for everything.

| Class | Surface | Recovery path |
| :--- | :--- | :--- |
| **Field/validation** (bad filter regex) | Inline, red, under the input | Self-clears on valid input; `esc` cancels |
| **Action failure** (clean blocked by protected path) | **Toast**, `danger` color, persists until acknowledged | Toast links to the audit reason; `enter` opens detail with the safety rule citation (Rule A1) |
| **View load failure** (dashboard contract returns error envelope) | **In-pane banner** replacing the table body, with the error envelope's `code`+`message` | `r` retries; `Ctrl-K → "open log"` shows the raw exchange |
| **Fatal** (terminal too small, render panic) | **Full-frame takeover**, single message, no chrome | `esc`/`q` exits cleanly to the shell, restoring the screen (alt-screen teardown guaranteed via deferred restore) |

**Hard rule (Rule A1 alignment):** a destructive action (`clean`, `guard --slay`, `hapi --kill-orphans`) **never executes from a single keystroke**. The key opens a **confirm modal** that names the exact targets and shows the dry-run delta; execution requires a deliberate second confirmation. Protected-path blocks are surfaced as `danger`, not swallowed — silent failure is a governance violation here as in the CLI.

No error is ever logged-and-hidden: every error has a visible surface **and** an entry reachable via `Ctrl-K → "open log"`.

---

## 5. Accessibility

TUI accessibility is real and easily broken; this proof treats it as a gate item, not an afterthought.

- **Screen readers.** Fullscreen TUIs are largely opaque to screen readers, so the console ships a **`--no-altscreen` linear mode**: the same data rendered as scrolling, labeled, line-oriented output that a screen reader can traverse (each row prefixed with its semantic label, e.g. `Finding 3 of 14: …`). The palette and all actions remain reachable. This mode is auto-suggested when `$TERM` indicates a screen-reader pairing or when started under one of the known SR-friendly emulators.
- **Color is never sole signal (§2.4).** Severity also carries a **text token** (`PASS` / `WARN` / `BLOCK`) and a shape sigil, so red/green-blind operators and `NO_COLOR` users lose nothing. WCAG contrast: gold `#C8A951` on black `#0F0F0F` ≈ 8.5:1, lapis text is never on black (used as a *background* chip with white text ≥ 7:1).
- **High-contrast toggle.** `Ctrl-K → "high contrast"` swaps the palette to a pure black/white/single-accent ramp meeting AAA, persisted in `tui.json`.
- **Reduced motion.** All animation (spinners, toast slide, focus-border pulse) is **off** when `prefers-reduced-motion` is detectable, when `NO_COLOR` is set, or via `Ctrl-K → "reduce motion"`. Spinners degrade to a static `⠿`/`*`; progress to a determinate `[####····]` bar with a numeric percent (never motion-only).
- **Focus is always visible.** The focused pane has a `accent`-colored border **and** a `▸` marker in its title — focus is never conveyed by color alone.

---

## 6. Sample Screens

Three canonical views, cell-aligned at 100 cols. These mocks **are** the visual proof (ADR-020 §6) — they stand in for a Figma prototype. Color is annotated in `‹brackets›` since ASCII cannot show it.

### 6.1 Scan results — Layout A "Survey" (Jackal)

```
┌ ◉ Horus ‹gold›  ·  Scan ‹white›  ·  ~/Development ─────────────────────────── ⏱ 1.2s ┐
│                                                                                       │
│  RULE                  FINDINGS   RECLAIMABLE   SEVERITY   LAST SEEN                   │
│  ────────────────────────────────────────────────────────────────────────────────── │
│ ▸parallels-remnants         12        4.2 GB    ‹warn›WARN   12s ago                   │
│  docker-dangling             8        1.1 GB    ‹warn›WARN   12s ago                   │
│  ghost-apps (Ka *)           5        318 MB    ‹dim› INFO   12s ago                   │
│  node-modules-orphan        31        9.7 GB    ‹warn›WARN   12s ago                   │
│  xcode-derived-data          3       14.0 GB    ‹warn›WARN   12s ago                   │
│  spotlight-corrupt           1            —     ‹danger›BLOCK 12s ago  (protected)     │
│                                                                                       │
│  6 rules · 60 findings · 29.3 GB reclaimable · 1 blocked                              │
└───────────────────────────────────────────────────────────────────────────────────┘
 ↑↓ move · enter inspect · c clean (dry-run) · / filter · r rescan      ‹lapis›SCAN ‹gold›29.3 GB
```

### 6.2 Ra deployment status — Layout C "Stream" (fleet)

```
┌ ◉ Horus  ·  Ra ^ Fleet ‹gold›  ·  scope: cleanup-sweep ──────────────────── 4 nodes ┐
│  NODE              SURFACE   STATE        STEP                    ELAPSED             │
│  ─────────────────────────────────────────────────────────────────────────────────  │
│ ▸mac-studio-01     claude    ‹ok›RUNNING  3/7 weighing findings   00:42              │
│  mac-mini-02       codex     ‹ok›RUNNING  5/7 purging             01:07              │
│  vm-ubuntu-03      claude    ‹warn›WAIT   2/7 awaiting confirm     00:51              │
│  vm-ubuntu-04      claude    ‹danger›FAIL 0/7 health-check failed  00:03              │
├─ log · vm-ubuntu-04 ────────────────────────────────────────────────────────────── │
│  01:03:12  ^ Ra      dispatch scope=cleanup-sweep pid=10521                          │
│  01:03:12  + Ma'at   pre-flight gate... [vm-ubuntu-04]                               │
│  01:03:15  ‹danger›  health-check: `sirsi --version` exit 127 (binary not on PATH)   │
│  01:03:15  ^ Ra      node marked FAIL; not reaped (remote host)                      │
└───────────────────────────────────────────────────────────────────────────────────┘
 ↑↓ node · enter detail · r redeploy node · esc back · / filter      ‹lapis›RA ‹gold›3 live / 1 fail
```

### 6.3 Router inbox — Layout B "Inspect" (idea-router)

```
┌ ◉ Horus  ·  Router Inbox ‹gold›  ·  claude-pantheon ───────────────────────── 2 open ┐
│  FROM → TO            LANE / TITLE              STATE  │ ▸ canon-correction v2          │
│  ──────────────────────────────────────────────────  │ ──────────────────────────────│
│ ▸codex → claude   B  canon-correction v2  ‹ok›CLOSED  │ from:  codex-pantheon          │
│  codex → claude   A  dispatch concurrency ‹warn›OPEN  │ to:    claude-pantheon         │
│  claude → codex   B  TUI design proof     ‹dim›DRAFT  │ status: closed · ✓ APPROVED    │
│                                                       │                                │
│                                                       │ /goal:                         │
│                                                       │  (a) ack patches      ‹ok›✓    │
│                                                       │  (b) ack commit scope ‹ok›✓    │
│                                                       │  (c) open proof gate  ‹ok›✓    │
│                                                       │                                │
│                                                       │ verdict: gate OPEN for         │
│                                                       │  docs/TUI_DESIGN_PROOF.md      │
└───────────────────────────────────────────────────────────────────────────────────┘
 ↑↓ item · enter open · a ack · / filter · r refresh          ‹lapis›INBOX ‹gold›1 needs reply
```

Every column in every mock aligns to a fixed cell grid; truncation (`…`) is used, never wrapping; deity identity uses safe sigils (`◉ ^ * +`), never hieroglyphs; severity carries a text token **and** a color annotation.

---

## 7. Explicit "Different From v0.22" Deltas

The gate (ADR-020 §7) names six v0.22 failure modes. Each is addressed structurally — not "we'll be more careful," but a mechanism that makes the failure unrepresentable.

| # | v0.22 failure | This design's structural fix |
| :--- | :--- | :--- |
| 1 | **Tabs were decorative** | No top-level tabs. Navigation is the binary-split pane tree (§1) + numbered views (§3.2). Tabs exist only *inside* a pane and genuinely swap that pane's table with real focus/selection. A decorative tab is not expressible. |
| 2 | **Key bindings did nothing** | Every binding resolves through the `Command` registry (§3.4). Status-bar hints are **generated from** the focused view's registered commands, so a shown key is provably wired — a dead hint cannot be rendered. |
| 3 | **Wrong color semantics** | Six semantic tokens, each meaning exactly one thing, with a truecolor→256→16→attribute ladder (§2.4). Color is assigned by meaning, not by decoration; `NO_COLOR` and colorblindness lose no information (§5). |
| 4 | **Wrong deity attribution** | Sigils and names bind to the **Deity Registry (Rule A25)** at render time from a single table (§2.3 G2). Ma'at is always `+`/green for QA; Ra is always `^`/gold for fleet. A deity cannot be mis-attributed because attribution is table-driven, not per-screen string literals. |
| 5 | **Stale CLI verb** | The TUI dispatches the **same `Command` IDs** the CLI cobra tree exposes (post-restructure verbs: `scan, clean, status, audit, risk, hardware, ra, thread, router, maat`). There is no parallel TUI verb list to drift; the palette is generated from the live command set. A verb the CLI dropped vanishes from the palette automatically. |
| 6 | **Glyph rendering failure** | Rules G1–G3 (§2.3): hieroglyphs are **forbidden in layout-bearing positions**; deity identity uses BMP-safe sigils; the hieroglyph appears only as probe-gated flair in non-layout cells. Box-drawing is the universally-present light set with an ASCII degrade. Tofu-breaks-the-grid is structurally impossible. |

**The meta-delta:** v0.22 conflated *declaring* UI with *wiring* it. This design makes the rendered surface a **projection of wired state** — status hints from registered commands, palette from the live verb set, sigils from the deity registry, severity from semantic tokens. You cannot render an affordance that isn't backed, because the render reads from the backing.

---

## Gate Readiness

This proof addresses all seven conditions of ADR-020 §"Why This TUI Will Be Different." Per the gate, the next artifact after **codex-pantheon + user approval** is the `internal/tui/` scaffold (Phase-2 batch-2 Gate 2) — **not** before. Mac native SwiftUI remains deferred to Phase-3 (Hybrid C), to activate once this TUI clears its v1 quality bar.

**Open questions surfaced for the user** (do not block codex's structural review, but shape v1 scope):

1. **Product surface vs. proof-of-craft** (ADR-020 OQ1): is Horus-TUI something users *operate Pantheon through*, or the demonstration that Sirsi can build one? This sets whether §6.3 (router inbox) is in v1 or v2.
2. **Default surface on first launch** (ADR-020 OQ3): with CLI + TUI + (later) Mac native, what does `sirsi` with no args do once the TUI exists? This proof assumes it still **prints help** (per current CLI_COMPATIBILITY); a TUI auto-launch is a separate decision.
3. **Reduced-motion / animation budget**: the Mole bar invoked "animation." How much motion is on-brand vs. noise for an infra tool operators keep open for hours?
