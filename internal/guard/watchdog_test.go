package guard

import (
	"context"
	"testing"
	"time"
)

func TestDefaultWatchConfig(t *testing.T) {
	cfg := DefaultWatchConfig()
	if cfg.Interval != 5*time.Second {
		t.Errorf("Interval = %v, want 5s", cfg.Interval)
	}
	if cfg.CPUThreshold != 80.0 {
		t.Errorf("CPUThreshold = %.1f, want 80.0", cfg.CPUThreshold)
	}
	if cfg.DurationSecs != 3 {
		t.Errorf("DurationSecs = %d, want 3", cfg.DurationSecs)
	}
}

func TestGetCPUSnapshot(t *testing.T) {
	procs, err := getCPUSnapshot()
	if err != nil {
		t.Fatalf("getCPUSnapshot: %v", err)
	}
	if len(procs) == 0 {
		t.Log("no processes returned (CI?)")
	}
	// Verify sorted by CPU descending
	for i := 1; i < len(procs); i++ {
		if procs[i].CPUPercent > procs[i-1].CPUPercent {
			t.Errorf("processes not sorted by CPU: [%d]%.1f > [%d]%.1f",
				i, procs[i].CPUPercent, i-1, procs[i-1].CPUPercent)
		}
	}
	// Should have at most 20
	if len(procs) > 20 {
		t.Errorf("expected <= 20 procs, got %d", len(procs))
	}
}

func TestWatch_ImmediateCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cfg := DefaultWatchConfig()
	cfg.Interval = 50 * time.Millisecond

	err := Watch(ctx, cfg, func(a WatchAlert) {
		t.Error("should not alert on immediate cancel")
	})
	if err != context.Canceled {
		t.Errorf("Watch error = %v, want context.Canceled", err)
	}
}

func TestWatch_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cfg := WatchConfig{
		Interval:     50 * time.Millisecond,
		CPUThreshold: 99999.0, // Unreachable — no alerts
		DurationSecs: 1,
	}

	err := Watch(ctx, cfg, func(a WatchAlert) {
		t.Error("threshold too high to trigger")
	})
	if err != context.DeadlineExceeded {
		t.Errorf("Watch error = %v, want DeadlineExceeded", err)
	}
}

func TestWatch_ZeroConfig_Defaults(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Zero config should apply defaults and not panic
	cfg := WatchConfig{}
	err := Watch(ctx, cfg, nil)
	if err != context.DeadlineExceeded {
		t.Errorf("Watch error = %v, want DeadlineExceeded", err)
	}
}

func TestFormatAlert(t *testing.T) {
	alert := WatchAlert{
		Process: ProcessInfo{
			PID:        1234,
			Name:       "test-proc",
			RSS:        512 * 1024 * 1024,
			CPUPercent: 95.3,
		},
		CPUPercent: 95.3,
		Duration:   15 * time.Second,
		Timestamp:  time.Now(),
	}
	s := FormatAlert(alert)
	if s == "" {
		t.Error("FormatAlert returned empty string")
	}
	if !containsAll(s, "SEKHMET", "test-proc", "1234", "95.3", "15s") {
		t.Errorf("FormatAlert missing expected content: %s", s)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
