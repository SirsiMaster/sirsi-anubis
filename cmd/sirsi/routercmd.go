package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SirsiMaster/sirsi-pantheon/internal/router"
	"github.com/SirsiMaster/sirsi-pantheon/internal/work"
	"github.com/spf13/cobra"
)

// workRoot resolves the router items directory for the current repo.
func workRoot() (string, error) {
	repoRoot, err := router.FindRepoRoot()
	if err != nil {
		return "", fmt.Errorf("no .agents/idea-router/ found: %w", err)
	}
	root := filepath.Join(repoRoot, ".agents", "idea-router")
	if err := work.EnsureRoot(root); err != nil {
		return "", err
	}
	return root, nil
}

// loadOrLiteral returns the literal value, or the contents of the file if it
// starts with @. Lets callers pass --instructions "text" or --instructions @file.
func loadOrLiteral(v string) (string, error) {
	if strings.HasPrefix(v, "@") {
		data, err := os.ReadFile(strings.TrimPrefix(v, "@"))
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return v, nil
}

var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "Pull-model work queue between agent threads",
	Long: `The router is a directory of markdown files. Senders write items
with sirsi router send; recipients pull with sirsi router pull and close
with sirsi router close. No daemon, no dispatch, no launch agents — each
agent session reads its own inbox on wake.

Thread registration is handled separately by sirsi thread register.`,
}

var routerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Summarize the work queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := workRoot()
		if err != nil {
			return err
		}
		all, err := work.ListAll(root)
		if err != nil {
			return err
		}
		var open, closed int
		perAgent := map[string]int{}
		for _, it := range all {
			if it.Status == "closed" {
				closed++
				continue
			}
			open++
			perAgent[it.To]++
		}
		fmt.Printf("  Items: %d open, %d closed\n", open, closed)
		if open > 0 {
			fmt.Println("\n  Open by recipient:")
			for agent, n := range perAgent {
				fmt.Printf("    %s: %d\n", agent, n)
			}
		}
		return nil
	},
}

var (
	sendFrom         string
	sendTo           string
	sendTitle        string
	sendInstructions string
)

var routerSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a work item from one agent to another",
	Long: `Writes a new open work item under .agents/idea-router/items/. The
recipient picks it up next time they run sirsi router pull <their-id>.

  sirsi router send --from claude-pantheon --to codex-pantheon \
    --title "review canon-sync" --instructions @proposal.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if sendFrom == "" || sendTo == "" {
			return fmt.Errorf("--from and --to are required")
		}
		if sendTitle == "" {
			return fmt.Errorf("--title is required")
		}
		instr, err := loadOrLiteral(sendInstructions)
		if err != nil {
			return fmt.Errorf("--instructions: %w", err)
		}
		root, err := workRoot()
		if err != nil {
			return err
		}
		id, err := work.Send(root, sendFrom, sendTo, sendTitle, instr)
		if err != nil {
			return err
		}
		fmt.Printf("  Sent %s → %s: %s\n", sendFrom, sendTo, id)
		return nil
	},
}

var routerPullCmd = &cobra.Command{
	Use:   "pull <agent>",
	Short: "Pull open work items addressed to an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := workRoot()
		if err != nil {
			return err
		}
		items, err := work.ListInbox(root, args[0])
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Printf("  No open items for %s.\n", args[0])
			return nil
		}
		fmt.Printf("  %d open items for %s:\n\n", len(items), args[0])
		for _, it := range items {
			fmt.Printf("  • %s\n      from: %s\n      title: %s\n      opened: %s\n\n", it.ID, it.From, it.Title, it.Opened)
		}
		fmt.Printf("  Read full: sirsi router show <id>\n")
		fmt.Printf("  Close when done: sirsi router close <id> --result @path/to/result.md\n")
		return nil
	},
}

var routerShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Print the full text of a work item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := workRoot()
		if err != nil {
			return err
		}
		path := filepath.Join(root, "items", args[0]+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read item: %w", err)
		}
		_, _ = os.Stdout.Write(data)
		return nil
	},
}

var closeResult string

var routerCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Mark a work item closed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := loadOrLiteral(closeResult)
		if err != nil {
			return fmt.Errorf("--result: %w", err)
		}
		root, err := workRoot()
		if err != nil {
			return err
		}
		if err := work.Close(root, args[0], result); err != nil {
			return err
		}
		fmt.Printf("  Closed %s\n", args[0])
		return nil
	},
}

func init() {
	routerSendCmd.Flags().StringVar(&sendFrom, "from", "", "Sender agent id (e.g., claude-pantheon)")
	routerSendCmd.Flags().StringVar(&sendTo, "to", "", "Recipient agent id (e.g., codex-pantheon)")
	routerSendCmd.Flags().StringVar(&sendTitle, "title", "", "Short title for the work item")
	routerSendCmd.Flags().StringVar(&sendInstructions, "instructions", "", "Instructions body (literal text, or @file)")
	routerCloseCmd.Flags().StringVar(&closeResult, "result", "", "Result body (literal text, or @file)")
	routerCmd.AddCommand(routerStatusCmd, routerSendCmd, routerPullCmd, routerShowCmd, routerCloseCmd)
}
