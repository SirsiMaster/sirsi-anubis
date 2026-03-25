package guard

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// ── AlertRing Tests ─────────────────────────────────────────────────────

func TestNewAlertRing(t *testing.T) {
	r := NewAlertRing()
	if r == nil {
		t.Fatal("NewAlertRing returned nil")
	}
	current, lifetime := r.Stats()
	if current != 0 || lifetime != 0 {
		t.Errorf("Fresh ring should have 0/0, got %d/%d", current, lifetime)
	}
}

func TestAlertRing_PushAndRecent(t *testing.T) {
	r := NewAlertRing()

	// Push 3 alerts
	for i := 0; i < 3; i++ {
		r.Push(AlertEntry{
			ProcessName: "test-process",
			PID:         1000 + i,
			CPUPercent:  float64(80 + i*10),
			Severity:    "warning",
			Timestamp:   time.Now().Format(time.RFC3339),
		})
	}

	current, lifetime := r.Stats()
	if current != 3 {
		t.Errorf("Expected 3 current, got %d", current)
	}
	if lifetime != 3 {
		t.Errorf("Expected 3 lifetime, got %d", lifetime)
	}

	// Recent 2 — should be newest first
	recent := r.Recent(2)
	if len(recent) != 2 {
		t.Fatalf("Expected 2 recent, got %d", len(recent))
	}
	if recent[0].PID != 1002 {
		t.Errorf("Most recent should be PID 1002, got %d", recent[0].PID)
	}
	if recent[1].PID != 1001 {
		t.Errorf("Second most recent should be PID 1001, got %d", recent[1].PID)
	}
}

func TestAlertRing_RecentEdgeCases(t *testing.T) {
	r := NewAlertRing()

	// Empty ring
	if got := r.Recent(5); got != nil {
		t.Errorf("Empty ring Recent should return nil, got %v", got)
	}

	// Request 0
	r.Push(AlertEntry{PID: 1})
	if got := r.Recent(0); got != nil {
		t.Errorf("Recent(0) should return nil, got %v", got)
	}

	// Request negative
	if got := r.Recent(-1); got != nil {
		t.Errorf("Recent(-1) should return nil, got %v", got)
	}

	// Request more than available
	got := r.Recent(100)
	if len(got) != 1 {
		t.Errorf("Should cap at current count, expected 1, got %d", len(got))
	}
}

func TestAlertRing_Overflow(t *testing.T) {
	r := NewAlertRing()

	// Push more than AlertRingSize
	for i := 0; i < AlertRingSize+10; i++ {
		r.Push(AlertEntry{PID: i, ProcessName: "overflow-test"})
	}

	current, lifetime := r.Stats()
	if current != AlertRingSize {
		t.Errorf("Expected current=%d after overflow, got %d", AlertRingSize, current)
	}
	if lifetime != AlertRingSize+10 {
		t.Errorf("Expected lifetime=%d, got %d", AlertRingSize+10, lifetime)
	}

	// Most recent should be the last pushed
	recent := r.Recent(1)
	if recent[0].PID != AlertRingSize+9 {
		t.Errorf("Most recent should be PID %d, got %d", AlertRingSize+9, recent[0].PID)
	}
}

func TestAlertRing_Concurrent(t *testing.T) {
	r := NewAlertRing()
	var wg sync.WaitGroup

	// 10 goroutines each pushing 100 alerts
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				r.Push(AlertEntry{
					PID:         gid*1000 + i,
					ProcessName: "concurrent",
				})
			}
		}(g)
	}

	// Concurrent readers
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				r.Recent(10)
				r.Stats()
			}
		}()
	}

	wg.Wait()

	_, lifetime := r.Stats()
	if lifetime != 1000 {
		t.Errorf("Expected 1000 lifetime after concurrent pushes, got %d", lifetime)
	}
}

// ── DefaultBridgeConfig Tests ───────────────────────────────────────────

func TestDefaultBridgeConfig(t *testing.T) {
	cfg := DefaultBridgeConfig()
	if cfg.WatchConfig.CPUThreshold != 60.0 {
		t.Errorf("Expected IDE threshold 60.0, got %.1f", cfg.WatchConfig.CPUThreshold)
	}
	if cfg.WatchConfig.SustainCount != 1 {
		t.Errorf("Expected sustain count 1, got %d", cfg.WatchConfig.SustainCount)
	}
	if cfg.WatchConfig.Interval != 800*time.Millisecond {
		t.Errorf("Expected 800ms interval, got %s", cfg.WatchConfig.Interval)
	}
	if cfg.CPUCritical != 120.0 {
		t.Errorf("Expected critical 120.0, got %.1f", cfg.CPUCritical)
	}
}

// ── Bridge Lifecycle Tests ──────────────────────────────────────────────

