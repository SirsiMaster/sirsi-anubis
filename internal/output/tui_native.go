package output

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal/rules"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/maat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/mirror"
	"github.com/SirsiMaster/sirsi-pantheon/internal/osiris"
	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ra"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seshat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/thoth"
)

// ── Native Deity Functions ───────────────────────────────────────────

func nativeScan() ([]string, string, []string, error) {
	engine := jackal.DefaultEngine()
	engine.RegisterAll(rules.AllRules()...)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// Grab the progress channel if set — streams per-rule updates to TUI
	pendingSelectMu.Lock()
	ch := scanProgressCh
	scanProgressCh = nil
	pendingSelectMu.Unlock()

	opts := jackal.ScanOptions{}
	if ch != nil {
		opts.OnProgress = func(ruleName string, found int, size int64, done, total int) {
			line := fmt.Sprintf("  %s %-30s",
				lipgloss.NewStyle().Foreground(Green).Render("✓"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(ruleName))
			if found > 0 {
				line += fmt.Sprintf("  %8s", jackal.FormatSize(size))
			}
			line += lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).
				Render(fmt.Sprintf("  [%d/%d]", done, total))
			ch <- line
		}
	}

	res, err := engine.Scan(ctx, opts)
	if ch != nil {
		ch <- ""
	}
	if err != nil {
		return nil, "anubis", nil, err
	}
	jackal.EnrichAdvisory(res)
	_ = jackal.Persist(res, 0)

	var fixCmds []string
	safeCount := 0
	for _, f := range res.Findings {
		if f.Severity == jackal.SeveritySafe && f.CanFix {
			safeCount++
		}
	}
	if safeCount > 0 {
		fixCmds = append(fixCmds, "anubis clean --dry-run")
	}
	return RenderScanResult(res), "anubis", fixCmds, nil
}

func nativeGhosts() ([]string, string, []string, error) {
	scanner := ka.NewScanner()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ghosts, err := scanner.Scan(ctx, false)
	if err != nil {
		return nil, "anubis", nil, err
	}
	if len(ghosts) == 0 {
		return RenderGhostResult(ghosts), "anubis", nil, nil
	}

	var items []selectItem
	for _, g := range ghosts {
		detail := fmt.Sprintf("%d residual files", g.TotalFiles)
		if len(g.Residuals) > 0 {
			detail += " · " + ShortenPath(g.Residuals[0].Path)
		}
		items = append(items, selectItem{
			Label: g.AppName, Detail: detail,
			Size: g.TotalSize, Selected: true, Data: g,
		})
	}

	pendingSelectMu.Lock()
	pendingSelectReq = &selectRequest{
		title: "𓃣 Ghosts — Select Hauntings to Exorcise",
		items: items,
		onConfirm: func(selected []selectItem) ([]string, string, []string, error) {
			s := ka.NewScanner()
			var totalFreed int64
			var totalCleaned int
			var names []string
			var cleanErrors []error
			for _, item := range selected {
				if g, ok := item.Data.(ka.Ghost); ok {
					freed, cleaned, err := s.Clean(g, false, true)
					if err != nil {
						cleanErrors = append(cleanErrors, fmt.Errorf("%s: %w", g.AppName, err))
						continue
					}
					totalFreed += freed
					totalCleaned += cleaned
					names = append(names, g.AppName)
				}
			}
			var lines []string
			lines = append(lines, "")
			if totalCleaned > 0 {
				bannerText := fmt.Sprintf("𓃣 Exorcised: %s freed", jackal.FormatSize(totalFreed))
				lines = append(lines, "  "+ResultBanner(bannerText, rGold, 50))
				lines = append(lines, "  "+rDim.Render(fmt.Sprintf("%d files from %d apps", totalCleaned, len(names))))
				lines = append(lines, "")
				for _, name := range names {
					lines = append(lines, "  "+rGreen.Render("✓")+"  "+rBody.Render(name))
				}
			} else {
				lines = append(lines, "  "+rDim.Render("No ghosts were cleaned."))
			}
			if len(cleanErrors) > 0 {
				lines = append(lines, "")
				for _, e := range cleanErrors {
					lines = append(lines, "  "+lipgloss.NewStyle().Foreground(Red).Render("✗ "+e.Error()))
				}
				return lines, "anubis", []string{"anubis scan"}, fmt.Errorf("%d ghost(s) failed to clean", len(cleanErrors))
			}
			return lines, "anubis", []string{"anubis scan"}, nil
		},
	}
	pendingSelectMu.Unlock()
	return nil, "anubis", nil, nil
}

