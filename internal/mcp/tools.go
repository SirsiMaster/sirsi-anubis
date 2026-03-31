package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/brain"
	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal/rules"
	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
	"github.com/SirsiMaster/sirsi-pantheon/internal/maat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/thoth"
)

// registerTools adds all Anubis tools to the MCP server.
func registerTools(s *Server) {
	s.RegisterTool(Tool{
		Name:        "scan_workspace",
		Description: "Scan a workspace directory for infrastructure waste (stale caches, orphaned build artifacts, unused dependencies). Read-only — never deletes anything.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"path": {
					Type:        "string",
					Description: "Absolute path to the workspace directory to scan. Defaults to current working directory.",
				},
				"category": {
					Type:        "string",
					Description: "Optional category filter: general, dev, ai, vms, ides, cloud, storage. Leave empty for all.",
				},
			},
		},
	}, handleScanWorkspace)

	s.RegisterTool(Tool{
		Name:        "ghost_report",
		Description: "Hunt for ghost apps — remnants of uninstalled applications lingering on the system (preferences, caches, launch agents, Spotlight registrations).",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"target": {
					Type:        "string",
					Description: "Optional: hunt for a specific ghost by app name or bundle ID.",
				},
			},
		},
	}, handleGhostReport)

	s.RegisterTool(Tool{
		Name:        "health_check",
		Description: "Quick system health summary — CPU, RAM, GPU, disk usage, and infrastructure hygiene score.",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]SchemaField{},
		},
	}, handleHealthCheck)

	s.RegisterTool(Tool{
		Name:        "classify_files",
		Description: "Classify files semantically using Anubis brain (junk, project, config, model, media, etc). Uses heuristic classifier or neural model if installed.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"paths": {
					Type:        "string",
					Description: "Comma-separated list of file paths to classify.",
				},
			},
			Required: []string{"paths"},
		},
	}, handleClassifyFiles)

	s.RegisterTool(Tool{
		Name:        "thoth_read_memory",
		Description: "𓁟 Read the project's Thoth memory file (.thoth/memory.yaml) for instant project context. Call this at the start of every conversation to understand the project without reading source files. Returns architecture, design decisions, limitations, and file map.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"path": {
					Type:        "string",
					Description: "Path to the project root. Defaults to current working directory.",
				},
			},
		},
	}, handleThothReadMemory)

	s.RegisterTool(Tool{
		Name:        "detect_hardware",
		Description: "Detect system hardware accelerators (GPU/NPU/TPU) and resource limits (RAM/CPU/Metal).",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]SchemaField{},
		},
	}, handleDetectHardware)

	s.RegisterTool(Tool{
		Name:        "thoth_sync",
		Description: "Trigger Thoth memory sync — auto-discovers codebase facts and updates .thoth/memory.yaml and journal.md.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"path": {
					Type:        "string",
					Description: "Path to the project root. Defaults to current working directory.",
				},
			},
		},
	}, handleThothSync)

	s.RegisterTool(Tool{
		Name:        "maat_audit",
		Description: "Run coverage governance audit across all modules. Returns per-module coverage and pass/fail verdicts.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"skip_tests": {
					Type:        "boolean",
					Description: "Skip running go test, use cached coverage only. Defaults to true for speed.",
				},
			},
		},
	}, handleMaatAudit)

	s.RegisterTool(Tool{
		Name:        "anubis_weigh",
		Description: "Run a full disk waste analysis and return a dashboard summary of infrastructure waste.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"path": {
					Type:        "string",
					Description: "Path to scan. Defaults to current working directory.",
				},
			},
		},
	}, handleAnubisWeigh)

	s.RegisterTool(Tool{
		Name:        "judge_cleanup",
		Description: "DRY RUN ONLY — Preview what would be cleaned without deleting anything. Use the CLI 'pantheon anubis judge --confirm' for actual cleanup.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]SchemaField{
				"path": {
					Type:        "string",
					Description: "Path to scan. Defaults to current working directory.",
				},
			},
		},
	}, handleJudgeCleanup)

	s.RegisterTool(Tool{
		Name:        "pantheon_status",
		Description: "Overall Pantheon system status — modules, version, brain status, and operational readiness.",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]SchemaField{},
		},
	}, makePantheonStatusHandler(s))
}

