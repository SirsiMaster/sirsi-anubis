// Package output — tui_render.go
//
// Native TUI renderers for deity results. Each function takes a result
// struct from a deity package and returns styled lines for the viewport.
// No subprocess output. No ANSI parsing. Clean lipgloss rendering.
package output

import (
	"fmt"
	"image/color"
	"strings"
	"syscall"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/maat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/mirror"
	"github.com/SirsiMaster/sirsi-pantheon/internal/osiris"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ra"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seshat"
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

	// Hero banner — Anubis weighs
	lines = append(lines, "")
	bannerText := fmt.Sprintf("𓃣 Weighed: %s reclaimable", jackal.FormatSize(res.TotalSize))
	bannerStyle := rGold
	if res.TotalSize > 10*1024*1024*1024 {
		bannerStyle = rRed
	} else if res.TotalSize > 5*1024*1024*1024 {
		bannerStyle = rWarn
	}
	lines = append(lines, "  "+ResultBanner(bannerText, bannerStyle, 50))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d findings across %d rules", len(res.Findings), res.RulesRan)))
	lines = append(lines, "")

	// Category breakdown
	if len(res.ByCategory) > 0 {
		lines = append(lines, "  "+rLabel.Render("CATEGORIES"))
		for cat, summary := range res.ByCategory {
			icon := categoryIcon(string(cat))
			lines = append(lines, fmt.Sprintf("  %s %-14s %8s  %s",
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
			lines = append(lines, "     "+rDim.Render(ShortenPath(f.Path)))
		}
		if len(res.Findings) > limit {
			lines = append(lines, "     "+rDim.Render(fmt.Sprintf("... and %d more", len(res.Findings)-limit)))
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
		lines = append(lines, fmt.Sprintf("  %-20s %8s  %s",
			rBody.Render(g.AppName),
			rGold.Render(jackal.FormatSize(g.TotalSize)),
			rDim.Render(fmt.Sprintf("%d files", g.TotalFiles))))
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
		lines = append(lines, "")
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
		rBig.Render(fmt.Sprintf("%3d", cp.UncommittedFiles)),
		rBody.Render(fmt.Sprintf("%3d", cp.StagedFiles)),
		rDim.Render(fmt.Sprintf("%3d", cp.UntrackedFiles))))
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

// ── Clean Preview ────────────────────────────────────────────────────

func RenderCleanPreview(findings []jackal.Finding) []string {
	var lines []string
	totalSize := int64(0)
	for _, f := range findings {
		totalSize += f.SizeBytes
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("CLEANUP PREVIEW"))
	lines = append(lines, "  "+rBig.Render(jackal.FormatSize(totalSize))+"  "+rDim.Render("will be moved to Trash"))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d safe items", len(findings))))
	lines = append(lines, "")

	// Show what will be cleaned
	limit := min(len(findings), 15)
	for _, f := range findings[:limit] {
		lines = append(lines, fmt.Sprintf("  %s  %8s  %s",
			rGreen.Render("✓"),
			jackal.FormatSize(f.SizeBytes),
			rBody.Render(f.Description)))
		lines = append(lines, "     "+rDim.Render(ShortenPath(f.Path)))
	}
	if len(findings) > limit {
		lines = append(lines, "     "+rDim.Render(fmt.Sprintf("... and %d more", len(findings)-limit)))
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rWarn.Render("Press 1 to confirm cleanup"))

	return lines
}

// RenderCleanResult shows what was freed after cleaning.
func RenderCleanResult(result *jackal.CleanResult) []string {
	var lines []string

	lines = append(lines, "")
	if result.BytesFreed > 0 {
		bannerText := fmt.Sprintf("𓃣 Purged: %s freed", jackal.FormatSize(result.BytesFreed))
		lines = append(lines, "")
		lines = append(lines, "  "+ResultBanner(bannerText, rGold, 50))
		lines = append(lines, "")
		// Show current free space — the payoff line
		var stat syscall.Statfs_t
		if err := syscall.Statfs("/", &stat); err == nil {
			freeBytes := int64(stat.Bavail) * int64(stat.Bsize)
			lines = append(lines, "  "+rBody.Render(fmt.Sprintf("%d items cleaned, %d skipped", result.Cleaned, result.Skipped))+
				"  "+rGold.Render("·")+"  "+rBig.Render(jackal.FormatSize(freeBytes))+" "+rDim.Render("free now"))
		} else {
			lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d items cleaned, %d skipped", result.Cleaned, result.Skipped)))
		}
	} else {
		lines = append(lines, "  "+rDim.Render("No space freed."))
		if result.Skipped > 0 {
			lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d items skipped (protected or in use)", result.Skipped)))
		}
	}

	if len(result.Errors) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+rLabel.Render("ERRORS"))
		for _, e := range result.Errors {
			lines = append(lines, "  "+rRed.Render("✗")+"  "+rDim.Render(e.Error()))
		}
	}

	return lines
}

// ── Network Audit ────────────────────────────────────────────────────

// RenderNetworkAudit renders the network security posture results.
// Returns lines for the viewport AND fixable items as post-run commands.
func RenderNetworkAudit(report *guard.NetworkReport) ([]string, []string) {
	var lines []string
	var fixCmds []string

	// Score hero — Isis guards the network
	scoreStyle := rGold
	scoreLabel := "FORTIFIED"
	switch {
	case report.Score < 50:
		scoreStyle = rRed
		scoreLabel = "EXPOSED"
	case report.Score < 75:
		scoreStyle = rWarn
		scoreLabel = "VULNERABLE"
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("𓁐 SECURITY POSTURE"))
	lines = append(lines, "  "+scoreStyle.Bold(true).Render(fmt.Sprintf("%d/100", report.Score))+"  "+scoreStyle.Render(scoreLabel))
	lines = append(lines, "  "+ScoreBar(report.Score, 20))
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

	// Score hero — the system's vital force
	scoreStyle := rGold
	scoreLabel := "THRIVING"
	switch {
	case report.Score < 50:
		scoreStyle = rRed
		scoreLabel = "AILING"
	case report.Score < 75:
		scoreStyle = rWarn
		scoreLabel = "STRAINED"
	case report.Score < 90:
		scoreStyle = rBody
		scoreLabel = "STABLE"
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("𓁐 VITALS"))
	lines = append(lines, "  "+scoreStyle.Bold(true).Render(fmt.Sprintf("%d/100", report.Score))+"  "+scoreStyle.Render(scoreLabel))
	lines = append(lines, "  "+ScoreBar(report.Score, 20))
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

// ── Mirror (Duplicates) ──────────────────────────────────────────────

func RenderMirrorResult(res *mirror.MirrorResult) []string {
	var lines []string
	lines = append(lines, "")
	if res.TotalDuplicates == 0 {
		lines = append(lines, "  "+rGreen.Render("✓")+"  "+rBody.Render("No duplicates found"))
		lines = append(lines, "  "+rDim.Render(fmt.Sprintf("Scanned %d files", res.TotalScanned)))
		return lines
	}

	lines = append(lines, "  "+rLabel.Render("DUPLICATES"))
	lines = append(lines, "  "+rBig.Render(fmt.Sprintf("%d", res.TotalDuplicates))+"  "+rDim.Render("duplicate files"))
	lines = append(lines, "  "+rGold.Render(jackal.FormatSize(res.TotalWasteBytes))+"  "+rDim.Render("reclaimable"))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("Scanned %d files in %s", res.TotalScanned, res.ScanDuration.Truncate(1e6))))
	lines = append(lines, "")

	limit := min(len(res.Groups), 10)
	for _, g := range res.Groups[:limit] {
		lines = append(lines, fmt.Sprintf("  %2d copies  %8s wasted",
			len(g.Files), jackal.FormatSize(g.WasteBytes)))
		for j, f := range g.Files {
			if j >= 3 {
				lines = append(lines, "     "+rDim.Render(fmt.Sprintf("... +%d more", len(g.Files)-3)))
				break
			}
			lines = append(lines, "     "+rDim.Render(ShortenPath(f.Path)))
		}
		lines = append(lines, "")
	}
	return lines
}

