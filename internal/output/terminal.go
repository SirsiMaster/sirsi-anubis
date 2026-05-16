// Package output handles terminal rendering for Sirsi Pantheon.
// Uses the Pantheon brand language: Gold + Black + Deep Lapis (Rule A10).
package output

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
)

// SuggestSteps is a convenience that calls suggest.After and formats the
// result for NextSteps. Use this at the end of every deity CLI command:
//
//	output.Footer(elapsed)
//	output.NextSteps(output.SuggestSteps(suggest.Context{Deity: "anubis", Subcommand: "weigh"}))
func SuggestSteps(ctx suggest.Context) [][]string {
	actions := suggest.After(ctx)
	steps := make([][]string, 0, len(actions))
	for _, a := range actions {
		steps = append(steps, []string{a.Command, a.Description})
	}
	return steps
}

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

	// Info style (Thoth Lapis)
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#51A9C8"))

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
			Padding(0, 2).
			MarginTop(1)

	// Column style
	ColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(Lapis).
			Padding(0, 1).
			MarginRight(2)

	// Value style
	ValueStyle = lipgloss.NewStyle().
			Foreground(White).
			Bold(true)

	// Dashboard style
	DashboardStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(Gold)

	// Severity styles
	SeveritySafe    = lipgloss.NewStyle().Foreground(Green)
	SeverityCaution = lipgloss.NewStyle().Foreground(Yellow)
	SeverityWarning = lipgloss.NewStyle().Foreground(Red)
)

// inTUI returns true when running as a subprocess inside the Pantheon TUI.
// CLI chrome (banners, section headers, footers) is suppressed so the TUI
// can provide its own consistent presentation.
func inTUI() bool {
	return os.Getenv("SIRSI_TUI") == "1"
}

// spinnerFrames are the animation frames for the CLI spinner.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner starts a CLI progress spinner with the given label.
// Returns a stop function that clears the spinner line.
// Suppressed in TUI, JSON, and quiet modes — returns a no-op stop.
func Spinner(label string) func() {
	if inTUI() || os.Getenv("SIRSI_JSON") == "1" {
		return func() {}
	}

	var once sync.Once
	done := make(chan struct{})
	gold := lipgloss.NewStyle().Foreground(Gold)

	go func() {
		i := 0
		for {
			select {
			case <-done:
				fmt.Fprintf(os.Stderr, "\r\033[K") // clear line
				return
			default:
				frame := gold.Render(spinnerFrames[i%len(spinnerFrames)])
				fmt.Fprintf(os.Stderr, "\r  %s %s", frame, DimStyle.Render(label))
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	return func() {
		once.Do(func() { close(done) })
		time.Sleep(100 * time.Millisecond) // let goroutine clear the line
	}
}

// Banner prints the Pantheon banner. Suppressed inside TUI.
func Banner() {
	if inTUI() {
		return
	}
	banner := TitleStyle.Render(`
   P A N T H E O N
   ───────────────────────────────
   Infrastructure Hygiene Platform
   "One Install. Everything Clean."
`)
	fmt.Fprintln(os.Stderr, banner)
}

// Header prints a section header with the 𓁢 prefix. Suppressed inside TUI.
func Header(text string) {
	if inTUI() {
		return
	}
	fmt.Fprintf(os.Stderr, "\n%s\n", HeaderStyle.Render("𓁢 "+text))
}

// Info prints a themed informational message.
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s %s\n", InfoStyle.Render("𓁟"), BodyStyle.Render(msg))
}

// Dim prints a dimmed message.
func Dim(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s\n", DimStyle.Render(msg))
}

// Success prints a themed success message.
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s %s\n", SuccessStyle.Render("𓆄"), BodyStyle.Render(msg))
}

// Warn prints a themed warning message.
func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s %s\n", WarningStyle.Render("⚠️"), BodyStyle.Render(msg))
}

// Error prints a themed error message.
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "  %s %s\n", ErrorStyle.Render("𓁵"), BodyStyle.Render(msg))
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

// Dashboard prints a multi-column summary dashboard.
func Dashboard(metrics map[string]string) {
	var cols []string

	for label, value := range metrics {
		col := ColumnStyle.Render(
			fmt.Sprintf("%s\n%s",
				DimStyle.Render(label),
				ValueStyle.Render(value),
			),
		)
		cols = append(cols, col)
	}

	dash := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	fmt.Fprintf(os.Stderr, "\n%s\n", DashboardStyle.Render(dash))
}

// Table displays results in a beautiful TUI table.
func Table(headers []string, rows [][]string) {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(Gold)).
		Headers(headers...).
		Rows(rows...)

	fmt.Fprintf(os.Stderr, "\n%s\n", t.Render())
}

// Summary prints a summary box with totals. Suppressed inside TUI.
func Summary(totalSize string, findingCount int, ruleCount int) {
	if inTUI() {
		return
	}
	content := fmt.Sprintf(
		"%s found across %s (%s rules matched)",
		SizeStyle.Render(totalSize),
		BodyStyle.Render(fmt.Sprintf("%d findings", findingCount)),
		DimStyle.Render(fmt.Sprintf("%d", ruleCount)),
	)
	fmt.Fprintf(os.Stderr, "\n%s\n", BoxStyle.Render("𓁢 "+content))
}

// Footer prints the completion ritual with elapsed time. Suppressed inside TUI.
func Footer(elapsed time.Duration) {
	if inTUI() {
		return
	}
	fmt.Fprintf(os.Stderr, "\n  %s %s\n",
		DimStyle.Render("Completed in"),
		ValueStyle.Render(elapsed.Round(time.Millisecond).String()),
	)
}

// FooterWithSuggestions renders the elapsed time footer followed by
// contextual "What's Next" suggestions from the suggest package.
// This is the standard way to end any deity CLI command.
func FooterWithSuggestions(elapsed time.Duration, actions [][]string) {
	Footer(elapsed)
	if len(actions) > 0 {
		NextSteps(actions)
	}
}

func Section(title string) {
	if inTUI() {
		return
	}
	fmt.Fprintf(os.Stderr, "\n%s\n", TitleStyle.Render("𓁢 "+title))
}

// NextSteps renders a contextual "What's Next" section in the terminal
// after a command completes. Each step is a command + description pair.
// Suppressed when running inside the TUI (SIRSI_TUI=1) since the TUI
// renders its own sticky suggestions.
func NextSteps(steps [][]string) {
	if os.Getenv("SIRSI_TUI") == "1" {
		return
	}
	dim := DimStyle
	gold := TitleStyle

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  %s\n\n", dim.Render("── What's Next ──────────────────────────"))
	for _, step := range steps {
		if len(step) >= 2 {
			fmt.Fprintf(os.Stderr, "  %s  %s\n", gold.Render(padRight(step[0], 22)), dim.Render(step[1]))
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// padRight pads a string to a minimum width with spaces.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// shortenPath replaces home dir with ~ and truncates long paths.
func ShortenPath(path string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(path, home) {
		path = "~" + strings.TrimPrefix(path, home)
	}
	if len(path) > 60 {
		return "..." + path[len(path)-57:]
	}
	return path
}

// Truncate shortens a string to a max length with an ellipsis.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
