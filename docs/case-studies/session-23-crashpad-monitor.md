# Case Study 012 — The Crashpad Monitor: Why Your IDE Doesn't Tell You It's Dying

**Date**: March 26, 2026  
**Module**: `extensions/vscode/src/crashpadMonitor.ts`  
**Lines**: 370+  
**Status**: Shipped in v0.7.0-alpha

---

## The Problem No One Monitors

Every VS Code-based IDE (VS Code, Cursor, Windsurf, Antigravity) is an Electron application. When a subprocess crashes — the Extension Host, a renderer, the GPU process — the IDE generates a **minidump** (`.dmp` file) and drops it in a `Crashpad/pending/` directory.

These dumps **never get cleaned up.**

They sit in your Application Support folder forever, silently accumulating. Nobody looks at them. No extension monitors them. No IDE surface shows you that your Extension Host has crashed 34 times in the last month.

**Until the day it all collapses.**

---

## What Happened (Session 22–23)

We patched two `package.json` files inside the Antigravity IDE's application bundle. Both patches were valid JSON. Both passed schema validation. Neither changed code — just manifest properties.

The Extension Host couldn't bind the declared commands to handlers. It entered a validation loop. The loop leaked memory. V8 ran out of heap `(electron.v8-oom.is_heap_oom)`. macOS killed the process. The app restarted. Same crash. macOS killed it again. **Two reinstalls and two restarts later, the IDE finally worked.**

The only evidence? **34 `.dmp` files** sitting in `~/Library/Application Support/Antigravity/Crashpad/pending/`, dating back weeks. The crash we triggered was just the most dramatic — the IDE had been chronically unstable for a month. Nobody noticed.

---

## What the Crashpad Monitor Does

The Crashpad Monitor is a module inside Pantheon's VS Code extension that does what no other extension does: **it watches the crash dump directory.**

### Core Capabilities

| Feature | How It Works |
|---------|-------------|
| **Auto-Detection** | Finds `Crashpad/` for Antigravity, VS Code, Cursor, or Windsurf automatically |
| **Pending Count** | Reads `pending/*.dmp` every 5 minutes via `fs.readdirSync` |
| **Trend Detection** | Tracks a 3-reading rolling window — classifies as `stable`, `growing`, or `critical` |
| **Extension Host ID** | Reads the first 8KB of recent dumps looking for `extensionHost` process type |
| **Status Bar** | Hidden when stable — 🟡 at 5+ dumps — 🔴 at 15+ |
| **Webview Report** | Premium dark-themed report with timeline, forensics commands, and recommendations |
| **Dump Cleanup** | Confirmation modal → deletes stale `.dmp` files you'll never need |
| **Session Warning** | One-time notification when trend shifts from stable to growing/critical |

### What Makes It Unique

1. **No other VS Code extension monitors Crashpad.** We checked. Extensions monitor CPU, memory, network. Nobody watches the crash dump directory.

2. **It's a leading indicator.** A growing dump count means your IDE is silently dying. Not today — but eventually. Without this monitor, you don't know until it's too late.

3. **Extension Host crash detection.** The monitor reads the first 8KB of recent dumps and checks for `VSCODE_CRASH_REPORTER_PROCESS_TYPE=extensionHost`. This is the most dangerous crash type — the one that caused our cascade.

4. **It doesn't try to fix anything.** It just tells you the truth. The Crashpad Monitor is an early warning system, not a repair tool. That's the lesson we learned from Rule A19 — trying to fix things inside the IDE bundle is what caused the crash in the first place.

---

## Architecture

```
CrashpadMonitor (5-minute polling)
│
├── detectCrashpadPath()
│   ├── ~/Library/Application Support/Antigravity/Crashpad
│   ├── ~/Library/Application Support/Code/Crashpad
│   ├── ~/Library/Application Support/Cursor/Crashpad
│   └── ~/Library/Application Support/Windsurf/Crashpad
│
├── checkCrashpad()
│   ├── Count .dmp files in pending/
│   ├── Categorize: recent (24h) vs stale
│   ├── Read first 8KB of recent dumps → detect Extension Host crashes
│   ├── Update rolling trend history (3 readings)
│   └── Classify: stable / growing / critical
│
├── updateStatusBar()
│   ├── Hide (stable, <3 dumps)
│   ├── 🟡 Warning (5+ dumps or growing trend)
│   └── 🔴 Critical (15+ dumps)
│
└── showReport() → Premium webview
    ├── Stats cards (pending, recent, trend)
    ├── Extension Host crash alert
    ├── Timeline (oldest/newest dump)
    └── Forensics reference (strings commands)
```

---

## Why It Matters

Most developers think IDE crashes are random. They are — in the moment. But the **pattern** is never random. A growing Crashpad directory tells a story:

- 5 dumps in a month? Normal. Extensions crash sometimes.
- 15 dumps? Something is chronically wrong. Probably an extension leaking memory.
- 34 dumps? Your IDE is in a death spiral. The next crash might require a reinstall.

Without the Crashpad Monitor, the first you hear about this is when your IDE won't launch.

---

## Verification

```bash
# Check your own Crashpad
ls ~/Library/Application\ Support/Antigravity/Crashpad/pending/*.dmp 2>/dev/null | wc -l

# Extract process type from a dump
strings <dump_file> | grep "VSCODE_CRASH_REPORTER_PROCESS_TYPE"

# Check for V8 memory exhaustion
strings <dump_file> | grep "electron.v8-oom"
```

---

*Born from Case Study 011. Built because nobody else will watch.*