func nativeHardware() ([]string, string, []string, error) {
	hw, err := seba.DetectHardware()
	if err != nil {
		return nil, "seba", nil, err
	}
	return RenderHardwareProfile(hw), "seba", nil, nil
}

func nativeFindings() ([]string, string, []string, error) {
	scan, err := jackal.LoadLatest()
	if err != nil {
		return []string{"", "  No scan results found. Press esc and run Scan first."}, "anubis", nil, nil
	}
	res := &jackal.ScanResult{
		Findings:   make([]jackal.Finding, len(scan.Findings)),
		TotalSize:  scan.TotalSize,
		RulesRan:   scan.RulesRan,
		ByCategory: make(map[jackal.Category]jackal.CategorySummary),
	}
	for i, f := range scan.Findings {
		res.Findings[i] = jackal.Finding{
			RuleName:    f.RuleName,
			Description: f.Description,
			Path:        f.Path,
			SizeBytes:   f.SizeBytes,
			Severity:    f.Severity,
			Category:    f.Category,
			Advisory:    f.Advisory,
			CanFix:      f.CanFix,
			Remediation: f.Remediation,
		}
	}
	for cat, s := range scan.ByCategory {
		res.ByCategory[cat] = s
	}
	return RenderScanResult(res), "anubis", nil, nil
}

func nativeNetworkAudit() ([]string, string, []string, error) {
	report, err := guard.NetworkAudit()
	if err != nil {
		return nil, "isis", nil, err
	}
	lines, fixCmds := RenderNetworkAudit(report)
	return lines, "isis", fixCmds, nil
}

func nativeNetworkFix() ([]string, string, []string, error) {
	report, err := guard.NetworkAuditFix()
	if err != nil {
		return nil, "isis", nil, err
	}
	lines, _ := RenderNetworkAudit(report)
	return lines, "isis", nil, nil
}

func nativeDoctor() ([]string, string, []string, error) {
	pendingSelectMu.Lock()
	ch := doctorProgressCh
	doctorProgressCh = nil
	pendingSelectMu.Unlock()

	opts := guard.DoctorOpts{}
	if ch != nil {
		opts.OnCheck = func(name string, sev guard.DiagnosticSeverity, msg string, done, total int) {
			icon := lipgloss.NewStyle().Foreground(Green).Render("✓")
			switch sev {
			case guard.SeverityWarn:
				icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Render("!")
			case guard.SeverityCritical:
				icon = lipgloss.NewStyle().Foreground(Red).Render("✗")
			}
			line := fmt.Sprintf("  %s %-20s %s",
				icon,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(name),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(fmt.Sprintf("[%d/%d]", done, total)))
			ch <- line
		}
	}

	report, err := guard.DoctorWithOpts(platform.Current(), opts)
	if ch != nil {
		ch <- ""
	}
	if err != nil {
		return nil, "isis", nil, err
	}
	lines, fixCmds := RenderDoctorReport(report)
	return lines, "isis", fixCmds, nil
}

func nativeCleanDryRun() ([]string, string, []string, error) {
	scan, err := jackal.LoadLatest()
	if err != nil {
		return []string{"", "  No scan results. Run Scan first."}, "anubis", nil, nil
	}

	// Filter to safe findings only
	var safeFindings []jackal.Finding
	for _, f := range scan.Findings {
		if f.Severity == jackal.SeveritySafe && f.CanFix {
			safeFindings = append(safeFindings, jackal.Finding{
				RuleName:    f.RuleName,
				Description: f.Description,
				Path:        f.Path,
				SizeBytes:   f.SizeBytes,
				Severity:    f.Severity,
				Category:    f.Category,
				CanFix:      f.CanFix,
			})
		}
	}

	if len(safeFindings) == 0 {
		return []string{"", "  No safe items to clean."}, "anubis", nil, nil
	}

	var items []selectItem
	for _, f := range safeFindings {
		items = append(items, selectItem{
			Label: f.Description, Detail: ShortenPath(f.Path),
			Size: f.SizeBytes, Selected: true, Data: f,
		})
	}

	pendingSelectMu.Lock()
	pendingSelectReq = &selectRequest{
		title: "𓃣 Clean — Select Items to Purge",
		items: items,
		onConfirm: func(selected []selectItem) ([]string, string, []string, error) {
			var findings []jackal.Finding
			for _, item := range selected {
				if f, ok := item.Data.(jackal.Finding); ok {
					findings = append(findings, f)
				}
			}
			if len(findings) == 0 {
				return []string{"", "  Nothing selected to clean."}, "anubis", nil, nil
			}
			engine := jackal.DefaultEngine()
			engine.RegisterAll(rules.AllRules()...)
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			result, err := engine.Clean(ctx, findings, jackal.CleanOptions{
				Confirm: true, UseTrash: true,
			})
			if err != nil {
				return nil, "anubis", nil, err
			}
			return RenderCleanResult(result), "anubis", []string{"anubis scan"}, nil
		},
	}
	pendingSelectMu.Unlock()
	return nil, "anubis", nil, nil
}

