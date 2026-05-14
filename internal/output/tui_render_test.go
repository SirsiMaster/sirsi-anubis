package output

import (
	"strings"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
)

func TestProgressBar(t *testing.T) {
	tests := []float64{0, 25, 50, 60, 85, 100, -5, 120}
	for _, pct := range tests {
		got := ProgressBar(pct, 20)
		if got == "" {
			t.Errorf("ProgressBar(%.0f, 20) returned empty", pct)
		}
		// Should contain a percentage label
		if !strings.Contains(got, "%") {
			t.Errorf("ProgressBar(%.0f) missing %% label: %s", pct, got)
		}
	}
}

func TestScoreBar(t *testing.T) {
	tests := []int{0, 25, 50, 75, 100}
	for _, score := range tests {
		got := ScoreBar(score, 20)
		if got == "" {
			t.Errorf("ScoreBar(%d, 20) returned empty", score)
		}
	}
}

func TestSparkline_Empty(t *testing.T) {
	got := Sparkline(nil, 10, Gold)
	if len(got) == 0 {
		t.Error("Sparkline(nil) should return floor sparks")
	}
}

func TestSparkline_Values(t *testing.T) {
	vals := []float64{0, 25, 50, 75, 100}
	got := Sparkline(vals, 5, Gold)
	if got == "" {
		t.Error("Sparkline returned empty for valid values")
	}
}

func TestSparkline_MoreValuesThanWidth(t *testing.T) {
	vals := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	got := Sparkline(vals, 5, Gold)
	if got == "" {
		t.Error("Sparkline should truncate to width")
	}
}

func TestSparkline_OutOfBounds(t *testing.T) {
	vals := []float64{-10, 200}
	got := Sparkline(vals, 5, Gold)
	if got == "" {
		t.Error("Sparkline should clamp out-of-bounds values")
	}
}

func TestResultBanner(t *testing.T) {
	style := TitleStyle
	got := ResultBanner("Test Complete", style, 50)
	if !strings.Contains(got, "Test Complete") {
		t.Errorf("ResultBanner missing message: %s", got)
	}
	if !strings.Contains(got, "𓊝") {
		t.Errorf("ResultBanner missing cartouche markers: %s", got)
	}
}

func TestResultBanner_NarrowWidth(t *testing.T) {
	got := ResultBanner("Hi", TitleStyle, 5)
	if !strings.Contains(got, "Hi") {
		t.Errorf("narrow ResultBanner missing message: %s", got)
	}
}

func TestRenderScanResult_Empty(t *testing.T) {
	result := &jackal.ScanResult{
		Findings: nil,
		RulesRan: 58,
	}
	lines := RenderScanResult(result)
	if len(lines) == 0 {
		t.Error("RenderScanResult returned empty for no findings")
	}
}

func TestRenderScanResult_WithFindings(t *testing.T) {
	result := &jackal.ScanResult{
		Findings: []jackal.Finding{
			{Description: "Test Cache", Path: "/tmp/cache", SizeBytes: 1024, Severity: jackal.SeveritySafe, Category: "cache"},
			{Description: "Big Build", Path: "/tmp/build", SizeBytes: 1048576, Severity: jackal.SeverityCaution, Category: "build"},
		},
		TotalSize:  1049600,
		RulesRan:   58,
		ByCategory: map[jackal.Category]jackal.CategorySummary{},
	}
	lines := RenderScanResult(result)
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines for 2 findings, got %d", len(lines))
	}
}

func TestRenderGhostResult_Empty(t *testing.T) {
	lines := RenderGhostResult(nil)
	if len(lines) == 0 {
		t.Error("RenderGhostResult returned empty for no ghosts")
	}
}

func TestRenderGhostResult_WithGhosts(t *testing.T) {
	ghosts := []ka.Ghost{
		{AppName: "OldApp", TotalSize: 5000, Residuals: []ka.Residual{{Path: "/tmp/oldapp", Type: ka.ResidualCaches, SizeBytes: 5000}}},
	}
	lines := RenderGhostResult(ghosts)
	if len(lines) < 2 {
		t.Errorf("expected at least 2 lines for 1 ghost, got %d", len(lines))
	}
}

func TestRenderCleanResult(t *testing.T) {
	result := &jackal.CleanResult{
		Cleaned:    3,
		BytesFreed: 2048,
		Skipped:    1,
	}
	lines := RenderCleanResult(result)
	if len(lines) == 0 {
		t.Error("RenderCleanResult returned empty")
	}
}

func TestRepeatRune(t *testing.T) {
	tests := []struct {
		r    rune
		n    int
		want string
	}{
		{'█', 3, "███"},
		{'━', 0, ""},
		{'━', -1, ""},
		{'A', 1, "A"},
	}
	for _, tt := range tests {
		got := repeatRune(tt.r, tt.n)
		if got != tt.want {
			t.Errorf("repeatRune(%q, %d) = %q, want %q", tt.r, tt.n, got, tt.want)
		}
	}
}
