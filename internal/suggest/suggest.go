package suggest

import "strings"

// After returns recommended actions after a deity command completes successfully.
// Returns nil if no suggestions are applicable.
func After(ctx Context) []Action {
	if ctx.Err != nil {
		return OnError(ctx)
	}

	switch ctx.Deity {
	case "anubis":
		return afterAnubis(ctx)
	case "isis":
		return afterIsis(ctx)
	case "maat":
		return afterMaat(ctx)
	case "ra":
		return afterRa(ctx)
	case "net":
		return afterNet(ctx)
	case "thoth":
		return afterThoth(ctx)
	case "seshat":
		return afterSeshat(ctx)
	case "seba":
		return afterSeba(ctx)
	case "osiris":
		return afterOsiris(ctx)
	case "horus":
		return afterHorus(ctx)
	}
	return nil
}

// OnError returns remediation suggestions when a command fails.
func OnError(ctx Context) []Action {
	errStr := ""
	if ctx.Err != nil {
		errStr = strings.ToLower(ctx.Err.Error())
	}

	// Pattern-match common error classes first.
	switch {
	case strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "operation not permitted"):
		return []Action{
			{Command: "", Short: "Check permissions", Description: "Re-run with --sudo if supported, or grant Full Disk Access in System Settings → Privacy", Priority: 0},
		}
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no such file"):
		return []Action{
			{Command: "sirsi doctor", Short: "Run doctor", Description: "Health diagnostic to identify missing dependencies", Priority: 0},
		}
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		return []Action{
			{Command: "", Short: "Retry", Description: "The operation timed out — check connectivity and try again", Priority: 0},
		}
	case strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no route"):
		return []Action{
			{Command: "sirsi isis network", Short: "Network audit", Description: "Run a network security audit to diagnose connectivity", Priority: 0},
		}
	}

	// Deity-specific error fallbacks.
	switch ctx.Deity {
	case "anubis":
		return []Action{
			{Command: "sirsi doctor", Short: "Run doctor", Description: "System health diagnostic", Priority: 0},
			{Command: "sirsi scan", Short: "Re-scan", Description: "Try a fresh scan", Priority: 1},
		}
	case "ra":
		return []Action{
			{Command: "sirsi ra status", Short: "Check status", Description: "Check orchestrator state", Priority: 0},
			{Command: "sirsi ra health", Short: "Health check", Description: "Health check all repos", Priority: 1},
		}
	case "thoth":
		return []Action{
			{Command: "sirsi thoth status", Short: "Check status", Description: "Check memory system health", Priority: 0},
			{Command: "sirsi thoth init", Short: "Re-initialize", Description: "Re-initialize .thoth/ if corrupted", Priority: 1},
		}
	case "maat":
		return []Action{
			{Command: "sirsi maat pulse", Short: "Quick pulse", Description: "Try a quick pulse check instead", Priority: 0},
		}
	case "seshat":
		return []Action{
			{Command: "sirsi seshat adapters", Short: "List adapters", Description: "Check available adapters", Priority: 0},
		}
	case "horus":
		return []Action{
			{Command: "sirsi horus scan", Short: "Rebuild graph", Description: "Rebuild the code graph", Priority: 0},
		}
	default:
		return []Action{
			{Command: "sirsi doctor", Short: "Run doctor", Description: "System health diagnostic", Priority: 0},
		}
	}
}

