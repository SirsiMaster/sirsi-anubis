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

// nativeScanWithProgress is called when a progress channel is available.
func nativeScanWithProgress(ch chan string) nativeResult {
	return nativeScanImpl(ch)
}

func nativeScan() nativeResult {
	return nativeScanImpl(nil)
}

func nativeScanImpl(ch chan string) nativeResult {
	engine := jackal.DefaultEngine()
	engine.RegisterAll(rules.AllRules()...)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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
		return nativeResult{deityKey: "anubis", err: err}
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
	return nativeResult{lines: RenderScanResult(res), deityKey: "anubis", fixCmds: fixCmds}
}

func nativeGhosts() nativeResult {
	scanner := ka.NewScanner()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ghosts, err := scanner.Scan(ctx, false)
	if err != nil {
		return nativeResult{deityKey: "anubis", err: err}
	}
	if len(ghosts) == 0 {
		return nativeResult{lines: RenderGhostResult(ghosts), deityKey: "anubis"}
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

	return nativeResult{
		deityKey: "anubis",
		selectReq: &selectRequest{
			title: "𓃣 Ghosts — Select Hauntings to Exorcise",
			items: items,
			onConfirm: func(selected []selectItem) nativeResult {
				// Route through safety gateway before any deletion
				if gw := getCleanGateway(); gw != nil {
					var gatewayItems []jackal.Finding
					for _, item := range selected {
						if g, ok := item.Data.(ka.Ghost); ok {
							gatewayItems = append(gatewayItems, jackal.Finding{
								RuleName:    "ghost-exorcism",
								Description: g.AppName,
								SizeBytes:   g.TotalSize,
							})
						}
					}
					if err := gw.ConfirmClean(gatewayItems, "ghost-exorcism"); err != nil {
						return nativeResult{err: err}
					}
				}
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
					return nativeResult{lines: lines, deityKey: "anubis", fixCmds: []string{"anubis scan"}, err: fmt.Errorf("%d ghost(s) failed to clean", len(cleanErrors))}
				}
				return nativeResult{lines: lines, deityKey: "anubis", fixCmds: []string{"anubis scan"}}
			},
		},
	}
}

func nativeHardware() nativeResult {
	hw, err := seba.DetectHardware()
	if err != nil {
		return nativeResult{deityKey: "seba", err: err}
	}
	return nativeResult{lines: RenderHardwareProfile(hw), deityKey: "seba"}
}

func nativeFindings() nativeResult {
	scan, err := jackal.LoadLatest()
	if err != nil {
		return nativeResult{lines: []string{"", "  No scan results found. Press esc and run Scan first."}, deityKey: "anubis"}
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
	return nativeResult{lines: RenderScanResult(res), deityKey: "anubis"}
}

func nativeNetworkAudit() nativeResult {
	report, err := guard.NetworkAudit()
	if err != nil {
		return nativeResult{deityKey: "isis", err: err}
	}
	lines, fixCmds := RenderNetworkAudit(report)
	return nativeResult{lines: lines, deityKey: "isis", fixCmds: fixCmds}
}

func nativeNetworkFix() nativeResult {
	report, err := guard.NetworkAuditFix()
	if err != nil {
		return nativeResult{deityKey: "isis", err: err}
	}
	lines, _ := RenderNetworkAudit(report)
	return nativeResult{lines: lines, deityKey: "isis"}
}

func nativeDoctorWithProgress(ch chan string) nativeResult {
	return nativeDoctorImpl(ch)
}

func nativeDoctor() nativeResult {
	return nativeDoctorImpl(nil)
}

func nativeDoctorImpl(ch chan string) nativeResult {
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
		return nativeResult{deityKey: "isis", err: err}
	}
	lines, fixCmds := RenderDoctorReport(report)
	return nativeResult{lines: lines, deityKey: "isis", fixCmds: fixCmds}
}

