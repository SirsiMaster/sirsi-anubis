package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-anubis/internal/output"
	"github.com/SirsiMaster/sirsi-anubis/internal/sight"
)

var (
	sightFix     bool
	sightDryRun  bool
	sightConfirm bool
	sightReindex bool
)

var sightCmd = &cobra.Command{
	Use:   "sight",
	Short: "👁️ Fix ghost apps in Spotlight and Launch Services",
	Long: `👁️ Sight — The All-Seeing Eye

Detect and remove phantom app registrations from macOS Launch Services.
Ghost apps pollute Spotlight search results and waste index space.

  anubis sight                Scan for ghost app registrations
  anubis sight --fix          Rebuild Launch Services database (removes all ghosts)
  anubis sight --reindex      Trigger Spotlight re-index after cleanup

Safety: --fix requires --dry-run or --confirm flag.
        Rebuilding Launch Services resets ALL file associations.`,
	Run: runSight,
}

func init() {
	sightCmd.Flags().BoolVar(&sightFix, "fix", false, "Rebuild Launch Services database to remove ghosts")
	sightCmd.Flags().BoolVar(&sightDryRun, "dry-run", false, "Show what would be fixed without making changes")
	sightCmd.Flags().BoolVar(&sightConfirm, "confirm", false, "Actually rebuild Launch Services (required for --fix)")
	sightCmd.Flags().BoolVar(&sightReindex, "reindex", false, "Trigger Spotlight re-index after fix")
}

func runSight(cmd *cobra.Command, args []string) {
	// Scan for ghosts
	result, err := sight.Scan()
	if err != nil {
		output.Error(fmt.Sprintf("Sight scan failed: %v", err))
		os.Exit(1)
	}

	// If --fix is requested
	if sightFix {
		if !sightDryRun && !sightConfirm {
			output.Error("--fix requires --dry-run or --confirm flag")
			output.Warn("Try: anubis sight --fix --dry-run")
			os.Exit(1)
		}

		isDryRun := sightDryRun || !sightConfirm
		if err := sight.Fix(isDryRun); err != nil {
			output.Error(fmt.Sprintf("Sight fix failed: %v", err))
			os.Exit(1)
		}

		if sightReindex {
			if err := sight.ReindexSpotlight(isDryRun); err != nil {
				output.Warn(fmt.Sprintf("Spotlight reindex failed (may need sudo): %v", err))
			}
		}

		if isDryRun {
			output.Header("👁️ Sight — Fix [DRY RUN]")
			fmt.Println()
			output.Info(fmt.Sprintf("Would rebuild Launch Services database"))
			output.Info(fmt.Sprintf("Would remove %d ghost app registrations", result.TotalGhosts))
			if sightReindex {
				output.Info("Would trigger Spotlight re-index")
			}
			fmt.Println()
			output.Warn("To actually fix: anubis sight --fix --confirm")
		} else {
			output.Header("👁️ Sight — Fix Complete")
			fmt.Println()
			output.Info("✅ Launch Services database rebuilt")
			output.Info(fmt.Sprintf("Removed %d ghost app registrations", result.TotalGhosts))
			if sightReindex {
				output.Info("✅ Spotlight re-index triggered")
			}
		}
		fmt.Println()
		return
	}

	// Default: show ghost scan results
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return
	}

	renderSightResult(result)
}

func renderSightResult(result *sight.SightResult) {
	output.Header("👁️ Sight — Ghost App Scan")
	fmt.Println()

	if result.TotalGhosts == 0 {
		output.Info("✅ No ghost app registrations found in Launch Services")
		fmt.Println()
		return
	}

	output.Warn(fmt.Sprintf("🔍 Found %d ghost app registrations in Launch Services", result.TotalGhosts))
	fmt.Println()

	for _, g := range result.GhostRegistrations {
		name := g.Name
		if len(name) > 35 {
			name = name[:32] + "..."
		}
		fmt.Printf("  👻 %-35s  %s\n", name, g.BundleID)
		fmt.Printf("     Missing: %s\n", g.Path)
		fmt.Println()
	}

	output.Info("To fix: anubis sight --fix --dry-run")
	fmt.Println()
}
