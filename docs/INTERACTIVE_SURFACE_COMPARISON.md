# Interactive Surface Comparison Matrix

**Governing ADR:** ADR-020 (Interactive Surface Reopened — Multi-Track Evaluation)
**Purpose:** Help the user pick the surface track(s) Pantheon ships. Evaluative, not prescriptive.
**Status:** Draft for codex-pantheon review.

## Tracks Evaluated

1. **TUI** — A new, Mole-grade operator console / command shell. **Not** a revival of v0.22 code. New design bar: typography, hierarchy, density, key flow, scroll affordances, keyboard-first command surface, router/log inspection, agent threading visibility.
2. **Mac Native (SwiftUI MenuBarExtra + optional main window)** — ADR-018's chosen direction. Native macOS app over the dashboard HTTP contract via unix socket.
3. **CLI + Dashboard** — Ship the surface that already works today: `sirsi <verb>` for everything, with the browser dashboard at `localhost:9119` for visual inspection. No new interactive chrome.
4. **Hybrid A — Mac Native + TUI for Win/Linux** — Mac users get the native app; Windows/Linux users get the new operator TUI; CLI everywhere as scripting fallback.
5. **Hybrid B — TUI everywhere + native Mac shell** — TUI is the cross-platform interactive default. Mac users additionally get a thin native shell that wraps the TUI (Terminal-in-a-window with brand chrome). Closer to "TUI is the wave."
6. **Hybrid C — TUI everywhere + Mac native as power-user upgrade** — TUI ships first on all platforms (fastest to market). Mac native ships later as the polish-bar upgrade.

## Evaluation Dimensions

Each track scored on 8 dimensions. Scores are 1–5 (5 = best); reasoning matters more than the number.

| Dimension | Meaning |
| :--- | :--- |
| **Quality bar** | Can it credibly clear the "Mole-grade" threshold the user has named? |
| **Cost efficiency** | Inverse of first-ship LOC + complexity (including Go/Swift backend and design work). **Higher score = cheaper to ship.** Renamed from "Dev cost" per codex condition 3 — a 5 means low cost, not high cost. |
| **Platform reach** | Coverage across macOS / Windows / Linux. |
| **Distribution** | Friction to install/update on each platform. |
| **Time to ship** | Wall-clock to a credible v1 from today. Assumes single-developer pace per repo conventions. |
| **Accessibility** | Keyboard-only operation, screen-reader compatibility, color contrast, low-vision support. |
| **Local-agent integration** | How well it integrates with the Pantheon agent surface (Ra deployment windows, router state, thread inspection, live logs). |
| **Failure modes** | What happens when it goes wrong; how visible is the failure to the user; how easy is recovery. |

## Track 1 — TUI (new operator console)

| Dimension | Score | Reasoning |
| :--- | :---: | :--- |
| Quality bar | 4 | Achievable at high quality, but **TUIs cannot match native typography/animation**. The Mole-grade bar in a TUI medium is a "best-in-class TUI," not visual parity with native. The user must explicitly accept the medium constraint. |
| Cost efficiency | 4 | ~2,000–4,000 LOC of new Go in `internal/tui/` or similar. Reuses existing `internal/dashboard` business logic. Charm libraries (`lipgloss`, `bubbletea` returning) are the standard. **Risk:** re-adopting BubbleTea is what we just deleted; needs an honest "why this time will be different" answer. |
| Platform reach | 5 | Single Go binary runs on macOS / Windows / Linux. Single artifact, single brand, single command. |
| Distribution | 5 | Already shipped via `brew install` and CLI distribution channels. No new pipeline. |
| Time to ship | 3 | A new credible TUI is months, not weeks, if the quality bar is real. The 4,800 LOC we deleted took prior sessions to write and was still unreleasable; rebuilding to Mole-grade is harder, not easier. |
| Accessibility | 4 | Terminal apps win on keyboard-only operation and screen-reader compatibility via standard terminal accessibility. Color contrast depends on user theme. |
| Local-agent integration | 5 | A TUI has natural seats for live router state, thread tables, Ra window logs, stele tail — these are tabular/text-native concepts. |
| Failure modes | 4 | Crashes are local to the process; restart is one command. No app-bundle corruption, no installer state, no widget extension surprises. |

**TUI verdict:** Strongest on reach + distribution + agent integration. Capped on visual quality bar by medium. The "TUIs are the wave" thesis lives or dies here.

## Track 2 — Mac Native (SwiftUI MenuBarExtra)

