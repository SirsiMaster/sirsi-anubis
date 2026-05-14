package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
)

var thothStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check brain and memory health",
	RunE:  runThothStatus,
}

func init() {
	thothCmd.AddCommand(thothStatusCmd)
}

func runThothStatus(cmd *cobra.Command, args []string) error {
	start := time.Now()

	output.Banner()
	output.Header("Memory Status")

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("thoth status: %w", err)
	}

	thothDir := filepath.Join(cwd, ".thoth")
	thothInfo, err := os.Stat(thothDir)
	if err != nil || !thothInfo.IsDir() {
		output.Info(".thoth/ directory: not found")
		output.Dim("  Run `sirsi thoth init` to initialize the knowledge system.")
	} else {
		output.Info(".thoth/ directory: present")

		// memory.yaml
		memPath := filepath.Join(thothDir, "memory.yaml")
		if memInfo, memErr := os.Stat(memPath); memErr == nil {
			output.Info("memory.yaml: %s  (%d bytes)", memInfo.ModTime().Format("2006-01-02 15:04:05"), memInfo.Size())
		} else {
			output.Dim("memory.yaml: not found")
		}

		// journal.md
		journalPath := filepath.Join(thothDir, "journal.md")
		if jInfo, jErr := os.Stat(journalPath); jErr == nil {
			entries := countJournalEntries(journalPath)
			output.Info("journal.md:  %s  (%d bytes, %d entries)", jInfo.ModTime().Format("2006-01-02 15:04:05"), jInfo.Size(), entries)
		} else {
			output.Dim("journal.md:  not found")
		}
	}

	fmt.Println()

	// Brain status
	home, _ := os.UserHomeDir()
	brainDir := filepath.Join(home, ".config", "pantheon", "brain")
	brainInfo, brainErr := os.Stat(brainDir)
	if brainErr != nil || !brainInfo.IsDir() {
		output.Info("Brain: not installed")
		output.Dim("  Run `sirsi thoth brain` to install neural weights.")
	} else {
		entries, _ := os.ReadDir(brainDir)
		var models []string
		for _, e := range entries {
			if e.IsDir() || strings.HasSuffix(e.Name(), ".onnx") || strings.HasSuffix(e.Name(), ".mlmodelc") || strings.HasSuffix(e.Name(), ".mlpackage") {
				models = append(models, e.Name())
			}
		}
		if len(models) == 0 {
			output.Info("Brain: directory exists, no models found")
		} else {
			output.Info("Brain: %d model(s) installed", len(models))
			for _, m := range models {
				output.Dim("  • %s", m)
			}
		}
	}

	output.Footer(time.Since(start))

	actions := suggest.After(suggest.Context{Deity: "thoth", Subcommand: "status"})
	var steps [][]string
	for _, a := range actions {
		steps = append(steps, []string{a.Command, a.Description})
	}
	output.NextSteps(steps)

	return nil
}

// countJournalEntries counts lines starting with "##" in the journal file.
func countJournalEntries(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "##") {
			count++
		}
	}
	return count
}
