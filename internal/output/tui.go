// Package output — Pantheon TUI
//
// Tab-based interface inspired by Mole (mole.fit). Each deity group gets
// its own page, its own job. No split panes, no REPL. Guided navigation
// with numbered actions. Press a number to act. Press esc to go back.
package output

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal/rules"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/maat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/mirror"
	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
	"github.com/SirsiMaster/sirsi-pantheon/internal/osiris"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ra"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seba"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seshat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
	"github.com/SirsiMaster/sirsi-pantheon/internal/thoth"
	"github.com/SirsiMaster/sirsi-pantheon/internal/vitals"
)

// ── Tab Definitions ──────────────────────────────────────────────────
// Five tabs, like Mole's five planets. Each has a purpose, a voice,
// and numbered actions. The user never types a command — they press
// a number.

// nativeResult is returned by native deity calls.
type nativeResult struct {
	lines     []string       // rendered output lines
	deityKey  string         // which deity ran
	fixCmds   []string       // actionable fix commands (override suggest engine)
	err       error
	selectReq *selectRequest // if non-nil, enter viewSelect mode instead of viewDone
}

type nativeResultMsg nativeResult

// tabAction defines one action on a tab. Native functions return
// (rendered lines, deityKey, fixCmds, error). fixCmds override
// the suggest engine — these become the numbered "What's Next" items.
type tabAction struct {
	Label  string
	Desc   string
	Args   []string                                   // CLI args (fallback if Native is nil)
	Native func() ([]string, string, []string, error) // (lines, deityKey, fixCmds, err)
}

type tabDef struct {
	Name    string
	Glyph   string
	Tagline string
	Actions []tabAction
}

var tabs = []tabDef{
	{
		Name:    "Scan",
		Glyph:   "𓃣",
		Tagline: "Find waste. Free space. Stay clean.",
		Actions: []tabAction{
			{"Scan", "Find infrastructure waste", []string{"anubis", "scan"}, nativeScan},
			{"Ghosts", "Find remnants of uninstalled apps", []string{"anubis", "ghosts"}, nativeGhosts},
			{"Clean", "Preview and remove safe items", []string{"anubis", "clean", "--dry-run"}, nativeCleanDryRun},
			{"Duplicates", "Find duplicate files", []string{"anubis", "duplicates"}, nativeMirror},
			{"Purge", "Remove project build artifacts", []string{"anubis", "purge"}, nativePurge},
			{"Analyze", "Visual disk space explorer", []string{"anubis", "analyze"}, nativeAnalyze},
			{"Installer", "Find and remove installer files", []string{"anubis", "installer"}, nativeInstaller},
		},
	},
	{
		Name:    "Health",
		Glyph:   "𓁐",
		Tagline: "Diagnose. Fix. Monitor.",
		Actions: []tabAction{
			{"Diagnose", "Full system health check", []string{"isis", "diagnose"}, nativeDoctor},
			{"Network", "Network security posture audit", []string{"isis", "network"}, nativeNetworkAudit},
			{"Fix", "Auto-fix DNS, firewall, and security", []string{"isis", "fix"}, nativeNetworkFix},
			{"Monitor", "Watch processes and RAM pressure", []string{"isis", "monitor"}, nil}, // long-running
		},
	},
	{
		Name:    "Quality",
		Glyph:   "𓆄",
		Tagline: "Audit. Assess. Enforce.",
		Actions: []tabAction{
			{"Audit", "Code quality and governance scan", []string{"maat", "audit"}, nativeMaatAudit},
			{"Risk", "Uncommitted work risk assessment", []string{"osiris", "risk"}, nativeRisk},
			{"Lint", "Run linters across the codebase", []string{"ra", "lint"}, nil},
			{"Test", "Run test suites fleet-wide", []string{"ra", "test"}, nil},
		},
	},
	{
		Name:    "Intel",
		Glyph:   "𓇽",
		Tagline: "Profile. Map. Learn.",
		Actions: []tabAction{
			{"Hardware", "Accelerator and architecture profile", []string{"seba", "hardware"}, nativeHardware},
			{"Diagram", "Generate architecture diagrams", []string{"seba", "diagram"}, nativeDiagram},
			{"Learn", "Ingest knowledge from sources", []string{"seshat", "learn"}, nativeSeshatIngest},
			{"Memory", "Sync project memory state", []string{"thoth", "sync"}, nativeThothSync},
		},
	},
	{
		Name:    "Status",
		Glyph:   "𓂀",
		Tagline: "Live vitals. Fleet health. Code graph.",
		Actions: []tabAction{
			{"Refresh", "Refresh system vitals", []string{"isis", "diagnose"}, nativeDoctor},
			{"Fleet", "Fleet orchestrator status", []string{"ra", "fleet"}, nativeRaStatus},
			{"Index", "Build code symbol index", []string{"horus", "index"}, nativeHorusScan},
		},
	},
}

// nativeCommands maps suggest command strings to native functions.
// When a post-run suggestion matches one of these, it runs natively
// instead of shelling out to a subprocess.
var nativeCommands = map[string]func() ([]string, string, []string, error){
	// Anubis — scan & clean
	"anubis scan":            nativeScan,
	"anubis ghosts":          nativeGhosts,
	"anubis clean --dry-run": nativeCleanDryRun,
	"anubis clean --confirm": nativeCleanConfirm,
	"anubis duplicates":      nativeMirror,
	// Isis — health
	"isis diagnose":          nativeDoctor,
	"isis network":           nativeNetworkAudit,
	"isis fix":               nativeNetworkFix,
	// Quality & Intel
	"seba hardware":          nativeHardware,
	"osiris risk":            nativeRisk,
	// Aliases for convenience
	"scan":                   nativeScan,
	"findings":               nativeFindings,
}

