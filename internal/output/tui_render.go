// Package output — tui_render.go
//
// Native TUI renderers for deity results. Each function takes a result
// struct from a deity package and returns styled lines for the viewport.
// No subprocess output. No ANSI parsing. Clean lipgloss rendering.
package output

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/osiris"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
)

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

// ── Scan Result ──────────────────────────────────────────────────────

func RenderScanResult(res *jackal.ScanResult) []string {
	var lines []string

	// Big number hero
	lines = append(lines, "")
	lines = append(lines, "  "+rBig.Render(jackal.FormatSize(res.TotalSize))+"  "+rDim.Render("reclaimable"))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d findings across %d rules", len(res.Findings), res.RulesRan)))
	lines = append(lines, "")

	// Category breakdown
	if len(res.ByCategory) > 0 {
		lines = append(lines, "  "+rLabel.Render("CATEGORIES"))
		for cat, summary := range res.ByCategory {
			icon := categoryIcon(string(cat))
			lines = append(lines, fmt.Sprintf("  %s %-14s %s  %s",
				icon,
				rBody.Render(string(cat)),
				rGold.Render(jackal.FormatSize(summary.TotalSize)),
				rDim.Render(fmt.Sprintf("%d items", summary.Findings))))
		}
		lines = append(lines, "")
	}

	// Top findings (max 10)
	limit := min(len(res.Findings), 10)
	if limit > 0 {
		lines = append(lines, "  "+rLabel.Render("TOP FINDINGS"))
		for _, f := range res.Findings[:limit] {
			severity := rGreen.Render("safe")
			switch f.Severity {
			case jackal.SeverityCaution:
				severity = rWarn.Render("caution")
			case jackal.SeverityWarning:
				severity = rRed.Render("warning")
			}
			lines = append(lines, fmt.Sprintf("  %-9s %8s  %s",
				severity,
				jackal.FormatSize(f.SizeBytes),
				rBody.Render(f.Description)))
			lines = append(lines, "  "+rDim.Render("         "+ShortenPath(f.Path)))
		}
		if len(res.Findings) > limit {
			lines = append(lines, "  "+rDim.Render(fmt.Sprintf("         ... and %d more", len(res.Findings)-limit)))
		}
	}

	return lines
}

// ── Ghost Result ─────────────────────────────────────────────────────

func RenderGhostResult(ghosts []ka.Ghost) []string {
	var lines []string

	if len(ghosts) == 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+rGreen.Render("✓")+"  "+rBody.Render("No ghost apps found. Clean machine."))
		lines = append(lines, "")
		return lines
	}

	// Hero
	totalSize := int64(0)
	totalFiles := 0
	for _, g := range ghosts {
		totalSize += g.TotalSize
		totalFiles += g.TotalFiles
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rBig.Render(fmt.Sprintf("%d", len(ghosts)))+"  "+rDim.Render("ghost apps found"))
	lines = append(lines, "  "+rGold.Render(jackal.FormatSize(totalSize))+"  "+rDim.Render(fmt.Sprintf("across %d residual files", totalFiles)))
	lines = append(lines, "")

	// Ghost list
	lines = append(lines, "  "+rLabel.Render("GHOSTS"))
	for _, g := range ghosts {
		lines = append(lines, fmt.Sprintf("  %s  %s  %s",
			rBody.Render(g.AppName),
			rGold.Render(jackal.FormatSize(g.TotalSize)),
			rDim.Render(fmt.Sprintf("%d files", g.TotalFiles))))
		// Show top residual paths (max 2)
		shown := 0
		for _, r := range g.Residuals {
			if shown >= 2 {
				break
			}
			lines = append(lines, "     "+rDim.Render(ShortenPath(r.Path)))
			shown++
		}
		if len(g.Residuals) > 2 {
			lines = append(lines, "     "+rDim.Render(fmt.Sprintf("... +%d more", len(g.Residuals)-2)))
		}
	}

	return lines
}

