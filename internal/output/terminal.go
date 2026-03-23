// Package output handles terminal rendering for Sirsi Pantheon.
// Uses the Pantheon brand language: Gold + Black + Deep Lapis (Rule A10).
package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Anubis brand colors (Rule A10)
var (
	Gold     = lipgloss.Color("#C8A951")
	Black    = lipgloss.Color("#0F0F0F")
	Lapis    = lipgloss.Color("#1A1A5E")
	White    = lipgloss.Color("#FAFAFA")
	DimWhite = lipgloss.Color("#888888")
	Red      = lipgloss.Color("#FF4444")
	Green    = lipgloss.Color("#44FF88")
	Yellow   = lipgloss.Color("#FFD700")
)

// Styles
var (
	// Title style — gold text, bold
	TitleStyle = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	// Header style — gold, underlined
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true).
			Underline(true)

	// Body text
	BodyStyle = lipgloss.NewStyle().
			Foreground(White)

	// Dim body text
	DimStyle = lipgloss.NewStyle().
			Foreground(DimWhite)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	// Size style — gold, for file sizes
	SizeStyle = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	// Category badge
	CategoryStyle = lipgloss.NewStyle().
			Foreground(Black).
			Background(Gold).
			Padding(0, 1).
			Bold(true)

	// Box style for major sections
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Gold).
			Padding(0, 2)

	// Severity styles
	SeveritySafe    = lipgloss.NewStyle().Foreground(Green)
	SeverityCaution = lipgloss.NewStyle().Foreground(Yellow)
	SeverityWarning = lipgloss.NewStyle().Foreground(Red)
)

// Banner prints the Pantheon banner.
func Banner() {
	banner := TitleStyle.Render(`
  🏛️  Sirsi Pantheon
  ═══════════════════════════════
  Unified DevOps Intelligence Platform
  "One Install. All Deities."

  𓂀 Anubis · 🪶 Ma'at · 𓁟 Thoth
`)
	fmt.Fprintln(os.Stderr, banner)
}

// Header prints a section header with the 𓂀 prefix.
func Header(text string) {
	fmt.Fprintf(os.Stderr, "\n%s\n", HeaderStyle.Render("𓂀 "+text))
}

// Info prints an informational message.
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s\n", BodyStyle.Render(msg))
}

// Dim prints a dimmed message.
func Dim(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s\n", DimStyle.Render(msg))
}

// Success prints a success message.
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s %s\n", SuccessStyle.Render("✓"), BodyStyle.Render(msg))
}

// Warn prints a warning message.
func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s %s\n", WarningStyle.Render("⚠"), BodyStyle.Render(msg))
}

// Error prints an error message.
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s %s\n", ErrorStyle.Render("✗"), BodyStyle.Render(msg))
}

// FindingRow formats a single finding as a table row.
func FindingRow(name, path, size, severity string) string {
	severityStyled := severity
	switch strings.ToLower(severity) {
	case "safe":
		severityStyled = SeveritySafe.Render(severity)
	case "caution":
		severityStyled = SeverityCaution.Render(severity)
	case "warning":
		severityStyled = SeverityWarning.Render(severity)
	}

	return fmt.Sprintf("  %-30s %s  %s  %s",
		BodyStyle.Render(name),
		SizeStyle.Render(fmt.Sprintf("%10s", size)),
		severityStyled,
		DimStyle.Render(path),
	)
}

// Summary prints a summary box with totals.
func Summary(totalSize string, findingCount int, ruleCount int) {
	content := fmt.Sprintf(
		"%s found across %s (%s rules matched)",
		SizeStyle.Render(totalSize),
		BodyStyle.Render(fmt.Sprintf("%d findings", findingCount)),
		DimStyle.Render(fmt.Sprintf("%d", ruleCount)),
	)
	fmt.Fprintf(os.Stderr, "\n%s\n", BoxStyle.Render("𓂀 "+content))
}