func nativeCleanDryRun() nativeResult {
	scan, err := jackal.LoadLatest()
	if err != nil {
		return nativeResult{lines: []string{"", "  No scan results. Run Scan first."}, deityKey: "anubis"}
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
		return nativeResult{lines: []string{"", "  No safe items to clean."}, deityKey: "anubis"}
	}

	var items []selectItem
	for _, f := range safeFindings {
		items = append(items, selectItem{
			Label: f.Description, Detail: ShortenPath(f.Path),
			Size: f.SizeBytes, Selected: true, Data: f,
		})
	}

	return nativeResult{
		deityKey: "anubis",
		selectReq: &selectRequest{
			title: "𓃣 Clean — Select Items to Purge",
			items: items,
			onConfirm: func(selected []selectItem) nativeResult {
				var findings []jackal.Finding
				for _, item := range selected {
					if f, ok := item.Data.(jackal.Finding); ok {
						findings = append(findings, f)
					}
				}
				if len(findings) == 0 {
					return nativeResult{lines: []string{"", "  Nothing selected to clean."}, deityKey: "anubis"}
				}
				if gw := getCleanGateway(); gw != nil {
					if err := gw.ConfirmClean(findings, "clean-select"); err != nil {
						return nativeResult{err: err}
					}
				}
				engine := jackal.DefaultEngine()
				engine.RegisterAll(rules.AllRules()...)
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				result, err := engine.Clean(ctx, findings, jackal.CleanOptions{
					Confirm: true, UseTrash: true,
				})
				if err != nil {
					return nativeResult{deityKey: "anubis", err: err}
				}
				return nativeResult{lines: RenderCleanResult(result), deityKey: "anubis", fixCmds: []string{"anubis scan"}}
			},
		},
	}
}

func nativeCleanConfirm() nativeResult {
	scan, err := jackal.LoadLatest()
	if err != nil {
		return nativeResult{lines: []string{"", "  No scan results. Run Scan first."}, deityKey: "anubis"}
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
		return nativeResult{lines: []string{"", "  Nothing to clean."}, deityKey: "anubis"}
	}

	if gw := getCleanGateway(); gw != nil {
		if err := gw.ConfirmClean(safeFindings, "clean-confirm"); err != nil {
			return nativeResult{deityKey: "anubis", err: err}
		}
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
		return nativeResult{deityKey: "anubis", err: err}
	}

	return nativeResult{lines: RenderCleanResult(result), deityKey: "anubis", fixCmds: []string{"anubis scan"}}
}

func nativeRisk() nativeResult {
	cp, err := osiris.Assess(".")
	if err != nil {
		return nativeResult{deityKey: "osiris", err: err}
	}
	return nativeResult{lines: RenderRiskAssessment(cp), deityKey: "osiris"}
}

func nativeMirror() nativeResult {
	home, _ := os.UserHomeDir()
	res, err := mirror.Scan(mirror.ScanOptions{
		Paths:   []string{filepath.Join(home, "Development"), filepath.Join(home, "Documents")},
		MinSize: 1024 * 100, // 100KB minimum
	})
	if err != nil {
		return nativeResult{deityKey: "anubis", err: err}
	}
	return nativeResult{lines: RenderMirrorResult(res), deityKey: "anubis"}
}

