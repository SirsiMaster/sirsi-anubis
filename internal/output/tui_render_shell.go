// Package output — tui_render_shell.go
//
// Reusable rendering primitives: banners, gauges, sparklines, helpers.
package output

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── Pantheon Gauge ───────────────────────────────────────────────
// Gold-filled fractional bars with Lapis empty track. 8 sub-blocks
// per character = 160-step resolution on a 20-char bar. Color
// escalation uses the Pantheon palette: Gold → Red at pressure.

var subBlocks = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}

// ProgressBar renders a Pantheon-branded gauge: ████████▒▒▒ 72%
// Gold fill below 60%, warm amber 60-85%, Red above 85%.
// Empty track uses Deep Lapis — not generic gray.
func ProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Pantheon color escalation
	barColor := Gold // brand gold is the default healthy state
	switch {
	case percent >= 85:
		barColor = Red
	case percent >= 60:
		barColor = Yellow // warm escalation before danger
	}
	filled := lipgloss.NewStyle().Foreground(barColor)
	track := lipgloss.NewStyle().Foreground(Lapis) // deep lapis empty track

	totalUnits := float64(width) * 8
	filledUnits := int(percent / 100 * totalUnits)
	fullBlocks := filledUnits / 8
	remainder := filledUnits % 8

	var bar string
	bar += filled.Render(repeatRune('█', fullBlocks))
	if fullBlocks < width {
		bar += filled.Render(string(subBlocks[remainder]))
		bar += track.Render(repeatRune('▒', width-fullBlocks-1))
	}

	pctLabel := lipgloss.NewStyle().Foreground(Gold).Bold(true)
	return bar + pctLabel.Render(fmt.Sprintf(" %3.0f%%", percent))
}

// ScoreBar renders a Pantheon gauge for 0-100 scores.
// Gold at top, Red at bottom — the Feather weighs favorably.
func ScoreBar(score int, width int) string {
	barColor := Red
	switch {
	case score >= 75:
		barColor = Gold
	case score >= 50:
		barColor = Yellow
	}
	filled := lipgloss.NewStyle().Foreground(barColor)
	track := lipgloss.NewStyle().Foreground(Lapis)

	pct := float64(score) / 100
	totalUnits := float64(width) * 8
	filledUnits := int(pct * totalUnits)
	fullBlocks := filledUnits / 8
	remainder := filledUnits % 8

	var bar string
	bar += filled.Render(repeatRune('█', fullBlocks))
	if fullBlocks < width {
		bar += filled.Render(string(subBlocks[remainder]))
		bar += track.Render(repeatRune('▒', width-fullBlocks-1))
	}
	return bar
}

// ── Sparkline ───────────────────────────────────────────────────
// Mini history chart using vertical block characters. 8 levels
// from ▁ (floor) to █ (ceiling). Gold-tinted by default.

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders a mini history chart. vals are 0-100, width is char count.
func Sparkline(vals []float64, width int, c color.Color) string {
	if len(vals) == 0 {
		return lipgloss.NewStyle().Foreground(Lapis).
			Render(strings.Repeat(string(sparkChars[0]), width))
	}

	// Take last `width` values
	start := 0
	if len(vals) > width {
		start = len(vals) - width
	}
	visible := vals[start:]

	var buf strings.Builder
	for _, v := range visible {
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		idx := int(v / 100 * 7)
		if idx > 7 {
			idx = 7
		}
		buf.WriteRune(sparkChars[idx])
	}
	// Pad with floor sparks if not enough data
	for i := len(visible); i < width; i++ {
		buf.WriteRune(sparkChars[0])
	}

	return lipgloss.NewStyle().Foreground(c).Render(buf.String())
}

func repeatRune(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]rune, n)
	for i := range out {
		out[i] = r
	}
	return string(out)
}

// ── Pantheon Banner ─────────────────────────────────────────────
// Cartouche-style decree banner. Gold borders with hieroglyphic
// markers — every result is a judgment from the scales.

// ResultBanner renders 𓊝━━━┫ message ┣━━━𓊝 in the given style.
func ResultBanner(message string, style lipgloss.Style, width int) string {
	msgLen := len(message) + 6 // ┫ + spaces + message + spaces + ┣
	if width < msgLen+8 {
		width = msgLen + 8
	}
	sideLen := (width - msgLen) / 2
	border := lipgloss.NewStyle().Foreground(Gold)
	left := border.Render("𓊝" + repeatRune('━', sideLen) + "┫ ")
	right := border.Render(" ┣" + repeatRune('━', sideLen) + "𓊝")
	return left + style.Render(message) + right
}

func categoryIcon(cat string) string {
	switch cat {
	case "cache":
		return "𓊗" // vessel — things poured out
	case "logs":
		return "𓏛" // papyrus scroll
	case "build":
		return "𓍹" // chisel
	case "containers":
		return "𓊖" // enclosure
	case "dev-tools", "dev":
		return "𓌙" // tool
	case "packages":
		return "𓎟" // bundle
	case "ai":
		return "𓂀" // eye of Horus
	case "ides":
		return "𓉔" // house/workshop
	case "cloud":
		return "𓇼" // star (sky)
	case "storage":
		return "𓋹" // ankh (life/data)
	case "vms":
		return "𓊝" // cartouche
	default:
		return "𓃀" // foot (path)
	}
}

func splitLines(s string) []string {
	result := []string{}
	start := 0
	for i, c := range s {
		if c == '\n' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
