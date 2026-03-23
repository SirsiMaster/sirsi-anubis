package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/horus"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal/rules"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

// Weigh flags
var (
	weighDev     bool
	weighAI      bool
	weighVMs     bool
	weighIDEs    bool
	weighCloud   bool
	weighStorage bool
	weighAll     bool
	weighFresh   bool
)

// weighCmd implements `pantheon weigh` — the scanning command.
var weighCmd = &cobra.Command{
	Use:   "weigh",
	Short: "𓂀 Scan your workstation (The Weighing)",
	Long: `Weigh your workstation against Ma'at's feather of truth.

Discovers infrastructure waste across your machine: stale caches,
orphaned build artifacts, unused dependencies, and more.

This command is READ-ONLY — it never deletes anything.
Use 'pantheon judge' to clean what was found.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWeigh()
	},
}

func init() {
	weighCmd.Flags().BoolVar(&weighDev, "dev", false, "Scan developer frameworks (Node, Rust, Go, Python)")
	weighCmd.Flags().BoolVar(&weighAI, "ai", false, "Scan AI/ML caches (MLX, CUDA, HuggingFace, Ollama)")
	weighCmd.Flags().BoolVar(&weighVMs, "vms", false, "Scan virtualization (Parallels, Docker, VMware)")
	weighCmd.Flags().BoolVar(&weighIDEs, "ides", false, "Scan IDEs (Xcode, VS Code, JetBrains)")
	weighCmd.Flags().BoolVar(&weighCloud, "cloud", false, "Scan cloud/infra (K8s, Terraform, gcloud)")
	weighCmd.Flags().BoolVar(&weighStorage, "storage", false, "Scan cloud storage (OneDrive, GDrive, iCloud)")
	weighCmd.Flags().BoolVar(&weighAll, "all", false, "Scan all categories")
	weighCmd.Flags().BoolVar(&weighFresh, "fresh", false, "Force fresh filesystem index (ignore Horus cache)")
}

func runWeigh() error {
	start := time.Now()

	if !quietMode {
		output.Banner()
		output.Header("THE WEIGHING — Scanning Your Machine")
	}

	// Determine categories to scan
	categories := buildCategories()

	if !quietMode {
		if len(categories) == 0 {
			output.Info("Scanning all categories...")
		} else {
			names := make([]string, len(categories))
			for i, c := range categories {
				names[i] = string(c)
			}
			output.Info("Scanning: %s", strings.Join(names, ", "))
		}
		fmt.Fprintln(os.Stderr)
	}

	// Build Horus index (shared filesystem manifest)
	if !quietMode {
		output.Dim("  👁️ Building Horus index...")
	}
	manifest, err := horus.Index(horus.IndexOptions{
		ForceRefresh: weighFresh,
	})
	if err != nil {
		// Non-fatal: fall back to per-rule filesystem walks
		if !quietMode {
			output.Warn("Horus index unavailable, using direct filesystem scan: %v", err)
		}
	}
	if manifest != nil && !quietMode {
		output.Dim("  %s", manifest.Summary())
		fmt.Fprintln(os.Stderr)
	}

	// Create engine and register all built-in rules
	engine := jackal.NewEngine()
	engine.RegisterAll(rules.AllRules()...)

	// Run scan with Horus index
	ctx := context.Background()
	opts := jackal.ScanOptions{
		Categories: categories,
		Manifest:   manifest,
	}

	result, err := engine.Scan(ctx, opts)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// JSON output mode
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Terminal output
	if len(result.Findings) == 0 {
		output.Success("Your machine is clean! Nothing found.")
		return nil
	}

	// Group findings by category for display
	byCat := make(map[jackal.Category][]jackal.Finding)
	for _, f := range result.Findings {
		byCat[f.Category] = append(byCat[f.Category], f)
	}

	// Display findings by category
	catOrder := []jackal.Category{
		jackal.CategoryGeneral,
		jackal.CategoryDev,
		jackal.CategoryVirtualization,
		jackal.CategoryAI,
		jackal.CategoryIDEs,
		jackal.CategoryCloud,
		jackal.CategoryStorage,
	}

	for _, cat := range catOrder {
		findings, ok := byCat[cat]
		if !ok || len(findings) == 0 {
			continue
		}

		catName := categoryDisplayName(cat)
		output.Header(catName)

		for _, f := range findings {
			output.Info("%s", output.FindingRow(
				f.Description,
				shortenPath(f.Path),
				jackal.FormatSize(f.SizeBytes),
				string(f.Severity),
			))
		}

		// Category subtotal
		catSummary := result.ByCategory[cat]
		output.Dim("  Subtotal: %s (%d items)",
			jackal.FormatSize(catSummary.TotalSize),
			catSummary.Findings,
		)
	}

	// Errors (non-fatal)
	if len(result.Errors) > 0 {
		output.Header("Scan Errors")
		for _, re := range result.Errors {
			output.Warn("%s: %s", re.RuleName, re.Err)
		}
	}

	// Summary
	elapsed := time.Since(start)
	output.Summary(
		jackal.FormatSize(result.TotalSize),
		len(result.Findings),
		result.RulesRan,
	)
	output.Dim("  Scanned in %s", elapsed.Round(time.Millisecond))
	fmt.Fprintln(os.Stderr)
	output.Info("Run %s to clean these artifacts.",
		output.SizeStyle.Render("pantheon judge --dry-run"))

	return nil
}

func buildCategories() []jackal.Category {
	if weighAll {
		return nil // nil = all categories
	}

	var cats []jackal.Category
	if weighDev {
		cats = append(cats, jackal.CategoryDev)
	}
	if weighAI {
		cats = append(cats, jackal.CategoryAI)
	}
	if weighVMs {
		cats = append(cats, jackal.CategoryVirtualization)
	}
	if weighIDEs {
		cats = append(cats, jackal.CategoryIDEs)
	}
	if weighCloud {
		cats = append(cats, jackal.CategoryCloud)
	}
	if weighStorage {
		cats = append(cats, jackal.CategoryStorage)
	}

	// If no specific flag, scan everything
	if len(cats) == 0 {
		return nil
	}
	return cats
}

func categoryDisplayName(cat jackal.Category) string {
	switch cat {
	case jackal.CategoryGeneral:
		return "General Mac"
	case jackal.CategoryDev:
		return "Developer Frameworks"
	case jackal.CategoryVirtualization:
		return "Virtualization"
	case jackal.CategoryAI:
		return "AI / ML"
	case jackal.CategoryIDEs:
		return "IDEs & AI Tools"
	case jackal.CategoryCloud:
		return "Cloud & Infrastructure"
	case jackal.CategoryStorage:
		return "Cloud Storage"
	default:
		return string(cat)
	}
}

func shortenPath(path string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
