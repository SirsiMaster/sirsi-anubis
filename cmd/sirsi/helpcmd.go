package main

import (
	"github.com/spf13/cobra"

	"github.com/SirsiMaster/sirsi-pantheon/internal/help"
	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
)

var (
	helpDocs bool
	helpList bool
)

var helpCmd = &cobra.Command{
	Use:   "guides [module]",
	Short: "Show rich guides for Pantheon modules",
	Long: `Show a styled terminal guide for any Pantheon module, or open the
web documentation in your browser.

  sirsi guides memory         Show terminal guide for the Memory module
  sirsi guides knowledge --docs  Open Knowledge web docs in browser
  sirsi guides --list         List all available guides`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if helpList || len(args) == 0 {
			help.ListGuides()
			return nil
		}

		deity := args[0]

		if helpDocs {
			output.Info("Opening docs for %s...", deity)
			return help.OpenDocs(deity)
		}

		return help.ShowGuide(deity)
	},
}

func init() {
	helpCmd.Flags().BoolVar(&helpDocs, "docs", false, "Open web documentation in browser")
	helpCmd.Flags().BoolVar(&helpList, "list", false, "List all available module guides")

	// Register as a named subcommand (not overriding cobra's built-in help)
	rootCmd.AddCommand(helpCmd)
}