// handleScanWorkspace runs the Jackal scan engine on a workspace.
func handleScanWorkspace(args map[string]interface{}) (*ToolResult, error) {
	// Parse path argument
	scanPath, _ := args["path"].(string)
	if scanPath == "" {
		var err error
		scanPath, err = os.Getwd()
		if err != nil {
			return textResult(fmt.Sprintf("Could not determine working directory: %v", err), true), nil
		}
	}

	// Parse category filter
	categoryStr, _ := args["category"].(string)
	var categories []jackal.Category
	if categoryStr != "" {
		cat, err := parseCategory(categoryStr)
		if err != nil {
			return textResult(fmt.Sprintf("Invalid category %q: %v", categoryStr, err), true), nil
		}
		categories = []jackal.Category{cat}
	}

	// Create engine and run scan with aggressive timeout.
	// MCP callers (AI IDEs) should not wait >5s for context.
	engine := jackal.NewEngine()
	engine.RegisterAll(rules.AllRules()...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := engine.Scan(ctx, jackal.ScanOptions{
		Categories: categories,
	})
	if err != nil {
		return textResult(fmt.Sprintf("Scan failed: %v", err), true), nil
	}

	// Format results
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("𓁢 Anubis Scan Results for: %s\n\n", scanPath))
	sb.WriteString(fmt.Sprintf("Total waste found: %s\n", jackal.FormatSize(result.TotalSize)))
	sb.WriteString(fmt.Sprintf("Findings: %d across %d rules\n\n", len(result.Findings), result.RulesRan))

	// Category breakdown
	if len(result.ByCategory) > 0 {
		sb.WriteString("Category Breakdown:\n")
		for cat, summary := range result.ByCategory {
			sb.WriteString(fmt.Sprintf("  • %s: %s (%d items)\n",
				string(cat), jackal.FormatSize(summary.TotalSize), summary.Findings))
		}
		sb.WriteString("\n")
	}

	// Top findings (up to 20)
	limit := 20
	if len(result.Findings) < limit {
		limit = len(result.Findings)
	}
	if limit > 0 {
		sb.WriteString(fmt.Sprintf("Top %d Findings:\n", limit))
		for _, f := range result.Findings[:limit] {
			sb.WriteString(fmt.Sprintf("  %s — %s (%s)\n",
				f.Description, shortenHomePath(f.Path), jackal.FormatSize(f.SizeBytes)))
		}
	}

	if len(result.Findings) > limit {
		sb.WriteString(fmt.Sprintf("\n  ... and %d more findings\n", len(result.Findings)-limit))
	}

	sb.WriteString("\nRun 'anubis judge --dry-run' to preview cleanup.")

	return textResult(sb.String(), false), nil
}