// scanProgressCh streams per-rule progress to the TUI during scans.
var scanProgressCh chan string

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
			for _, item := range selected {
				if g, ok := item.Data.(ka.Ghost); ok {
					freed, cleaned, err := s.Clean(g, false, true)
					if err != nil {
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

// doctorProgressCh streams per-check progress to the TUI.
var doctorProgressCh chan string

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

var (
	pendingAnalyzeMu  sync.Mutex
	pendingAnalyzeRes *jackal.AnalyzeResult
)

type analyzeResultMsg struct {
	result *jackal.AnalyzeResult
	err    error
}

// ── View Mode ────────────────────────────────────────────────────────

type viewMode int

const (
	viewTabs    viewMode = iota // Showing a tab landing page
	viewRunning                 // Command executing
	viewDone                    // Command finished, output + next actions
	viewPrompt                  // Power-user command prompt (: key)
	viewSelect                  // Interactive checkbox selection
	viewAnalyze                 // Disk space analyzer with drill-down
)

// ── Selection Types ──────────────────────────────────────────────────

type selectItem struct {
	Label    string
	Detail   string      // secondary line (path, size, etc.)
	Size     int64       // for size display
	Selected bool
	Data     interface{} // opaque payload for the confirm handler
}

type selectRequest struct {
	title     string
	items     []selectItem
	onConfirm func(selected []selectItem) ([]string, string, []string, error)
}

var (
	pendingSelectMu  sync.Mutex
	pendingSelectReq *selectRequest
)

// ── Model ────────────────────────────────────────────────────────────

type TUIModel struct {
	width  int
	height int

	activeTab int      // 0-4 index into tabs
	mode      viewMode // current view state

	// Checkbox selection state
	selectItems     []selectItem
	selectCursor    int
	selectTitle     string
	selectOnConfirm func(selected []selectItem) ([]string, string, []string, error)

	// Command execution
	input        textinput.Model
	viewport     viewport.Model
	outputLines  []string
	runningDeity string
	runningCmd   string
	runningArgs  []string
	cmdStartTime time.Time
	spinner      spinner.Model
	streamCh     chan string
	runningProc  *atomic.Pointer[os.Process]

	// Post-run suggestions
	postRunCmds    []string
	postRunActions []suggest.Action // full actions with descriptions
	tabIdx         int
	lastDeity      string // deity key preserved for done view

	// History
	history    []historyEntry
	cmdHistory []string
	historyIdx int

	// State
	activeDeity map[string]bool
	deityState  map[string]deityRunState
	steleReader *stele.Reader
	quitting    bool

	// Notifications
	notifyStore       *notify.Store
	recentNotify      []notify.Notification
	notifyRefreshTime time.Time

	// System vitals
	vitals systemVitals

	// Live dashboard history (ring buffers for sparklines)
	cpuHistory  []float64 // last 60 samples
	memHistory  []float64 // last 60 samples
	netDownHist []float64 // last 60 samples
	netUpHist   []float64 // last 60 samples

	// Disk analyzer state
	analyzePath    string
	analyzeEntries []jackal.DirEntry
	analyzeCursor  int
	analyzeTotal   int64
	analyzeHistory []analyzeLevel
}

type analyzeLevel struct {
	path    string
	entries []jackal.DirEntry
	total   int64
	cursor  int
}

type systemVitals = vitals.Snapshot

type deityRunState = deity.RunState

const (
	stateNeverRun  = deity.StateNeverRun
	stateSucceeded = deity.StateSucceeded
	stateFailed    = deity.StateFailed
	stateHasData   = deity.StateHasData
)

type historyEntry struct {
	deity, command, output string
}

// ── Messages ─────────────────────────────────────────────────────────

type refreshMsg time.Time
type elapsedTickMsg time.Time
type liveTickMsg time.Time

type streamLineMsg struct {
	line string
	done bool
	err  error
}

func refreshTick() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

func liveTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return liveTickMsg(t) })
}

func elapsedTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return elapsedTickMsg(t) })
}

func waitForStreamLine(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return streamLineMsg{done: true}
		}
		return streamLineMsg{line: line}
	}
}

// ── Constructor ──────────────────────────────────────────────────────

func NewTUIModel() TUIModel {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.CharLimit = 256
	ti.Prompt = "𓉴 "
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Gold).Bold(true)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(White)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	ti.SetStyles(styles)

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(Gold)

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(10))

	m := TUIModel{
		input:       ti,
		viewport:    vp,
		spinner:     sp,
		width:       100,
		height:      40,
		mode:        viewTabs,
		activeTab:   0,
		historyIdx:  -1,
		tabIdx:      -1,
		activeDeity: make(map[string]bool),
		deityState:  make(map[string]deityRunState),
		streamCh:    make(chan string, 100),
		runningProc: &atomic.Pointer[os.Process]{},
		steleReader: stele.NewReader("tui"),
	}
	m.refreshActive()
	return m
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(refreshTick())
}