// ── Hardware Profile ─────────────────────────────────────────────────

func RenderHardwareProfile(hw *seba.HardwareProfile) []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("CPU"))
	lines = append(lines, "  "+rBig.Render(hw.CPUModel))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d cores · %s", hw.CPUCores, hw.CPUArch)))
	lines = append(lines, "")

	lines = append(lines, "  "+rLabel.Render("MEMORY"))
	lines = append(lines, "  "+rBig.Render(fmt.Sprintf("%.0f GB", float64(hw.TotalRAM)/(1024*1024*1024))))
	lines = append(lines, "")

	lines = append(lines, "  "+rLabel.Render("GPU"))
	lines = append(lines, "  "+rBig.Render(hw.GPU.Name))
	if hw.GPU.MetalFamily != "" {
		lines = append(lines, "  "+rDim.Render(hw.GPU.MetalFamily))
	}
	lines = append(lines, "")

	if hw.NeuralEngine {
		lines = append(lines, "  "+rLabel.Render("NEURAL ENGINE"))
		lines = append(lines, "  "+rGreen.Render("✓")+"  "+rBody.Render("Active"))
		lines = append(lines, "")
	}

	lines = append(lines, "  "+rLabel.Render("PLATFORM"))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%s · kernel %s", hw.OS, hw.Kernel)))

	return lines
}

// ── Risk Assessment ──────────────────────────────────────────────────

func RenderRiskAssessment(cp *osiris.Checkpoint) []string {
	var lines []string

	// Risk level hero
	riskStyle := rGreen
	riskLabel := "CLEAN"
	switch cp.Risk {
	case osiris.RiskLow:
		riskStyle = lipgloss.NewStyle().Foreground(Green)
		riskLabel = "LOW"
	case osiris.RiskModerate:
		riskStyle = rWarn
		riskLabel = "MODERATE"
	case osiris.RiskHigh:
		riskStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600"))
		riskLabel = "HIGH"
	case osiris.RiskCritical:
		riskStyle = rRed
		riskLabel = "CRITICAL"
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("RISK"))
	lines = append(lines, "  "+riskStyle.Bold(true).Render(riskLabel))
	if cp.Warning != "" {
		lines = append(lines, "  "+rWarn.Render(cp.Warning))
	}
	lines = append(lines, "")

	lines = append(lines, "  "+rLabel.Render("BRANCH"))
	lines = append(lines, "  "+rBody.Render(cp.Branch))
	lines = append(lines, "")

	lines = append(lines, "  "+rLabel.Render("CHANGES"))
	lines = append(lines, fmt.Sprintf("  %s uncommitted  %s staged  %s untracked",
		rBig.Render(fmt.Sprintf("%d", cp.UncommittedFiles)),
		rBody.Render(fmt.Sprintf("%d", cp.StagedFiles)),
		rDim.Render(fmt.Sprintf("%d", cp.UntrackedFiles))))
	if cp.LinesAdded > 0 || cp.LinesDeleted > 0 {
		lines = append(lines, fmt.Sprintf("  %s  %s",
			rGreen.Render(fmt.Sprintf("+%d", cp.LinesAdded)),
			rRed.Render(fmt.Sprintf("-%d", cp.LinesDeleted))))
	}
	lines = append(lines, "")

	lines = append(lines, "  "+rLabel.Render("LAST COMMIT"))
	lines = append(lines, "  "+rDim.Render(cp.LastCommitHash)+"  "+rBody.Render(cp.LastCommitMessage))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%s ago", cp.TimeSinceCommit.Truncate(1e9))))

	return lines
}

// ── Network Audit ────────────────────────────────────────────────────

