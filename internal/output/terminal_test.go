package output

import (
	"testing"
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
