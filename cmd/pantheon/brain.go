package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/brain"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

var (
	brainRemove bool
	brainUpdate bool
)

var installBrainCmd = &cobra.Command{
	Use:   "install-brain",
	Short: "🧠 Download neural weights for Anubis Pro",
	Long: `🧠 Install Brain — Neural Weight Manager

Downloads on-demand CoreML/ONNX model weights to ~/.pantheon/weights/.
The base Anubis binary ships without models — install-brain adds
neural capabilities for semantic file classification.

  pantheon install-brain             Install default model
  pantheon install-brain --update    Fetch latest model version
  pantheon install-brain --remove    Delete installed weights

Pro tier feature: Enables semantic search, neural dedup, and
context sanitization for AI development environments.

Size budget: < 100 MB quantized model.`,
	Run: runInstallBrain,
}

var uninstallBrainCmd = &cobra.Command{
	Use:   "uninstall-brain",
	Short: "🧠 Remove neural weights",
	Long: `🧠 Uninstall Brain — Remove Neural Weights

Completely removes all downloaded model weights from ~/.pantheon/weights/.
The base Anubis CLI continues to work without neural capabilities.`,
	Run: func(cmd *cobra.Command, args []string) {
		brainRemove = true
		runInstallBrain(cmd, args)
	},
}

func init() {
	installBrainCmd.Flags().BoolVar(&brainRemove, "remove", false, "Remove installed brain weights")
	installBrainCmd.Flags().BoolVar(&brainUpdate, "update", false, "Check for and install latest model")
}

func runInstallBrain(cmd *cobra.Command, args []string) {
	// Remove mode
	if brainRemove {
		runBrainRemove()
		return
	}

	// Update mode
	if brainUpdate {
		runBrainUpdate()
		return
	}

	// Default: install
	runBrainInstall()
}

func runBrainInstall() {
	output.Header("🧠 Install Brain — Neural Weights")
	fmt.Println()

	// Check if already installed
	if brain.IsInstalled() {
		status, err := brain.GetStatus()
		if err == nil && status.Model != nil {
			if JsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				_ = enc.Encode(status)
				return
			}
			output.Info(fmt.Sprintf("Brain already installed: %s v%s", status.Model.InstalledModel, status.Model.Version))
			output.Info(fmt.Sprintf("Size: %s", brain.FormatBytes(status.Model.SizeBytes)))
			output.Info(fmt.Sprintf("Installed: %s", status.Model.InstalledAt.Format(time.RFC3339)))
			fmt.Println()
			output.Info("Use --update to check for newer version")
			output.Info("Use --remove to delete weights")
			return
		}
	}

	output.Info("Fetching brain manifest from GitHub...")
	fmt.Println()

	// Progress tracking
	var lastPercent int
	onProgress := func(downloaded, total int64) {
		if quietMode {
			return
		}
		if total <= 0 {
			return
		}
		percent := int(float64(downloaded) / float64(total) * 100)
		if percent != lastPercent && percent%5 == 0 {
			lastPercent = percent
			bar := renderProgressBar(percent)
			fmt.Fprintf(os.Stderr, "\r  %s %s/%s (%d%%)",
				bar,
				brain.FormatBytes(downloaded),
				brain.FormatBytes(total),
				percent,
			)
		}
	}

	local, err := brain.Install(onProgress)
	if err != nil {
		if JsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(map[string]string{"error": err.Error()})
			return
		}

		// Handle "no models available" gracefully — expected before first release
		if strings.Contains(err.Error(), "HTTP") || strings.Contains(err.Error(), "manifest") {
			fmt.Println()
			output.Warn("Neural brain models are not yet published.")
			output.Info("Models will be available in a future release.")
			output.Info("The Anubis CLI works fully without neural weights.")
			fmt.Println()
			output.Dim("Anubis Free → scan, clean, guard, ghost hunt")
			output.Dim("Anubis Pro  → + neural classification, semantic search")
			fmt.Println()
			printProUpsell()
			return
		}

		output.Error(fmt.Sprintf("Install failed: %v", err))
		os.Exit(1)
	}

	if JsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(local)
		return
	}

	fmt.Println() // Clear progress line
	fmt.Println()
	output.Success(fmt.Sprintf("Brain installed: %s v%s", local.InstalledModel, local.Version))
	output.Info(fmt.Sprintf("Format:    %s", local.Format))
	output.Info(fmt.Sprintf("Size:      %s", brain.FormatBytes(local.SizeBytes)))
	output.Info(fmt.Sprintf("Location:  %s", weightsLocation()))
	output.Info(fmt.Sprintf("Checksum:  %s", truncateHash(local.SHA256)))
	fmt.Println()
	printProUpsell()
}

