package output

import (
	"context"
	"fmt"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal/rules"
)

func nativeCleanDryRun() nativeResult {
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