// ── Ma'at (Quality Audit) ────────────────────────────────────────────

func RenderMaatReport(report *maat.Report) ([]string, []string) {
	var lines []string
	var fixCmds []string

	verdictStyle := rGold
	verdictLabel := "WORTHY"
	switch report.OverallVerdict {
	case maat.VerdictWarning:
		verdictStyle = rWarn
		verdictLabel = "WANTING"
	case maat.VerdictFail:
		verdictStyle = rRed
		verdictLabel = "UNWORTHY"
	}

	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("𓆄 FEATHER WEIGHT"))
	lines = append(lines, "  "+verdictStyle.Bold(true).Render(fmt.Sprintf("%d/100", report.OverallWeight))+"  "+verdictStyle.Render(verdictLabel))
	lines = append(lines, "  "+ScoreBar(report.OverallWeight, 20))
	lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d passed  %d warnings  %d failures", report.Passes, report.Warnings, report.Failures)))
	lines = append(lines, "")

	lines = append(lines, "  "+rLabel.Render("ASSESSMENTS"))
	for _, a := range report.Assessments {
		icon := rGreen.Render("✓")
		if a.Verdict == maat.VerdictWarning {
			icon = rWarn.Render("!")
		} else if a.Verdict == maat.VerdictFail {
			icon = rRed.Render("✗")
		}
		lines = append(lines, fmt.Sprintf("  %s  %s  %s",
			icon, rBody.Render(a.Subject), rDim.Render(fmt.Sprintf("%d/100", a.FeatherWeight))))
		lines = append(lines, "     "+rDim.Render(a.Message))
		if a.Remediation != "" && a.Verdict != maat.VerdictPass {
			lines = append(lines, "     "+rWarn.Render("→ "+a.Remediation))
		}
		lines = append(lines, "")
	}

	if report.Failures > 0 {
		fixCmds = append(fixCmds, "maat heal")
	}
	return lines, fixCmds
}

