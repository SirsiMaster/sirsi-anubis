package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

// dependency represents a tool Pantheon uses.
type dependency struct {
	Name        string // binary name
	Description string // what it's for
	Required    bool   // true = Pantheon won't work without it, false = optional but recommended
	InstallCmd  string // brew/apt command to install
	CheckCmd    string // command to verify installation (empty = just check PATH)
}

// macOS dependencies.
var macDeps = []dependency{
	{Name: "git", Description: "Version control (used by Thoth, Ma'at, Ra)", Required: true, InstallCmd: "xcode-select --install"},
	{Name: "go", Description: "Go compiler (build from source, Ma'at coverage)", Required: false, InstallCmd: "brew install go"},
	{Name: "golangci-lint", Description: "Linter (Ma'at pre-push gate, matches CI)", Required: false, InstallCmd: "brew install golangci-lint"},
	{Name: "gh", Description: "GitHub CLI (Ma'at pipeline checks, Ra CI status)", Required: false, InstallCmd: "brew install gh"},
	{Name: "python3", Description: "Ra deployment agent (scope orchestration)", Required: false, InstallCmd: "brew install python3"},
}

// linuxDeps would differ (apt-get, etc.) — extend when shipping Linux.

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Check and install Pantheon dependencies",
	Long: `Check which tools Pantheon depends on and install any that are missing.

  sirsi setup            Check all dependencies
  sirsi setup --install  Install missing dependencies automatically
  sirsi setup --json     Machine-readable output`,
	RunE: runSetup,
}

var (
	setupInstall bool
	setupJSON    bool
)

func init() {
	setupCmd.Flags().BoolVar(&setupInstall, "install", false, "Install missing dependencies automatically")
	setupCmd.Flags().BoolVar(&setupJSON, "json", false, "Output as JSON")
}

type depStatus struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
	Version     string `json:"version,omitempty"`
	Required    bool   `json:"required"`
	InstallCmd  string `json:"install_cmd"`
}

func runSetup(_ *cobra.Command, _ []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("setup currently supports macOS only (detected %s)", runtime.GOOS)
	}

	deps := macDeps
	var statuses []depStatus
	missing := 0

	for _, d := range deps {
		s := depStatus{
			Name:        d.Name,
			Description: d.Description,
			Required:    d.Required,
			InstallCmd:  d.InstallCmd,
		}

		path, err := exec.LookPath(d.Name)
		if err == nil {
			s.Installed = true
			// Try to get version.
			if ver, err := exec.Command(path, "--version").Output(); err == nil {
				firstLine := strings.Split(strings.TrimSpace(string(ver)), "\n")[0]
				if len(firstLine) < 80 {
					s.Version = firstLine
				}
			}
		} else {
			missing++
		}

		statuses = append(statuses, s)
	}

	if setupJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(statuses)
	}

	output.Banner()
	fmt.Println("Pantheon Dependency Check")
	fmt.Println()

	var rows [][]string
	for _, s := range statuses {
		icon := "✅"
		status := "installed"
		if !s.Installed {
			if s.Required {
				icon = "❌"
				status = "MISSING (required)"
			} else {
				icon = "⚠️"
				status = "not installed"
			}
		}
		_ = s.Version // version available for --json output
		rows = append(rows, []string{icon, s.Name, s.Description, status})
	}
	output.Table([]string{"", "Tool", "Purpose", "Status"}, rows)

	if missing == 0 {
		fmt.Println("\n  All dependencies satisfied.")
		return nil
	}

	fmt.Printf("\n  %d missing dependency(ies)\n", missing)

	if !setupInstall {
		fmt.Println("\n  Run 'sirsi setup --install' to install missing tools automatically.")
		return nil
	}

	// Install missing dependencies.
	fmt.Println()
	for _, s := range statuses {
		if s.Installed {
			continue
		}

		fmt.Printf("  Installing %s... ", s.Name)

		parts := strings.Fields(s.InstallCmd)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		start := time.Now()
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ failed: %v\n", err)
			fmt.Printf("    Manual install: %s\n", s.InstallCmd)
			continue
		}

		// Verify it's now in PATH.
		if _, err := exec.LookPath(s.Name); err == nil {
			fmt.Printf("✅ (%s)\n", time.Since(start).Truncate(time.Second))
		} else {
			fmt.Printf("⚠️  installed but not in PATH (restart terminal)\n")
		}
	}

	return nil
}
