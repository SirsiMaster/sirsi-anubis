// Package guard — antigravity.go
//
// Antigravity IPC Bridge: Connects the Sekhmet watchdog to MCP consumers.
//
// Problem: IDE Plugin Host CPU starvation is detected by the watchdog but
// the alert stays in the Go process — no way to surface it to the IDE/AI.
//
// Solution: A thread-safe alert ring buffer that MCP resources/tools can query.
// The bridge runs alongside the watchdog, draining its alert channel and
// storing alerts in a bounded, lock-free-read ring buffer.
//
// Architecture:
//
//	┌──────────┐     ┌────────────┐     ┌──────────┐     ┌─────────┐
//	│ Watchdog │────▶│ Bridge     │────▶│  Ring     │────▶│ MCP     │
//	│ (alerts) │     │ (consumer) │     │  Buffer   │     │ resource│
//	└──────────┘     └────────────┘     └──────────┘     └─────────┘
//	                       │
//	                       ├──▶ IDE notification (future)
//	                       └──▶ Log output
package guard

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AlertRingSize is the max alerts retained in the ring buffer.
const AlertRingSize = 64

// AlertEntry is a serializable alert record for MCP consumers.
type AlertEntry struct {
	ProcessName string  `json:"process_name"`
	PID         int     `json:"pid"`
	CPUPercent  float64 `json:"cpu_percent"`
	RSSB        int64   `json:"rss_bytes"`
	RSSHuman    string  `json:"rss_human"`
	Duration    string  `json:"duration"`
	Timestamp   string  `json:"timestamp"`
	Severity    string  `json:"severity"` // "warning", "critical"
}

// AlertRing is a thread-safe bounded ring buffer of recent alerts.
type AlertRing struct {
	mu      sync.RWMutex
	entries []AlertEntry
	head    int
	count   int
	total   int64 // lifetime total alerts received
}

// NewAlertRing creates an empty ring buffer.
func NewAlertRing() *AlertRing {
	return &AlertRing{
		entries: make([]AlertEntry, AlertRingSize),
	}
}

// Push adds an alert to the ring buffer, overwriting the oldest if full.
func (r *AlertRing) Push(entry AlertEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[r.head] = entry
	r.head = (r.head + 1) % AlertRingSize
	if r.count < AlertRingSize {
		r.count++
	}
	r.total++
}

// Recent returns the N most recent alerts, newest first.
func (r *AlertRing) Recent(n int) []AlertEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if n <= 0 || r.count == 0 {
		return nil
	}
	if n > r.count {
		n = r.count
	}

	result := make([]AlertEntry, n)
	for i := 0; i < n; i++ {
		// Walk backwards from head
		idx := (r.head - 1 - i + AlertRingSize) % AlertRingSize
		result[i] = r.entries[idx]
	}
	return result
}

// Stats returns ring buffer statistics.
func (r *AlertRing) Stats() (current, lifetime int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count, int(r.total)
}

// ── Bridge ──────────────────────────────────────────────────────────────

// AntigravityBridge connects a watchdog to the alert ring buffer.
type AntigravityBridge struct {
	watchdog *Watchdog
	ring     *AlertRing
	ctx      context.Context
	cancel   context.CancelFunc
	stopped  chan struct{}

	// Callback for custom alert handling (e.g., logging, MCP notification).
	OnAlert func(AlertEntry)
}

// BridgeConfig configures the Antigravity bridge.
type BridgeConfig struct {
	WatchConfig WatchConfig
	CPUCritical float64 // CPU% threshold for "critical" severity (default: 150.0)
	OnAlert     func(AlertEntry)
}

// DefaultBridgeConfig returns sensible defaults for IDE protection.
func DefaultBridgeConfig() BridgeConfig {
	wc := DefaultWatchConfig()
	wc.CPUThreshold = 60.0 // Lower threshold for IDE — catch Plugin Host early
	wc.SustainCount = 1    // Trigger on first violation for snappier response
	wc.Interval = 800 * time.Millisecond
	return BridgeConfig{
		WatchConfig: wc,
		CPUCritical: 120.0, // Mark critical earlier
	}
}

// StartBridge launches the watchdog and the bridge consumer.
// Returns the bridge handle. Ring buffer is queryable immediately.
func StartBridge(ctx context.Context, cfg BridgeConfig) *AntigravityBridge {
	if cfg.CPUCritical <= 0 {
		cfg.CPUCritical = 150.0
	}

	bCtx, cancel := context.WithCancel(ctx)
	ring := NewAlertRing()
	watchdog := StartWatch(bCtx, cfg.WatchConfig)

	b := &AntigravityBridge{
		watchdog: watchdog,
		ring:     ring,
		ctx:      bCtx,
		cancel:   cancel,
		stopped:  make(chan struct{}),
		OnAlert:  cfg.OnAlert,
	}

	go b.consume(cfg.CPUCritical)

	return b
}

// Ring returns the alert ring buffer for MCP consumers to read.
func (b *AntigravityBridge) Ring() *AlertRing {
	return b.ring
}

// Watchdog returns the underlying watchdog instance.
func (b *AntigravityBridge) Watchdog() *Watchdog {
	return b.watchdog
}

// Stop shuts down both the watchdog and the bridge.
func (b *AntigravityBridge) Stop() {
	b.cancel()
	<-b.stopped
}

// consume drains the watchdog alert channel and pushes to the ring buffer.
func (b *AntigravityBridge) consume(cpuCritical float64) {
	defer close(b.stopped)

	for alert := range b.watchdog.Alerts() {
		severity := "warning"
		if alert.CPUPercent >= cpuCritical {
			severity = "critical"
		}

		entry := AlertEntry{
			ProcessName: alert.Process.Name,
			PID:         alert.Process.PID,
			CPUPercent:  alert.CPUPercent,
			RSSB:        alert.Process.RSS,
			RSSHuman:    FormatBytes(alert.Process.RSS),
			Duration:    alert.Duration.Truncate(time.Second).String(),
			Timestamp:   alert.Timestamp.Format(time.RFC3339),
			Severity:    severity,
		}

		b.ring.Push(entry)

		if b.OnAlert != nil {
			b.OnAlert(entry)
		}
	}
}

// ── MCP Serialization ──────────────────────────────────────────────────

// BridgeStatus is the JSON payload for the MCP watchdog-alerts resource.
type BridgeStatus struct {
	Active           bool         `json:"active"`
	RecentAlerts     []AlertEntry `json:"recent_alerts"`
	BufferedCount    int          `json:"buffered_count"`
	LifetimeAlerts   int          `json:"lifetime_alerts"`
	WatchdogPolls    int64        `json:"watchdog_polls"`
	WatchdogBackoffs int64        `json:"watchdog_backoffs"`
}

// Status returns a complete bridge status snapshot for MCP.
func (b *AntigravityBridge) Status() *BridgeStatus {
	buffered, lifetime := b.ring.Stats()
	polls, _, backoffs := b.watchdog.Stats()

	return &BridgeStatus{
		Active:           b.watchdog.IsRunning(),
		RecentAlerts:     b.ring.Recent(10),
		BufferedCount:    buffered,
		LifetimeAlerts:   lifetime,
		WatchdogPolls:    polls,
		WatchdogBackoffs: backoffs,
	}
}

// StatusJSON returns the bridge status as formatted JSON.
func (b *AntigravityBridge) StatusJSON() (string, error) {
	status := b.Status()
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal bridge status: %w", err)
	}
	return string(data), nil
}