func nativeCleanConfirm() ([]string, string, []string, error) {
	scan, err := jackal.LoadLatest()
	if err != nil {
		return []string{"", "  No scan results. Run Scan first."}, "anubis", nil, nil
	}

	var safeFindings []jackal.Finding
	for _, f := range scan.Findings {
		if f.Severity == jackal.SeveritySafe && f.CanFix {
			safeFindings = append(safeFindings, jackal.Finding{
				RuleName:    f.RuleName,
				Description: f.Description,
				Path:        f.Path,
				SizeBytes:   f.SizeBytes,
				Severity:    f.Severity,
				Category:    f.Category,
				IsDir:       f.IsDir,
				CanFix:      f.CanFix,
			})
		}
	}

	if len(safeFindings) == 0 {
		return []string{"", "  Nothing to clean."}, "anubis", nil, nil
	}

	engine := jackal.DefaultEngine()
	engine.RegisterAll(rules.AllRules()...)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := engine.Clean(ctx, safeFindings, jackal.CleanOptions{
		Confirm:  true,
		UseTrash: true,
	})
	if err != nil {
		return nil, "anubis", nil, err
	}

	lines := RenderCleanResult(result)
	return lines, "anubis", []string{"anubis scan"}, nil
}

func nativeRisk() ([]string, string, []string, error) {
	cp, err := osiris.Assess(".")
	if err != nil {
		return nil, "osiris", nil, err
	}
	return RenderRiskAssessment(cp), "osiris", nil, nil
}

func nativeMirror() ([]string, string, []string, error) {
	home, _ := os.UserHomeDir()
	res, err := mirror.Scan(mirror.ScanOptions{
		Paths:   []string{filepath.Join(home, "Development"), filepath.Join(home, "Documents")},
		MinSize: 1024 * 100, // 100KB minimum
	})
	if err != nil {
		return nil, "anubis", nil, err
	}
	return RenderMirrorResult(res), "anubis", nil, nil
}

func nativePurge() ([]string, string, []string, error) {
	roots := jackal.DefaultPurgeRoots()
	if len(roots) == 0 {
		return []string{"", "  No project directories found (~/Development, ~/Projects, ~/Documents)."}, "anubis", nil, nil
	}

	res, err := jackal.ScanArtifacts(roots)
	if err != nil {
		return nil, "anubis", nil, err
	}
	if len(res.Artifacts) == 0 {
		return []string{
			"",
			"  " + lipgloss.NewStyle().Foreground(Green).Render("✓") + "  " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render("No build artifacts found. Projects are clean."),
		}, "anubis", nil, nil
	}

	var items []selectItem
	for _, a := range res.Artifacts {
		detail := string(a.Type)
		if a.IsRecent {
			detail += "  | Recent"
		}
		items = append(items, selectItem{
			Label:    a.ProjectName,
			Detail:   detail,
			Size:     a.Size,
			Selected: !a.IsRecent,
			Data:     a,
		})
	}

	pendingSelectMu.Lock()
	pendingSelectReq = &selectRequest{
		title: fmt.Sprintf("𓃣 Purge — Select Artifacts to Remove — %s", jackal.FormatSize(res.TotalSize)),
		items: items,
		onConfirm: func(selected []selectItem) ([]string, string, []string, error) {
			var toClean []jackal.ProjectArtifact
			for _, item := range selected {
				if a, ok := item.Data.(jackal.ProjectArtifact); ok {
					toClean = append(toClean, a)
				}
			}
			if len(toClean) == 0 {
				return []string{"", "  Nothing selected to purge."}, "anubis", nil, nil
			}
			result, err := jackal.PurgeArtifacts(toClean, true)
			if err != nil {
				return nil, "anubis", nil, err
			}
			return RenderCleanResult(result), "anubis", []string{"anubis scan"}, nil
		},
	}
	pendingSelectMu.Unlock()
	return nil, "anubis", nil, nil
}

