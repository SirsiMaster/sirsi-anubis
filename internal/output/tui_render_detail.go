// Package output — tui_render_detail.go
//
// Detailed report renderers: network, doctor, maat, diagram, knowledge,
// mirror, ra status, symbol graph.
package output

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/maat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/mirror"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ra"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seshat"
)

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