// ── Update ───────────────────────────────────────────────────────────

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcViewport()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case nativeResultMsg:
		return m.handleNativeResult(msg)

	case analyzeResultMsg:
		if msg.err != nil {
			if len(m.analyzeHistory) > 0 {
				last := m.analyzeHistory[len(m.analyzeHistory)-1]
				m.analyzeHistory = m.analyzeHistory[:len(m.analyzeHistory)-1]
				m.analyzePath = last.path
				m.analyzeEntries = last.entries
				m.analyzeTotal = last.total
				m.analyzeCursor = last.cursor
			}
			m.mode = viewAnalyze
			return m, nil
		}
		m.mode = viewAnalyze
		m.analyzePath = msg.result.Path
		m.analyzeEntries = msg.result.Entries
		m.analyzeTotal = msg.result.TotalSize
		m.analyzeCursor = 0
		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil

	case streamLineMsg:
		return m.handleStreamLine(msg)

	case liveTickMsg:
		if m.mode == viewTabs && m.activeTab == 4 {
			m.refreshVitals()
			m.appendHistory()
			return m, liveTick()
		}
		return m, nil

	case elapsedTickMsg:
		if m.mode == viewRunning {
			return m, elapsedTick()
		}
		return m, nil

	case refreshMsg:
		m.refreshActive()
		return m, refreshTick()

	case spinner.TickMsg:
		if m.mode == viewRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Pass through to active component
	if m.mode == viewPrompt {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// ── Key Handling ─────────────────────────────────────────────────────

func (m TUIModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "ctrl+c":
		if m.mode == viewRunning {
			if proc := m.runningProc.Load(); proc != nil {
				_ = proc.Kill()
				m.runningProc.Store(nil)
				return m, nil
			}
		}
		m.quitting = true
		return m, tea.Quit

	case "q":
		if m.mode == viewTabs {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.mode {
	case viewTabs:
		return m.handleTabKey(key)
	case viewRunning:
		// Scroll output while running
		switch key {
		case "up", "pgup":
			m.viewport.PageUp()
		case "down", "pgdown":
			m.viewport.PageDown()
		}
		return m, nil
	case viewDone:
		return m.handleDoneKey(key)
	case viewPrompt:
		return m.handlePromptKey(key, msg)
	case viewSelect:
		return m.handleSelectKey(key)
	case viewAnalyze:
		return m.handleAnalyzeKey(key)
	}

	return m, nil
}

func (m TUIModel) handleTabKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		if m.activeTab > 0 {
			m.activeTab--
		}
		if m.activeTab == 4 {
			return m, liveTick()
		}
		return m, nil
	case "right", "l":
		if m.activeTab < len(tabs)-1 {
			m.activeTab++
		}
		if m.activeTab == 4 {
			return m, liveTick()
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7":
		idx := int(key[0]-'0') - 1
		tab := tabs[m.activeTab]
		if idx < len(tab.Actions) {
			return m.executeAction(tab.Actions[idx])
		}
		return m, nil
	case ":":
		// Power-user command prompt
		m.mode = viewPrompt
		m.input.Focus()
		m.input.Reset()
		return m, textinput.Blink
	case "esc":
		m.quitting = true
		return m, tea.Quit
	}

	// Tab switching by first letter
	for i, tab := range tabs {
		if strings.EqualFold(key, tab.Name[:1]) {
			m.activeTab = i
			if i == 4 {
				return m, liveTick()
			}
			return m, nil
		}
	}

	return m, nil
}

func (m TUIModel) handleDoneKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		// Back to tab
		m.mode = viewTabs
		m.outputLines = nil
		m.postRunCmds = nil
		m.recalcViewport()
		return m, nil
	case "up", "pgup":
		m.viewport.PageUp()
		return m, nil
	case "down", "pgdown":
		m.viewport.PageDown()
		return m, nil
	case "1", "2", "3":
		idx := int(key[0]-'0') - 1
		if idx < len(m.postRunCmds) {
			cmd := m.postRunCmds[idx]
			// Check for native handlers first
			if fn, ok := nativeCommands[cmd]; ok {
				action := tabAction{
					Label:  cmd,
					Args:   strings.Fields(cmd),
					Native: fn,
				}
				return m.executeAction(action)
			}
			// Fallback to subprocess
			args := strings.Fields(cmd)
			if len(args) > 0 {
				return m.executeArgs(args)
			}
		}
		return m, nil
	case ":":
		m.mode = viewPrompt
		m.input.Focus()
		m.input.Reset()
		return m, textinput.Blink
	}
	return m, nil
}

func (m TUIModel) handleSelectKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.selectCursor > 0 {
			m.selectCursor--
		}
		return m, nil
	case "down", "j":
		if m.selectCursor < len(m.selectItems)-1 {
			m.selectCursor++
		}
		return m, nil
	case " ":
		if m.selectCursor < len(m.selectItems) {
			m.selectItems[m.selectCursor].Selected = !m.selectItems[m.selectCursor].Selected
		}
		return m, nil
	case "a":
		allSelected := true
		for _, item := range m.selectItems {
			if !item.Selected {
				allSelected = false
				break
			}
		}
		for i := range m.selectItems {
			m.selectItems[i].Selected = !allSelected
		}
		return m, nil
	case "enter":
		var selected []selectItem
		for _, item := range m.selectItems {
			if item.Selected {
				selected = append(selected, item)
			}
		}
		if len(selected) == 0 {
			m.mode = viewTabs
			return m, nil
		}
		if m.selectOnConfirm != nil {
			m.mode = viewRunning
			m.cmdStartTime = time.Now()
			m.outputLines = nil
			m.postRunCmds = nil
			onConfirm := m.selectOnConfirm
			return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
				lines, deityKey, fixCmds, err := onConfirm(selected)
				return nativeResultMsg{lines: lines, deityKey: deityKey, fixCmds: fixCmds, err: err}
			})
		}
		m.mode = viewTabs
		return m, nil
	case "esc":
		m.mode = viewTabs
		m.selectItems = nil
		m.selectOnConfirm = nil
		return m, nil
	}
	return m, nil
}