// Placeholder returns a concise input-bar hint for the given context.
// Used by the TUI as the textinput placeholder text.
func Placeholder(ctx Context) string {
	if ctx.Err != nil {
		return "doctor · help  (diagnose or see all commands)"
	}

	switch ctx.Deity {
	case "anubis":
		switch ctx.Subcommand {
		case "weigh", "scan":
			return "findings · clean · judge  (or type a category like dev, ai, cloud)"
		case "judge", "clean":
			return "findings · scan  (verify cleanup with a fresh scan)"
		case "ka", "ghosts":
			return "findings · clean  (remove ghost residuals)"
		case "mirror", "duplicates":
			return "scan · clean  (full scan or reclaim space)"
		case "apps":
			return "ghosts · scan · clean  (check residuals or scan waste)"
		}
	case "isis":
		switch ctx.Subcommand {
		case "network":
			return "heal · doctor  (remediate issues or full health check)"
		default:
			return "isis network · doctor · heal"
		}
	case "maat":
		switch ctx.Subcommand {
		case "audit":
			return "maat pulse · heal  (quick summary or auto-remediate)"
		case "pulse":
			return "maat audit · heal  (full audit or auto-remediate)"
		case "heal":
			return "maat audit · maat pulse  (verify fixes)"
		default:
			return "maat audit · maat pulse · heal"
		}
	case "ra":
		switch ctx.Subcommand {
		case "deploy":
			return "ra status · ra health  (check progress or health)"
		case "status":
			return "ra deploy · ra health · ra test"
		case "health":
			return "ra deploy · ra test · ra lint"
		case "test", "lint":
			return "ra status · heal  (check results or auto-remediate)"
		default:
			return "ra status · ra deploy · ra health"
		}
	case "net":
		switch ctx.Subcommand {
		case "align":
			return "net status · maat audit  (check alignment or run QA)"
		default:
			return "net align · maat audit"
		}
	case "thoth":
		switch ctx.Subcommand {
		case "sync":
			return "thoth compact · maat audit  (persist state or check quality)"
		case "compact":
			return "sirsi thoth sync · sirsi risk  (sync memory or check risk)"
		case "init":
			return "sirsi thoth sync  (populate memory from source files)"
		default:
			return "sirsi thoth sync · thoth compact"
		}
	case "seshat":
		switch ctx.Subcommand {
		case "ingest":
			return "seshat list · seshat export  (review or export knowledge)"
		case "notebooklm":
			return "seshat ingest · seshat list"
		default:
			return "seshat ingest · seshat list · seshat export"
		}
	case "seba":
		switch ctx.Subcommand {
		case "hardware":
			return "seba diagram · seba scan · scan  (visualize, map, or scan waste)"
		case "diagram":
			return "seba scan · seba hardware"
		case "fleet":
			return "seba diagram · isis network  (visualize fleet or audit network)"
		default:
			return "seba scan · seba diagram · seba hardware"
		}
	case "osiris":
		switch ctx.Subcommand {
		case "assess", "risk":
			return "osiris status · sirsi thoth sync  (quick status or sync memory)"
		default:
			return "sirsi risk · sirsi thoth sync"
		}
	case "horus":
		switch ctx.Subcommand {
		case "scan":
			return "horus outline · horus symbols · horus stats"
		default:
			return "horus scan · horus outline · horus stats"
		}
	}
	return "What next?"
}

