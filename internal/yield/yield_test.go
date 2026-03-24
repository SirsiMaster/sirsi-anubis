package yield

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"testing"
)

// ── Mock load providers for deterministic testing ────────────────────────

func mockLoad(load1, load5 float64) LoadProvider {
	return func() (float64, float64, error) {
		return load1, load5, nil
	}
}

func mockLoadErr() LoadProvider {
	return func() (float64, float64, error) {
		return 0, 0, errors.New("mock load error")
	}
}

// Helper to compute a load1 that produces a given ratio on this machine
func loadForRatio(ratio float64) float64 {
	return ratio * float64(runtime.NumCPU())
}

// ── CheckWith ────────────────────────────────────────────────────────────

func TestCheckWith_Healthy(t *testing.T) {
	load, err := CheckWith(mockLoad(loadForRatio(0.3), 0.2))
	if err != nil {
		t.Fatalf("CheckWith: %v", err)
	}
	if load.Verdict != VerdictHealthy {
		t.Errorf("Verdict = %q, want healthy (ratio=%.2f)", load.Verdict, load.LoadRatio)
	}
}

func TestCheckWith_Caution(t *testing.T) {
	load, err := CheckWith(mockLoad(loadForRatio(0.7), 0.5))
	if err != nil {
		t.Fatalf("CheckWith: %v", err)
	}
	if load.Verdict != VerdictCaution {
		t.Errorf("Verdict = %q, want caution (ratio=%.2f)", load.Verdict, load.LoadRatio)
	}
}

func TestCheckWith_Yield(t *testing.T) {
	load, err := CheckWith(mockLoad(loadForRatio(1.5), 1.0))
	if err != nil {
		t.Fatalf("CheckWith: %v", err)
	}
	if load.Verdict != VerdictYield {
		t.Errorf("Verdict = %q, want yield (ratio=%.2f)", load.Verdict, load.LoadRatio)
	}
}

func TestCheckWith_Error(t *testing.T) {
	_, err := CheckWith(mockLoadErr())
	if err == nil {
		t.Error("Expected error from failed load provider")
	}
}

func TestCheckWith_ZeroLoad(t *testing.T) {
	load, err := CheckWith(mockLoad(0, 0))
	if err != nil {
		t.Fatalf("CheckWith: %v", err)
	}
	if load.Verdict != VerdictHealthy {
		t.Errorf("Zero load should be healthy, got %q", load.Verdict)
	}
	if load.LoadRatio != 0 {
		t.Errorf("LoadRatio = %f, want 0", load.LoadRatio)
	}
}

func TestCheckWith_ExactBoundary_060(t *testing.T) {
	// 0.6 is NOT > 0.6, so should be healthy
	load, err := CheckWith(mockLoad(loadForRatio(0.6), 0.5))
	if err != nil {
		t.Fatalf("CheckWith: %v", err)
	}
	if load.Verdict != VerdictHealthy {
		t.Errorf("0.60 ratio should be healthy, got %q", load.Verdict)
	}
}

func TestCheckWith_ExactBoundary_085(t *testing.T) {
	// 0.85 is NOT > 0.85, so should be caution
	load, err := CheckWith(mockLoad(loadForRatio(0.85), 0.5))
	if err != nil {
		t.Fatalf("CheckWith: %v", err)
	}
	if load.Verdict != VerdictCaution {
		t.Errorf("0.85 ratio should be caution, got %q", load.Verdict)
	}
}

// ── WarnIfHeavyWith ─────────────────────────────────────────────────────

func captureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestWarnIfHeavyWith_Healthy(t *testing.T) {
	output := captureStderr(func() {
		abort := WarnIfHeavyWith("test-cmd", false, mockLoad(loadForRatio(0.3), 0.2))
		if abort {
			t.Error("Healthy load should not abort")
		}
	})
	// Healthy = no output
	if output != "" {
		t.Errorf("Healthy should produce no stderr, got: %q", output)
	}
}

