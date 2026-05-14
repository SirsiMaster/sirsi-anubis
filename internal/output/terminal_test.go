package output

import (
	"os"
	"strings"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
)

// ── Style Tests ──────────────────────────────────────────────────────────
// Verify all brand styles are properly initialized (not nil/zero).

func TestColorConstants(t *testing.T) {
	colors := map[string]interface{}{
		"Gold":     Gold,
		"Black":    Black,
		"Lapis":    Lapis,
		"White":    White,
		"DimWhite": DimWhite,
		"Red":      Red,
		"Green":    Green,
		"Yellow":   Yellow,
	}
	for name, c := range colors {
		if c == nil {
			t.Errorf("Color %s should not be nil", name)
		}
	}
}

func TestStyles_Render(t *testing.T) {
	// All styles should render without panic
	styles := map[string]interface {
		Render(...string) string
	}{
		"TitleStyle":      TitleStyle,
		"HeaderStyle":     HeaderStyle,
		"BodyStyle":       BodyStyle,
		"DimStyle":        DimStyle,
		"ErrorStyle":      ErrorStyle,
		"SuccessStyle":    SuccessStyle,
		"WarningStyle":    WarningStyle,
		"SizeStyle":       SizeStyle,
		"CategoryStyle":   CategoryStyle,
		"BoxStyle":        BoxStyle,
		"SeveritySafe":    SeveritySafe,
		"SeverityCaution": SeverityCaution,
		"SeverityWarning": SeverityWarning,
	}

	for name, style := range styles {
		result := style.Render("test content")
		if result == "" {
			t.Errorf("Style %s rendered empty string", name)
		}
	}
}

// ── FindingRow ───────────────────────────────────────────────────────────

func TestFindingRow(t *testing.T) {
	tests := []struct {
		name     string
		severity string
	}{
		{"safe", "safe"},
		{"caution", "caution"},
		{"warning", "warning"},
		{"unknown", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := FindingRow("TestApp", "/path/to/app", "1.2 GB", tt.severity)
			if row == "" {
				t.Error("FindingRow should not be empty")
			}
		})
	}
}

// ── Banner / Header / Info / Dim / Success / Warn / Error ───────────────
// These all write to stderr. We verify they don't panic.

func TestBanner(t *testing.T) {
	// Should not panic
	Banner()
}

func TestHeader(t *testing.T) {
	Header("Test Section")
}

func TestInfo(t *testing.T) {
	Info("Test message: %s", "hello")
}

func TestDim(t *testing.T) {
	Dim("Dim message: %d items", 42)
}

func TestSuccess(t *testing.T) {
	Success("Operation %s", "completed")
}

func TestWarn(t *testing.T) {
	Warn("Warning: %s", "low disk space")
}

func TestError(t *testing.T) {
	Error("Error: %s", "file not found")
}

// ── Summary ──────────────────────────────────────────────────────────────

func TestSummary(t *testing.T) {
	Summary("2.4 GB", 15, 8)
}

// ── ShortenPath ─────────────────────────────────────────────────────────

func TestShortenPath_HomeDirReplacement(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{home + "/Documents/file.txt", "~/Documents/file.txt"},
		{"/usr/local/bin/sirsi", "/usr/local/bin/sirsi"},
		{"relative/path", "relative/path"},
	}
	for _, tt := range tests {
		got := ShortenPath(tt.input)
		if got != tt.want {
			t.Errorf("ShortenPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShortenPath_LongPathTruncation(t *testing.T) {
	long := "/very/long/path/" + strings.Repeat("a", 60) + "/file.txt"
	got := ShortenPath(long)
	if len(got) > 60 {
		t.Errorf("ShortenPath should truncate to ≤60 chars, got %d", len(got))
	}
	if !strings.HasPrefix(got, "...") {
		t.Errorf("truncated path should start with '...': %s", got)
	}
}

// ── Truncate ────────────────────────────────────────────────────────────

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"abcdefghij", 6, "abc..."},
	}
	for _, tt := range tests {
		got := Truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

// ── padRight ────────────────────────────────────────────────────────────

func TestPadRight(t *testing.T) {
	tests := []struct {
		input string
		width int
		want  string
	}{
		{"hi", 5, "hi   "},
		{"hello", 3, "hello"},
		{"exact", 5, "exact"},
		{"", 3, "   "},
	}
	for _, tt := range tests {
		got := padRight(tt.input, tt.width)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
		}
	}
}

// ── SuggestSteps ────────────────────────────────────────────────────────

func TestSuggestSteps(t *testing.T) {
	ctx := suggest.Context{Deity: "anubis", Subcommand: "scan"}
	steps := SuggestSteps(ctx)
	if len(steps) == 0 {
		t.Error("SuggestSteps() returned empty for anubis/scan")
	}
	for _, s := range steps {
		if len(s) != 2 {
			t.Errorf("each step should have 2 elements, got %d", len(s))
		}
	}
}
