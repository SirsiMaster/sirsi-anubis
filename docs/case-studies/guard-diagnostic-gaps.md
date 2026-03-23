# Guard Diagnostic Coverage Gap Analysis

**Current State (v0.3.0-alpha) → Required State (v0.5.0)**

Guard currently captures RAM and process information but misses five critical
pressure domains that would have caught the Antigravity IPC issue immediately.

## Current Coverage

| Metric | Captured? | Source | Module |
|:---|:---:|:---|:---|
| Total / Used / Free RAM | ✅ | vm_stat, sysctl | guard/audit.go |
| Per-process RSS (memory) | ✅ | ps -axo rss | guard/audit.go |
| Per-process VSZ (virtual) | ✅ | ps -axo vsz | guard/audit.go |
| Per-process CPU % | ✅ | ps -axo %cpu | guard/audit.go |
| Process grouping | ✅ | heuristic classifier | guard/audit.go |
| Orphan detection | ✅ | RSS > 50 MB threshold | guard/audit.go |
| Load average check | ✅ | sysctl vm.loadavg | yield/yield.go (NEW) |

## Missing Pressure Domains

### 1. 🔴 CPU Pressure (CRITICAL — caused this incident)
| Metric | Source (macOS) | Source (Linux) |
|:---|:---|:---|
| System load vs core ratio | `sysctl vm.loadavg` vs `runtime.NumCPU()` | `/proc/loadavg` |
| Per-core utilization | `top -l 1 -stats cpu` | `/proc/stat` |
| Sustained high CPU detection | Poll every 5s, flag >80% for >30s | Same |
| CPU pressure stall info | N/A on macOS | `/proc/pressure/cpu` |

**Impact**: Would have immediately flagged the Plugin Host processes at 104% sustained CPU.

### 2. 🟡 Swap / Memory Pressure
| Metric | Source (macOS) | Source (Linux) |
|:---|:---|:---|
| Swap total / used / free | `sysctl vm.swapusage` | `/proc/swaps` |
| Swap I/O rate (swapins/outs) | `vm_stat` (Swapins/Swapouts) | `/proc/vmstat` |
| Compressor ratio | `vm_stat` (compressed/occupied) | N/A |
| Memory pressure level | `memory_pressure` command | `/proc/pressure/memory` |
| Kernel kill count | `sysctl kern.memorystatus.kill_on_sustained_pressure_count` | OOM killer log |

**Impact**: Would have proven RAM was NOT the issue (88% free, minimal swap).

### 3. 🟡 I/O Pressure (Disk)
| Metric | Source (macOS) | Source (Linux) |
|:---|:---|:---|
| Disk read/write bytes/sec | `iostat -d` | `/proc/diskstats` |
| I/O wait percentage | `iostat -c` | `/proc/stat` (iowait) |
| Per-process I/O | `ioreg` / `fs_usage` | `/proc/[pid]/io` |
| I/O pressure stall info | N/A | `/proc/pressure/io` |

**Impact**: Would rule out disk I/O as a contributor to IDE lag.

### 4. 🟡 Network Pressure
| Metric | Source (macOS) | Source (Linux) |
|:---|:---|:---|
| Bytes in/out per interface | `netstat -ib` | `/proc/net/dev` |
| Active connections count | `netstat -an \| wc -l` | `ss -s` |
| DNS resolution latency | `dig +time=2 google.com` | Same |
| Per-process network I/O | `nettop` (macOS only) | `/proc/[pid]/net/dev` |

**Impact**: Would identify if agent processes are making excessive API calls.

### 5. 🔴 IPC Pressure (CRITICAL — root cause of button failure)
| Metric | Source (macOS) | Source (Linux) |
|:---|:---|:---|
| Mach port count per process | `lsmp -p [pid]` | N/A |
| IPC message queue depth | Not directly observable | `/proc/sysvipc/msg` |
| Pipe buffer utilization | `lsof -p [pid] \| grep PIPE` | `/proc/[pid]/fdinfo` |
| XPC/IPC connection count | `launchctl print system` | N/A |

**Impact**: Would have directly identified the Electron IPC bus saturation between
Plugin Hosts and Renderer. This is the hardest to instrument but most valuable for
IDE-specific diagnostics.

## Implementation Priority

| Domain | Priority | ADR | Effort |
|:---|:---:|:---:|:---:|
| CPU Pressure | P0 | ADR-006 | 2 hours |
| Swap/Memory Pressure | P1 | ADR-006 | 1 hour |
| I/O Pressure | P2 | — | 2 hours |
| Network Pressure | P3 | — | 3 hours |
| IPC Pressure | P3 | — | 4 hours (macOS-specific) |

## The Lesson

> "I blamed RAM all this time."

The user's instinct was wrong — and that's exactly why Pantheon needs to capture ALL
five pressure domains. Developers default to "out of RAM" because it's the most
commonly discussed bottleneck. But modern systems (especially Apple Silicon with
unified memory and aggressive compression) rarely run out of RAM. The real killers
are CPU contention, IPC starvation, and I/O wait — and without proper diagnostics,
users will never know the difference.

**Pantheon's value proposition**: We don't just tell you "your system is slow."
We tell you *which pressure domain* is the bottleneck and *which deity* can help.
