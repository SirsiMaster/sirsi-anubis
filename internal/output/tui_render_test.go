package output

import (
	"strings"
	"testing"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/maat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/mirror"
	"github.com/SirsiMaster/sirsi-pantheon/internal/osiris"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seshat"
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

// ── RenderCleanPreview ──────────────────────────────────────────────

func TestRenderCleanPreview_Empty(t *testing.T) {
	lines := RenderCleanPreview(nil)
	if len(lines) == 0 {
		t.Error("RenderCleanPreview(nil) returned empty")
	}
}

func TestRenderCleanPreview_WithFindings(t *testing.T) {
	findings := make([]jackal.Finding, 5)
	for i := range findings {
		findings[i] = jackal.Finding{
			Description: "Test cache", Path: "/tmp/cache", SizeBytes: 1024, Severity: jackal.SeveritySafe,
		}
	}
	lines := RenderCleanPreview(findings)
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(lines))
	}
}

func TestRenderCleanPreview_ManyFindings(t *testing.T) {
	findings := make([]jackal.Finding, 20)
	for i := range findings {
		findings[i] = jackal.Finding{Description: "Item", Path: "/tmp/x", SizeBytes: 100}
	}
	lines := RenderCleanPreview(findings)
	found := false
	for _, l := range lines {
		if strings.Contains(l, "more") {
			found = true
		}
	}
	if !found {
		t.Error("expected '... and N more' for 20 findings")
	}
}

// ── RenderMirrorResult ──────────────────────────────────────────────

func TestRenderMirrorResult_NoDuplicates(t *testing.T) {
	res := &mirror.MirrorResult{TotalScanned: 100, TotalDuplicates: 0}
	lines := RenderMirrorResult(res)
	if len(lines) < 2 {
		t.Error("expected at least 2 lines for no-duplicate result")
	}
}

func TestRenderMirrorResult_WithDuplicates(t *testing.T) {
	res := &mirror.MirrorResult{
		TotalScanned: 100, TotalDuplicates: 3, TotalWasteBytes: 5000,
		ScanDuration: 2 * time.Second,
		Groups: []mirror.DuplicateGroup{
			{ID: "abc", WasteBytes: 2000, Files: []mirror.FileEntry{
				{Path: "/a/file1"}, {Path: "/a/file2"},
			}},
		},
	}
	lines := RenderMirrorResult(res)
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(lines))
	}
}

// ── RenderMaatReport ────────────────────────────────────────────────

func TestRenderMaatReport_Pass(t *testing.T) {
	report := &maat.Report{
		OverallVerdict: maat.VerdictPass, OverallWeight: 90,
		Passes: 10, Warnings: 0, Failures: 0,
		Assessments: []maat.Assessment{
			{Subject: "jackal", Verdict: maat.VerdictPass, FeatherWeight: 95, Message: "good"},
		},
	}
	lines, fixCmds := RenderMaatReport(report)
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(lines))
	}
	if len(fixCmds) != 0 {
		t.Errorf("expected no fix commands for passing report, got %v", fixCmds)
	}
}

func TestRenderMaatReport_Fail(t *testing.T) {
	report := &maat.Report{
		OverallVerdict: maat.VerdictFail, OverallWeight: 40,
		Passes: 2, Warnings: 3, Failures: 5,
		Assessments: []maat.Assessment{
			{Subject: "output", Verdict: maat.VerdictFail, FeatherWeight: 20, Message: "low coverage", Remediation: "add tests"},
		},
	}
	lines, fixCmds := RenderMaatReport(report)
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(lines))
	}
	if len(fixCmds) == 0 {
		t.Error("expected fix commands for failing report")
	}
}

// ── RenderDiagram ───────────────────────────────────────────────────

func TestRenderDiagram(t *testing.T) {
	res := &seba.DiagramResult{Title: "Test Diagram", Mermaid: "graph LR\n  A --> B\n  B --> C"}
	lines := RenderDiagram(res)
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines, got %d", len(lines))
	}
}

// ── RenderKnowledgeItems ────────────────────────────────────────────

func TestRenderKnowledgeItems_Empty(t *testing.T) {
	lines := RenderKnowledgeItems(nil)
	if len(lines) < 2 {
		t.Error("expected at least 2 lines for empty knowledge items")
	}
}

func TestRenderKnowledgeItems_WithItems(t *testing.T) {
	items := []seshat.KnowledgeItem{
		{Title: "Item 1", Summary: "Summary of item 1"},
		{Title: "Item 2", Summary: "Summary of item 2"},
	}
	lines := RenderKnowledgeItems(items)
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines for 2 items, got %d", len(lines))
	}
}

// ── RenderRiskAssessment ────────────────────────────────────────────

func TestRenderRiskAssessment(t *testing.T) {
	cp := &osiris.Checkpoint{
		Risk: osiris.RiskHigh, TotalChanges: 20, Branch: "main",
		UncommittedFiles: 15, LinesAdded: 200, LinesDeleted: 50,
	}
	lines := RenderRiskAssessment(cp)
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

// ── splitLines ──────────────────────────────────────────────────────

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"one line", 1},
		{"line1\nline2\nline3", 3},
	}
	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitLines(%q) = %d lines, want %d", tt.input, len(got), tt.want)
		}
	}
}

// ── RenderHardwareProfile ───────────────────────────────────────────

func TestRenderHardwareProfile(t *testing.T) {
	hw := &seba.HardwareProfile{
		CPUModel: "Apple M4 Max", CPUCores: 16, CPUArch: "arm64",
		TotalRAM: 128 * 1024 * 1024 * 1024,
		GPU:      seba.GPUInfo{Name: "Apple M4 Max GPU", MetalFamily: "Metal 3"},
		NeuralEngine: true,
		OS: "macOS", Kernel: "25.4.0",
	}
	lines := RenderHardwareProfile(hw)
	if len(lines) < 8 {
		t.Errorf("expected at least 8 lines, got %d", len(lines))
	}
}

func TestRenderHardwareProfile_NoNeuralEngine(t *testing.T) {
	hw := &seba.HardwareProfile{
		CPUModel: "Intel i9", CPUCores: 8, CPUArch: "x86_64",
		TotalRAM: 32 * 1024 * 1024 * 1024,
		GPU:      seba.GPUInfo{Name: "NVIDIA RTX 4090"},
		OS: "Linux", Kernel: "6.1.0",
	}
	lines := RenderHardwareProfile(hw)
	for _, l := range lines {
		if strings.Contains(l, "NEURAL ENGINE") {
			t.Error("should not show Neural Engine section when disabled")
		}
	}
}

// ── RenderMaatReport additional ─────────────────────────────────────

func TestRenderMaatReport_Warning(t *testing.T) {
	report := &maat.Report{
		OverallVerdict: maat.VerdictWarning, OverallWeight: 60,
		Passes: 5, Warnings: 3, Failures: 0,
		Assessments: []maat.Assessment{
			{Subject: "ka", Verdict: maat.VerdictWarning, FeatherWeight: 60, Message: "below target", Remediation: "add tests"},
		},
	}
	lines, _ := RenderMaatReport(report)
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d", len(lines))
	}
}

// ── categoryIcon ────────────────────────────────────────────────────

func TestCategoryIcon(t *testing.T) {
	tests := []string{"cache", "logs", "build", "containers", "ai", "general", "unknown"}
	for _, cat := range tests {
		icon := categoryIcon(cat)
		if icon == "" {
			t.Errorf("categoryIcon(%q) returned empty", cat)
		}
	}
}