// Commands returns just the command strings for tab-cycling.
// This is a convenience wrapper over After().
func Commands(ctx Context) []string {
	actions := After(ctx)
	cmds := make([]string, 0, len(actions))
	for _, a := range actions {
		if a.Command != "" {
			// Strip "sirsi " prefix for TUI input bar (TUI dispatches internally).
			cmd := strings.TrimPrefix(a.Command, "sirsi ")
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

// ── Deity-specific suggestion sets ──────────────────────────────────

func afterAnubis(ctx Context) []Action {
	switch ctx.Subcommand {
	case "weigh", "scan":
		actions := []Action{
			{Command: "sirsi clean", Short: "Review & clean", Description: "Review and clean safe items (dry-run by default)", Priority: 0},
			{Command: "sirsi clean --confirm", Short: "Apply cleanup", Description: "Apply cleanup (moves to Trash)", Priority: 1},
			{Command: "sirsi ghosts", Short: "Hunt ghosts", Description: "Hunt ghost app residuals", Priority: 2},
		}
		if ctx.FindingsCount > 0 {
			// Prepend findings drill-down when there are results.
			actions = append([]Action{
				{Command: "findings", Short: "View findings", Description: "View full breakdown with advisories", Priority: 0},
			}, actions...)
		}
		return actions
	case "judge", "clean":
		return []Action{
			{Command: "sirsi scan", Short: "Re-scan", Description: "Run a fresh scan to verify cleanup", Priority: 0},
			{Command: "sirsi ghosts", Short: "Hunt ghosts", Description: "Hunt remaining ghost app residuals", Priority: 1},
		}
	case "ka", "ghosts":
		return []Action{
			{Command: "findings", Short: "View findings", Description: "View all findings including ghosts", Priority: 0},
			{Command: "sirsi clean", Short: "Clean up", Description: "Remove ghost residuals", Priority: 1},
		}
	case "mirror", "duplicates":
		return []Action{
			{Command: "sirsi scan", Short: "Full scan", Description: "Run full waste scan", Priority: 0},
		}
	case "apps":
		return []Action{
			{Command: "sirsi ghosts", Short: "Hunt ghosts", Description: "Deep ghost residual scan", Priority: 0},
			{Command: "sirsi scan", Short: "Full scan", Description: "Run full waste scan", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi scan", Short: "Scan", Description: "Scan for infrastructure waste", Priority: 0},
		}
	}
}

func afterIsis(ctx Context) []Action {
	switch ctx.Subcommand {
	case "network":
		return []Action{
			{Command: "sirsi maat heal", Short: "Auto-heal", Description: "Auto-remediate failed checks", Priority: 0},
			{Command: "sirsi diagnose", Short: "Full diagnostic", Description: "Full system health diagnostic", Priority: 1},
		}
	case "guard", "doctor", "isis":
		return []Action{
			{Command: "sirsi isis network", Short: "Network audit", Description: "Network security audit", Priority: 0},
			{Command: "sirsi scan", Short: "Scan", Description: "Scan for infrastructure waste", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi isis network", Short: "Network audit", Description: "Network security audit", Priority: 0},
			{Command: "sirsi doctor", Short: "Diagnostic", Description: "Full system health diagnostic", Priority: 1},
		}
	}
}

func afterMaat(ctx Context) []Action {
	switch ctx.Subcommand {
	case "audit":
		return []Action{
			{Command: "sirsi maat pulse", Short: "Quick pulse", Description: "Quick coverage summary", Priority: 0},
			{Command: "sirsi maat heal", Short: "Auto-heal", Description: "Auto-remediate quality issues", Priority: 1},
		}
	case "pulse":
		return []Action{
			{Command: "sirsi maat audit", Short: "Full audit", Description: "Full governance assessment", Priority: 0},
			{Command: "sirsi maat heal", Short: "Auto-heal", Description: "Auto-remediate quality issues", Priority: 1},
		}
	case "heal":
		return []Action{
			{Command: "sirsi maat audit", Short: "Verify", Description: "Full audit to verify fixes", Priority: 0},
			{Command: "sirsi maat pulse", Short: "Quick check", Description: "Quick coverage summary", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi maat audit", Short: "Audit", Description: "Governance and quality scan", Priority: 0},
			{Command: "sirsi maat pulse", Short: "Pulse", Description: "Quick coverage summary", Priority: 1},
		}
	}
}

func afterRa(ctx Context) []Action {
	switch ctx.Subcommand {
	case "deploy":
		return []Action{
			{Command: "sirsi ra status", Short: "Check progress", Description: "Check deployment progress", Priority: 0},
			{Command: "sirsi ra health", Short: "Health check", Description: "Health check across all repos", Priority: 1},
			{Command: "sirsi ra collect", Short: "Collect logs", Description: "Collect logs from agents", Priority: 2},
		}
	case "status":
		return []Action{
			{Command: "sirsi ra deploy", Short: "Deploy", Description: "Deploy a task to repos", Priority: 0},
			{Command: "sirsi ra health", Short: "Health check", Description: "Health check across all repos", Priority: 1},
			{Command: "sirsi ra test", Short: "Run tests", Description: "Run tests across all repos", Priority: 2},
		}
	case "health":
		return []Action{
			{Command: "sirsi ra deploy", Short: "Deploy", Description: "Deploy a task to repos", Priority: 0},
			{Command: "sirsi ra test", Short: "Run tests", Description: "Run tests across all repos", Priority: 1},
			{Command: "sirsi maat heal", Short: "Auto-heal", Description: "Auto-remediate failures", Priority: 2},
		}
	case "test", "lint":
		return []Action{
			{Command: "sirsi ra status", Short: "Check status", Description: "Check overall repo status", Priority: 0},
			{Command: "sirsi maat heal", Short: "Auto-heal", Description: "Auto-remediate failures", Priority: 1},
		}
	case "kill":
		return []Action{
			{Command: "sirsi ra status", Short: "Check status", Description: "Check which windows are still running", Priority: 0},
			{Command: "sirsi ra deploy", Short: "Redeploy", Description: "Deploy scopes again", Priority: 1},
		}
	case "collect":
		return []Action{
			{Command: "sirsi ra status", Short: "Check status", Description: "Check fleet status", Priority: 0},
			{Command: "sirsi ra deploy", Short: "Deploy", Description: "Deploy a new task", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi ra status", Short: "Status", Description: "Check orchestrator status", Priority: 0},
			{Command: "sirsi ra deploy", Short: "Deploy", Description: "Deploy a task to repos", Priority: 1},
		}
	}
}

func afterNet(ctx Context) []Action {
	switch ctx.Subcommand {
	case "align":
		return []Action{
			{Command: "sirsi net status", Short: "Check score", Description: "Check plan alignment score", Priority: 0},
			{Command: "sirsi maat audit", Short: "QA check", Description: "Run governance quality check", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi net align", Short: "Align check", Description: "Validate cross-module consistency", Priority: 0},
			{Command: "sirsi maat audit", Short: "QA check", Description: "Run governance quality check", Priority: 1},
		}
	}
}

func afterThoth(ctx Context) []Action {
	switch ctx.Subcommand {
	case "sync":
		return []Action{
			{Command: "sirsi thoth compact", Short: "Compact", Description: "Persist state before context compression", Priority: 0},
			{Command: "sirsi sirsi risk", Short: "Risk check", Description: "Check uncommitted work risk", Priority: 1},
			{Command: "sirsi maat audit", Short: "QA check", Description: "Run quality assessment", Priority: 2},
		}
	case "compact":
		return []Action{
			{Command: "sirsi sirsi thoth sync", Short: "Sync memory", Description: "Sync memory from source files", Priority: 0},
			{Command: "sirsi sirsi risk", Short: "Risk check", Description: "Check uncommitted work risk", Priority: 1},
		}
	case "init":
		return []Action{
			{Command: "sirsi sirsi thoth sync", Short: "Sync", Description: "Populate memory from source + git history", Priority: 0},
			{Command: "sirsi scan", Short: "Scan", Description: "Scan for infrastructure waste", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi sirsi thoth sync", Short: "Sync", Description: "Sync project memory", Priority: 0},
			{Command: "sirsi thoth compact", Short: "Compact", Description: "Persist state for continuations", Priority: 1},
		}
	}
}

func afterSeshat(ctx Context) []Action {
	switch ctx.Subcommand {
	case "ingest":
		return []Action{
			{Command: "sirsi seshat list", Short: "Browse", Description: "Browse ingested knowledge items", Priority: 0},
			{Command: "sirsi seshat export", Short: "Export", Description: "Export knowledge to a target", Priority: 1},
			{Command: "sirsi seshat notebooklm", Short: "NotebookLM", Description: "Export to Google NotebookLM", Priority: 2},
		}
	case "notebooklm":
		return []Action{
			{Command: "sirsi seshat ingest", Short: "Ingest more", Description: "Ingest from more sources", Priority: 0},
			{Command: "sirsi seshat list", Short: "Browse", Description: "Browse knowledge items", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi seshat ingest", Short: "Ingest", Description: "Ingest knowledge from sources", Priority: 0},
			{Command: "sirsi seshat list", Short: "Browse", Description: "Browse knowledge items", Priority: 1},
		}
	}
}

func afterSeba(ctx Context) []Action {
	switch ctx.Subcommand {
	case "hardware":
		return []Action{
			{Command: "sirsi seba diagram", Short: "Diagram", Description: "Generate architecture diagram from scan", Priority: 0},
			{Command: "sirsi seba scan", Short: "Topology", Description: "Full infrastructure topology map", Priority: 1},
			{Command: "sirsi scan", Short: "Waste scan", Description: "Scan for infrastructure waste", Priority: 2},
		}
	case "diagram":
		return []Action{
			{Command: "sirsi seba scan", Short: "Topology", Description: "Full infrastructure topology map", Priority: 0},
			{Command: "sirsi seba hardware", Short: "Hardware", Description: "Hardware and accelerator profile", Priority: 1},
		}
	case "fleet":
		return []Action{
			{Command: "sirsi seba diagram", Short: "Diagram", Description: "Visualize fleet architecture", Priority: 0},
			{Command: "sirsi isis network", Short: "Network audit", Description: "Network security audit", Priority: 1},
		}
	default:
		return []Action{
			{Command: "sirsi seba diagram", Short: "Diagram", Description: "Generate architecture diagram", Priority: 0},
			{Command: "sirsi seba hardware", Short: "Hardware", Description: "Hardware and accelerator profile", Priority: 1},
		}
	}
}

func afterOsiris(ctx Context) []Action {
	switch ctx.Subcommand {
	case "assess", "risk":
		return []Action{
			{Command: "sirsi osiris status", Short: "Quick status", Description: "One-line risk summary", Priority: 0},
			{Command: "sirsi sirsi thoth sync", Short: "Sync memory", Description: "Sync memory before committing", Priority: 1},
			{Command: "sirsi scan", Short: "Scan", Description: "Scan for infrastructure waste", Priority: 2},
		}
	default:
		return []Action{
			{Command: "sirsi sirsi risk", Short: "Full assess", Description: "Full checkpoint assessment", Priority: 0},
			{Command: "sirsi sirsi thoth sync", Short: "Sync memory", Description: "Sync project memory", Priority: 1},
		}
	}
}

func afterHorus(ctx Context) []Action {
	switch ctx.Subcommand {
	case "scan":
		return []Action{
			{Command: "sirsi horus outline", Short: "Outline", Description: "Print file declaration outline", Priority: 0},
			{Command: "sirsi horus symbols", Short: "Symbols", Description: "Search symbols in the graph", Priority: 1},
			{Command: "sirsi horus stats", Short: "Stats", Description: "Graph statistics", Priority: 2},
		}
	default:
		return []Action{
			{Command: "sirsi horus scan", Short: "Build graph", Description: "Build code symbol graph", Priority: 0},
			{Command: "sirsi horus stats", Short: "Stats", Description: "Graph statistics", Priority: 1},
		}
	}
}
