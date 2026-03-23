package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal/rules"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

// Judge flags
var (
	judgeDryRun  bool
	judgeConfirm bool
	judgeTrash   bool
)

// judgeCmd implements `pantheon judge` — the cleaning command.
var judgeCmd = &cobra.Command{
	Use:   "judge",
	Short: "𓂀 Clean artifacts (The Judgment)",
	Long: `Render the verdict on your machine's artifacts.

Scans first, then cleans the findings. This IS a destructive operation.
You MUST specify either --dry-run or --confirm.

  pantheon judge --dry-run      Preview what would be cleaned
  pantheon judge --confirm      Actually clean (moves to Trash by default)
  pantheon judge --confirm --permanent  Actually delete (no Trash)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runJudge()
	},
}

func init() {
	judgeCmd.Flags().BoolVar(&judgeDryRun, "dry-run", false, "Preview what would be cleaned without deleting")
	judgeCmd.Flags().BoolVar(&judgeConfirm, "confirm", false, "Actually perform the cleaning")
	judgeCmd.Flags().BoolVar(&judgeTrash, "trash", true, "Move to Trash instead of deleting (default: true)")
}

func runJudge() error {
	// Rule A1: Must specify --dry-run or --confirm
	if !judgeDryRun && !judgeConfirm {
		output.Error("You must specify --dry-run or --confirm")
		output.Info("")
		output.Info("  pantheon judge --dry-run      Preview what would be cleaned")
		output.Info("  pantheon judge --confirm      Actually clean")
		output.Info("")
		output.Dim("Safety first. Anubis never deletes without your explicit command.")
		return fmt.Errorf("missing required flag: --dry-run or --confirm")
	}

	start := time.Now()

	if !quietMode {
		output.Banner()
		if judgeDryRun {
			output.Header("THE JUDGMENT — Dry Run (Preview Only)")
		} else {
			output.Header("THE JUDGMENT — Cleaning Your Machine")
		}
	}

	// First: scan
	engine := jackal.NewEngine()
	engine.RegisterAll(rules.AllRules()...)
	ctx := context.Background()

	categories := buildCategories()
	scanOpts := jackal.ScanOptions{
		Categories: categories,
	}

	if !quietMode {
		output.Info("Scanning...")
	}

	scanResult, err := engine.Scan(ctx, scanOpts)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(scanResult.Findings) == 0 {
		output.Success("Your machine is clean! Nothing to judge.")
		return nil
	}

	if !quietMode {
		output.Info("Found %s across %d items",
			jackal.FormatSize(scanResult.TotalSize),
			len(scanResult.Findings))
		fmt.Fprintln(os.Stderr)
	}

	// Then: clean
	cleanOpts := jackal.CleanOptions{
		DryRun:   judgeDryRun,
		UseTrash: judgeTrash,
		Confirm:  judgeConfirm,
	}

	cleanResult, err := engine.Clean(ctx, scanResult.Findings, cleanOpts)
	if err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

	// Report results
	elapsed := time.Since(start)

	if judgeDryRun {
		output.Header("DRY RUN RESULTS")
		output.Info("Would clean %d items", cleanResult.Cleaned)
		output.Info("Would free  %s", jackal.FormatSize(cleanResult.BytesFreed))
		if cleanResult.Skipped > 0 {
			output.Warn("Would skip  %d items (protected or errors)", cleanResult.Skipped)
		}
		fmt.Fprintln(os.Stderr)
		output.Info("Run %s to actually clean.",
			output.SizeStyle.Render("pantheon judge --confirm"))
	} else {
		output.Header("JUDGMENT RENDERED")
		output.Success("Cleaned %d items", cleanResult.Cleaned)
		output.Success("Freed   %s", jackal.FormatSize(cleanResult.BytesFreed))
		if cleanResult.Skipped > 0 {
			output.Warn("Skipped %d items (protected or errors)", cleanResult.Skipped)
		}
	}

	// Report errors
	if len(cleanResult.Errors) > 0 {
		output.Header("Errors")
		for _, e := range cleanResult.Errors {
			output.Error("%s", e)
		}
	}

	output.Dim("  Completed in %s", elapsed.Round(time.Millisecond))

	return nil
}
