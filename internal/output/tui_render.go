// Package output — tui_render.go
//
// Shared render styles used across all tui_render_*.go files.
// Actual renderers live in:
//   tui_render_shell.go       — reusable primitives (banners, gauges, sparklines)
//   tui_render_status.go      — command result renderers (scan, ghost, clean, hw, risk)
//   tui_render_detail.go      — detailed report renderers (network, doctor, maat, etc.)
//   tui_render_interactive.go — TUI-interactive renderers (select, analyze)
package output

import "charm.land/lipgloss/v2"

// Shared render styles
var (
	rGold  = lipgloss.NewStyle().Foreground(Gold).Bold(true)
	rDim   = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	rBody  = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	rBig   = lipgloss.NewStyle().Foreground(White).Bold(true)
	rGreen = lipgloss.NewStyle().Foreground(Green)
	rRed   = lipgloss.NewStyle().Foreground(Red)
	rWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00"))
	rLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)
