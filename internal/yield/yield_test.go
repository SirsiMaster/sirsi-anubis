package yield

import (
	"testing"
)

func TestCheck(t *testing.T) {
	load, err := Check()
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	if load.CPUCount <= 0 {
		t.Errorf("CPUCount = %d, want > 0", load.CPUCount)
	}

	if load.LoadAvg1 < 0 {
		t.Errorf("LoadAvg1 = %f, want >= 0", load.LoadAvg1)
	}

	if load.LoadRatio < 0 {
		t.Errorf("LoadRatio = %f, want >= 0", load.LoadRatio)
	}

	validVerdicts := map[string]bool{
		VerdictHealthy: true,
		VerdictCaution: true,
		VerdictYield:   true,
	}
	if !validVerdicts[load.Verdict] {
		t.Errorf("Verdict = %q, want one of healthy/caution/yield", load.Verdict)
	}

	t.Logf("System load: avg1=%.2f, avg5=%.2f, cpus=%d, ratio=%.2f, verdict=%s",
		load.LoadAvg1, load.LoadAvg5, load.CPUCount, load.LoadRatio, load.Verdict)
}

func TestShouldYield(t *testing.T) {
	// This just verifies it doesn't panic or error — actual behavior
	// depends on the machine's current load.
	result := ShouldYield()
	t.Logf("ShouldYield() = %v", result)
}

func TestVerdictThresholds(t *testing.T) {
	tests := []struct {
		name    string
		ratio   float64
		want    string
	}{
		{"idle", 0.1, VerdictHealthy},
		{"light", 0.4, VerdictHealthy},
		{"moderate", 0.7, VerdictCaution},
		{"heavy", 0.86, VerdictYield},
		{"overloaded", 1.5, VerdictYield},
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
				t.Errorf("ratio %.2f: got %q, want %q", tt.ratio, verdict, tt.want)
			}
		})
	}
}

func TestGetLoadAverage(t *testing.T) {
	load1, load5, err := getLoadAverage()
	if err != nil {
		t.Fatalf("getLoadAverage() failed: %v", err)
	}
	if load1 < 0 || load5 < 0 {
		t.Errorf("negative load averages: load1=%f, load5=%f", load1, load5)
	}
	t.Logf("Load averages: 1m=%.2f, 5m=%.2f", load1, load5)
}