func nativePurge() nativeResult {
	roots := jackal.DefaultPurgeRoots()
	if len(roots) == 0 {
		return nativeResult{lines: []string{"", "  No project directories found (~/Development, ~/Projects, ~/Documents)."}, deityKey: "anubis"}
	}

	res, err := jackal.ScanArtifacts(roots)
	if err != nil {
		return nativeResult{deityKey: "anubis", err: err}
	}
	if len(res.Artifacts) == 0 {
		return nativeResult{lines: []string{
			"",
			"  " + lipgloss.NewStyle().Foreground(Green).Render("✓") + "  " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render("No build artifacts found. Projects are clean."),
		}, deityKey: "anubis"}
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

	return nativeResult{
		deityKey: "anubis",
		selectReq: &selectRequest{
			title: fmt.Sprintf("𓃣 Purge — Select Artifacts to Remove — %s", jackal.FormatSize(res.TotalSize)),
			items: items,
			onConfirm: func(selected []selectItem) nativeResult {
				var toClean []jackal.ProjectArtifact
				for _, item := range selected {
					if a, ok := item.Data.(jackal.ProjectArtifact); ok {
						toClean = append(toClean, a)
					}
				}
				if len(toClean) == 0 {
					return nativeResult{lines: []string{"", "  Nothing selected to purge."}, deityKey: "anubis"}
				}
				// Route through safety gateway before any deletion
				if gw := getCleanGateway(); gw != nil {
					var gatewayItems []jackal.Finding
					for _, a := range toClean {
						gatewayItems = append(gatewayItems, jackal.Finding{
							RuleName:    "purge-artifact",
							Description: a.ProjectName,
							Path:        a.ArtifactDir,
							SizeBytes:   a.Size,
						})
					}
					if err := gw.ConfirmClean(gatewayItems, "purge"); err != nil {
						return nativeResult{err: err}
					}
				}
				result, err := jackal.PurgeArtifacts(toClean, true)
				if err != nil {
					return nativeResult{deityKey: "anubis", err: err}
				}
				return nativeResult{lines: RenderCleanResult(result), deityKey: "anubis", fixCmds: []string{"anubis scan"}}
			},
		},
	}
}

func nativeMaatAudit() nativeResult {
	report, err := maat.Weigh()
	if err != nil {
		return nativeResult{deityKey: "maat", err: err}
	}
	lines, fixCmds := RenderMaatReport(report)
	return nativeResult{lines: lines, deityKey: "maat", fixCmds: fixCmds}
}

func nativeDiagram() nativeResult {
	res, err := seba.GenerateDiagram(".", seba.DiagramHierarchy)
	if err != nil {
		return nativeResult{deityKey: "seba", err: err}
	}
	return nativeResult{lines: RenderDiagram(res), deityKey: "seba"}
}

func nativeSeshatIngest() nativeResult {
	reg := seshat.DefaultRegistry()
	items, err := reg.IngestAll(time.Now().Add(-24 * time.Hour))
	if err != nil {
		return nativeResult{deityKey: "seshat", err: err}
	}
	return nativeResult{lines: RenderKnowledgeItems(items), deityKey: "seshat"}
}

func nativeThothSync() nativeResult {
	err := thoth.Sync(thoth.SyncOptions{RepoRoot: ".", UpdateDate: true})
	if err != nil {
		return nativeResult{deityKey: "thoth", err: err}
	}
	return nativeResult{lines: []string{
		"",
		"  " + lipgloss.NewStyle().Foreground(Green).Render("✓") + "  " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render("Memory synced"),
		"",
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("Updated .thoth/memory.yaml with current project state"),
	}, deityKey: "thoth"}
}

func nativeRaStatus() nativeResult {
	home, _ := os.UserHomeDir()
	raDir := filepath.Join(home, ".config", "ra")
	status, err := ra.Monitor(raDir)
	if err != nil {
		return nativeResult{deityKey: "ra", err: err}
	}
	return nativeResult{lines: RenderRaStatus(status), deityKey: "ra"}
}

func nativeHorusScan() nativeResult {
	p := horus.NewGoParser()
	graph, err := p.ParseDir(".")
	if err != nil {
		return nativeResult{deityKey: "horus", err: err}
	}
	return nativeResult{lines: RenderSymbolGraph(graph), deityKey: "horus"}
}

func nativeAnalyze() nativeResult {
	home, _ := os.UserHomeDir()
	res, err := jackal.Analyze(home, 0)
	if err != nil {
		return nativeResult{deityKey: "anubis", err: err}
	}
	return nativeResult{deityKey: "anubis", analyzeRes: res}
}

func nativeInstaller() nativeResult {
	res, err := jackal.ScanInstallers()
	if err != nil {
		return nativeResult{deityKey: "anubis", err: err}
	}
	if len(res.Files) == 0 {
		return nativeResult{lines: []string{
			"",
			"  " + lipgloss.NewStyle().Foreground(Green).Render("✓") + "  " +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render("No installer files found. Clean machine."),
		}, deityKey: "anubis"}
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

	return nativeResult{
		deityKey: "anubis",
		selectReq: &selectRequest{
			title: fmt.Sprintf("𓃣 Installers — Select Files to Remove — %s", jackal.FormatSize(res.TotalSize)),
			items: items,
			onConfirm: func(selected []selectItem) nativeResult {
				var toRemove []jackal.InstallerFile
				for _, item := range selected {
					if f, ok := item.Data.(jackal.InstallerFile); ok {
						toRemove = append(toRemove, f)
					}
				}
				if len(toRemove) == 0 {
					return nativeResult{lines: []string{"", "  Nothing selected to remove."}, deityKey: "anubis"}
				}
				// Route through safety gateway before any deletion
				if gw := getCleanGateway(); gw != nil {
					var gatewayItems []jackal.Finding
					for _, f := range toRemove {
						gatewayItems = append(gatewayItems, jackal.Finding{
							RuleName:    "installer-removal",
							Description: f.Name,
							Path:        f.Path,
							SizeBytes:   f.Size,
						})
					}
					if err := gw.ConfirmClean(gatewayItems, "installer"); err != nil {
						return nativeResult{err: err}
					}
				}
				result, err := jackal.RemoveInstallers(toRemove, true)
				if err != nil {
					return nativeResult{deityKey: "anubis", err: err}
				}
				return nativeResult{lines: RenderCleanResult(result), deityKey: "anubis", fixCmds: []string{"anubis scan"}}
			},
		},
	}
}
