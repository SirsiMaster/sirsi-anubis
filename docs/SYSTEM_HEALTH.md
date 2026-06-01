# System Health Notes — M5 Max

Quick reference for diagnosing Mac "lockups / system failures."
First stop, always: **`sirsi diagnose`** — reads Jetsam events + kernel panics directly.
Since 2026-05-31, every Claude session also prints a one-line `health:` score at start.

---

## Reading a kernel panic

Panic reports live in `/Library/Logs/DiagnosticReports/*.panic` (and `.ips`).
The `panic_string` names the cause. Match it below.

### `watchdog timeout: no checkins from watchdogd in N seconds`
Caller `AppleARMWatchdogTimer`. The kernel stopped getting heartbeats and hard-reset to recover.
**This is almost never RAM or your dev work.** Prime suspects, in order:

1. **Display / Thunderbolt / dock stall** — a flaky TB5 dock, cable, or hub that wedges the
   display pipeline. **Most likely culprit given the hardware-rig work** (PCIe 5.0 / TB5 test
   rigs, pucks, receivers). Correlate panic timestamps against hardware-testing sessions.
2. **GPU driver hang** — external GPU path or a heavy Metal/render workload stalling the GPU.
3. **Firmware / SMC hiccup** — transient; clears on reboot.

**Action if it recurs:** unplug external displays / TB docks and run a day bare. If the panics
stop, it's the dock/cable path — swap the cable first (cheapest fix), then the dock.

### `Jetsam` / `memorystatus` kills
macOS force-killing processes under a RAM spike. Victims like `zsh`, `weatherd`, `wifid` =
transient system-daemon spike, not a leak. Victims that are *your* big apps (Chrome helpers,
Codex, multiple `claude`) repeatedly = genuine memory pressure — close duplicate heavy apps.
Check live: `sirsi diagnose` (RAM Pressure + Jetsam Events rows), `sysctl vm.swapusage`.

---

## Baseline (healthy)
- 48 GB RAM, typically ~90% free, **zero swap** at rest.
- Clean uptime is weeks (e.g. Apr 16 → May 31 = 6 weeks, one reboot).
- A *cluster* of reboots in a single day = a bad-driver/thermal day, usually self-resolving.

## 2026-05-31 incident (resolved)
4 reboots in one afternoon (15:23, 18:37, 19:11, 20:17) after 6 clean weeks.
- 18:37: **watchdog-timeout kernel panic** (AppleARMWatchdogTimer) — display/driver stall.
- 15:08: JetsamEvent, victims were system daemons (zsh/wifid/weatherd) — transient spike.
- No swap, RAM healthy throughout. Not memory exhaustion. Cluster passed; no recurrence.
- Root mitigation: wired `sirsi diagnose` into Claude SessionStart so future clusters are
  visible immediately instead of going undiagnosed. See
  `~/.claude/hooks/health-line.sh` and the Pantheon Health Surface memory.