func nativeMaatAudit() ([]string, string, []string, error) {
	report, err := maat.Weigh()
	if err != nil {
		return nil, "maat", nil, err
	}
	lines, fixCmds := RenderMaatReport(report)
	return lines, "maat", fixCmds, nil
}

func nativeDiagram() ([]string, string, []string, error) {
	res, err := seba.GenerateDiagram(".", seba.DiagramHierarchy)
	if err != nil {
		return nil, "seba", nil, err
	}
	return RenderDiagram(res), "seba", nil, nil
}

func nativeSeshatIngest() ([]string, string, []string, error) {
	reg := seshat.DefaultRegistry()
	items, err := reg.IngestAll(time.Now().Add(-24 * time.Hour))
	if err != nil {
		return nil, "seshat", nil, err
	}
	return RenderKnowledgeItems(items), "seshat", nil, nil
}

func nativeThothSync() ([]string, string, []string, error) {
	err := thoth.Sync(thoth.SyncOptions{RepoRoot: ".", UpdateDate: true})
	if err != nil {
		return nil, "thoth", nil, err
	}
	return []string{
		"",
		"  " + lipgloss.NewStyle().Foreground(Green).Render("✓") + "  " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render("Memory synced"),
		"",
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("Updated .thoth/memory.yaml with current project state"),
	}, "thoth", nil, nil
}

func nativeRaStatus() ([]string, string, []string, error) {
	home, _ := os.UserHomeDir()
	raDir := filepath.Join(home, ".config", "ra")
	status, err := ra.Monitor(raDir)
	if err != nil {
		return nil, "ra", nil, err
	}
	return RenderRaStatus(status), "ra", nil, nil
}

func nativeHorusScan() ([]string, string, []string, error) {
	p := horus.NewGoParser()
	graph, err := p.ParseDir(".")
	if err != nil {
		return nil, "horus", nil, err
	}
	return RenderSymbolGraph(graph), "horus", nil, nil
}

func nativeAnalyze() ([]string, string, []string, error) {
	home, _ := os.UserHomeDir()
	res, err := jackal.Analyze(home, 0)
	if err != nil {
		return nil, "anubis", nil, err
	}
	pendingAnalyzeMu.Lock()
	pendingAnalyzeRes = res
	pendingAnalyzeMu.Unlock()
	return nil, "anubis", nil, nil
}

func nativeInstaller() ([]string, string, []string, error) {
	res, err := jackal.ScanInstallers()
	if err != nil {
		return nil, "anubis", nil, err
	}
	if len(res.Files) == 0 {
		return []string{
			"",
			"  " + lipgloss.NewStyle().Foreground(Green).Render("✓") + "  " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render("No installer files found. Clean machine."),
		}, "anubis", nil, nil
	}

	var items []selectItem
	for _, f := range res.Files {
		items = append(items, selectItem{
			Label:    f.Name,
			Detail:   f.Source,
			Size:     f.Size,
			Selected: true,
			Data:     f,
		})
	}

	pendingSelectMu.Lock()
	pendingSelectReq = &selectRequest{
		title: fmt.Sprintf("𓃣 Installers — Select Files to Remove — %s", jackal.FormatSize(res.TotalSize)),
		items: items,
		onConfirm: func(selected []selectItem) ([]string, string, []string, error) {
			var toRemove []jackal.InstallerFile
			for _, item := range selected {
				if f, ok := item.Data.(jackal.InstallerFile); ok {
					toRemove = append(toRemove, f)
				}
			}
			if len(toRemove) == 0 {
				return []string{"", "  Nothing selected to remove."}, "anubis", nil, nil
			}
			result, err := jackal.RemoveInstallers(toRemove, true)
			if err != nil {
				return nil, "anubis", nil, err
			}
			return RenderCleanResult(result), "anubis", []string{"anubis scan"}, nil
		},
	}
	pendingSelectMu.Unlock()
	return nil, "anubis", nil, nil
}