func runBrainUpdate() {
	output.Header("🧠 Update Brain — Check for Latest Model")
	fmt.Println()

	output.Info("Checking for updates...")

	var lastPercent int
	onProgress := func(downloaded, total int64) {
		if quietMode {
			return
		}
		if total <= 0 {
			return
		}
		percent := int(float64(downloaded) / float64(total) * 100)
		if percent != lastPercent && percent%5 == 0 {
			lastPercent = percent
			bar := renderProgressBar(percent)
			fmt.Fprintf(os.Stderr, "\r  %s %s/%s (%d%%)",
				bar,
				brain.FormatBytes(downloaded),
				brain.FormatBytes(total),
				percent,
			)
		}
	}

	local, updated, err := brain.Update(onProgress)
	if err != nil {
		if JsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(map[string]interface{}{"error": err.Error(), "updated": false})
			return
		}

		if strings.Contains(err.Error(), "HTTP") || strings.Contains(err.Error(), "manifest") {
			output.Warn("Unable to check for updates (manifest not published yet).")
			output.Info("Models will be available in a future release.")
			return
		}

		output.Error(fmt.Sprintf("Update check failed: %v", err))
		os.Exit(1)
	}

	if JsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]interface{}{
			"updated": updated,
			"model":   local,
		})
		return
	}

	fmt.Println() // Clear progress line
	if updated {
		output.Success(fmt.Sprintf("Brain updated to %s v%s", local.InstalledModel, local.Version))
		output.Info(fmt.Sprintf("Size: %s", brain.FormatBytes(local.SizeBytes)))
	} else {
		output.Success("Brain is already up to date")
		if local != nil {
			output.Info(fmt.Sprintf("Current: %s v%s", local.InstalledModel, local.Version))
		}
	}
}

func runBrainRemove() {
	output.Header("🧠 Remove Brain — Deleting Neural Weights")
	fmt.Println()

	if !brain.IsInstalled() {
		if JsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(map[string]interface{}{"removed": false, "reason": "not installed"})
			return
		}
		output.Info("No brain weights installed — nothing to remove.")
		return
	}

	// Get info before removing
	status, _ := brain.GetStatus()

	err := brain.Remove()
	if err != nil {
		if JsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(map[string]string{"error": err.Error()})
			return
		}
		output.Error(fmt.Sprintf("Remove failed: %v", err))
		os.Exit(1)
	}

	if JsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]interface{}{
			"removed":     true,
			"freed_bytes": status.Model.SizeBytes,
		})
		return
	}

	output.Success("Brain weights removed")
	if status != nil && status.Model != nil {
		output.Info(fmt.Sprintf("Freed: %s", brain.FormatBytes(status.Model.SizeBytes)))
	}
	output.Info(fmt.Sprintf("Deleted: %s", weightsLocation()))
	fmt.Println()
	output.Dim("Anubis CLI continues to work fully without neural weights.")
	output.Dim("Run 'pantheon install-brain' to re-install anytime.")
}

// renderProgressBar creates an ASCII progress bar.
func renderProgressBar(percent int) string {
	const width = 30
	filled := width * percent / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s]", bar)
}

// truncateHash shows first 12 chars of a hash for display.
func truncateHash(hash string) string {
	if len(hash) > 12 {
		return hash[:12] + "..."
	}
	return hash
}

// weightsLocation returns the weights directory path for display.
func weightsLocation() string {
	dir, err := brain.WeightsDir()
	if err != nil {
		return "~/.pantheon/weights/"
	}
	return dir
}

// printProUpsell prints the Anubis Pro upsell footer.
func printProUpsell() {
	output.Dim("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	output.Dim("𓂀 Anubis Pro — Neural classification, semantic search,")
	output.Dim("  context sanitization for AI development environments.")
	output.Dim("  https://github.com/SirsiMaster/sirsi-pantheon")
	output.Dim("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