func (m TUIModel) handleAnalyzeKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.analyzeCursor > 0 {
			m.analyzeCursor--
		}
		return m, nil
	case "down", "j":
		if m.analyzeCursor < len(m.analyzeEntries)-1 {
			m.analyzeCursor++
		}
		return m, nil
	case "enter", "right", "l":
		if m.analyzeCursor < len(m.analyzeEntries) {
			entry := m.analyzeEntries[m.analyzeCursor]
			if entry.IsDir {
				m.analyzeHistory = append(m.analyzeHistory, analyzeLevel{
					path: m.analyzePath, entries: m.analyzeEntries,
					total: m.analyzeTotal, cursor: m.analyzeCursor,
				})
				childPath := entry.Path
				m.mode = viewRunning
				m.runningCmd = "analyze " + ShortenPath(childPath)
				m.runningDeity = "anubis"
				m.cmdStartTime = time.Now()
				m.outputLines = nil
				m.postRunCmds = nil
				return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
					res, err := jackal.Analyze(childPath, 0)
					if err != nil {
						return analyzeResultMsg{err: err}
					}
					return analyzeResultMsg{result: res}
				})
			}
		}
		return m, nil
	case "esc", "left", "h":
		if len(m.analyzeHistory) > 0 {
			last := m.analyzeHistory[len(m.analyzeHistory)-1]
			m.analyzeHistory = m.analyzeHistory[:len(m.analyzeHistory)-1]
			m.analyzePath = last.path
			m.analyzeEntries = last.entries
			m.analyzeTotal = last.total
			m.analyzeCursor = last.cursor
		} else {
			m.mode = viewTabs
			m.analyzeEntries = nil
			m.analyzeHistory = nil
		}
		return m, nil
	case "q":
		m.mode = viewTabs
		m.analyzeEntries = nil
		m.analyzeHistory = nil
		return m, nil
	}
	return m, nil
}

func (m TUIModel) handlePromptKey(key string, msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.mode = viewTabs
		m.input.Blur()
		return m, nil
	case "enter":
		raw := strings.TrimSpace(m.input.Value())
		if raw == "" {
			m.mode = viewTabs
			m.input.Blur()
			return m, nil
		}
		m.input.Blur()
		args := strings.Fields(raw)
		return m.executeArgs(args)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// ── Command Execution ────────────────────────────────────────────────

func (m TUIModel) executeAction(action tabAction) (TUIModel, tea.Cmd) {
	if action.Native != nil {
		m.mode = viewRunning
		m.runningCmd = strings.Join(action.Args, " ")
		m.runningArgs = action.Args
		m.runningDeity = ""
		if len(action.Args) > 0 {
			m.runningDeity = action.Args[0]
		}
		m.cmdStartTime = time.Now()
		m.outputLines = nil
		m.postRunCmds = nil

		// Stream per-step progress for scan and doctor
		isScan := len(action.Args) >= 2 && action.Args[0] == "anubis" && action.Args[1] == "scan"
		isDoctor := len(action.Args) >= 2 && action.Args[0] == "isis" && action.Args[1] == "diagnose"
		if isScan || isDoctor {
			ch := make(chan string, 100)
			m.streamCh = ch
			pendingSelectMu.Lock()
			if isScan {
				scanProgressCh = ch
			} else {
				doctorProgressCh = ch
			}
			pendingSelectMu.Unlock()

			label := "Scanning..."
			if isDoctor {
				label = "Diagnosing..."
			}
			m.outputLines = []string{
				"",
				"  " + lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(label),
				"",
			}
			m.viewport.SetContent(strings.Join(m.outputLines, "\n"))

			fn := action.Native
			return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
				go func() {
					fn()
					close(ch)
				}()
				line, ok := <-ch
				if !ok {
					return streamLineMsg{done: true}
				}
				return streamLineMsg{line: line}
			})
		}

		fn := action.Native
		return m, tea.Batch(m.spinner.Tick, elapsedTick(), func() tea.Msg {
			lines, deityKey, fixCmds, err := fn()
			pendingSelectMu.Lock()
			selReq := pendingSelectReq
			pendingSelectReq = nil
			pendingSelectMu.Unlock()
			return nativeResultMsg{lines: lines, deityKey: deityKey, fixCmds: fixCmds, err: err, selectReq: selReq}
		})
	}
	return m.executeArgs(action.Args)
}

func (m TUIModel) executeArgs(args []string) (TUIModel, tea.Cmd) {
	m.mode = viewRunning
	m.runningCmd = strings.Join(args, " ")
	m.runningArgs = args
	m.cmdStartTime = time.Now()
	m.streamCh = make(chan string, 100)
	m.outputLines = nil
	m.postRunCmds = nil

	// Determine deity from first arg
	m.runningDeity = ""
	for _, d := range deity.Roster {
		if len(args) > 0 && args[0] == d.Key {
			m.runningDeity = d.Key
			break
		}
	}
	// Check CLI aliases
	aliases := map[string]string{
		"scan": "anubis", "ghosts": "anubis", "clean": "anubis",
		"duplicates": "anubis", "purge": "anubis", "analyze": "anubis",
		"diagnose": "isis", "network": "isis", "fix": "isis", "monitor": "isis",
		"audit": "maat", "risk": "osiris",
		"hardware": "seba", "diagram": "seba",
		"learn": "seshat", "fleet": "ra", "index": "horus",
	}
	if m.runningDeity == "" && len(args) > 0 {
		if d, ok := aliases[args[0]]; ok {
			m.runningDeity = d
		}
	}

	m.recalcViewport()

	exe, _ := os.Executable()
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "SIRSI_TUI=1")

	return m, tea.Batch(m.spinner.Tick, elapsedTick(), m.runCommandStreaming(cmd))
}

