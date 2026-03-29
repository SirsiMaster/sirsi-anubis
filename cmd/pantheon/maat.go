package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/isis"
	"github.com/SirsiMaster/sirsi-pantheon/internal/maat"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

var (
	maatSudo bool
	maatFix  bool

	// Isis / Heal flags
	healFull bool
	healLint bool
)

var maatCmd = &cobra.Command{
	Use:   "maat",
	Short: "𓁐 Ma'at — QA/QC Governance & Policy Enforcement",
	Long: `𓁐 Ma'at — The Goddess of Truth, Balance, and Cosmic Order

Ma'at manages your workstation's governance and ensures all infrastructure
complies with the Pantheon Charter. It balances the Scale of Truth.

  pantheon maat audit            Run full governance assessment
  pantheon maat scales           Enforce infrastructure policies (Scales)
  pantheon maat heal             Autonomous remediation cycle (Isis)`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var maatAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "𓁐 Full workstation governance and compliance scan",
	RunE:  runMaatAudit,
}

var maatScalesCmd = &cobra.Command{
	Use:   "scales",
	Short: "𓆄 Enforce infrastructure policies and resolve drifts",
	RunE:  runMaatScales,
}

var maatHealCmd = &cobra.Command{
	Use:   "heal",
	Short: "𓁐 Autonomous remediation cycle (Ma'at → Isis)",
	RunE:  runMaatHeal,
}

func init() {
	maatAuditCmd.Flags().BoolVar(&maatSudo, "sudo", false, "Scan system-level governance")
	maatScalesCmd.Flags().BoolVar(&maatFix, "fix", false, "Actually apply policy fixes")

	maatHealCmd.Flags().BoolVar(&maatFix, "fix", false, "Apply healing remedies")
	maatHealCmd.Flags().BoolVar(&healFull, "full", false, "Run full (slow) test suite")

	maatCmd.AddCommand(maatAuditCmd)
	maatCmd.AddCommand(maatScalesCmd)
	maatCmd.AddCommand(maatHealCmd)
}

func runMaatAudit(cmd *cobra.Command, args []string) error {
	start := time.Now()
	output.Banner()
	output.Header("MA'AT — Governance Audit")

	report, err := maat.Weigh(&maat.CoverageAssessor{Thresholds: maat.DefaultThresholds(), DiffOnly: true})
	if err != nil {
		return err
	}

	output.Dashboard(map[string]string{
		"Verdict": fmt.Sprintf("%s", report.OverallVerdict.Icon()),
		"Weight":  fmt.Sprintf("%d/100", report.OverallWeight),
		"Status":  report.OverallVerdict.String(),
	})
	output.Footer(time.Since(start))
	return nil
}

func runMaatScales(cmd *cobra.Command, args []string) error {
	start := time.Now()
	output.Banner()
	output.Header("MA'AT — The Scales of Balance")
	output.Footer(time.Since(start))
	return nil
}

func runMaatHeal(cmd *cobra.Command, args []string) error {
	start := time.Now()
	output.Banner()
	output.Header("MA'AT — The Healing Pulse (Isis)")

	// Step 1: Weigh
	report, _ := maat.Weigh(&maat.CoverageAssessor{Thresholds: maat.DefaultThresholds(), DiffOnly: !healFull})

	// Step 2: Heal
	findings := isis.FromMaatReport(report)
	if len(findings) == 0 {
		output.Success("The feather is balanced. No healing required.")
		return nil
	}

	healer := isis.NewHealer(".")
	res := healer.Heal(findings, !maatFix)

	output.Dashboard(map[string]string{
		"Findings": fmt.Sprintf("%d", len(findings)),
		"Healed":   fmt.Sprintf("%d", res.Healed),
		"Failed":   fmt.Sprintf("%d", res.Failed),
	})
	output.Footer(time.Since(start))
	return nil
}