// handleGhostReport runs Ka ghost detection.
func handleGhostReport(args map[string]interface{}) (*ToolResult, error) {
	target, _ := args["target"].(string)

	scanner := ka.NewScanner()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ghosts, err := scanner.Scan(ctx, false) // no sudo in MCP mode
	if err != nil {
		return textResult(fmt.Sprintf("Ghost scan failed: %v", err), true), nil
	}

	// Filter by target if specified
	if target != "" {
		var filtered []ka.Ghost
		target = strings.ToLower(target)
		for _, g := range ghosts {
			if strings.Contains(strings.ToLower(g.AppName), target) ||
				strings.Contains(strings.ToLower(g.BundleID), target) {
				filtered = append(filtered, g)
			}
		}
		ghosts = filtered
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚠️ Ka Ghost Report — %d ghosts detected\n\n", len(ghosts)))

	if len(ghosts) == 0 {
		sb.WriteString("No ghost apps found. Machine is spiritually clean.")
		return textResult(sb.String(), false), nil
	}

	var totalSize int64
	var totalResiduals int

	for i, ghost := range ghosts {
		totalSize += ghost.TotalSize
		totalResiduals += len(ghost.Residuals)

		spotlightTag := ""
		if ghost.InLaunchServices {
			spotlightTag = " 👻 (in Spotlight)"
		}

		sizeStr := ""
		if ghost.TotalSize > 0 {
			sizeStr = fmt.Sprintf(" — %s", jackal.FormatSize(ghost.TotalSize))
		}

		sb.WriteString(fmt.Sprintf("[%d] %s%s%s\n", i+1, ghost.AppName, sizeStr, spotlightTag))
		sb.WriteString(fmt.Sprintf("    Bundle: %s\n", ghost.BundleID))

		for _, r := range ghost.Residuals {
			sb.WriteString(fmt.Sprintf("    ├─ %s %s %s\n",
				string(r.Type), jackal.FormatSize(r.SizeBytes), shortenHomePath(r.Path)))
		}
		sb.WriteString("\n")

		// Limit output for MCP (avoid huge context)
		if i >= 29 {
			sb.WriteString(fmt.Sprintf("  ... and %d more ghosts\n\n", len(ghosts)-30))
			break
		}
	}

	sb.WriteString(fmt.Sprintf("Summary: %d ghosts, %d residuals, %s total waste\n",
		len(ghosts), totalResiduals, jackal.FormatSize(totalSize)))
	sb.WriteString("Run 'anubis ka --clean --dry-run' to preview cleanup.")

	return textResult(sb.String(), false), nil
}

// handleHealthCheck provides an instant system health summary.
// PERFORMANCE: Uses cached Horus index + static system info. No live scans.
// Target: <10ms response time (was 17s with live Jackal+Ka scans).
func handleHealthCheck(_ map[string]interface{}) (*ToolResult, error) {
	start := time.Now()
	var sb strings.Builder
	sb.WriteString("𓁢 Anubis Health Check\n\n")

	// System info — instant (runtime constants)
	sb.WriteString(fmt.Sprintf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("CPUs: %d\n", runtime.NumCPU()))
	sb.WriteString(fmt.Sprintf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0)))

	// Horus index — read cache only, NEVER trigger a build.
	// LoadManifest returns in <1ms (gob decode) or fails instantly (no file).
	m, err := horus.LoadManifest(horus.DefaultCachePath())
	if err == nil {
		sb.WriteString(fmt.Sprintf("Indexed: %d dirs, %d files\n",
			m.Stats.DirsWalked, m.Stats.FilesIndexed))
		sb.WriteString(fmt.Sprintf("Index age: %s\n",
			time.Since(m.Timestamp).Truncate(time.Second)))
	} else {
		sb.WriteString("Horus index: not cached (run 'pantheon weigh' to build)\n")
	}

	// Brain status — instant (file existence check)
	if brain.IsInstalled() {
		sb.WriteString("Neural brain: ✅ Installed\n")
	} else {
		sb.WriteString("Neural brain: Not installed (run 'pantheon install-brain')\n")
	}

	// Watchdog status — instant (ring buffer read)
	bridge := GetWatchdogBridge()
	if bridge != nil {
		buffered, lifetime := bridge.Ring().Stats()
		sb.WriteString(fmt.Sprintf("Watchdog: active (%d alerts, %d lifetime)\n", buffered, lifetime))
	} else {
		sb.WriteString("Watchdog: dormant\n")
	}

	sb.WriteString(fmt.Sprintf("\nResponse time: %s\n", time.Since(start).Round(time.Microsecond)))
	sb.WriteString("\nFor full scan: 'pantheon weigh' or call scan_workspace tool.")

	return textResult(sb.String(), false), nil
}

// handleClassifyFiles classifies files using the brain module.
func handleClassifyFiles(args map[string]interface{}) (*ToolResult, error) {
	pathsRaw, _ := args["paths"].(string)
	if pathsRaw == "" {
		return textResult("Missing required parameter: paths", true), nil
	}

	paths := strings.Split(pathsRaw, ",")
	for i := range paths {
		paths[i] = strings.TrimSpace(paths[i])
	}

	classifier, err := brain.GetClassifier()
	if err != nil {
		return textResult(fmt.Sprintf("Classifier error: %v", err), true), nil
	}
	defer classifier.Close()

	result, err := classifier.ClassifyBatch(paths, 4)
	if err != nil {
		return textResult(fmt.Sprintf("Classification failed: %v", err), true), nil
	}

	// Format as JSON for structured consumption
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return textResult(fmt.Sprintf("Marshal error: %v", err), true), nil
	}

	return textResult(string(data), false), nil
}

// handleThothReadMemory reads the project's Thoth memory and journal files.
// This gives AI assistants instant project context without reading source files.
func handleThothReadMemory(args map[string]interface{}) (*ToolResult, error) {
	projectPath, _ := args["path"].(string)
	if projectPath == "" {
		var err error
		projectPath, err = os.Getwd()
		if err != nil {
			return textResult("Could not determine working directory", true), nil
		}
	}

	var sb strings.Builder
	sb.WriteString("𓁟 Thoth Memory\n\n")

	// Try .thoth/memory.yaml first, then legacy .anubis-memory.yaml
	memoryPaths := []string{
		projectPath + "/.thoth/memory.yaml",
		projectPath + "/.anubis-memory.yaml",
	}

	found := false
	for _, mp := range memoryPaths {
		data, err := os.ReadFile(mp)
		if err == nil {
			sb.WriteString("=== Memory ===\n")
			sb.WriteString(string(data))
			sb.WriteString("\n")
			found = true
			break
		}
	}

	if !found {
		sb.WriteString("No Thoth memory file found.\n")
		sb.WriteString("Create one: mkdir -p .thoth && touch .thoth/memory.yaml\n")
		sb.WriteString("See: docs/THOTH.md for the specification.\n")
		return textResult(sb.String(), false), nil
	}

	// Try to read journal (optional)
	journalPaths := []string{
		projectPath + "/.thoth/journal.md",
		projectPath + "/.anubis-journal.md",
	}
	for _, jp := range journalPaths {
		data, err := os.ReadFile(jp)
		if err == nil {
			sb.WriteString("\n=== Journal (last 2000 chars) ===\n")
			content := string(data)
			if len(content) > 2000 {
				content = content[len(content)-2000:]
			}
			sb.WriteString(content)
			break
		}
	}

	return textResult(sb.String(), false), nil
}