func (m TUIModel) runCommandStreaming(cmd *exec.Cmd) tea.Cmd {
	ch := m.streamCh
	procPtr := m.runningProc
	return func() tea.Msg {
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			close(ch)
			return streamLineMsg{done: true, err: err}
		}
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			close(ch)
			return streamLineMsg{done: true, err: err}
		}
		combined := io.MultiReader(stdoutPipe, stderrPipe)
		if err := cmd.Start(); err != nil {
			close(ch)
			return streamLineMsg{done: true, err: err}
		}
		procPtr.Store(cmd.Process)
		go func() {
			scanner := bufio.NewScanner(combined)
			for scanner.Scan() {
				ch <- scanner.Text()
			}
			_ = cmd.Wait()
			close(ch)
		}()
		line, ok := <-ch
		if !ok {
			return streamLineMsg{done: true}
		}
		return streamLineMsg{line: line}
	}
}

func (m TUIModel) handleStreamLine(msg streamLineMsg) (TUIModel, tea.Cmd) {
	if msg.done {
		m.mode = viewDone
		m.runningProc.Store(nil)
		m.lastDeity = m.runningDeity

		if m.runningDeity != "" {
			m.activeDeity[m.runningDeity] = true
		}
		m.history = append(m.history, historyEntry{
			deity: m.runningDeity, command: m.runningCmd,
			output: strings.Join(m.outputLines, "\n"),
		})
		m.cmdHistory = deduplicateHistory(m.history)

		if msg.err != nil {
			m.outputLines = append(m.outputLines, "",
				lipgloss.NewStyle().Foreground(Red).Render("  ✗ "+msg.err.Error()))
			if m.runningDeity != "" {
				m.deityState[m.runningDeity] = stateFailed
			}
		} else {
			if m.runningDeity != "" {
				state := stateSucceeded
				if m.runningDeity == "anubis" {
					// Replace streaming progress with final rendered scan result
					if scan, loadErr := jackal.LoadLatest(); loadErr == nil && len(scan.Findings) > 0 {
						state = stateHasData
						res := &jackal.ScanResult{
							Findings:   make([]jackal.Finding, len(scan.Findings)),
							TotalSize:  scan.TotalSize,
							RulesRan:   scan.RulesRan,
							ByCategory: make(map[jackal.Category]jackal.CategorySummary),
						}
						for i, f := range scan.Findings {
							res.Findings[i] = jackal.Finding{
								Description: f.Description, Path: f.Path,
								SizeBytes: f.SizeBytes, Severity: f.Severity,
								Category: f.Category, Advisory: f.Advisory,
								CanFix: f.CanFix, Remediation: f.Remediation,
							}
						}
						for cat, s := range scan.ByCategory {
							res.ByCategory[cat] = s
						}
						m.outputLines = RenderScanResult(res)
						// Offer cleanup
						safeCount := 0
						for _, f := range res.Findings {
							if f.Severity == jackal.SeveritySafe && f.CanFix {
								safeCount++
							}
						}
						if safeCount > 0 {
							m.postRunCmds = []string{"anubis clean --dry-run"}
							m.postRunActions = []suggest.Action{{
								Command:     "anubis clean --dry-run",
								Description: "Preview safe items to clean",
							}}
						}
					}
				}
				m.deityState[m.runningDeity] = state
			}
		}

		// Build post-run suggestions
		ctx := m.buildSuggestContext()
		if msg.err != nil {
			ctx.Err = msg.err
		}
		m.postRunCmds = suggest.Commands(ctx)
		m.postRunActions = suggest.After(ctx)
		if msg.err != nil {
			m.postRunActions = suggest.OnError(ctx)
		}

		m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
		m.recalcViewport() // re-fit viewport to actual content size
		m.viewport.GotoTop()
		m.savePersistedState()

		// Record notification
		if m.notifyStore != nil && m.runningDeity != "" {
			sev := notify.SeveritySuccess
			summary := fmt.Sprintf("%s completed", m.runningCmd)
			if msg.err != nil {
				sev = notify.SeverityError
				summary = fmt.Sprintf("%s failed", m.runningCmd)
			}
			_ = m.notifyStore.Record(notify.Notification{
				Source: m.runningDeity, Action: m.runningCmd,
				Severity: sev, Summary: summary,
			})
			m.notifyRefreshTime = time.Time{}
			m.refreshNotifications()
		}

		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil
	}

	m.outputLines = append(m.outputLines, "  "+msg.line)
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.viewport.GotoBottom()
	return m, waitForStreamLine(m.streamCh)
}

