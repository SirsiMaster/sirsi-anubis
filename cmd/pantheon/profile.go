package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/profile"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "📋 Manage scan profiles (general, developer, ai-engineer, devops)",
	Long: `📋 Profile — Scan Configuration Profiles

Profiles define which scan categories are active for your workflow.

  pantheon profile list           Show all available profiles
  pantheon profile use <name>     Set the active profile
  pantheon profile show <name>    Show profile details

Built-in profiles: general, developer, ai-engineer, devops`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available profiles",
	Run:   runProfileList,
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the active scan profile",
	Args:  cobra.ExactArgs(1),
	Run:   runProfileUse,
}

var profileShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show details of a profile",
	Args:  cobra.ExactArgs(1),
	Run:   runProfileShow,
}

func init() {
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileShowCmd)
}

func runProfileList(cmd *cobra.Command, args []string) {
	profiles, err := profile.ListProfiles()
	if err != nil {
		output.Error(fmt.Sprintf("Failed to list profiles: %v", err))
		os.Exit(1)
	}

	cfg, _ := profile.LoadConfig()
	active := cfg.ActiveProfile

	output.Header("📋 Available Profiles")
	fmt.Println()

	for _, p := range profiles {
		marker := "  "
		if p.Name == active {
			marker = "▶ "
		}
		fmt.Printf("  %s%-15s  %s\n", marker, p.Name, p.Description)
	}
	fmt.Println()
	output.Info(fmt.Sprintf("Active: %s", active))
	output.Info("Change with: pantheon profile use <name>")
	fmt.Println()
}

func runProfileUse(cmd *cobra.Command, args []string) {
	name := args[0]

	// Verify profile exists
	_, err := profile.LoadProfile(name)
	if err != nil {
		output.Error(fmt.Sprintf("Profile not found: %s", name))
		output.Info("Run: pantheon profile list")
		os.Exit(1)
	}

	// Update config
	cfg, _ := profile.LoadConfig()
	cfg.ActiveProfile = name
	if err := profile.SaveConfig(cfg); err != nil {
		output.Error(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	output.Info(fmt.Sprintf("✅ Active profile set to: %s", name))
}

func runProfileShow(cmd *cobra.Command, args []string) {
	name := args[0]

	p, err := profile.LoadProfile(name)
	if err != nil {
		output.Error(fmt.Sprintf("Profile not found: %s", name))
		os.Exit(1)
	}

	output.Header(fmt.Sprintf("📋 Profile: %s", p.Name))
	fmt.Println()
	output.Info(fmt.Sprintf("Description:  %s", p.Description))
	fmt.Printf("  Categories:   %v\n", p.Categories)
	if p.MinAgeDays > 0 {
		fmt.Printf("  Min Age:      %d days\n", p.MinAgeDays)
	}
	if len(p.ExcludeRules) > 0 {
		fmt.Printf("  Excluded:     %v\n", p.ExcludeRules)
	}
	fmt.Println()
}
