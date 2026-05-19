package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SirsiMaster/sirsi-pantheon/internal/router"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Register AI agents with the local router",
}

var (
	agentRegisterID        string
	agentRegisterCLI       string
	agentRegisterCWD       string
	agentRegisterURL       string
	agentRegisterAPIKey    string
	agentRegisterMechanism string
	agentRegisterMCPServer string
)

var agentRegisterCmd = &cobra.Command{
	Use:   "register <type>",
	Short: "Register an agent wake profile",
	Long: `Register a new Idea Router agent wake profile.

Examples:
  sirsi agent register gemini --api-key env:GEMINI_API_KEY
  sirsi agent register qwen --cli qwen-cli --cwd ~/Development/myproject
  sirsi agent register webhook --url https://example.test/pantheon-hook`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")
		agentType := args[0]
		cwd := agentRegisterCWD
		if cwd == "" {
			cwd = repoRoot
		}
		id := agentRegisterID
		if id == "" {
			id = agentType + "-pantheon"
		}

		cfg := router.AgentConfig{
			ID:         id,
			Type:       agentType,
			Cwd:        cwd,
			Workstream: "pantheon",
		}
		mechanism := agentRegisterMechanism
		if mechanism == "" {
			switch {
			case agentRegisterURL != "":
				mechanism = router.WakeAPICall
			case agentRegisterMCPServer != "":
				mechanism = router.WakeMCPNotification
			default:
				mechanism = router.WakeCLISpawn
			}
		}
		cfg.Wake.Mechanism = mechanism

		switch mechanism {
		case router.WakeCLISpawn:
			cli := agentRegisterCLI
			if cli == "" {
				cli = agentType
			}
			cfg.Command = []string{cli}
			cfg.Wake.HealthCheck = []string{cli, "--version"}
		case router.WakeAPICall:
			if agentRegisterURL == "" {
				return fmt.Errorf("--url is required for api-call registration")
			}
			cfg.Wake.Endpoint = agentRegisterURL
			cfg.Wake.Auth = agentRegisterAPIKey
		case router.WakeMCPNotification:
			server := agentRegisterMCPServer
			if server == "" {
				server = "sirsi"
			}
			cfg.Wake.MCPServer = server
		default:
			return fmt.Errorf("unsupported wake mechanism %q", mechanism)
		}

		if err := router.RegisterAgent(routerRoot, cfg); err != nil {
			return err
		}
		if !quietMode {
			fmt.Printf("Registered %s with %s wake", id, mechanism)
			if cfg.Wake.Endpoint != "" {
				fmt.Printf(" at %s", cfg.Wake.Endpoint)
			}
			if len(cfg.Command) > 0 {
				fmt.Printf(" via %s", strings.Join(cfg.Command, " "))
			}
			fmt.Println()
		}
		return nil
	},
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered agents and their wake mechanisms",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")
		reg, err := router.LoadRegistry(routerRoot)
		if err != nil {
			return err
		}

		fmt.Printf("  Registered agents: %d\n\n", len(reg.Agents))
		for id, cfg := range reg.Agents {
			wake := cfg.WakeMechanism()
			if wake == "" {
				wake = "cli-spawn"
			}
			fmt.Printf("  %-28s type:%-8s wake:%-16s\n", id, cfg.Type, wake)
		}
		return nil
	},
}

func init() {
	agentRegisterCmd.Flags().StringVar(&agentRegisterID, "id", "", "Agent id (default: <type>-pantheon)")
	agentRegisterCmd.Flags().StringVar(&agentRegisterCLI, "cli", "", "CLI binary for cli-spawn wake")
	agentRegisterCmd.Flags().StringVar(&agentRegisterCWD, "cwd", "", "Working directory for cli-spawn wake")
	agentRegisterCmd.Flags().StringVar(&agentRegisterURL, "url", "", "HTTP endpoint for api-call wake")
	agentRegisterCmd.Flags().StringVar(&agentRegisterAPIKey, "api-key", "", "API auth token or env:VARIABLE reference")
	agentRegisterCmd.Flags().StringVar(&agentRegisterMechanism, "mechanism", "", "Wake mechanism: cli-spawn, api-call, mcp-notification")
	agentRegisterCmd.Flags().StringVar(&agentRegisterMCPServer, "mcp-server", "", "MCP server name for mcp-notification wake")
	agentCmd.AddCommand(agentRegisterCmd, agentListCmd)
}