// handleNativeResult processes results from native deity function calls.
func (m TUIModel) handleNativeResult(msg nativeResultMsg) (TUIModel, tea.Cmd) {
	// If the result carries an analyze result, enter analyze mode.
	pendingAnalyzeMu.Lock()
	analyzeRes := pendingAnalyzeRes
	pendingAnalyzeRes = nil
	pendingAnalyzeMu.Unlock()
	if analyzeRes != nil && msg.err == nil {
		m.mode = viewAnalyze
		m.analyzePath = analyzeRes.Path
		m.analyzeEntries = analyzeRes.Entries
		m.analyzeTotal = analyzeRes.TotalSize
		m.analyzeCursor = 0
		m.analyzeHistory = nil
		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil
	}

	// If the result carries a select request, enter checkbox mode.
	if msg.selectReq != nil && msg.err == nil {
		m.mode = viewSelect
		m.selectTitle = msg.selectReq.title
		m.selectItems = msg.selectReq.items
		m.selectCursor = 0
		m.selectOnConfirm = msg.selectReq.onConfirm
		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil
	}

	m.mode = viewDone
	m.lastDeity = msg.deityKey
	m.runningDeity = msg.deityKey

	if msg.err != nil {
		m.outputLines = []string{
			"",
			"  " + lipgloss.NewStyle().Foreground(Red).Render("✗ "+msg.err.Error()),
		}
		if msg.deityKey != "" {
			m.deityState[msg.deityKey] = stateFailed
		}
	} else {
		m.outputLines = msg.lines
		if msg.deityKey != "" {
			state := stateSucceeded
			if msg.deityKey == "anubis" {
				if scan, err := jackal.LoadLatest(); err == nil && len(scan.Findings) > 0 {
					state = stateHasData
				}
			}
			m.deityState[msg.deityKey] = state
		}
	}

	// Use fix commands from the renderer if provided (actionable results).
	// Otherwise fall back to the generic suggest engine.
	// Known fix descriptions
	fixDescs := map[string]string{
		"anubis clean --dry-run": "Preview safe items to clean",
		"anubis clean --confirm": "Clean safe items (move to Trash)",
		"anubis scan":            "Run a fresh scan",
		"isis fix":               "Auto-fix DNS, firewall, security",
		"isis diagnose":          "Full system health check",
	}

	if len(msg.fixCmds) > 0 {
		m.postRunCmds = msg.fixCmds
		m.postRunActions = nil
		for _, cmd := range msg.fixCmds {
			desc := fixDescs[cmd]
			if desc == "" {
				desc = "Run " + cmd
			}
			m.postRunActions = append(m.postRunActions, suggest.Action{
				Command:     cmd,
				Description: desc,
			})
		}
	} else {
		ctx := m.buildSuggestContext()
		ctx.Deity = msg.deityKey
		if msg.err != nil {
			ctx.Err = msg.err
			m.postRunActions = suggest.OnError(ctx)
		} else {
			m.postRunActions = suggest.After(ctx)
		}
		m.postRunCmds = suggest.Commands(ctx)
		if msg.deityKey == "anubis" {
			if scan, loadErr := jackal.LoadLatest(); loadErr == nil {
				ctx.FindingsCount = len(scan.Findings)
			}
			m.postRunActions = suggest.After(ctx)
			m.postRunCmds = suggest.Commands(ctx)
		}
	}

	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.recalcViewport()
	m.viewport.GotoTop()
	m.savePersistedState()

	// Record notification
	if m.notifyStore != nil && msg.deityKey != "" {
		sev := notify.SeveritySuccess
		summary := fmt.Sprintf("%s completed", m.runningCmd)
		if msg.err != nil {
			sev = notify.SeverityError
			summary = fmt.Sprintf("%s failed", m.runningCmd)
		}
		_ = m.notifyStore.Record(notify.Notification{
			Source: msg.deityKey, Action: m.runningCmd,
			Severity: sev, Summary: summary,
		})
		m.notifyRefreshTime = time.Time{}
		m.refreshNotifications()
	}

	m.runningDeity = ""
	m.runningCmd = ""
	m.runningArgs = nil
	return m, nil
}

// ── View ─────────────────────────────────────────────────────────────

func (m TUIModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var b strings.Builder
	maxW := min(m.width-2, 120)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	dimText := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	gold := lipgloss.NewStyle().Foreground(Gold)
	heavyDiv := dim.Render(strings.Repeat("━", maxW))
	lightDiv := dim.Render(strings.Repeat("─", maxW))

	// ── Persistent Header — "This is Pantheon" ──
	ver := readVersionFile()
	titleLine := "  " + gold.Bold(true).Render("𓉴 PANTHEON")
	if ver != "" {
		pad := maxW - visibleLen(titleLine) - len(ver) - 2
		if pad < 1 {
			pad = 1
		}
		titleLine += strings.Repeat(" ", pad) + dimText.Render(ver)
	}
	tagline := "  " + dimText.Render("Unified DevOps Intelligence")
	urlPad := maxW - visibleLen(tagline) - len("sirsi.ai") - 2
	if urlPad < 1 {
		urlPad = 1
	}
	tagline += strings.Repeat(" ", urlPad) + dimText.Render("sirsi.ai")

	b.WriteString("\n")
	b.WriteString(titleLine + "\n")
	b.WriteString(tagline + "\n")
	b.WriteString(" " + heavyDiv + "\n")

	// ── Tab bar ──
	b.WriteString(m.renderTabBar())
	b.WriteString(" " + lightDiv + "\n")

	// ── Content ──
	switch m.mode {
	case viewTabs:
		b.WriteString(m.renderTabPage())
	case viewRunning:
		b.WriteString(m.renderRunning())
	case viewDone:
		b.WriteString(m.renderDone())
	case viewPrompt:
		b.WriteString(m.renderTabPage())
	case viewSelect:
		b.WriteString(m.renderSelect())
	case viewAnalyze:
		b.WriteString(m.renderAnalyze())
	}

	// ── Bottom bar ──
	b.WriteString(" " + lightDiv + "\n")
	if m.mode == viewPrompt {
		b.WriteString(" " + m.input.View() + "\n")
	} else {
		b.WriteString(m.renderBottomHints() + "\n")
	}

	// Push footer to bottom
	content := b.String()
	lines := strings.Count(content, "\n")
	remaining := m.height - lines - 2
	if remaining > 0 {
		content += strings.Repeat("\n", remaining)
	}
	// Footer: faint brand + uptime if available
	footerRight := ""
	if m.vitals.UptimeStr != "" {
		footerRight = dimText.Render("up " + m.vitals.UptimeStr)
	}
	footerLeft := dimText.Render(" 𓉴 sirsi-pantheon")
	footPad := maxW - visibleLen(footerLeft) - visibleLen(footerRight)
	if footPad < 1 {
		footPad = 1
	}
	content += footerLeft + strings.Repeat(" ", footPad) + footerRight

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// readVersionFile reads the VERSION file from the repo root or ~/.config.
func readVersionFile() string {
	for _, p := range []string{"VERSION", filepath.Join(os.Getenv("HOME"), ".config", "sirsi", "VERSION")} {
		if data, err := os.ReadFile(p); err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return "v" + v
			}
		}
	}
	return ""
}

