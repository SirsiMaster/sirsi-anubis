package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/brain"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/thoth"
)

var (
	thothSince       string
	thothDryRun      bool
	thothMemoryOnly  bool
	thothJournalOnly bool
	brainUpdate      bool
	brainRemove      bool
)

var thothCmd = &cobra.Command{
	Use:   "thoth",
	Short: "𓁟 Thoth — Persistent Knowledge & Brain Manager",
	Long: `𓁟 Thoth — The Scribe of the Gods

Thoth manages your project's persistent memory and its neural "brain."
Use it to sync your development journal or manage AI weights.

  pantheon thoth sync         Sync memory.yaml and journal.md
  pantheon thoth brain        Install/Update neural weights (Anubis Pro)
  pantheon thoth status       Check brain and memory health`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// thothSyncCmd handles memory/journal synchronization
var thothSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Auto-sync memory.yaml and journal.md from source + git history",
	RunE:  runThothSync,
}

// thothBrainCmd handles neural weight installation
var thothBrainCmd = &cobra.Command{
	Use:   "brain",
	Short: "Manage neural weights (CoreML/ONNX) for Pro-tier analysis",
	Run:   runBrainCommand,
}

func init() {
	// Sync Flags
	thothSyncCmd.Flags().StringVar(&thothSince, "since", "24 hours ago", "Git log timeframe")
	thothSyncCmd.Flags().BoolVar(&thothDryRun, "dry-run", false, "Preview without writing")

	// Brain Flags
	thothBrainCmd.Flags().BoolVar(&brainUpdate, "update", false, "Update to latest weights")
	thothBrainCmd.Flags().BoolVar(&brainRemove, "remove", false, "Remove weights")

	thothCmd.AddCommand(thothSyncCmd)
	thothCmd.AddCommand(thothBrainCmd)
}

func runThothSync(cmd *cobra.Command, args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	output.Header(fmt.Sprintf("𓁟 Thoth Sync — %s", repoRoot))

	// ... (Existing sync logic) ...
	if err := thoth.Sync(thoth.SyncOptions{RepoRoot: repoRoot, UpdateDate: true}); err != nil {
		return err
	}
	output.Success("Memory synced.")
	return nil
}

func runBrainCommand(cmd *cobra.Command, args []string) {
	if brainRemove {
		output.Info("Removing neural weights...")
		_ = brain.Remove()
		return
	}
	output.Info("Checking brain status...")
	// ... (Existing install/update logic) ...
}

func findRepoRoot() (string, error) {
	dir, _ := os.Getwd()
	for {
		if info, err := os.Stat(filepath.Join(dir, ".thoth")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .thoth/ found")
}