| Dimension | Score | Reasoning |
| :--- | :---: | :--- |
| Quality bar | 5 | Native SwiftUI clears the Mole bar at the medium's ceiling — typography, animation, motion, density. |
| Cost efficiency | 3 | ~3,850 LOC Swift (per Phase-1 step 3 audit), plus unix-socket transport in `internal/dashboard`, plus Xcode project, plus signing/notarization path. |
| Platform reach | 1 | macOS only. Windows/Linux unserved. |
| Distribution | 3 | First cut via direct download (Sparkle path noted). Homebrew cask + Mac App Store are later asks. macOS Gatekeeper / notarization is non-trivial. |
| Time to ship | 3 | Phase-1 audits + Phase-2 doc batch + socket transport + bridge + new endpoints + UI = months. |
| Accessibility | 4 | Native macOS accessibility stack is the strongest available. VoiceOver, Switch Control, contrast modes — all for free. |
| Local-agent integration | 4 | Good — but requires bridging the dashboard contract into SwiftUI views. The bridge work is real (per `DASHBOARD_API_GAP.md`: 19 new endpoints + 6 adapters needed). |
| Failure modes | 3 | App-bundle errors, signing failures, sandbox surprises, Sonoma/Sequoia API changes. Mole-grade quality demands handling each. |

**Mac Native verdict:** Best quality bar at the cost of platform reach. Largest dev surface. Was ADR-018's choice for good reason; the cost is unchanged.

## Track 3 — CLI + Dashboard (status quo)

| Dimension | Score | Reasoning |
| :--- | :---: | :--- |
| Quality bar | 2 | The browser dashboard is functional but not aimed at the Mole bar. CLI output is utility-grade. Neither satisfies the user's "interactive surface" frame. |
| Cost efficiency | 5 | Zero new work. Already shipped. (Note: prior draft had this scored 1, which contradicted the reasoning — corrected.) |
| Platform reach | 5 | All platforms. |
| Distribution | 5 | Already shipped. |
| Time to ship | 5 | Already shipped (today). |
| Accessibility | 3 | CLI = terminal accessibility. Browser dashboard = browser accessibility. Both fine, neither curated. |
| Local-agent integration | 3 | CLI commands per deity; dashboard reads JSON. No live operator console. |
| Failure modes | 5 | Lowest blast radius — CLI commands are independent, no long-running process state in the UI layer. |

**CLI + Dashboard verdict:** The "do nothing" track. Cheap, shipping, but does not answer the user's "TUIs are the wave" question. **High score is an artifact of the dimensions weighting effort/risk; it does not reflect strategic intent.** Codex condition 2: do not let the score table decide.

## Track 4 — Hybrid A — Mac Native + TUI for Win/Linux

| Dimension | Score | Reasoning |
| :--- | :---: | :--- |
| Quality bar | 5 | Each medium at its own ceiling. |
| Cost efficiency | 2 | Sum of Tracks 1 and 2. Two codebases, two brand expressions, two release pipelines. |
| Platform reach | 5 | All platforms served by an interactive surface. |
| Distribution | 3 | Mac via direct download / Sparkle; Win/Linux via brew + binary. Two stories. |
| Time to ship | 2 | Latest of the two tracks. Mac native is the long pole. |
| Accessibility | 5 | Best-of-both. |
| Local-agent integration | 5 | Best-of-both. |
| Failure modes | 3 | Two surfaces to maintain; risk of feature drift between them. |

**Hybrid A verdict:** "Right tool for each platform." Highest quality + reach at the highest dev cost.

## Track 5 — Hybrid B — TUI everywhere + native Mac shell wrapping the TUI

| Dimension | Score | Reasoning |
| :--- | :---: | :--- |
| Quality bar | 3 | Mac users see a terminal inside a brand window. Quality is TUI-quality with native chrome. **Risk:** this is what `cmd/sirsi-menubar/`'s `spawnTUIWithCommand` did and the user already rejected. |
| Cost efficiency | 4 | Mostly Track 1 + a thin SwiftUI wrapper. |
| Platform reach | 5 | All platforms. |
| Distribution | 4 | TUI via brew; Mac wrapper via direct download. |
| Time to ship | 3 | TUI is the long pole; Mac wrapper is a few hundred LOC. |
| Accessibility | 4 | Terminal accessibility + native window chrome. |
| Local-agent integration | 5 | TUI seats handle everything. |
| Failure modes | 4 | Single business logic surface; failure modes converge on TUI. |

**Hybrid B verdict:** LEAN cross-platform. The "Mac shell over TUI" pattern is exactly what we just deleted because the underlying TUI was unreleasable. Only viable if Track 1's redesign is genuinely Mole-grade.

## Track 6 — Hybrid C — TUI first cross-platform; Mac native later as polish-bar upgrade