// renderTabBar draws the horizontal tab switcher.
// Tab names only — no hieroglyphs here (inconsistent terminal widths).
// Glyphs live in the tab landing pages where alignment doesn't matter.
func (m TUIModel) renderTabBar() string {
	active := lipgloss.NewStyle().Foreground(Gold).Bold(true)
	inactive := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	dot := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("·")

	var parts []string
	for i, tab := range tabs {
		if i == m.activeTab {
			parts = append(parts, active.Render("▸ "+tab.Name))
		} else {
			parts = append(parts, inactive.Render("  "+tab.Name))
		}
	}

	return "  " + strings.Join(parts, "  "+dot+"  ") + "\n"
}

// renderTabPage draws the landing page for the active tab.
func (m TUIModel) renderTabPage() string {
	tab := tabs[m.activeTab]
	gold := lipgloss.NewStyle().Foreground(Gold)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))


	var b strings.Builder

	if tab.Name == "Status" {
		b.WriteString(m.renderStatusPage(gold, dim))
	} else {
		b.WriteString("\n")
		b.WriteString("  " + gold.Bold(true).Render(tab.Glyph+"  "+tab.Name) + "\n")
		b.WriteString("  " + lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#555555")).
			Render(tab.Tagline) + "\n")
		b.WriteString("\n")

		for i, action := range tab.Actions {
			keyBadge := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F0F0F")).
				Background(Gold).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf("%d", i+1))
			b.WriteString("  " + keyBadge +
				"  " + body.Bold(true).Render(action.Label) + "\n")
			b.WriteString("       " + dim.Render(action.Desc) + "\n\n")
		}
	}

	return b.String()
}

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

// visibleLen estimates the visible length of a string, stripping ANSI escapes.
func visibleLen(s string) int {
	n := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		n++
	}
	return n
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

// renderCard renders a small bento card with a label, big value, and subtitle.
func (m TUIModel) renderCard(labelText, value, subtitle, icon string, width int) string {
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	bigNum := lipgloss.NewStyle().Foreground(White).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	card := label.Render(icon+" "+labelText) + "\n" +
		"  " + bigNum.Render(value) + "\n"
	if subtitle != "" {
		card += "  " + dim.Render(subtitle) + "\n"
	}

	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(0, 1).
		Render(card)
}

// renderRunning shows the command execution screen.
func (m TUIModel) renderRunning() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	glyph, name := deity.Display(m.runningDeity)
	elapsed := time.Since(m.cmdStartTime).Truncate(time.Second)
	elapsedStr := ""
	if elapsed >= time.Second {
		elapsedStr = " " + dim.Render(fmt.Sprintf("(%s)", elapsed))
	}

	b.WriteString("\n")
	b.WriteString("  " + m.spinner.View() + " " +
		lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(glyph+" "+name) +
		"  " + dim.Render(m.runningCmd) + elapsedStr + "\n")
	b.WriteString("\n")
	b.WriteString(m.viewport.View() + "\n")

	return b.String()
}

// renderDone shows command output + numbered next actions.
func (m TUIModel) renderDone() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	green := lipgloss.NewStyle().Foreground(Green)
	body := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))


	b.WriteString("\n")
	b.WriteString(m.viewport.View() + "\n")

	// ── Completion banner — decree from the deity that ran ──
	bannerMsg := "Judgment Complete"
	if m.lastDeity != "" {
		glyph, name := deity.Display(m.lastDeity)
		bannerMsg = glyph + " " + name + " — Complete"
	}
	b.WriteString("\n")
	b.WriteString("  " + ResultBanner(bannerMsg, green, 50) + "\n")
	b.WriteString("\n")

	// ── Numbered next actions ──
	if len(m.postRunCmds) > 0 {
		b.WriteString("  " + dim.Render("What's next?") + "\n\n")
	}
	shown := 0
	for i, cmd := range m.postRunCmds {
		if i >= 3 {
			break
		}
		desc := ""
		if i < len(m.postRunActions) {
			desc = m.postRunActions[i].Description
		}
		keyBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F0F0F")).
			Background(Gold).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("%d", i+1))
		line := "   " + keyBadge + "  " + body.Render(cmd)
		if desc != "" {
			line += "  " + dim.Render(desc)
		}
		b.WriteString(line + "\n")
		shown++
	}

	b.WriteString("\n")
	if shown > 0 {
		b.WriteString("   " + dim.Render(fmt.Sprintf("press 1-%d to continue  ·  esc back  ·  : command", shown)) + "\n")
	} else {
		b.WriteString("   " + dim.Render("esc back  ·  : command") + "\n")
	}

	return b.String()
}

