package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/router"
	"github.com/spf13/cobra"
)

var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "Ra — multi-agent work queue and dispatch",
	Long: `Ra's Idea Router: multi-agent work queue for autonomous collaboration.

Routes work to registered agents (Claude, Codex, Gemini, Qwen, etc.),
launches them with context, and verifies writeback. Thoth preserves
router continuity; Ma'at validates governance.`,
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
		state.NormalizePending()

		output.Header("Ra — Router Status")
		fmt.Println()

		// Registered inboxes
		entries := state.PendingEntries(false)
		if len(entries) == 0 {
			fmt.Println("  📥 Registered inboxes: clear")
		} else {
			fmt.Printf("  📥 Registered inboxes: %d active\n", len(entries))
			for _, entry := range entries {
				fmt.Printf("     %s: %d pending\n", entry.Agent, len(entry.IDs))
				for _, id := range entry.IDs {
					fmt.Printf("       • %s\n", id)
				}
			}
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
				entries := state.PendingEntries(false)
				total := 0
				for _, entry := range entries {
					total += len(entry.IDs)
				}
				ts := time.Now().Format("15:04:05")
				if total > 0 {
					fmt.Printf("  [%s] %d pending items\n", ts, total)
					for _, entry := range entries {
						for _, id := range entry.IDs {
							fmt.Printf("     Pending for %s: %s\n", entry.Agent, id)
						}
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
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")

		// Validate target against registry
		if runTarget != "all" {
			reg, err := router.LoadRegistry(routerRoot)
			if err == nil && !reg.IsRegistered(runTarget) {
				// Allow legacy "codex"/"claude" for backwards compat
				if runTarget != "codex" && runTarget != "claude" {
					return fmt.Errorf("agent %q not registered in agents.json", runTarget)
				}
			}
		}

		// Build v3 executor
		reg, _ := router.LoadRegistry(routerRoot)
		wq, _ := router.LoadWorkQueue(routerRoot)
		exec := router.NewExecutor(reg, r, wq, os.Stdout)

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt) //nolint:govet
		defer stop()

		rr := router.NewRunner(r, router.RunnerOptions{
			RepoRoot: repoRoot,
			Agent:    runTarget,
			DryRun:   runDryRun,
			Once:     runOnce,
			Interval: runInterval,
			Out:      os.Stdout,
			Executor: exec,
		})
		return rr.Run(ctx)
	},
}

var (
	daemonDryRun   bool
	daemonTarget   string
	daemonInterval time.Duration
	workDryRun     bool
	workTarget     string
	workPoll       bool
	workInterval   time.Duration
	installRepo    string
	installLoad    bool
	serviceRepo    string
	smokeDryRun    bool
	smokeAgentPair bool
)

var routerDaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the always-on autorouter daemon",
	Long: `Run the resident autorouter daemon for immediate Codex ↔ Claude dispatch.

The daemon watches .agents/idea-router/ with fsnotify and also polls as a
fallback. Live dispatch requires SIRSI_ROUTER_NOTIFY=1. It never acknowledges
inbox items for an agent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")

		// Validate target against registry
		if daemonTarget != "all" {
			reg, err := router.LoadRegistry(routerRoot)
			if err == nil && !reg.IsRegistered(daemonTarget) {
				if daemonTarget != "codex" && daemonTarget != "claude" {
					return fmt.Errorf("agent %q not registered in agents.json", daemonTarget)
				}
			}
		}

		// Build v3 executor for daemon
		reg, _ := router.LoadRegistry(routerRoot)
		wq, _ := router.LoadWorkQueue(routerRoot)
		exec := router.NewExecutor(reg, r, wq, os.Stdout)

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt) //nolint:govet
		defer stop()

		d := router.NewDaemon(r, router.DaemonOptions{
			RepoRoot:    repoRoot,
			Agent:       daemonTarget,
			DryRun:      daemonDryRun,
			Interval:    daemonInterval,
			Out:         os.Stdout,
			Executor:    exec,
			UseFSNotify: true,
		})
		return d.Run(ctx)
	},
}

var routerWorkCmd = &cobra.Command{
	Use:   "work",
	Short: "Check the router, then launch runnable registered-agent work",
	Long: `Check the Idea Router queue and immediately dispatch runnable work to the
registered target agents. This is the operator verb for "check the router,
then work."

