package guard

import (
	"context"
	"strings"
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
	if cfg.SustainCount != 3 {
		t.Errorf("SustainCount = %d, want 3", cfg.SustainCount)
	}
	if cfg.SampleSize != 15 {
		t.Errorf("SampleSize = %d, want 15", cfg.SampleSize)
	}
	if cfg.SelfBudget != 5.0 {
		t.Errorf("SelfBudget = %.1f, want 5.0", cfg.SelfBudget)
	}
}

func TestApplyDefaults_ZeroConfig(t *testing.T) {
	cfg := WatchConfig{}
	applyDefaults(&cfg)
	if cfg.Interval == 0 {
		t.Error("Interval should be set")
	}
	if cfg.CPUThreshold == 0 {
		t.Error("CPUThreshold should be set")
	}
	if cfg.SustainCount == 0 {
		t.Error("SustainCount should be set")
	}
	if cfg.SampleSize == 0 {
		t.Error("SampleSize should be set")
	}
	if cfg.SelfBudget == 0 {
		t.Error("SelfBudget should be set")
	}
}

func TestSampleTopCPU(t *testing.T) {
	procs, err := sampleTopCPU(10)
	if err != nil {
		t.Fatalf("sampleTopCPU: %v", err)
	}
	if len(procs) == 0 {
		t.Log("no processes returned (CI?)")
		return
	}
	if len(procs) > 10 {
		t.Errorf("expected <= 10 procs, got %d", len(procs))
	}
	// Verify sorted by CPU descending
	for i := 1; i < len(procs); i++ {
		if procs[i].CPUPercent > procs[i-1].CPUPercent {
			t.Errorf("not sorted: [%d]%.1f > [%d]%.1f",
				i, procs[i].CPUPercent, i-1, procs[i-1].CPUPercent)
		}
	}
}

func TestSampleTopCPU_DefaultN(t *testing.T) {
	procs, err := sampleTopCPU(0) // Should default to 15
	if err != nil {
		t.Fatalf("sampleTopCPU: %v", err)
	}
	if len(procs) > 15 {
		t.Errorf("default should cap at 15, got %d", len(procs))
	}
}

func TestStartWatch_AndStop(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultWatchConfig()
	cfg.Interval = 50 * time.Millisecond
	cfg.CPUThreshold = 99999.0 // Unreachable

	w := StartWatch(ctx, cfg)
	if !w.IsRunning() {
		t.Error("watchdog should be running")
	}

	// Let it poll a couple times
	time.Sleep(150 * time.Millisecond)

	w.Stop()
	if w.IsRunning() {
		t.Error("watchdog should be stopped")
	}

	polls, alerts, _ := w.Stats()
	if polls == 0 {
		t.Error("should have polled at least once")
	}
	if alerts != 0 {
		t.Errorf("should have 0 alerts with unreachable threshold, got %d", alerts)
	}
}

func TestStartWatch_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := DefaultWatchConfig()
	cfg.Interval = 50 * time.Millisecond

	w := StartWatch(ctx, cfg)
	time.Sleep(100 * time.Millisecond)

	cancel()     // Cancel should stop the watchdog
	<-w.Alerts() // Drain until closed

	if w.IsRunning() {
		t.Error("watchdog should stop on context cancel")
	}
}

func TestStartWatch_AlertChannel_NonBlocking(t *testing.T) {
	// Verify the alert channel has a buffer and doesn't block
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultWatchConfig()
	cfg.Interval = 50 * time.Millisecond
	cfg.CPUThreshold = 0.0 // Everything triggers
	cfg.SustainCount = 1   // Immediate
	cfg.MaxAlerts = 3      // Stop after 3

	w := StartWatch(ctx, cfg)

	// Don't consume alerts — verify the watchdog doesn't deadlock
	time.Sleep(300 * time.Millisecond)

	w.Stop() // Should not hang even if we never read alerts
}

// Legacy API compatibility
func TestWatch_ImmediateCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

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
		CPUThreshold: 99999.0,
		SustainCount: 1,
	}

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
	for _, want := range []string{"SEKHMET", "test-proc", "1234", "95.3", "15s"} {
		if !strings.Contains(s, want) {
			t.Errorf("FormatAlert missing %q in: %s", want, s)
		}
	}
}

func TestMin64(t *testing.T) {
	if min64(5*time.Second, 10*time.Second) != 5*time.Second {
		t.Error("min64 should return smaller")
	}
	if min64(30*time.Second, 10*time.Second) != 10*time.Second {
		t.Error("min64 should return smaller")
	}
}