func TestWarnIfHeavyWith_Caution(t *testing.T) {
	output := captureStderr(func() {
		abort := WarnIfHeavyWith("test-cmd", false, mockLoad(loadForRatio(0.7), 0.5))
		if abort {
			t.Error("Caution should not abort")
		}
	})
	if output == "" {
		t.Error("Caution should produce stderr warning")
	}
	if !bytes.Contains([]byte(output), []byte("moderately loaded")) {
		t.Errorf("Caution message missing 'moderately loaded', got: %q", output)
	}
}

func TestWarnIfHeavyWith_Yield_NoForce(t *testing.T) {
	output := captureStderr(func() {
		abort := WarnIfHeavyWith("scan", false, mockLoad(loadForRatio(1.5), 1.0))
		if !abort {
			t.Error("Yield without force should abort")
		}
	})
	if output == "" {
		t.Error("Yield should produce stderr warning")
	}
	if !bytes.Contains([]byte(output), []byte("heavy load")) {
		t.Errorf("Yield message missing 'heavy load', got: %q", output)
	}
	if !bytes.Contains([]byte(output), []byte("--force")) {
		t.Errorf("Yield message should mention --force, got: %q", output)
	}
}

func TestWarnIfHeavyWith_Yield_WithForce(t *testing.T) {
	output := captureStderr(func() {
		abort := WarnIfHeavyWith("scan", true, mockLoad(loadForRatio(1.5), 1.0))
		if abort {
			t.Error("Yield with force should NOT abort")
		}
	})
	if !bytes.Contains([]byte(output), []byte("--force")) {
		t.Errorf("Force override message expected, got: %q", output)
	}
}

func TestWarnIfHeavyWith_Error(t *testing.T) {
	output := captureStderr(func() {
		abort := WarnIfHeavyWith("test-cmd", false, mockLoadErr())
		if abort {
			t.Error("Error should not abort (fail-open)")
		}
	})
	if output != "" {
		t.Errorf("Error path should produce no output, got: %q", output)
	}
}

// ── ShouldYield integration ─────────────────────────────────────────────

func TestShouldYield_Integration(t *testing.T) {
	result := ShouldYield()
	t.Logf("ShouldYield = %v", result)
}

// ── Check integration ───────────────────────────────────────────────────

func TestCheck_Integration(t *testing.T) {
	load, err := Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if load.CPUCount <= 0 {
		t.Errorf("CPUCount = %d", load.CPUCount)
	}
	if load.Verdict == "" {
		t.Error("Verdict empty")
	}
	expectedRatio := load.LoadAvg1 / float64(load.CPUCount)
	if fmt.Sprintf("%.4f", load.LoadRatio) != fmt.Sprintf("%.4f", expectedRatio) {
		t.Errorf("LoadRatio mismatch")
	}
}

// ── WarnIfHeavy integration ─────────────────────────────────────────────

func TestWarnIfHeavy_Integration(t *testing.T) {
	captureStderr(func() {
		WarnIfHeavy("test-cmd", false)
	})
}

func TestWarnIfHeavy_ForceIntegration(t *testing.T) {
	captureStderr(func() {
		abort := WarnIfHeavy("test-cmd", true)
		if abort {
			t.Error("Force should never abort")
		}
	})
}

// ── getLoadAverage ──────────────────────────────────────────────────────

func TestGetLoadAverage(t *testing.T) {
	l1, l5, err := getLoadAverage()
	if err != nil {
		t.Fatalf("getLoadAverage: %v", err)
	}
	if l1 < 0 || l5 < 0 {
		t.Errorf("Negative loads: l1=%f, l5=%f", l1, l5)
	}
}

func TestGetLoadAverageUnix(t *testing.T) {
	l1, l5, err := getLoadAverageUnix()
	if err != nil {
		t.Fatalf("getLoadAverageUnix: %v", err)
	}
	t.Logf("Unix load: 1m=%.2f, 5m=%.2f", l1, l5)
}

// ── Constants ───────────────────────────────────────────────────────────

func TestVerdictConstants(t *testing.T) {
	if VerdictHealthy != "healthy" {
		t.Errorf("VerdictHealthy = %q", VerdictHealthy)
	}
	if VerdictCaution != "caution" {
		t.Errorf("VerdictCaution = %q", VerdictCaution)
	}
	if VerdictYield != "yield" {
		t.Errorf("VerdictYield = %q", VerdictYield)
	}
}
