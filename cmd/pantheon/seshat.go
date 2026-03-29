package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/mcp"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
	"github.com/SirsiMaster/sirsi-pantheon/internal/seshat"
)

var seshatCmd = &cobra.Command{
	Use:   "seshat",
	Short: "𓁆 Seshat — Gemini Bridge & AI Scribe Engine",
	Long: `𓁆 Seshat — The Scribe. Goddess of writing, wisdom, and measurement.

Seshat manages bidirectional knowledge sync and AI developer context.
Use it to sync Gemini conversations or start the MCP context server.

  pantheon seshat sync           Knowledge sync (Gemini ↔ NotebookLM)
  pantheon seshat list           List Antigravity brain items
  pantheon seshat mcp            Start the Model Context Protocol (MCP) server`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var seshatSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "𓁆 Bidirectional knowledge sync and extraction",
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		output.Banner()
		output.Header("SESHAT — Knowledge Sync")

		paths := seshat.DefaultPaths()
		kiName, _ := cmd.Flags().GetString("ki")
		target, _ := cmd.Flags().GetString("target")

		if kiName != "" && target != "" {
			if err := seshat.SyncKIToGeminiMD(paths, kiName, target); err != nil {
				return err
			}
			output.Success("Synced KI '%s' → %s", kiName, target)
		}

		output.Footer(time.Since(start))
		return nil
	},
}

var seshatListCmd = &cobra.Command{
	Use:   "list",
	Short: "𓁆 List Antigravity brain Knowledge Items",
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		output.Banner()
		output.Header("SESHAT — Knowledge Library")

		paths := seshat.DefaultPaths()
		items, _ := seshat.ListKnowledgeItems(paths)

		for i, item := range items {
			ki, _ := seshat.ReadKnowledgeItem(paths, item)
			fmt.Printf("  %d. %s\n", i+1, ki.Title)
			fmt.Printf("     %s\n", output.Truncate(ki.Summary, 80))
		}
		output.Footer(time.Since(start))
		return nil
	},
}

var seshatMcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "𓁆 Start the Model Context Protocol (MCP) context server",
	Run: func(cmd *cobra.Command, args []string) {
		// Singleton for MCP server
		unlock, err := platform.TryLock("mcp-cli")
		if err != nil {
			output.Error("Pantheon MCP Server is already active.")
			return
		}
		defer unlock()

		output.Header("SESHAT — Scribe's Voice (MCP Server)")

		server := mcp.NewServer()
		if err := server.Run(); err != nil {
			output.Error("Server error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	seshatSyncCmd.Flags().String("ki", "", "Knowledge Item name to sync")
	seshatSyncCmd.Flags().String("target", "", "Target GEMINI.md file path")

	seshatCmd.AddCommand(seshatSyncCmd)
	seshatCmd.AddCommand(seshatListCmd)
	seshatCmd.AddCommand(seshatMcpCmd)
}