func TestStartBridge_LifecycleWithAlerts(t *testing.T) {
	// Mock the sampler to produce high-CPU alerts
	alertCount := 0
	old := sampleTopCPUFn
	sampleTopCPUFn = func(topN int) ([]ProcessInfo, error) {
		alertCount++
		return []ProcessInfo{
			{PID: 42, Name: "Plugin Host", CPUPercent: 103.9, RSS: 512 * 1024 * 1024},
		}, nil
	}
	defer func() { sampleTopCPUFn = old }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var received []AlertEntry
	var mu sync.Mutex

	bridge := StartBridge(ctx, BridgeConfig{
		WatchConfig: WatchConfig{
			Interval:     50 * time.Millisecond,
			CPUThreshold: 80.0,
			SustainCount: 2,
			SampleSize:   5,
			SelfBudget:   50.0,
		},
		CPUCritical: 150.0,
		OnAlert: func(entry AlertEntry) {
			mu.Lock()
			received = append(received, entry)
			mu.Unlock()
		},
	})

	// Wait for some alerts to flow through
	time.Sleep(400 * time.Millisecond)
	bridge.Stop()

	// Verify ring buffer has alerts
	current, _ := bridge.Ring().Stats()
	if current == 0 {
		t.Error("Expected alerts in ring buffer after sustained CPU spike")
	}

	// Verify callback was called
	mu.Lock()
	rcvCount := len(received)
	mu.Unlock()
	if rcvCount == 0 {
		t.Error("OnAlert callback was never called")
	}

	// Verify severity classification
	mu.Lock()
	for _, entry := range received {
		if entry.Severity != "warning" {
			t.Errorf("CPU 103.9%% should be 'warning' (< 150.0 critical), got %s", entry.Severity)
		}
		if entry.ProcessName != "Plugin Host" {
			t.Errorf("Expected 'Plugin Host', got %s", entry.ProcessName)
		}
	}
	mu.Unlock()

	t.Logf("Bridge lifecycle: %d polls, %d alerts received", alertCount, rcvCount)
}

func TestStartBridge_CriticalSeverity(t *testing.T) {
	old := sampleTopCPUFn
	sampleTopCPUFn = func(topN int) ([]ProcessInfo, error) {
		return []ProcessInfo{
			{PID: 99, Name: "runaway", CPUPercent: 200.0, RSS: 1024 * 1024},
		}, nil
	}
	defer func() { sampleTopCPUFn = old }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var gotCritical bool
	var mu sync.Mutex

	bridge := StartBridge(ctx, BridgeConfig{
		WatchConfig: WatchConfig{
			Interval:     50 * time.Millisecond,
			CPUThreshold: 80.0,
			SustainCount: 2,
			SampleSize:   5,
			SelfBudget:   50.0,
		},
		CPUCritical: 150.0,
		OnAlert: func(entry AlertEntry) {
			mu.Lock()
			if entry.Severity == "critical" {
				gotCritical = true
			}
			mu.Unlock()
		},
	})

	time.Sleep(400 * time.Millisecond)
	bridge.Stop()

	mu.Lock()
	if !gotCritical {
		t.Error("Expected 'critical' severity for 200% CPU")
	}
	mu.Unlock()
}

func TestStartBridge_DefaultCPUCritical(t *testing.T) {
	old := sampleTopCPUFn
	sampleTopCPUFn = func(topN int) ([]ProcessInfo, error) {
		return nil, nil // no processes
	}
	defer func() { sampleTopCPUFn = old }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bridge := StartBridge(ctx, BridgeConfig{
		WatchConfig: WatchConfig{
			Interval:     50 * time.Millisecond,
			CPUThreshold: 80.0,
			SustainCount: 3,
			SampleSize:   5,
			SelfBudget:   50.0,
		},
		CPUCritical: 0, // Should default to 150.0
	})

	time.Sleep(100 * time.Millisecond)
	bridge.Stop()

	// Should still work (no panic, clean shutdown)
	current, _ := bridge.Ring().Stats()
	t.Logf("Bridge with defaults: %d alerts", current)
}

// ── Status & JSON Tests ─────────────────────────────────────────────────

func TestBridge_StatusJSON(t *testing.T) {
	old := sampleTopCPUFn
	sampleTopCPUFn = func(topN int) ([]ProcessInfo, error) {
		return []ProcessInfo{
			{PID: 42, Name: "vscode", CPUPercent: 95.0, RSS: 256 * 1024 * 1024},
		}, nil
	}
	defer func() { sampleTopCPUFn = old }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bridge := StartBridge(ctx, BridgeConfig{
		WatchConfig: WatchConfig{
			Interval:     50 * time.Millisecond,
			CPUThreshold: 80.0,
			SustainCount: 2,
			SampleSize:   5,
			SelfBudget:   50.0,
		},
		CPUCritical: 150.0,
	})

	time.Sleep(300 * time.Millisecond)

	// Get Status struct
	status := bridge.Status()
	if status == nil {
		t.Fatal("Status returned nil")
	}
	if !status.Active {
		t.Error("Expected watchdog to be active")
	}

	// Get JSON
	jsonStr, err := bridge.StatusJSON()
	if err != nil {
		t.Fatalf("StatusJSON error: %v", err)
	}
	if jsonStr == "" {
		t.Error("StatusJSON returned empty string")
	}

	// Verify valid JSON
	var parsed BridgeStatus
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("StatusJSON produced invalid JSON: %v", err)
	}

	bridge.Stop()

	t.Logf("StatusJSON: active=%v, buffered=%d, lifetime=%d, polls=%d",
		parsed.Active, parsed.BufferedCount, parsed.LifetimeAlerts, parsed.WatchdogPolls)
}

func TestBridge_WatchdogAccessor(t *testing.T) {
	old := sampleTopCPUFn
	sampleTopCPUFn = func(topN int) ([]ProcessInfo, error) {
		return nil, nil
	}
	defer func() { sampleTopCPUFn = old }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bridge := StartBridge(ctx, BridgeConfig{
		WatchConfig: WatchConfig{
			Interval:     100 * time.Millisecond,
			CPUThreshold: 80.0,
			SustainCount: 3,
			SampleSize:   5,
			SelfBudget:   50.0,
		},
	})

	if bridge.Watchdog() == nil {
		t.Error("Watchdog accessor should not be nil")
	}
	if bridge.Ring() == nil {
		t.Error("Ring accessor should not be nil")
	}

	bridge.Stop()
}
