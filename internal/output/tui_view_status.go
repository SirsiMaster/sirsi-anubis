// Package output — tui_view_status.go
//
// Status dashboard rendering: renderStatusPage, sideBySide, computeHealthScore.
// Extracted from tui_view.go to keep files under ~500 lines.
package output

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
)

// renderStatusPage renders the live real-time status dashboard.
func (m TUIModel) renderStatusPage(gold, dim lipgloss.Style) string {
	var b strings.Builder
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	maxW := min(m.width-4, 116)
	colW := maxW/2 - 2
	barW := colW - 18
	if barW < 10 {
		barW = 10
	}
	sparkW := colW - 8
	if sparkW < 10 {
		sparkW = 10
	}

	// ── Health score ──
	healthScore := m.computeHealthScore()
	healthLabel := "THRIVING"
	healthColor := Gold
	switch {
	case healthScore < 50:
		healthLabel = "AILING"
		healthColor = Red
	case healthScore < 70:
		healthLabel = "STRAINED"
		healthColor = lipgloss.Color("#FFAA00")
	case healthScore < 85:
		healthLabel = "STABLE"
		healthColor = Yellow
	}

	machineInfo := m.vitals.ModelName
	if m.vitals.Accelerator != "" {
		machineInfo += " · " + m.vitals.Accelerator
	}
	if m.vitals.RAMTotalGB > 0 {
		machineInfo += fmt.Sprintf(" · %.0fGB", m.vitals.RAMTotalGB)
	}
	if m.vitals.OSVersion != "" {
		machineInfo += " · " + m.vitals.OSVersion
	}

	b.WriteString("\n")
	b.WriteString("  " + gold.Render("𓂀 Status") + "  " +
		lipgloss.NewStyle().Foreground(healthColor).Bold(true).Render(fmt.Sprintf("Health ● %d", healthScore)) +
		"  " + lipgloss.NewStyle().Foreground(healthColor).Render(healthLabel) +
		"  " + dim.Render(machineInfo) + "\n")
	b.WriteString("\n")

	// All labels are 8 chars: "Total   ", "Load    ", "Used    ", "Free    "
	// This creates a strict grid like Mole's mo status.
	lbl := func(s string) string { return fmt.Sprintf("%-8s", s) }

	// ── Row 1: CPU | Memory ──
	leftCol := "  " + gold.Render("CPU") + "\n"
	leftCol += fmt.Sprintf("  %s%s\n", dim.Render(lbl("Total")), ProgressBar(m.vitals.CPUPercent, barW))
	leftCol += fmt.Sprintf("  %s%s\n", dim.Render(lbl("Load")),
		body.Render(fmt.Sprintf("%.2f / %.2f / %.2f", m.vitals.CPULoadAvg[0], m.vitals.CPULoadAvg[1], m.vitals.CPULoadAvg[2])))
	leftCol += fmt.Sprintf("  %s%s\n", lbl(""), Sparkline(m.cpuHistory, sparkW, Gold))

	rightCol := "  " + gold.Render("Memory") + "\n"
	rightCol += fmt.Sprintf("  %s%s\n", dim.Render(lbl("Used")), ProgressBar(m.vitals.RAMPercent, barW))
	rightCol += fmt.Sprintf("  %s%s\n", dim.Render(lbl("Total")),
		body.Render(fmt.Sprintf("%.1f / %.1f GB", m.vitals.RAMUsedGB, m.vitals.RAMTotalGB)))
	rightCol += fmt.Sprintf("  %s%s\n", lbl(""), Sparkline(m.memHistory, sparkW, Gold))

	b.WriteString(sideBySide(leftCol, rightCol, colW))
	b.WriteString("\n")

	// ── Row 2: Disk | Network ──
	leftCol = "  " + gold.Render("Disk") + "\n"
	leftCol += fmt.Sprintf("  %s%s\n", dim.Render(lbl("Used")), ProgressBar(m.vitals.DiskPercent, barW))
	leftCol += fmt.Sprintf("  %s%s\n", dim.Render(lbl("Free")),
		body.Render(fmt.Sprintf("%.1f GB", m.vitals.DiskFreeGB)))

	downMBs := m.vitals.NetDownBps / (1024 * 1024)
	upMBs := m.vitals.NetUpBps / (1024 * 1024)
	rightCol = "  " + gold.Render("Network") + "\n"
	rightCol += fmt.Sprintf("  %s%s  %s\n", dim.Render(lbl("Down")),
		Sparkline(m.netDownHist, sparkW, Green),
		body.Render(fmt.Sprintf("%.2f MB/s", downMBs)))
	rightCol += fmt.Sprintf("  %s%s  %s\n", dim.Render(lbl("Up")),
		Sparkline(m.netUpHist, sparkW, lipgloss.Color("#51A9C8")),
		body.Render(fmt.Sprintf("%.2f MB/s", upMBs)))

	b.WriteString(sideBySide(leftCol, rightCol, colW))
	b.WriteString("\n")

	// ── Top Processes ──
	if len(m.vitals.TopProcs) > 0 {
		b.WriteString("  " + gold.Render("Processes") + "\n")
		maxCPU := m.vitals.TopProcs[0].CPUPercent
		if maxCPU < 1 {
			maxCPU = 1
		}
		for _, p := range m.vitals.TopProcs {
			pctNorm := int(p.CPUPercent / maxCPU * 100)
			name := p.Name
			if len(name) > 14 {
				name = name[:14]
			}
			b.WriteString(fmt.Sprintf("  %-14s %s  %5.1f%%\n",
				body.Render(name),
				ScoreBar(pctNorm, 5),
				p.CPUPercent))
		}
		b.WriteString("\n")
	}

	// ── Row 3: Deities | Recent ──
	leftCol = "  " + gold.Render("Deities") + "\n"
	for _, d := range deity.Roster {
		state := m.deityState[d.Key]
		var indicator, status string
		switch state {
		case stateSucceeded:
			indicator = lipgloss.NewStyle().Foreground(Green).Render("✓")
			status = dim.Render("healthy")
		case stateFailed:
			indicator = lipgloss.NewStyle().Foreground(Red).Render("✗")
			status = lipgloss.NewStyle().Foreground(Red).Render("failed")
		case stateHasData:
			indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Render("◆")
			status = gold.Render("has data")
		default:
			indicator = dim.Render("·")
			status = dim.Render("—")
		}
		leftCol += fmt.Sprintf("  %s %-10s %s\n", indicator, body.Render(d.Name), status)
	}

	rightCol = "  " + gold.Render("𓏛 Recent") + "\n"
	if len(m.recentNotify) > 0 {
		for i, n := range m.recentNotify {
			if i >= 5 {
				break
			}
			icon := notify.SeverityIcon(n.Severity)
			summary := n.Summary
			if len(summary) > 40 {
				summary = summary[:37] + "…"
			}
			rightCol += fmt.Sprintf("  %s %s\n", icon, dim.Render(summary))
		}
	} else {
		rightCol += "  " + dim.Render("No recent activity") + "\n"
	}

	b.WriteString(sideBySide(leftCol, rightCol, colW))
	b.WriteString("\n")

	// ── Numbered actions ──
	tab := tabs[m.activeTab]
	for i, action := range tab.Actions {
		keyBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F0F0F")).
			Background(Gold).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("%d", i+1))
		b.WriteString("  " + keyBadge +
			"  " + body.Render(action.Label) +
			"  " + dim.Render(action.Desc) + "\n")
	}

	return b.String()
}

// sideBySide places two multi-line strings side by side.
func sideBySide(left, right string, colWidth int) string {
	leftLines := strings.Split(strings.TrimRight(left, "\n"), "\n")
	rightLines := strings.Split(strings.TrimRight(right, "\n"), "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var b strings.Builder
	for i := 0; i < maxLines; i++ {
		l := ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		r := ""
		if i < len(rightLines) {
			r = rightLines[i]
		}
		padded := l + strings.Repeat(" ", max(0, colWidth-visibleLen(l)))
		b.WriteString(padded + r + "\n")
	}
	return b.String()
}

// computeHealthScore returns a 0-100 weighted health score.
func (m TUIModel) computeHealthScore() int {
	cpuScore := 100 - m.vitals.CPUPercent
	ramScore := 100 - m.vitals.RAMPercent
	diskScore := 100 - m.vitals.DiskPercent
	score := cpuScore*0.30 + ramScore*0.40 + diskScore*0.30
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return int(score)
}
