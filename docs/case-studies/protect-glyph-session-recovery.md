# Case Study 018 — The Eye of Horus: How a Crash Became a Protection System

**Date**: April 3-4, 2026
**Severity**: Session-killing — Claude Code terminal destroyed mid-work
**Root Cause**: `KillAll` AppleScript closed every Terminal.app window, including the one running Claude Code
**Resolution**: The ProtectGlyph (`𓂀`) inoculation pattern

---

## What Happened

At 11:55 PM on April 3rd, Ra's multi-window orchestrator was being built. The feature: spawn 4+ terminal windows, each running an autonomous Claude Code agent against a different Sirsi repository. When Ra redeploys, it needs to kill the old agent windows first.

The `KillAll` function did its job too well.

It iterated Terminal.app windows looking for ones matching "Ra:", "claude", or "python3" in the window name. It found the user's own Claude Code session — the one actively building Ra — and killed it. The session crashed. Work in progress was lost. The conversation context was gone.

The irony was precise: the tool being built to orchestrate agents destroyed the agent building it.

---

## The First Fix (Wrong)

The first instinct was positional: "just skip the front window." The code grabbed the frontmost window's TTY identifier and excluded it from the kill loop.

```applescript
set keepTTY to tty of selected tab of front window
-- skip if wTTY is keepTTY
```

This works exactly once. It works when you're sitting in Terminal.app and the Claude Code window happens to be in front. It fails the moment you switch to a browser, or the Command Center steals focus, or you're on a different desktop. Position is not identity.

---

## The Real Fix: Inoculation Over Exclusion

The insight was to flip the logic. Instead of maintaining a growing list of things NOT to kill ("don't kill claude", "don't kill the command center", "don't kill the front window", "don't kill..."), we mark what's protected and kill everything else.

One constant:

```go
const ProtectGlyph = "𓂀"  // Eye of Horus
```

One rule: if a Terminal.app window's custom title contains `𓂀`, it is untouchable. KillAll doesn't check names, doesn't check positions, doesn't maintain exclusion lists. It checks for the glyph.

Two stamps:
1. `ProtectFrontWindow()` — called before deploy, stamps the user's Claude Code session
2. `SpawnWatchWindow()` — the Command Center's title includes `𓂀` from birth

The kill loop became trivial:

```applescript
if cTitle contains "𓂀" then
    -- inoculated, do not touch
else if cTitle contains "Ra:" or wName contains "Ra:" then
    do script "exit" in w
    close w saving no
end if
```

Any future window that needs protection just includes the glyph in its title. No code changes to KillAll. No new exclusion rules. The protection is declarative.

---

## Why This Matters Beyond Terminal Windows

The ProtectGlyph pattern is a microcosm of something we keep learning in infrastructure: **exclusion lists grow until they break; inclusion markers scale forever.**

Firewalls learned this (default-deny beats default-allow). Kubernetes learned this (labels beat name-matching). We learned it at 11:55 PM when our own orchestrator killed the session building it.

The `𓂀` is just a Unicode character in a window title. But the pattern — stamp what's sacred, sweep everything else — is the same pattern that governs The Stele, the governance loop, and eventually the Hedera-backed distributed ledger that will replace this local implementation.

Mark. Verify. Trust the mark.

---

## Technical Details

| Component | Change |
|-----------|--------|
| `ProtectGlyph` | `const ProtectGlyph = "𓂀"` in `internal/ra/terminal.go` |
| `ProtectFrontWindow()` | Stamps current Claude Code window via osascript before deploy |
| `SpawnWatchWindow()` | Command Center title: `𓂀 𓇶 Ra Command Center` |
| `KillAll()` | Single `contains "𓂀"` check replaces all exclusion logic |
| `buildTerminalScript()` | Agent windows now `; exit` on completion (auto-close) |

**Commits**: `034f139`, `1fc206a`
**Version**: v0.10.0
**Rule created**: Rule A25 candidate — ProtectGlyph Protocol

---

*The best safety systems don't try to enumerate danger. They mark what's safe and assume everything else is fair game.*
