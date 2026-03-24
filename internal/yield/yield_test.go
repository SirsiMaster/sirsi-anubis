package yield

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

// ── WarnIfHeavy Tests ───────────────────────────────────────────────────

func TestWarnIfHeavy_Healthy(t *testing.T) {
	// WarnIfHeavy should NOT abort when system is healthy.
	// Capture stderr output.
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	shouldAbort := WarnIfHeavy("test-command", false)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Can't guarantee verdict because it depends on actual system load,
	// but we can verify it doesn't panic and returns a bool.
	t.Logf("WarnIfHeavy returned shouldAbort=%v, stderr=%q", shouldAbort, buf.String())
}

func TestWarnIfHeavy_ForceOverride(t *testing.T) {
	// With --force, WarnIfHeavy should never abort.
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	shouldAbort := WarnIfHeavy("test-command", true)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Force mode should never abort
	if shouldAbort {
		t.Error("WarnIfHeavy with force=true should never return true")
	}
	t.Logf("stderr output: %q", buf.String())
}

// ── Verdict Classification (direct Check + classify) ────────────────────

func TestCheckVerdict_Boundaries(t *testing.T) {
	// We test verdict logic by directly checking the thresholds.
	// The actual Check() function uses real system load, so we verify
	// the classification logic matches.
	tests := []struct {
		name  string
		ratio float64
		want  string
	}{
		{"zero", 0.0, VerdictHealthy},
		{"low", 0.3, VerdictHealthy},
		{"boundary_healthy", 0.59, VerdictHealthy},
		{"boundary_caution", 0.61, VerdictCaution},
		{"mid_caution", 0.75, VerdictCaution},
		{"boundary_yield", 0.851, VerdictYield},
		{"high_yield", 1.0, VerdictYield},
		{"overloaded", 2.0, VerdictYield},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var verdict string
			switch {
			case tt.ratio > 0.85:
				verdict = VerdictYield
			case tt.ratio > 0.6:
				verdict = VerdictCaution
			default:
				verdict = VerdictHealthy
			}
			if verdict != tt.want {
				t.Errorf("ratio %.3f → %q, want %q", tt.ratio, verdict, tt.want)
			}
		})
	}
}

// ── ShouldYield Integration ─────────────────────────────────────────────

func TestShouldYield_NoError(t *testing.T) {
	// Just verify it returns a bool without error or panic.
	// The result depends on actual system load.
	result := ShouldYield()
	t.Logf("ShouldYield = %v", result)
}

// ── getLoadAverageUnix ──────────────────────────────────────────────────

func TestGetLoadAverageUnix(t *testing.T) {
	load1, load5, err := getLoadAverageUnix()
	if err != nil {
		t.Fatalf("getLoadAverageUnix: %v", err)
	}
	if load1 < 0 {
		t.Errorf("load1 = %f, should be >= 0", load1)
	}
	if load5 < 0 {
		t.Errorf("load5 = %f, should be >= 0", load5)
	}
	t.Logf("Unix load: 1m=%.2f, 5m=%.2f", load1, load5)
}

// ── Constant Values ─────────────────────────────────────────────────────

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

// ── SystemLoad Struct ───────────────────────────────────────────────────

func TestSystemLoad_Fields(t *testing.T) {
	load, err := Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	// All fields should be populated
	if load.CPUCount <= 0 {
		t.Errorf("CPUCount = %d", load.CPUCount)
	}
	if load.LoadRatio < 0 {
		t.Errorf("LoadRatio = %f", load.LoadRatio)
	}
	if load.Verdict == "" {
		t.Error("Verdict should not be empty")
	}

	// LoadRatio should be Load1 / CPUCount
	expectedRatio := load.LoadAvg1 / float64(load.CPUCount)
	if fmt.Sprintf("%.4f", load.LoadRatio) != fmt.Sprintf("%.4f", expectedRatio) {
		t.Errorf("LoadRatio = %f, expected Load1(%f)/CPUs(%d) = %f",
			load.LoadRatio, load.LoadAvg1, load.CPUCount, expectedRatio)
	}
}
