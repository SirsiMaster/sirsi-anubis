// Package output — tui_render_status.go
//
// Command result renderers: scan, ghost, clean, hardware, risk.
package output

import (
	"fmt"
	"strings"
	"syscall"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/osiris"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
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

	// Category breakdown — checkmark + name + right-aligned size
	if len(res.ByCategory) > 0 {
		for cat, summary := range res.ByCategory {
			name := string(cat)
			// Pad name to 40 chars, right-align size at col 50
			padLen := 40 - len(name)
			if padLen < 2 {
				padLen = 2
			}
			lines = append(lines, fmt.Sprintf("  %s %s%s%s",
				rGreen.Render("✓"),
				rBody.Render(name),
				strings.Repeat(" ", padLen),
				rGold.Render(fmt.Sprintf("%8s", jackal.FormatSize(summary.TotalSize)))))
		}
		lines = append(lines, "")
	}

	// Top findings (max 10)
	limit := min(len(res.Findings), 10)
	if limit > 0 {
		lines = append(lines, "  "+rLabel.Render("FINDINGS"))
		for _, f := range res.Findings[:limit] {
			severity := rGreen.Render("safe   ")
			switch f.Severity {
			case jackal.SeverityCaution:
				severity = rWarn.Render("caution")
			case jackal.SeverityWarning:
				severity = rRed.Render("warning")
			}
			desc := f.Description
			if len(desc) > 36 {
				desc = desc[:33] + "..."
			}
			lines = append(lines, fmt.Sprintf("  %s  %-36s  %8s",
				severity,
				rBody.Render(desc),
				jackal.FormatSize(f.SizeBytes)))
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

	// Ghost list — name + right-aligned size
	for _, g := range ghosts {
		name := g.AppName
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		padLen := 40 - len(name)
		if padLen < 2 {
			padLen = 2
		}
		lines = append(lines, fmt.Sprintf("  %s %s%s%s  %s",
			rWarn.Render("◆"),
			rBody.Render(name),
			strings.Repeat(" ", padLen),
			rGold.Render(fmt.Sprintf("%8s", jackal.FormatSize(g.TotalSize))),
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
		riskStyle = rGreen
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
