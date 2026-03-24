package guard

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WatchConfig configures the Sekhmet watchdog.
type WatchConfig struct {
	Interval     time.Duration // How often to poll (default: 5s)
	CPUThreshold float64       // Alert when a process exceeds this % (default: 80.0)
	DurationSecs int           // Alert only if sustained for this many consecutive checks (default: 3)
	MaxAlerts    int           // Stop after this many alerts (0 = unlimited)
}

// DefaultWatchConfig returns sensible defaults for the watchdog.
func DefaultWatchConfig() WatchConfig {
	return WatchConfig{
		Interval:     5 * time.Second,
		CPUThreshold: 80.0,
		DurationSecs: 3,
		MaxAlerts:    0,
	}
}

// WatchAlert is emitted when a process exceeds the CPU threshold.
type WatchAlert struct {
	Process    ProcessInfo
	CPUPercent float64
	Duration   time.Duration // How long it's been sustained
	Timestamp  time.Time
}

// AlertFunc is the callback for watch alerts.
type AlertFunc func(alert WatchAlert)

// Watch monitors system processes for CPU/memory pressure.
// It calls onAlert when a process sustains CPU > threshold for duration checks.
// It blocks until the context is cancelled.
func Watch(ctx context.Context, cfg WatchConfig, onAlert AlertFunc) error {
	if cfg.Interval == 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.CPUThreshold == 0 {
		cfg.CPUThreshold = 80.0
	}
	if cfg.DurationSecs == 0 {
		cfg.DurationSecs = 3
	}

	// Track how many consecutive checks each PID has been hot
	hotStreak := make(map[int]int)       // PID -> consecutive checks above threshold
	firstSeen := make(map[int]time.Time) // PID -> when first exceeded threshold
	alertCount := 0

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			procs, err := getCPUSnapshot()
			if err != nil {
				continue // Transient error — try again next tick
			}

			// Track hot processes
			currentHot := make(map[int]bool)
			for _, p := range procs {
				if p.CPUPercent >= cfg.CPUThreshold {
					currentHot[p.PID] = true
					hotStreak[p.PID]++

					if _, exists := firstSeen[p.PID]; !exists {
						firstSeen[p.PID] = time.Now()
					}

					// Alert if sustained for enough consecutive checks
					if hotStreak[p.PID] >= cfg.DurationSecs && onAlert != nil {
						duration := time.Since(firstSeen[p.PID])
						onAlert(WatchAlert{
							Process:    p,
							CPUPercent: p.CPUPercent,
							Duration:   duration,
							Timestamp:  time.Now(),
						})
						alertCount++
						// Reset streak so we don't spam — re-alert if it continues
						hotStreak[p.PID] = 0

						if cfg.MaxAlerts > 0 && alertCount >= cfg.MaxAlerts {
							return nil
						}
					}
				}
			}

			// Reset streaks for processes that cooled down
			for pid := range hotStreak {
				if !currentHot[pid] {
					delete(hotStreak, pid)
					delete(firstSeen, pid)
				}
			}
		}
	}
}

// getCPUSnapshot returns a quick snapshot of top CPU consumers.
func getCPUSnapshot() ([]ProcessInfo, error) {
	// Use ps sorted by CPU, top 20 only
	out, err := exec.Command("ps", "-arcxo", "pid,rss,%cpu,comm").Output()
	if err != nil {
		return nil, err
	}

	var procs []ProcessInfo
	lines := strings.Split(string(out), "\n")

	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pid, _ := strconv.Atoi(fields[0])
		rss, _ := strconv.ParseInt(fields[1], 10, 64)
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		name := strings.Join(fields[3:], " ")

		procs = append(procs, ProcessInfo{
			PID:        pid,
			Name:       name,
			RSS:        rss * 1024, // ps reports in KB
			CPUPercent: cpu,
		})

		// Only track top 20 — no need to scan everything
		if len(procs) >= 20 {
			break
		}
	}

	sort.Slice(procs, func(i, j int) bool {
		return procs[i].CPUPercent > procs[j].CPUPercent
	})

	return procs, nil
}

// FormatAlert formats a WatchAlert for terminal display.
func FormatAlert(a WatchAlert) string {
	return fmt.Sprintf("⚠️  𓁵 SEKHMET ALERT: %s (PID %d) at %.1f%% CPU for %s — using %s RAM",
		a.Process.Name, a.Process.PID, a.CPUPercent,
		a.Duration.Truncate(time.Second), FormatBytes(a.Process.RSS))
}