| Dimension | Score | Reasoning |
| :--- | :---: | :--- |
| Quality bar | 4 | TUI clears the bar at its ceiling first; Mac native adds visual ceiling later. |
| Cost efficiency | 4 | Sequential. TUI first (Track 1), Mac native second (Track 2). |
| Platform reach | 5 | All platforms served immediately by TUI; Mac users get an upgrade later. |
| Distribution | 4 | TUI via brew first; Mac native added when ready. |
| Time to ship | 4 | TUI ships first → fastest cross-platform v1. Mac native is a v2. |
| Accessibility | 5 | TUI accessibility from day one; native accessibility upgrade for Mac users. |
| Local-agent integration | 5 | TUI has the operator seats; Mac native gets the same data over the dashboard contract. |
| Failure modes | 4 | One surface at a time. Lower concurrent maintenance burden. |

**Hybrid C verdict:** The "ship a great TUI first, then add Mac polish" path. Honors the user's "TUIs are the wave" thesis. Minimizes cross-platform regret.

## Score Summary

|  | TUI | Mac Native | CLI+Dash | Hybrid A | Hybrid B | Hybrid C |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| Quality bar | 4 | **5** | 2 | **5** | 3 | 4 |
| Cost efficiency | 4 | 3 | **5** | 2 | 4 | 4 |
| Platform reach | **5** | 1 | **5** | **5** | **5** | **5** |
| Distribution | **5** | 3 | **5** | 3 | 4 | 4 |
| Time to ship | 3 | 3 | **5** | 2 | 3 | 4 |
| Accessibility | 4 | 4 | 3 | **5** | 4 | **5** |
| Local-agent integration | **5** | 4 | 3 | **5** | **5** | **5** |
| Failure modes | 4 | 3 | **5** | 3 | 4 | 4 |
| **Sum** | **34** | **26** | **37** | **30** | **32** | **35** |

**The sums do not decide.** Per codex condition 2, the strategic-frames table below is the primary instrument. The summed scores are visible because hiding them would be dishonest about the dimensions' default weighting — but read in this order:

1. **Strategic frame** (next table) — what is your priority?
2. **Per-dimension reasoning** (track sections above) — does the score reasoning match your context?
3. **Sum** — sanity check only.

The sums weight all eight dimensions equally. Under that weighting, Track 3 (CLI + Dashboard / do nothing) wins. That is the right answer **only** if the strategic frame is "ship nothing new." Under the user's "TUIs are the wave / if we can't build one it calls Sirsi into question" frame, Track 3 fails the strategic test even though it dominates the sum. **The scoring table is the wrong tool for the user's actual question; the strategic-frames table is the right tool.**

## Recommended Picks (Per Strategic Frame)

| If your priority is… | Pick |
| :--- | :--- |
| Highest single-surface quality on the platform you use most | **Track 2 (Mac Native)** |
| "TUIs are the wave; prove we can build one" | **Track 1 (TUI)** or **Hybrid C** |
| Fastest cross-platform interactive surface | **Track 1 (TUI)** or **Hybrid C** |
| Both native ceiling AND reach, willing to pay the cost | **Hybrid A** |
| LEAN cross-platform with native polish later | **Hybrid C** |
| Ship nothing new, return to other work | **Track 3 (CLI+Dashboard)** |

## My Recommendation (Claude-Pantheon, Lane B)

**Hybrid C** — TUI first across all platforms; Mac native ships later as the polish-bar upgrade.

Reasoning:

1. **Honors the user's stated thesis.** "TUIs are the wave… if we can't build one, it calls into question our ability to build Sirsi overall." A great TUI is the proof point. Shipping Mac native first would dodge that proof.
2. **Fastest meaningful cross-platform v1.** TUI is one Go binary on all three OSes. Track 2's Mac-only ceiling locks out 60%+ of devs Pantheon could serve.
3. **LEAN.** Sequential, not concurrent. Track 1's quality bar is the gate; Track 2 follows when the TUI is good enough that "Mac native polish" actually means something. No premature parallelism.
4. **Survives the v0.22 lesson.** The thing we deleted was bad TUI code, not a bad medium. Rebuilding to Mole-grade in the same medium is the honest answer to the user's "if we can't build one…" challenge. Shipping native instead would be conceding the challenge.
5. **Preserves Phase-2 batch-1 work.** The dashboard contract (`docs/DASHBOARD_API.md`, gap, envelope) is surface-independent. The TUI consumes it; Mac native consumes it later. Same JSON, two clients.

**Counter to my own recommendation:** if the user weights *quality ceiling* above *cross-platform reach* and considers Mac the only Pantheon-relevant platform for v1, **Track 2** is the right call and Hybrid C becomes premature parallelism.

The user picks. Codex reviews.

## /goal

User decision on surface track(s). On pick, ADR-020 closes with the chosen direction and Phase-2 batch-2 reshapes accordingly.
