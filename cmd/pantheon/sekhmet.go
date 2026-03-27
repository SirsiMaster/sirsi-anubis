package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/guard"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

var sekhmetTokenizeText string

var sekhmetCmd = &cobra.Command{
	Use:   "sekhmet",
	Short: "𓁵 Sekhmet — The Warrior: ANE-accelerated tokenization & compute",
	Long: `𓁵 Sekhmet — The Warrior
 
Offload intensive tokenization and ML inference to the Apple Neural Engine (ANE).
Phase II: Move tokenization from Node.js to native Go.
 
  pantheon sekhmet --tokenize "text to tokenize"`,
	Run: runSekhmet,
}

func init() {
	sekhmetCmd.Flags().StringVar(&sekhmetTokenizeText, "tokenize", "", "Text string to tokenize via ANE/CPU")
	rootCmd.AddCommand(sekhmetCmd)
}

func runSekhmet(cmd *cobra.Command, args []string) {
	if sekhmetTokenizeText == "" {
		_ = cmd.Help()
		return
	}

	result, err := guard.Tokenize(sekhmetTokenizeText)
	if err != nil {
		output.Error("Sekhmet tokenization failed: %v", err)
		os.Exit(1)
	}

	if JsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return
	}

	output.Header("𓁵 Sekhmet — Tokenization")
	fmt.Println()
	output.Info("Accelerator: %s", result.Accel)
	output.Info("Count:       %d tokens", result.Count)
	output.Info("Text Length: %d chars", len(sekhmetTokenizeText))
	fmt.Println()

	// Show first 10 tokens
	limit := 10
	if len(result.Tokens) < limit {
		limit = len(result.Tokens)
	}
	fmt.Printf("  Tokens: %v", result.Tokens[:limit])
	if len(result.Tokens) > limit {
		fmt.Printf(" ... (+%d more)", len(result.Tokens)-limit)
	}
	fmt.Println()
	fmt.Println()
}
