// Package output — tui_render_interactive.go
//
// TUI-interactive render functions: checkbox selection, disk analyzer.
package output

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
)

// ── Checkbox Selection ──────────────────────────────────────────

func (m TUIModel) renderSelect() string {
	var b strings.Builder
	gold := lipgloss.NewStyle().Foreground(Gold)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	sizeStyle := lipgloss.NewStyle().Foreground(Gold)
	selectedBox := gold.Render("☑")
	unselectedBox := dim.Render("☐")
	cursor := gold.Render("▸")
	sep := dim.Render(strings.Repeat("━", min(m.width-6, 60)))

	b.WriteString("\n")
	b.WriteString("  " + gold.Render(m.selectTitle) + "\n\n")

	maxVisible := m.height - 14
	if maxVisible < 5 {
		maxVisible = 5
	}
	startIdx := 0
	if len(m.selectItems) > maxVisible && m.selectCursor >= maxVisible {
		startIdx = m.selectCursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(m.selectItems) {
		endIdx = len(m.selectItems)
	}

	for i := startIdx; i < endIdx; i++ {
		item := m.selectItems[i]
		prefix := "    "
		if i == m.selectCursor {
			prefix = "  " + cursor + " "
		}
		box := unselectedBox
		if item.Selected {
			box = selectedBox
		}
		sizeStr := ""
		if item.Size > 0 {
			sizeStr = sizeStyle.Render(fmt.Sprintf("%8s", jackal.FormatSize(item.Size)))
		}
		labelStyle := body
		if i == m.selectCursor {
			labelStyle = lipgloss.NewStyle().Foreground(White).Bold(true)
		}
		line := prefix + box + " " + labelStyle.Render(item.Label)
		if sizeStr != "" {
			line += "  " + sizeStr
		}
		b.WriteString(line + "\n")
		if item.Detail != "" {
			b.WriteString("     " + dim.Render(item.Detail) + "\n")
		}
	}

	if len(m.selectItems) > maxVisible {
		if startIdx > 0 {
			b.WriteString("    " + dim.Render(fmt.Sprintf("↑ %d more above", startIdx)) + "\n")
		}
		if endIdx < len(m.selectItems) {
			b.WriteString("    " + dim.Render(fmt.Sprintf("↓ %d more below", len(m.selectItems)-endIdx)) + "\n")
		}
	}

	b.WriteString("\n  " + sep + "\n")

	selectedCount := 0
	var selectedSize int64
	for _, item := range m.selectItems {
		if item.Selected {
			selectedCount++
			selectedSize += item.Size
		}
	}
	summary := fmt.Sprintf("%d selected", selectedCount)
	if selectedSize > 0 {
		summary += " · " + jackal.FormatSize(selectedSize) + " to purge"
	}
	b.WriteString("  " + gold.Render(summary) + "\n\n")

	return b.String()
}

// ── Disk Space Analyzer ─────────────────────────────────────────

func (m TUIModel) renderAnalyze() string {
	var b strings.Builder
	gold := lipgloss.NewStyle().Foreground(Gold)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	sizeStyle := lipgloss.NewStyle().Foreground(Gold).Bold(true)
	pctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	cursor := gold.Render("▸")
	barFill := lipgloss.NewStyle().Foreground(Gold)
	barTrack := lipgloss.NewStyle().Foreground(Lapis)

	displayPath := ShortenPath(m.analyzePath)
	b.WriteString("\n")
	b.WriteString("  " + gold.Bold(true).Render("Analyze") + "  " +
		body.Render(displayPath) + "  " +
		dim.Render("|") + "  " +
		dim.Render("Total: ") + sizeStyle.Render(jackal.FormatSize(m.analyzeTotal)) + "\n")
	b.WriteString("\n")

	if len(m.analyzeEntries) == 0 {
		b.WriteString("  " + dim.Render("Empty directory") + "\n")
		return b.String()
	}

	barWidth := 20
	maxVisible := 15

	startIdx := 0
	if len(m.analyzeEntries) > maxVisible && m.analyzeCursor >= maxVisible {
		startIdx = m.analyzeCursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(m.analyzeEntries) {
		endIdx = len(m.analyzeEntries)
	}

	if startIdx > 0 {
		b.WriteString("    " + dim.Render(fmt.Sprintf("  ↑ %d more above", startIdx)) + "\n")
	}

	for i := startIdx; i < endIdx; i++ {
		entry := m.analyzeEntries[i]
		pct := float64(0)
		if m.analyzeTotal > 0 {
			pct = float64(entry.Size) / float64(m.analyzeTotal) * 100
		}

		// Proportional bar with sub-block precision
		filledUnits := int(pct / 100 * float64(barWidth) * 8)
		fullBlocks := filledUnits / 8
		remainder := filledUnits % 8
		var bar string
		bar += barFill.Render(repeatRune('█', fullBlocks))
		if fullBlocks < barWidth {
			bar += barFill.Render(string(subBlocks[remainder]))
			bar += barTrack.Render(repeatRune('▒', barWidth-fullBlocks-1))
		}

		name := entry.Name
		if len(name) > 24 {
			name = name[:21] + "..."
		}

		prefix := "    "
		numStyle := dim
		nameStyle := body
		if i == m.analyzeCursor {
			prefix = "  " + cursor + " "
			numStyle = gold
			nameStyle = lipgloss.NewStyle().Foreground(White).Bold(true)
		}

		line := prefix +
			numStyle.Render(fmt.Sprintf("%2d.", i+1)) + " " +
			bar + "  " +
			pctStyle.Render(fmt.Sprintf("%5.1f%%", pct)) + "  " +
			dim.Render("|") + "  " +
			nameStyle.Render(fmt.Sprintf("%-24s", name)) + " " +
			sizeStyle.Render(fmt.Sprintf("%8s", jackal.FormatSize(entry.Size)))
		if entry.IsDir {
			line += "  " + dim.Render(">")
		}
		b.WriteString(line + "\n")
	}

	if endIdx < len(m.analyzeEntries) {
		b.WriteString("    " + dim.Render(fmt.Sprintf("  ↓ %d more below", len(m.analyzeEntries)-endIdx)) + "\n")
	}
	b.WriteString("\n")

	// Breadcrumb trail
	if len(m.analyzeHistory) > 0 {
		trail := dim.Render("  Path: ")
		for i, lvl := range m.analyzeHistory {
			if i > 0 {
				trail += dim.Render(" > ")
			}
			parts := strings.Split(lvl.path, string(filepath.Separator))
			if len(parts) > 0 {
				trail += dim.Render(parts[len(parts)-1])
			}
		}
		parts := strings.Split(m.analyzePath, string(filepath.Separator))
		if len(parts) > 0 {
			trail += dim.Render(" > ") + gold.Render(parts[len(parts)-1])
		}
		b.WriteString(trail + "\n")
	}

	return b.String()
}