Unlike router run, this command is live by default because invoking it is the
operator approval to launch registered agents. Use --dry-run to preview.

  sirsi router work                       Check once, launch runnable work
  sirsi router work --dry-run             Check once, preview launches
  sirsi router work --poll                Keep polling and launching work
  sirsi router work --target claude-finalwishes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		r, err := router.New(repoRoot)
		if err != nil {
			return err
		}
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")
		if err := validateRouterTarget(routerRoot, workTarget); err != nil {
			return err
		}

		reg, err := router.LoadRegistry(routerRoot)
		if err != nil {
			return err
		}
		wq, err := router.LoadWorkQueue(routerRoot)
		if err != nil {
			return err
		}
		exec := router.NewExecutor(reg, r, wq, os.Stdout)

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt) //nolint:govet
		defer stop()

		rr := router.NewRunner(r, router.RunnerOptions{
			RepoRoot: repoRoot,
			Agent:    workTarget,
			DryRun:   workDryRun,
			Once:     !workPoll,
			Interval: workInterval,
			Out:      os.Stdout,
			Executor: exec,
		})

		pending, err := rr.PendingDispatches()
		if err != nil {
			return err
		}
		if workDryRun {
			fmt.Printf("Router work check: %d runnable dispatches would launch.\n", len(pending))
		} else {
			fmt.Printf("Router work check: %d runnable dispatches ready.\n", len(pending))
		}
		if workPoll {
			fmt.Printf("Polling every %s. Press Ctrl+C to stop.\n", workInterval)
		}
		return rr.Run(ctx)
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
		exe, err = router.ResolveStableBinary(repoRoot, exe)
		if err != nil {
			return err
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
			if program, err := router.LaunchAgentProgram(opts.PlistPath); err == nil {
				fmt.Printf("Configured binary: %s\n", program)
				if _, err := os.Stat(program); err == nil {
					fmt.Println("Configured binary exists: yes")
				} else {
					fmt.Printf("Configured binary exists: no (%v)\n", err)
				}
				if router.IsGoRunBinary(program) {
					fmt.Println("Configured binary is temporary go-run output: yes")
				} else {
					fmt.Println("Configured binary is temporary go-run output: no")
				}
			} else {
				fmt.Printf("Configured binary: unreadable (%v)\n", err)
			}
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

func validateRouterTarget(routerRoot, target string) error {
	if target == "" || target == "all" {
		return nil
	}
	reg, err := router.LoadRegistry(routerRoot)
	if err == nil && reg.IsRegistered(target) {
		return nil
	}
	if target == "codex" || target == "claude" {
		return nil
	}
	return fmt.Errorf("agent %q not registered in agents.json", target)
}

var routerNodeStatusCmd = &cobra.Command{
	Use:   "node-status",
	Short: "Horus — local node status (agents, queue, daemon health)",
	Long: `Horus local-node status aggregation. Shows everything happening on this
machine's router node: registered agents, pending work by agent, work-queue
item statuses, daemon health, and recent dispatch failures.

Ra owns the queue and dispatch. Horus owns this per-desktop view.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}

		domain, err := launchctlUserDomain()
		if err != nil {
			domain = ""
		}
		launchctlCheck := func(lcArgs ...string) error {
			if domain == "" {
				return fmt.Errorf("no user domain")
			}
			// Replace empty domain prefix with the actual domain
			if len(lcArgs) >= 2 && lcArgs[0] == "print" {
				lcArgs[1] = domain + "/" + lcArgs[1]
			}
			return router.Launchctl(lcArgs...)
		}

		ns, err := router.CollectNodeStatus(repoRoot, launchctlCheck)
		if err != nil {
			return err
		}

		output.Header("𓂀 Horus — Local Node Status")
		fmt.Println()

		// Router home
		fmt.Printf("  Router home:  %s\n", ns.RouterHome)
		fmt.Printf("  Repo root:    %s\n", ns.RepoRoot)
		fmt.Println()

		// Registered agents
		fmt.Printf("  Registered agents: %d\n", ns.AgentCount)
		for _, id := range ns.RegisteredAgents {
			fmt.Printf("    • %s\n", id)
		}
		fmt.Println()

		// Pending work
		if ns.TotalPending == 0 {
			fmt.Println("  Pending work: none")
		} else {
			fmt.Printf("  Pending work: %d items\n", ns.TotalPending)
			for agent, ids := range ns.PendingByAgent {
				fmt.Printf("    %s: %d\n", agent, len(ids))
				for _, id := range ids {
					fmt.Printf("      • %s\n", id)
				}
			}
		}
		fmt.Println()

		// Active topics
		if len(ns.ActiveTopics) > 0 {
			fmt.Printf("  Active topics: %d\n", len(ns.ActiveTopics))
			for _, t := range ns.ActiveTopics {
				fmt.Printf("    • %s\n", t)
			}
		}
		fmt.Printf("  Completed topics: %d\n", ns.CompletedCount)
		fmt.Println()

		// Work queue summary
		if len(ns.WorkItemSummary) > 0 {
			fmt.Println("  Work queue:")
			for status, count := range ns.WorkItemSummary {
				fmt.Printf("    %s: %d\n", status, count)
			}
			fmt.Println()
		}

		// Daemon health
		fmt.Println("  Daemon:")
		fmt.Printf("    Label:     %s\n", ns.DaemonLabel)
		if ns.DaemonInstalled {
			fmt.Println("    Installed: yes")
		} else {
			fmt.Println("    Installed: no")
		}
		if ns.DaemonLoaded {
			fmt.Println("    Loaded:    yes")
		} else {
			fmt.Println("    Loaded:    no")
		}
		if ns.ConfiguredBinary != "" {
			fmt.Printf("    Binary:    %s\n", ns.ConfiguredBinary)
			if !ns.BinaryExists {
				fmt.Println("    Binary exists: no (stale plist?)")
			}
			if ns.BinaryIsGoRun {
				fmt.Println("    Warning: binary is temporary go-run output")
			}
		}
		fmt.Println()

		// Last reads
		fmt.Printf("  Last Claude read: %s\n", ns.LastClaudeRead)
		fmt.Printf("  Last Codex read:  %s\n", ns.LastCodexRead)

		// Agent CLI health
		if len(ns.AgentHealth) > 0 {
			fmt.Println("  Agent CLI health:")
			for _, h := range ns.AgentHealth {
				if h.CLIFound && h.AuthOK {
					fmt.Printf("    ✅ %s: ready (%s)\n", h.AgentType, h.CLIPath)
				} else if h.CLIFound && !h.AuthOK {
					fmt.Printf("    ❌ %s: found but auth failed — run '%s' then /login\n", h.AgentType, h.AgentType)
					if h.AuthError != "" {
						fmt.Printf("       %s\n", h.AuthError)
					}
				} else {
					fmt.Printf("    ⚠️  %s: not found in PATH\n", h.AgentType)
				}
			}
			fmt.Println()
		}

		// Recent failures
		if len(ns.RecentFailures) > 0 {
			fmt.Println()
			fmt.Printf("  Recent failures: %d\n", len(ns.RecentFailures))
			for _, f := range ns.RecentFailures {
				fmt.Printf("    %s → %s: %s (%s)\n", f.Agent, f.ItemID, f.Error, f.FailedAt.Format(time.RFC3339))
			}
		}

		return nil
	},
}

var routerSmokeCmd = &cobra.Command{
	Use:   "smoke",
	Short: "Verify both agents can launch and write to the router directory",
	Long: `Smoke test for the autorouter relay. Launches each agent (Claude and Codex)
with a minimal prompt that writes a token file into .agents/idea-router/smoke-test/.
Verifies write access succeeds, then cleans up.

  sirsi router smoke --dry-run       Check CLIs exist without launching
  sirsi router smoke                 Full probe (launches agents, verifies writes)
  sirsi router smoke --agent-pair    Full relay: seed → Claude → Codex → verify`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}

		results, err := router.RunSmoke(cmd.Context(), router.SmokeOptions{
			RepoRoot:  repoRoot,
			DryRun:    smokeDryRun,
			AgentPair: smokeAgentPair,
			Out:       os.Stdout,
		})
		if err != nil {
			return err
		}

		allPassed := true
		fmt.Println()
		output.Header("Ra — Router Smoke Test")
		fmt.Println()
		for _, r := range results {
			status := "PASS"
			if !r.Passed {
				status = "FAIL"
				allPassed = false
			}
			fmt.Printf("  %-8s [%s] %s (%s)\n", r.Agent, status, r.Detail, r.Elapsed.Round(time.Millisecond))
		}
		fmt.Println()
		if !allPassed {
			return fmt.Errorf("smoke test failed — one or more agents cannot write to the router")
		}
		if smokeDryRun {
			fmt.Println("  Dry-run complete. Agent CLIs found. Live writeback was NOT tested.")
			fmt.Println("  Run without --dry-run from your terminal to verify live relay.")
		} else {
			fmt.Println("  Live smoke passed. Agents launched and wrote to the router.")
			fmt.Println("  Note: full relay proof requires both agents to read, act, and advance the queue.")
			fmt.Println("  Run 'sirsi router status' to verify pending items cleared after relay.")
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
	routerWorkCmd.Flags().BoolVar(&workDryRun, "dry-run", false, "Print dispatches without launching agents")
	routerWorkCmd.Flags().StringVar(&workTarget, "target", "all", "Dispatch target: registered agent id, codex, claude, or all")
	routerWorkCmd.Flags().BoolVar(&workPoll, "poll", false, "Keep polling and dispatching until interrupted")
	routerWorkCmd.Flags().DurationVar(&workInterval, "interval", 10*time.Second, "Polling interval for --poll")
	routerInstallAgentCmd.Flags().StringVar(&installRepo, "repo", "", "Repository root (defaults to current repo)")
	routerInstallAgentCmd.Flags().BoolVar(&installLoad, "load", false, "Load/start the launch agent after writing it")
	routerUninstallAgentCmd.Flags().StringVar(&serviceRepo, "repo", "", "Repository root (defaults to current repo)")
	routerServiceStatusCmd.Flags().StringVar(&serviceRepo, "repo", "", "Repository root (defaults to current repo)")
	routerSmokeCmd.Flags().BoolVar(&smokeDryRun, "dry-run", false, "Check CLIs exist without launching agents")
	routerSmokeCmd.Flags().BoolVar(&smokeAgentPair, "agent-pair", false, "Full relay test: seed router item, launch both agents, verify writeback")
	routerCmd.AddCommand(routerStatusCmd, routerWatchCmd, routerInboxCmd, routerRunCmd, routerDaemonCmd, routerWorkCmd, routerInstallAgentCmd, routerUninstallAgentCmd, routerServiceStatusCmd, routerNodeStatusCmd, routerSmokeCmd)
}
