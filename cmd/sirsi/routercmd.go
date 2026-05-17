package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/router"
	"github.com/spf13/cobra"
)

var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "Cross-agent collaboration router",
	Long:  `Manage the idea-router for Codex ↔ Claude collaboration.`,
}

var routerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show router inbox and topic status",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		r, err := router.New(repoRoot)
		if err != nil {
			return err
		}

		state, err := r.ReadState()
		if err != nil {
			return err
		}

		output.Header("Router Status")
		fmt.Println()

		// Inbox
		if len(state.PendingForClaude) > 0 {
			fmt.Printf("  📥 Claude inbox: %d pending\n", len(state.PendingForClaude))
			for _, id := range state.PendingForClaude {
				fmt.Printf("     • %s\n", id)
			}
		} else {
			fmt.Println("  📥 Claude inbox: clear")
		}

		if len(state.PendingForCodex) > 0 {
			fmt.Printf("  📥 Codex inbox:  %d pending\n", len(state.PendingForCodex))
			for _, id := range state.PendingForCodex {
				fmt.Printf("     • %s\n", id)
			}
		} else {
			fmt.Println("  📥 Codex inbox:  clear")
		}

		fmt.Println()

		// Active topics
		if len(state.ActiveTopics) > 0 {
			fmt.Printf("  Active topics: %d\n", len(state.ActiveTopics))
			for _, t := range state.ActiveTopics {
				fmt.Printf("     • %s\n", t)
			}
		}

		// Completed topics
		if len(state.CompletedTopics) > 0 {
			fmt.Printf("  Completed: %d topics\n", len(state.CompletedTopics))
		}

		fmt.Println()
		fmt.Printf("  Last Codex read:  %s\n", state.LastCodexRead)
		fmt.Printf("  Last Claude read: %s\n", state.LastClaudeRead)
		return nil
	},
}

var (
	watchOnce bool
	inboxAck  bool
)

var routerWatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Poll the router inbox for pending work (Ctrl+C to stop)",
	Long: `Monitor mode — polls the idea-router state every 10 seconds and
prints pending inbox items. This is a human-visible monitor, not
automatic agent triggering. Use --once for a single poll cycle.

True automatic Codex ↔ Claude wakeup requires a router runner (v1),
MCP server, or external automation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		r, err := router.New(repoRoot)
		if err != nil {
			return err
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)

		interval := 10 * time.Second
		if !watchOnce {
			fmt.Printf("  Watching router inbox every %s (Ctrl+C to stop)\n", interval)
			fmt.Println("  Note: this is monitor-only, not auto-triggering.")
			fmt.Println()
		}

		for {
			state, err := r.ReadState()
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
			} else {
				total := len(state.PendingForClaude) + len(state.PendingForCodex)
				ts := time.Now().Format("15:04:05")
				if total > 0 {
					fmt.Printf("  [%s] %d pending items\n", ts, total)
					for _, id := range state.PendingForClaude {
						fmt.Printf("     Pending for Claude: %s\n", id)
					}
					for _, id := range state.PendingForCodex {
						fmt.Printf("     Pending for Codex:  %s\n", id)
					}
				} else {
					fmt.Printf("  [%s] No pending items\n", ts)
				}
			}

			if watchOnce {
				return nil
			}

			select {
			case <-sig:
				fmt.Println("\n  Stopped.")
				return nil
			case <-time.After(interval):
			}
		}
	},
}

var routerInboxCmd = &cobra.Command{
	Use:   "inbox <agent>",
	Short: "Show pending items for an agent (codex or claude)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agent := args[0]

		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		r, err := router.New(repoRoot)
		if err != nil {
			return err
		}

		pending, err := r.PollInbox(agent)
		if err != nil {
			return err
		}

		if len(pending) == 0 {
			fmt.Printf("  No pending items for %s.\n", agent)
			return nil
		}

		fmt.Printf("  Pending for %s: %d items\n\n", agent, len(pending))
		for _, id := range pending {
			doc, err := r.Get(id)
			if err != nil {
				fmt.Printf("  • %s (could not load: %v)\n", id, err)
				continue
			}
			fmt.Printf("  • [%s] %s — %s\n", doc.Type, doc.ID, doc.Title)
		}

		if inboxAck {
			if err := r.AckInbox(agent, pending); err != nil {
				return fmt.Errorf("ack failed: %w", err)
			}
			fmt.Printf("\n  Acknowledged %d items.\n", len(pending))
		} else {
			fmt.Printf("\n  Use --ack to acknowledge and clear these items.\n")
		}

		return nil
	},
}

var (
	runOnce     bool
	runDryRun   bool
	runTarget   string
	runInterval time.Duration
)

var routerRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Autorouter v1 — dispatch pending inbox items to agents",
	Long: `Autorouter v1 dispatches pending Idea Router inbox items to the target agent.

