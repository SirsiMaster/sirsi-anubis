# Antigravity IDE Bug Report: Plugin Host IPC Starvation Under Extended Sessions

**Filed by:** Sirsi Pantheon Diagnostics (Guard + Yield modules)
**Date:** 2026-03-23
**Severity:** P2 — UI becomes unresponsive
**Product:** Antigravity IDE (Electron-based)
**Platform:** macOS 15.x, Apple M1 Max, 32 GB RAM

---

## Summary

During extended AI agent sessions (>15 minutes), Antigravity's Plugin Host processes
sustain 100%+ CPU indefinitely, starving the Renderer process and making UI elements
(including permission dialogs) unclickable. The system has 88% free RAM — this is
purely CPU/IPC contention, not memory pressure.

## Steps to Reproduce

1. Open Antigravity IDE
2. Start an AI agent conversation (e.g., code generation, multi-file editing)
3. Allow the conversation to run for 15+ minutes with continuous agent activity
4. Observe that UI buttons (specifically "Allow This Conversation" permission dialog)
   become unresponsive to clicks
5. Check Activity Monitor: Plugin Host processes show sustained 100%+ CPU

## Expected Behavior

- Plugin Host CPU usage should be **bursty**: high during active tool calls, idle between
- The Renderer process should always have priority for UI event handling
- Permission dialogs should remain responsive regardless of background agent load

## Actual Behavior

### Process Snapshot (captured during incident)

```
PID     RSS      CPU     Process
37363   787 MB   103.9%  Antigravity Helper (Plugin)    ← Agent Worker A
37364   633 MB    76.6%  Antigravity Helper (Plugin)    ← Agent Worker B  
36974   788 MB    43.2%  Antigravity Helper (Renderer)  ← UI Paint Loop
36971   156 MB    24.6%  Antigravity Helper (GPU)
37425  1677 MB    15.7%  language_server_macos_arm       ← LSP Server
37423   824 MB     0.3%  language_server_macos_arm       ← LSP Server
36967   411 MB     0.6%  Electron (Main)
```

**Total Antigravity footprint:** ~7.2 GB RAM, ~264% CPU across 32 processes

### System Context

```
Total RAM:        32 GB
Free RAM:         88% (NOT a memory issue)
Swap Used:        253 MB / 1 GB (negligible)
Load Average:     3.08 / 3.63 / 4.29 (10-core M1 Max)
Compressions:     4,166,290
Decompressions:   2,170,517
```

### Process Tree Analysis

```
PID 36967 — Electron Main Process (0.6% CPU)
├── PID 36974 — Renderer (43.2% CPU) ← UI STARVED
├── PID 37363 — Plugin Host A (103.9% CPU) ← SUSTAINED
│   ├── PID 37423 — Language Server (0.3% CPU)
│   ├── PID 37366, 38418, 39135 — Plugin children
├── PID 37364 — Plugin Host B (76.6% CPU) ← SUSTAINED
│   └── PID 37425 — Language Server (15.7% CPU)
└── 6 other helper processes
```

Both Plugin Hosts had been running for **17 minutes 29 seconds** at capture time.

## Root Cause Analysis

This appears to be **Electron main-thread / IPC bus starvation**:

1. Plugin Host processes run in separate Electron processes (correct architecture)
2. However, click events on UI elements like "Allow This Conversation" are IPC messages
   from the Renderer process → Plugin Host process
3. When Plugin Hosts are at 100%+ CPU, the IPC message queue builds up
4. Click events are received but not processed — the button appears "broken"
5. Rapid clicking by the user floods the IPC bus with duplicate events, making it worse

## Why It's Not RAM

The initial user hypothesis was "out of RAM." This is incorrect because:
- System reports **88% free memory**
- Swap usage is only 253 MB on a 1 GB swap file (25%)
- macOS memory compressor is active but not under pressure
- No `kern.memorystatus.kill_on_sustained_pressure_count` kills occurred

The actual bottleneck is **CPU saturation of the Plugin Host processes**, which
cascades into IPC starvation of the Renderer.

## Suggested Fix

1. **Plugin Host CPU Yielding**: Plugin Hosts should yield CPU between tool call
   executions. If the agent work is complete and waiting for the next user/model
   message, CPU usage should drop to near-zero.

2. **IPC Priority for UI Events**: Permission dialogs and security-critical UI
   elements should use a high-priority IPC channel that bypasses the standard
   message queue. A user unable to click "Allow" is a security UX failure.

3. **Plugin Host Timeout**: If a Plugin Host sustains >80% CPU for >5 minutes
   without active tool calls, the IDE should warn the user and offer to restart
   the plugin worker.

4. **Process Affinity**: On multi-core systems (M1 Max has 10 cores), Plugin
   Hosts could be pinned to efficiency cores, leaving performance cores free
   for the Renderer and user-facing operations.

## Environment Details

- **IDE:** Antigravity (version from screenshot: latest as of 2026-03-23)
- **OS:** macOS 15.x (Apple Silicon)
- **CPU:** Apple M1 Max (10 cores: 8P + 2E)
- **RAM:** 32 GB unified memory
- **GPU:** Apple M1 Max (32 cores, Metal 4)
- **Active Extensions:** Antigravity AI agent, Language Server

## Diagnostic Method

This issue was diagnosed using [Sirsi Pantheon](https://github.com/SirsiMaster/sirsi-pantheon),
an open-source workstation hygiene and diagnostic tool. Specifically:

- **Guard module**: Process auditing and RAM analysis
- **Yield module**: CPU pressure detection via load average analysis (ADR-006)
- **Platform module**: macOS-specific system metric collection

The user initially blamed RAM pressure because that is the most commonly cited
cause of IDE slowdowns. Pantheon's diagnostics proved it was CPU/IPC contention —
a distinction that matters for the fix.

---

**Discovery Credit:** Sirsi Pantheon v0.4.0-alpha — Guard + Yield diagnostic modules
**Reporter:** Sirsi Technologies (github.com/SirsiMaster)