// ---- Helpers ----

// textResult creates a simple text ToolResult.
func textResult(text string, isError bool) *ToolResult {
	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: text},
		},
		IsError: isError,
	}
}

// parseCategory converts a string to a jackal.Category.
func parseCategory(s string) (jackal.Category, error) {
	switch strings.ToLower(s) {
	case "general":
		return jackal.CategoryGeneral, nil
	case "dev", "developer":
		return jackal.CategoryDev, nil
	case "ai", "ml", "ai-ml":
		return jackal.CategoryAI, nil
	case "vms", "virtualization":
		return jackal.CategoryVirtualization, nil
	case "ides", "ide":
		return jackal.CategoryIDEs, nil
	case "cloud", "infra":
		return jackal.CategoryCloud, nil
	case "storage":
		return jackal.CategoryStorage, nil
	default:
		return "", fmt.Errorf("unknown category: %s. Valid: general, dev, ai, vms, ides, cloud, storage", s)
	}
}

// shortenHomePath replaces the home directory with ~.
func shortenHomePath(path string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// handleThothSync triggers Thoth memory and journal sync.
func handleThothSync(args map[string]interface{}) (*ToolResult, error) {
	projectPath, _ := args["path"].(string)
	if projectPath == "" {
		var err error
		projectPath, err = os.Getwd()
		if err != nil {
			return textResult("Could not determine working directory", true), nil
		}
	}

	// Sync memory
	err := thoth.Sync(thoth.SyncOptions{RepoRoot: projectPath, UpdateDate: true})
	if err != nil {
		return textResult(fmt.Sprintf("Thoth memory sync failed: %v", err), true), nil
	}

	// Sync journal
	commitCount, err := thoth.SyncJournal(thoth.JournalSyncOptions{
		RepoRoot: projectPath,
		Since:    "24 hours ago",
	})
	if err != nil {
		// Journal sync failure is non-fatal
		return textResult(fmt.Sprintf("Memory synced. Journal sync failed: %v", err), false), nil
	}

	return textResult(fmt.Sprintf("Thoth sync complete.\n- Memory updated: %s/.thoth/memory.yaml\n- Journal: %d commits processed", projectPath, commitCount), false), nil
}

// handleMaatAudit runs a coverage governance audit.
func handleMaatAudit(args map[string]interface{}) (*ToolResult, error) {
	skipTests := true // default to true for MCP (speed)
	if v, ok := args["skip_tests"].(bool); ok {
		skipTests = v
	}

	assessor := &maat.CoverageAssessor{
		Thresholds: maat.DefaultThresholds(),
		SkipTests:  skipTests,
	}

	report, err := maat.Weigh(assessor)
	if err != nil {
		return textResult(fmt.Sprintf("Ma'at audit failed: %v", err), true), nil
	}

	var sb strings.Builder
	sb.WriteString("Ma'at Governance Audit\n\n")
	sb.WriteString(fmt.Sprintf("Verdict: %s\n", report.OverallVerdict))
	sb.WriteString(fmt.Sprintf("Weight: %d / 100\n", report.OverallWeight))
	sb.WriteString(fmt.Sprintf("Assessments: %d passed, %d warned, %d failed\n\n",
		report.Passes, report.Warnings, report.Failures))

	for _, a := range report.Assessments {
		sb.WriteString(fmt.Sprintf("  [%s] %s — weight %d\n",
			a.Verdict, a.Subject, a.FeatherWeight))
	}

	return textResult(sb.String(), false), nil
}

// handleAnubisWeigh runs a full disk waste scan.
func handleAnubisWeigh(args map[string]interface{}) (*ToolResult, error) {
	scanPath, _ := args["path"].(string)
	if scanPath == "" {
		var err error
		scanPath, err = os.Getwd()
		if err != nil {
			return textResult("Could not determine working directory", true), nil
		}
	}

	engine := jackal.NewEngine()
	engine.RegisterAll(rules.AllRules()...)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := engine.Scan(ctx, jackal.ScanOptions{})
	if err != nil {
		return textResult(fmt.Sprintf("Weigh failed: %v", err), true), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Anubis Weigh — %s\n\n", scanPath))
	sb.WriteString(fmt.Sprintf("Total waste: %s\n", jackal.FormatSize(result.TotalSize)))
	sb.WriteString(fmt.Sprintf("Findings: %d across %d rules\n\n", len(result.Findings), result.RulesRan))

	if len(result.ByCategory) > 0 {
		sb.WriteString("By Category:\n")
		for cat, summary := range result.ByCategory {
			sb.WriteString(fmt.Sprintf("  %s: %s (%d items)\n",
				string(cat), jackal.FormatSize(summary.TotalSize), summary.Findings))
		}
	}

	return textResult(sb.String(), false), nil
}

// handleJudgeCleanup returns a dry-run cleanup preview. NEVER deletes anything.
func handleJudgeCleanup(_ map[string]interface{}) (*ToolResult, error) {
	engine := jackal.NewEngine()
	engine.RegisterAll(rules.AllRules()...)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := engine.Scan(ctx, jackal.ScanOptions{})
	if err != nil {
		return textResult(fmt.Sprintf("Scan failed: %v", err), true), nil
	}

	var sb strings.Builder
	sb.WriteString("Judge Cleanup Preview (DRY RUN)\n\n")
	sb.WriteString(fmt.Sprintf("Would clean: %s across %d items\n\n",
		jackal.FormatSize(result.TotalSize), len(result.Findings)))

	limit := 30
	if len(result.Findings) < limit {
		limit = len(result.Findings)
	}

	for _, f := range result.Findings[:limit] {
		sb.WriteString(fmt.Sprintf("  [REMOVE] %s — %s\n",
			shortenHomePath(f.Path), jackal.FormatSize(f.SizeBytes)))
	}

	if len(result.Findings) > limit {
		sb.WriteString(fmt.Sprintf("\n  ... and %d more items\n", len(result.Findings)-limit))
	}

	sb.WriteString("\nThis is a DRY RUN preview. No files were deleted.")
	sb.WriteString("\nRun 'pantheon anubis judge --confirm' in terminal for actual cleanup.")

	return textResult(sb.String(), false), nil
}

// makePantheonStatusHandler creates a handler that captures the server reference.
func makePantheonStatusHandler(s *Server) func(map[string]interface{}) (*ToolResult, error) {
	return func(_ map[string]interface{}) (*ToolResult, error) {
		var sb strings.Builder
		sb.WriteString("Pantheon System Status\n\n")
		sb.WriteString(fmt.Sprintf("Version: %s\n", ServerVersion))
		sb.WriteString(fmt.Sprintf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH))
		sb.WriteString(fmt.Sprintf("CPUs: %d\n", runtime.NumCPU()))
		sb.WriteString(fmt.Sprintf("MCP Tools: %d registered\n", len(s.tools)))
		sb.WriteString(fmt.Sprintf("MCP Resources: %d registered\n", len(s.resources)))

		// Brain status
		if brain.IsInstalled() {
			sb.WriteString("Brain: installed\n")
		} else {
			sb.WriteString("Brain: not installed\n")
		}

		// Horus index
		m, err := horus.LoadManifest(horus.DefaultCachePath())
		if err == nil {
			sb.WriteString(fmt.Sprintf("Horus Index: %d files, age %s\n",
				m.Stats.FilesIndexed, time.Since(m.Timestamp).Truncate(time.Second)))
		} else {
			sb.WriteString("Horus Index: not cached\n")
		}

		// Watchdog
		bridge := GetWatchdogBridge()
		if bridge != nil {
			sb.WriteString("Watchdog: active\n")
		} else {
			sb.WriteString("Watchdog: dormant\n")
		}

		sb.WriteString("\nModules: jackal, ka, guard, sight, hapi, seba, scarab, thoth, seshat, neith, maat, mcp, brain")

		return textResult(sb.String(), false), nil
	}
}

// handleDetectHardware returns the system hardware profile via Hapi.
func handleDetectHardware(_ map[string]interface{}) (*ToolResult, error) {
	bridge, err := brain.NewHapiBridge()
	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{
				{Type: "text", Text: fmt.Sprintf("Hardware detection failed: %v", err)},
			},
			IsError: true,
		}, nil
	}

	profile := bridge.Profile()
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: fmt.Sprintf("𓁢 Hardware Profile Detected:\n\n```json\n%s\n```\nRecommended Backend: %s", string(data), bridge.BackendPreference())},
		},
	}, nil
}