It does NOT acknowledge inbox items. The target agent must ack after reading.
Requires SIRSI_ROUTER_NOTIFY=1 to actually launch agents. Without it, only
--dry-run is allowed (safe preview mode).

  sirsi router run --once --dry-run                 Show what would dispatch
  SIRSI_ROUTER_NOTIFY=1 sirsi router run --once     Dispatch once and exit
  SIRSI_ROUTER_NOTIFY=1 sirsi router run            Poll forever (Ctrl+C to stop)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate target
		switch runTarget {
		case "all", "codex", "claude":
			// valid
		default:
			return fmt.Errorf("invalid --target %q: must be 'all', 'codex', or 'claude'", runTarget)
		}

		// Gate: notification requires SIRSI_ROUTER_NOTIFY=1 unless dry-run
		if !runDryRun && os.Getenv("SIRSI_ROUTER_NOTIFY") != "1" {
			return fmt.Errorf("autorouter dispatch requires SIRSI_ROUTER_NOTIFY=1 (use --dry-run to preview without launching agents)")
		}

		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		r, err := router.New(repoRoot)
		if err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt) //nolint:govet
		defer stop()

		rr := router.NewRunner(r, router.RunnerOptions{
			RepoRoot: repoRoot,
			Agent:    runTarget,
			DryRun:   runDryRun,
			Once:     runOnce,
			Interval: runInterval,
			Out:      os.Stdout,
		})
		return rr.Run(ctx)
	},
}

var (
	daemonDryRun   bool
	daemonTarget   string
	daemonInterval time.Duration
	installRepo    string
	installLoad    bool
	serviceRepo    string
)

var routerDaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the always-on autorouter daemon",
	Long: `Run the resident autorouter daemon for immediate Codex ↔ Claude dispatch.

The daemon watches .agents/idea-router/ with fsnotify and also polls as a
fallback. Live dispatch requires SIRSI_ROUTER_NOTIFY=1. It never acknowledges
inbox items for an agent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch daemonTarget {
		case "all", "codex", "claude":
		default:
			return fmt.Errorf("invalid --target %q: must be 'all', 'codex', or 'claude'", daemonTarget)
		}
		if !daemonDryRun && os.Getenv("SIRSI_ROUTER_NOTIFY") != "1" {
			return fmt.Errorf("autorouter daemon requires SIRSI_ROUTER_NOTIFY=1 (use --dry-run to preview without launching agents)")
		}
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		r, err := router.New(repoRoot)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt) //nolint:govet
		defer stop()

		d := router.NewDaemon(r, router.DaemonOptions{
			RepoRoot:    repoRoot,
			Agent:       daemonTarget,
			DryRun:      daemonDryRun,
			Interval:    daemonInterval,
			Out:         os.Stdout,
			UseFSNotify: true,
		})
		return d.Run(ctx)
	},
}

var routerInstallAgentCmd = &cobra.Command{
	Use:   "install-agent",
	Short: "Install the autorouter launch agent for this repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot := installRepo
		if repoRoot == "" {
			var err error
			repoRoot, err = router.FindRepoRoot()
			if err != nil {
				return fmt.Errorf("no idea-router found: %w", err)
			}
		}
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable: %w", err)
		}
		opts := router.DefaultServiceOptions(repoRoot, exe)
		if err := router.InstallLaunchAgent(opts); err != nil {
			return err
		}
		fmt.Printf("Installed autorouter launch agent: %s\n", opts.PlistPath)
		fmt.Printf("Label: %s\n", opts.Label)
		if installLoad {
			domain, err := launchctlUserDomain()
			if err != nil {
				return err
			}
			if err := router.Launchctl("bootout", domain, opts.PlistPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: prior service was not loaded or could not be unloaded: %v\n", err)
			}
			if err := router.Launchctl("bootstrap", domain, opts.PlistPath); err != nil {
				return err
			}
			fmt.Println("Autorouter launch agent loaded.")
		} else {
			fmt.Println("Use --load to start it now.")
		}
		return nil
	},
}

var routerUninstallAgentCmd = &cobra.Command{
	Use:   "uninstall-agent",
	Short: "Unload and remove the autorouter launch agent for this repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot := serviceRepo
		if repoRoot == "" {
			var err error
			repoRoot, err = router.FindRepoRoot()
			if err != nil {
				return fmt.Errorf("no idea-router found: %w", err)
			}
		}
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable: %w", err)
		}
		opts := router.DefaultServiceOptions(repoRoot, exe)
		domain, err := launchctlUserDomain()
		if err != nil {
			return err
		}
		if err := router.Launchctl("bootout", domain, opts.PlistPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: service was not loaded or could not be unloaded: %v\n", err)
		}
		if err := os.Remove(opts.PlistPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove launch agent plist: %w", err)
		}
		fmt.Printf("Removed autorouter launch agent: %s\n", opts.PlistPath)
		return nil
	},
}

var routerServiceStatusCmd = &cobra.Command{
	Use:   "service-status",
	Short: "Show autorouter launch agent status",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot := serviceRepo
		if repoRoot == "" {
			var err error
			repoRoot, err = router.FindRepoRoot()
			if err != nil {
				return fmt.Errorf("no idea-router found: %w", err)
			}
		}
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable: %w", err)
		}
		opts := router.DefaultServiceOptions(repoRoot, exe)
		fmt.Printf("Label: %s\n", opts.Label)
		fmt.Printf("Plist: %s\n", opts.PlistPath)
		if _, err := os.Stat(opts.PlistPath); err == nil {
			fmt.Println("Installed: yes")
		} else {
			fmt.Println("Installed: no")
		}
		domain, err := launchctlUserDomain()
		if err != nil {
			return err
		}
		if err := router.Launchctl("print", domain+"/"+opts.Label); err != nil {
			fmt.Println("Loaded: no")
		} else {
			fmt.Println("Loaded: yes")
		}
		return nil
	},
}

func launchctlUserDomain() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("resolve current user: %w", err)
	}
	return "gui/" + u.Uid, nil
}

func init() {
	routerWatchCmd.Flags().BoolVar(&watchOnce, "once", false, "Poll once and exit (for testing/CI)")
	routerInboxCmd.Flags().BoolVar(&inboxAck, "ack", false, "Acknowledge and clear pending items")
	routerRunCmd.Flags().BoolVar(&runOnce, "once", false, "Run one dispatch pass and exit")
	routerRunCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Print dispatches without launching agents")
	routerRunCmd.Flags().StringVar(&runTarget, "target", "all", "Dispatch target: codex, claude, or all")
	routerRunCmd.Flags().DurationVar(&runInterval, "interval", 10*time.Second, "Polling interval")
	routerDaemonCmd.Flags().BoolVar(&daemonDryRun, "dry-run", false, "Print dispatches without launching agents")
	routerDaemonCmd.Flags().StringVar(&daemonTarget, "target", "all", "Dispatch target: codex, claude, or all")
	routerDaemonCmd.Flags().DurationVar(&daemonInterval, "interval", time.Second, "Fallback polling interval")
	routerInstallAgentCmd.Flags().StringVar(&installRepo, "repo", "", "Repository root (defaults to current repo)")
	routerInstallAgentCmd.Flags().BoolVar(&installLoad, "load", false, "Load/start the launch agent after writing it")
	routerUninstallAgentCmd.Flags().StringVar(&serviceRepo, "repo", "", "Repository root (defaults to current repo)")
	routerServiceStatusCmd.Flags().StringVar(&serviceRepo, "repo", "", "Repository root (defaults to current repo)")
	routerCmd.AddCommand(routerStatusCmd, routerWatchCmd, routerInboxCmd, routerRunCmd, routerDaemonCmd, routerInstallAgentCmd, routerUninstallAgentCmd, routerServiceStatusCmd)
}