// RenderNetworkAudit renders the network security posture results.
// Returns lines for the viewport AND fixable items as post-run commands.
func RenderNetworkAudit(report *guard.NetworkReport) ([]string, []string) {
	var lines []string
	var fixCmds []string

	// Score hero
	scoreStyle := rGreen
	scoreLabel := "HEALTHY"
	switch {
	case report.Score < 50:
		scoreStyle = rRed
		scoreLabel = "AT RISK"
	case report.Score < 75:
		scoreStyle = rWarn
		scoreLabel = "NEEDS ATTENTION"
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("SECURITY SCORE"))
	lines = append(lines, "  "+scoreStyle.Bold(true).Render(fmt.Sprintf("%d/100", report.Score))+"  "+scoreStyle.Render(scoreLabel))
	lines = append(lines, "")

	// Findings with per-item status
	lines = append(lines, "  "+rLabel.Render("CHECKS"))
	fixIdx := 0
	for _, f := range report.Findings {
		var icon string
		switch f.Severity {
		case guard.SeverityOK:
			icon = rGreen.Render("✓")
		case guard.SeverityInfo:
			icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#51A9C8")).Render("ℹ")
		case guard.SeverityWarn:
			icon = rWarn.Render("!")
		default:
			icon = rRed.Render("✗")
		}

		line := fmt.Sprintf("  %s  %s", icon, rBody.Render(f.Check))
		lines = append(lines, line)
		lines = append(lines, "     "+rDim.Render(f.Message))

		// If failed and fixable, add a numbered fix action
		if f.Severity >= guard.SeverityWarn {
			fixIdx++
			if f.Detail != "" {
				lines = append(lines, "     "+rDim.Render(f.Detail))
			}
		}
		lines = append(lines, "")
	}

	// Build fix commands — "isis network --fix" handles all fixable items
	hasFailed := false
	for _, f := range report.Findings {
		if f.Severity >= guard.SeverityWarn {
			hasFailed = true
			break
		}
	}
	if hasFailed {
		fixCmds = append(fixCmds, "isis network --fix")
	}

	return lines, fixCmds
}

// ── Doctor ───────────────────────────────────────────────────────────

// RenderDoctorReport renders the system health diagnostic.
func RenderDoctorReport(report *guard.DoctorReport) ([]string, []string) {
	var lines []string
	var fixCmds []string

	// Score hero
	scoreStyle := rGreen
	scoreLabel := "EXCELLENT"
	switch {
	case report.Score < 50:
		scoreStyle = rRed
		scoreLabel = "POOR"
	case report.Score < 75:
		scoreStyle = rWarn
		scoreLabel = "FAIR"
	case report.Score < 90:
		scoreStyle = rBody
		scoreLabel = "GOOD"
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("HEALTH SCORE"))
	lines = append(lines, "  "+scoreStyle.Bold(true).Render(fmt.Sprintf("%d/100", report.Score))+"  "+scoreStyle.Render(scoreLabel))
	lines = append(lines, "")

	lines = append(lines, "  "+rLabel.Render("CHECKS"))
	for _, f := range report.Findings {
		var icon string
		switch f.Severity {
		case guard.SeverityOK:
			icon = rGreen.Render("✓")
		case guard.SeverityInfo:
			icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#51A9C8")).Render("ℹ")
		case guard.SeverityWarn:
			icon = rWarn.Render("!")
		default:
			icon = rRed.Render("✗")
		}

		lines = append(lines, fmt.Sprintf("  %s  %s", icon, rBody.Render(f.Check)))
		lines = append(lines, "     "+rDim.Render(f.Message))
		lines = append(lines, "")
	}

	return lines, fixCmds
}

// ── Helpers ──────────────────────────────────────────────────────────

func categoryIcon(cat string) string {
	switch cat {
	case "cache":
		return "🗑"
	case "logs":
		return "📋"
	case "build":
		return "🔨"
	case "containers":
		return "🐳"
	case "dev-tools", "dev":
		return "🔧"
	case "packages":
		return "📦"
	case "ai":
		return "🤖"
	case "ides":
		return "💻"
	case "cloud":
		return "☁️"
	case "storage":
		return "💾"
	case "vms":
		return "🖥"
	default:
		return "📁"
	}
}
