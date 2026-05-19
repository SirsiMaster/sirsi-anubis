package output

// ── Tab Definitions ──────────────────────────────────────────────────
// Five tabs, like Mole's five planets. Each has a purpose, a voice,
// and numbered actions. The user never types a command — they press
// a number.

// tabAction defines one action on a tab. Native functions return a
// nativeResult carrying rendered lines, deity key, fix cmds, and
// optional select/analyze state. fixCmds override the suggest engine.
type tabAction struct {
	Label  string
	Desc   string
	Args   []string              // CLI args (fallback if Native is nil)
	Native func() nativeResult   // returns rendered lines, deity key, fix cmds, select/analyze state
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
var nativeCommands = map[string]func() nativeResult{
	// Anubis — scan & clean
	"anubis scan":            nativeScan,
	"anubis ghosts":          nativeGhosts,
	"anubis clean --dry-run": nativeCleanDryRun,
	"anubis clean --confirm": nativeCleanConfirm,
	"anubis duplicates":      nativeMirror,
	// Isis — health
	"isis diagnose": nativeDoctor,
	"isis network":  nativeNetworkAudit,
	"isis fix":      nativeNetworkFix,
	// Quality & Intel
	"seba hardware": nativeHardware,
	"osiris risk":   nativeRisk,
	// Aliases for convenience
	"scan":     nativeScan,
	"findings": nativeFindings,
}