// renderBottomHints shows context-appropriate key hints.
func (m TUIModel) renderBottomHints() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	var hints []string

	switch m.mode {
	case viewTabs:
		n := len(tabs[m.activeTab].Actions)
		hints = []string{"←/→ switch tabs", fmt.Sprintf("1-%d act", n), ": command", "q quit"}
	case viewRunning:
		hints = []string{"↑/↓ scroll", "ctrl+c cancel"}
	case viewDone:
		if len(m.postRunCmds) > 0 {
			hints = []string{"1-3 next action", "↑/↓ scroll", ": command", "esc back"}
		} else {
			hints = []string{"↑/↓ scroll", ": command", "esc back"}
		}
	case viewSelect:
		hints = []string{"↑/↓ move", "space toggle", "a all", "enter confirm", "esc cancel"}
	case viewAnalyze:
		hints = []string{"↑/↓ navigate", "enter drill-down", "esc back", "q quit"}
	}

	return " " + dim.Render(strings.Join(hints, "  ·  "))
}

// ── Suggest Context ──────────────────────────────────────────────────

func (m TUIModel) buildSuggestContext() suggest.Context {
	sub := ""
	if len(m.runningArgs) >= 2 {
		sub = m.runningArgs[1]
	}
	ctx := suggest.Context{
		Deity:      m.runningDeity,
		Subcommand: sub,
	}
	if m.runningDeity == "anubis" {
		if scan, err := jackal.LoadLatest(); err == nil {
			ctx.FindingsCount = len(scan.Findings)
		}
	}
	return ctx
}

// ── Background Refresh ───────────────────────────────────────────────

func (m *TUIModel) refreshActive() {
	for k := range m.activeDeity {
		delete(m.activeDeity, k)
	}
	entries, _ := m.steleReader.ReadNew()
	now := time.Now()
	for _, e := range entries {
		ts, err := time.Parse(time.RFC3339, e.TS)
		if err != nil {
			continue
		}
		if now.Sub(ts) < 5*time.Minute {
			dKey := strings.ToLower(e.Deity)
			if !strings.Contains(dKey, ":") {
				m.activeDeity[dKey] = true
			}
		}
	}
	m.refreshNotifications()
	m.refreshVitals()

	home, _ := os.UserHomeDir()
	pidDir := filepath.Join(home, ".config", "ra", "pids")
	pidEntries, _ := os.ReadDir(pidDir)
	for _, f := range pidEntries {
		if f.IsDir() {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".pid")
		for _, d := range deity.Roster {
			if strings.Contains(strings.ToLower(name), d.Key) {
				m.activeDeity[d.Key] = true
			}
		}
	}
}

func (m *TUIModel) refreshVitals() {
	m.vitals = vitals.Collect()
}

const maxHistory = 60

func (m *TUIModel) appendHistory() {
	m.cpuHistory = appendCapped(m.cpuHistory, m.vitals.CPUPercent)
	m.memHistory = appendCapped(m.memHistory, m.vitals.RAMPercent)
	downNorm := m.vitals.NetDownBps / (10 * 1024 * 1024) * 100
	upNorm := m.vitals.NetUpBps / (10 * 1024 * 1024) * 100
	m.netDownHist = appendCapped(m.netDownHist, downNorm)
	m.netUpHist = appendCapped(m.netUpHist, upNorm)
}

func appendCapped(buf []float64, val float64) []float64 {
	buf = append(buf, val)
	if len(buf) > maxHistory {
		buf = buf[len(buf)-maxHistory:]
	}
	return buf
}

func (m *TUIModel) refreshNotifications() {
	if m.notifyStore == nil {
		return
	}
	now := time.Now()
	if now.Sub(m.notifyRefreshTime) < 5*time.Second {
		return
	}
	m.notifyRefreshTime = now
	items, err := m.notifyStore.Recent(5)
	if err != nil {
		return
	}
	m.recentNotify = items
}

// ── Layout ───────────────────────────────────────────────────────────

func (m *TUIModel) recalcViewport() {
	// Reserve: tab bar(2) + divider(1) + running header(2) + bottom divider(1) + hints(1) + padding(1)
	vpHeight := m.height - 8
	if m.mode == viewDone && len(m.postRunCmds) > 0 {
		shown := min(len(m.postRunCmds), 3)
		vpHeight -= (shown * 3) + 3 // each action ~3 lines + header + spacing
	}
	// Cap viewport to content size — don't waste space with blank lines
	if len(m.outputLines) > 0 && len(m.outputLines) < vpHeight {
		vpHeight = len(m.outputLines) + 1
	}
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewport.SetHeight(vpHeight)

	vpWidth := m.width - 4
	if vpWidth < 20 {
		vpWidth = 20
	}
	m.viewport.SetWidth(vpWidth)
}

// ── Helpers ──────────────────────────────────────────────────────────

func pluralize(word string, n int) string {
	if n == 1 {
		return word
	}
	if strings.HasSuffix(word, "y") {
		return word[:len(word)-1] + "ies"
	}
	return word + "s"
}

// ── Launcher ─────────────────────────────────────────────────────────

func LaunchTUI() error {
	return LaunchTUIWithNotify(nil)
}

func LaunchTUIWithNotify(store *notify.Store) error {
	m := NewTUIModel()
	m.notifyStore = store
	m.refreshNotifications()
	m.loadPersistedState()
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// ── Persistent State ─────────────────────────────────────────────────

func (m *TUIModel) loadPersistedState() {
	state, err := deity.LoadState()
	if err != nil {
		return
	}
	for k, v := range state.DeityState {
		m.deityState[k] = v
	}
}

func (m *TUIModel) savePersistedState() {
	_ = deity.SaveState(deity.PersistedState{DeityState: m.deityState})
}

// ── Unused but required by tests ─────────────────────────────────────
// These are no-ops preserved for test compilation. The old REPL functions
// (showFindings, showHelp, renderRosterColumns, etc.) are removed.

var _ = color.RGBA{} // keep image/color import for lipgloss
