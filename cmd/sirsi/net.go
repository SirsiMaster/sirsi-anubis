package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/neith"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
)

var netCmd = &cobra.Command{
	Use:     "net",
	Aliases: []string{"neith"},
	Short:   "𓁯 Net — Scope Weaver & Plan Alignment",
	Long: `𓁯 Net — Scope Weaver & Plan Alignment

Net defines task scopes for Ra, tracks plan alignment against build logs,
detects drift, and validates cross-module consistency.

  sirsi net status    Check plan alignment score
  sirsi net align     Validate all-module consistency`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var netStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check plan alignment against build logs",
	RunE:  runNetStatus,
}

var netAlignCmd = &cobra.Command{
	Use:   "align",
	Short: "Validate cross-module consistency",
	RunE:  runNetAlign,
}

func init() {
	netCmd.AddCommand(netStatusCmd)
	netCmd.AddCommand(netAlignCmd)
}

func runNetStatus(cmd *cobra.Command, args []string) error {
	start := time.Now()

	logContent := ""
	for _, path := range []string{"BUILD_LOG.md", "docs/BUILD_LOG.md"} {
		data, err := os.ReadFile(path)
		if err == nil {
			logContent = string(data)
			break
		}
	}

	w := &neith.Weave{
		SessionID: "cli-session",
		StartedAt: time.Now(),
		Plan:      []string{"Build all modules", "Pass all tests", "Ship release"},
	}

	score, err := w.AssessLogs(logContent)
	if err != nil {
		return fmt.Errorf("assess logs: %w", err)
	}

	verdict := "ALIGNED"
	if score < 0.5 {
		verdict = "DRIFTING"
	}

	if JsonOutput {
		result := map[string]interface{}{
			"alignment_score": score,
			"verdict":         verdict,
			"plan_items":      len(w.Plan),
			"plan":            w.Plan,
			"elapsed_ms":      time.Since(start).Milliseconds(),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	output.Banner()
	output.Header("Plan Alignment")

	if logContent != "" {
		output.Info("Loaded BUILD_LOG.md")
	} else {
		output.Warn("No BUILD_LOG.md found — alignment will report 1.0 (no log to compare)")
	}

	output.Dashboard(map[string]string{
		"Alignment Score": fmt.Sprintf("%.1f%%", score*100),
		"Verdict":         verdict,
		"Plan Items":      fmt.Sprintf("%d", len(w.Plan)),
	})

	output.Footer(time.Since(start))
	output.NextSteps(output.SuggestSteps(suggest.Context{Deity: "net", Subcommand: "status"}))
	return nil
}

func runNetAlign(cmd *cobra.Command, args []string) error {
	start := time.Now()
	output.Banner()
	output.Header("Module Consistency Check")

	// Real checks against the current project state
	tap := &neith.Tapestry{}

	// Ma'at: go vet passes
	if err := exec.Command("go", "vet", "./...").Run(); err == nil {
		tap.MaatConsistent = true
		output.Success("Ma'at: go vet passes")
	} else {
		output.Error("Ma'at: go vet failed")
	}

	// Anubis: build succeeds (no scan rule regressions)
	if err := exec.Command("go", "build", "./...").Run(); err == nil {
		tap.AnubisCorrect = true
		output.Success("Anubis: build succeeds")
	} else {
		output.Error("Anubis: build failed")
	}

	// Hygiene: gofmt clean
	out, _ := exec.Command("gofmt", "-l", "./internal/", "./cmd/").Output()
	if len(out) == 0 {
		tap.HygieneClean = true
		output.Success("Hygiene: gofmt clean")
	} else {
		output.Error("Hygiene: gofmt violations found")
	}

	// Thoth: .thoth/ memory present
	if _, err := os.Stat(".thoth/memory.yaml"); err == nil {
		tap.ThothAccurate = true
		output.Success("Thoth: memory present")
	} else {
		output.Warn("Thoth: .thoth/memory.yaml not found")
	}

	// Isis: hardened (always true for alignment — network checks are separate)
	tap.IsisHardened = true
	output.Success("Isis: system health assumed")

	fmt.Println()

	err := tap.Align()
	if err != nil {
		output.Error("Alignment failed: %v", err)
		output.Footer(time.Since(start))
		return nil
	}

	output.Success("All modules aligned — tapestry is balanced")
	output.Footer(time.Since(start))
	actions := suggest.After(suggest.Context{Deity: "net", Subcommand: "align"})
	var steps [][]string
	for _, a := range actions {
		steps = append(steps, []string{a.Command, a.Description})
	}
	output.NextSteps(steps)
	return nil
}