// ── Seba Diagram ─────────────────────────────────────────────────────

func RenderDiagram(res *seba.DiagramResult) []string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render(fmt.Sprintf("DIAGRAM — %s", res.Title)))
	lines = append(lines, "")
	for _, line := range splitLines(res.Mermaid) {
		lines = append(lines, "  "+rDim.Render(line))
	}
	lines = append(lines, "")
	lines = append(lines, "  "+rDim.Render("Copy the Mermaid code above into mermaid.live to visualize"))
	return lines
}

// ── Seshat Knowledge ─────────────────────────────────────────────────

func RenderKnowledgeItems(items []seshat.KnowledgeItem) []string {
	var lines []string
	lines = append(lines, "")
	if len(items) == 0 {
		lines = append(lines, "  "+rDim.Render("No new knowledge items found in the last 24 hours"))
		return lines
	}
	lines = append(lines, "  "+rLabel.Render("KNOWLEDGE"))
	lines = append(lines, "  "+rBig.Render(fmt.Sprintf("%d", len(items)))+"  "+rDim.Render("items ingested"))
	lines = append(lines, "")

	limit := min(len(items), 10)
	for _, item := range items[:limit] {
		lines = append(lines, "  "+rBody.Render(item.Title))
		if item.Summary != "" {
			summary := item.Summary
			if len(summary) > 80 {
				summary = summary[:77] + "…"
			}
			lines = append(lines, "     "+rDim.Render(summary))
		}
		lines = append(lines, "")
	}
	if len(items) > limit {
		lines = append(lines, "  "+rDim.Render(fmt.Sprintf("... +%d more", len(items)-limit)))
	}
	return lines
}

// ── Ra Status ────────────────────────────────────────────────────────

func RenderRaStatus(status *ra.DeploymentStatus) []string {
	var lines []string
	lines = append(lines, "")

	if len(status.Windows) == 0 {
		lines = append(lines, "  "+rLabel.Render("RA STATUS"))
		lines = append(lines, "  "+rDim.Render("No active deployments"))
		return lines
	}

	allDoneLabel := rWarn.Render("IN PROGRESS")
	if status.AllDone {
		allDoneLabel = rGreen.Render("ALL DONE")
	}
	lines = append(lines, "  "+rLabel.Render("RA STATUS")+"  "+allDoneLabel)
	lines = append(lines, "")

	for _, w := range status.Windows {
		stateStyle := rDim
		icon := "·"
		switch w.State {
		case "running":
			stateStyle = rGold
			icon = "⟳"
		case "completed":
			stateStyle = rGreen
			icon = "✓"
		case "failed", "crashed":
			stateStyle = rRed
			icon = "✗"
		}
		lines = append(lines, fmt.Sprintf("  %s  %s  %s  %s",
			stateStyle.Render(icon),
			rBody.Render(w.Name),
			stateStyle.Render(w.State),
			rDim.Render(w.Duration.Truncate(1e9).String())))
	}
	return lines
}

// ── Horus Symbol Graph ───────────────────────────────────────────────

func RenderSymbolGraph(graph *horus.SymbolGraph) []string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, "  "+rLabel.Render("CODE GRAPH"))
	lines = append(lines, "")

	s := graph.Stats
	lines = append(lines, fmt.Sprintf("  %s files   %s packages   %s lines",
		rBig.Render(fmt.Sprintf("%5d", s.Files)),
		rBody.Render(fmt.Sprintf("%4d", s.Packages)),
		rDim.Render(fmt.Sprintf("%6d", s.TotalLines))))
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("  %s types   %s functions   %s methods   %s interfaces",
		rGold.Render(fmt.Sprintf("%4d", s.Types)),
		rBody.Render(fmt.Sprintf("%4d", s.Functions)),
		rBody.Render(fmt.Sprintf("%4d", s.Methods)),
		rDim.Render(fmt.Sprintf("%4d", s.Interfaces))))
	lines = append(lines, "")

	// Show top packages
	limit := min(len(graph.Packages), 10)
	if limit > 0 {
		lines = append(lines, "  "+rLabel.Render("PACKAGES"))
		for _, pkg := range graph.Packages[:limit] {
			lines = append(lines, "  "+rDim.Render("  "+pkg))
		}
		if len(graph.Packages) > limit {
			lines = append(lines, "  "+rDim.Render(fmt.Sprintf("  ... +%d more", len(graph.Packages)-limit)))
		}
	}
	return lines
}

// ── Helpers ──────────────────────────────────────────────────────────

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
