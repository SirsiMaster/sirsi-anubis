package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/output"
	"github.com/SirsiMaster/sirsi-pantheon/internal/router"
	"github.com/spf13/cobra"
)

// reapDeadPIDThreads marks threads as closed when their recorded PID no
// longer exists on this host. Only acts on threads whose Host matches the
// current hostname — refuses to reap remote-host threads since we can't
// observe other hosts' process tables. Returns count reaped.
//
// Called automatically at the top of `sirsi thread list` so orphans get
// swept whenever anyone reads the registry — no daemon, no polling,
// per AGENTS.md §Lean #1 (the read IS the event).
func reapDeadPIDThreads(routerRoot string) int {
	reg, err := router.LoadThreadRegistry(routerRoot)
	if err != nil {
		return 0
	}
	host, _ := os.Hostname()
	reaped := 0
	for _, t := range reg.SortedThreads() {
		if t.Status == router.ThreadStatusClosed {
			continue
		}
		if t.PID <= 0 || t.Host != host {
			continue // missing PID or remote host — can't verify, leave alone
		}
		if err := syscall.Kill(t.PID, syscall.Signal(0)); err != nil && errors.Is(err, syscall.ESRCH) {
			t.Status = router.ThreadStatusClosed
			t.LastError = fmt.Sprintf("reaped: PID %d not alive at %s", t.PID, time.Now().UTC().Format(time.RFC3339))
			reaped++
		}
	}
	if reaped > 0 {
		_ = router.SaveThreadRegistry(routerRoot, reg)
	}
	return reaped
}

var (
	threadRegAgent      string
	threadRegSurface    string
	threadRegRepo       string
	threadRegWorkstream string
	threadRegWatches    []string
	threadRegWake       string
	threadRegID         string

	threadHbID      string
	threadHbStatus  string
	threadHbItem    string
	threadHbError   string
	threadCloseID   string
	threadListAll   bool
	threadListStale time.Duration
)

var threadCmd = &cobra.Command{
	Use:   "thread",
	Short: "CTR — register and track live agent threads",
	Long: `CTR thread registry. Every active agent thread/session (claude, codex,
gemini, gemma, qwen, mcp, api, webhook) should register a thread so
Horus can show which conversations are alive on this workstation.

  sirsi thread register --agent claude-pantheon --surface claude --repo .
  sirsi thread heartbeat --thread thr-abcd1234
  sirsi thread list
  sirsi thread close --thread thr-abcd1234`,
}

var threadRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register the current thread/session with CTR",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := threadRegRepo
		if repo == "" {
			rr, err := router.FindRepoRoot()
			if err != nil {
				return fmt.Errorf("no idea-router found and --repo not provided: %w", err)
			}
			repo = rr
		}
		absRepo, err := filepath.Abs(repo)
		if err == nil {
			repo = absRepo
		}
		routerRoot := filepath.Join(repo, ".agents", "idea-router")
		if _, err := os.Stat(routerRoot); err != nil {
			return fmt.Errorf("router directory not found at %s", routerRoot)
		}

		if threadRegAgent == "" {
			return fmt.Errorf("--agent is required")
		}
		if threadRegSurface == "" {
			return fmt.Errorf("--surface is required (claude|codex|gemini|gemma|qwen|mcp|api|webhook|worker)")
		}

		host, _ := os.Hostname()
		thr := &router.Thread{
			ThreadID:      threadRegID,
			AgentID:       threadRegAgent,
			Surface:       threadRegSurface,
			Repo:          repo,
			Workstream:    threadRegWorkstream,
			Watches:       threadRegWatches,
			WakeMechanism: threadRegWake,
			PID:           os.Getpid(),
			Host:          host,
		}
		// Fill wake mechanism from registry if not provided.
		if thr.WakeMechanism == "" {
			if reg, err := router.LoadRegistry(routerRoot); err == nil {
				if cfg, ok := reg.Agents[threadRegAgent]; ok {
					thr.WakeMechanism = cfg.WakeMechanism()
					if thr.Workstream == "" {
						thr.Workstream = cfg.Workstream
					}
				}
			}
		}
		// Default watches to the agent's own inbox.
		if len(thr.Watches) == 0 {
			thr.Watches = []string{threadRegAgent}
		}

		out, err := router.RegisterThread(routerRoot, thr)
		if err != nil {
			return err
		}

		if JsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(out)
		}
		fmt.Printf("Registered thread %s\n", out.ThreadID)
		fmt.Printf("  agent: %s (surface=%s)\n", out.AgentID, out.Surface)
		fmt.Printf("  watches: %s\n", strings.Join(out.Watches, ", "))
		fmt.Printf("  repo:  %s\n", out.Repo)
		if out.WakeMechanism != "" {
			fmt.Printf("  wake:  %s\n", out.WakeMechanism)
		}
		fmt.Printf("  status: %s\n", out.Status)
		fmt.Println()
		fmt.Println("Send heartbeats with:")
		fmt.Printf("  sirsi thread heartbeat --thread %s\n", out.ThreadID)
		return nil
	},
}

var threadHeartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Send a heartbeat for a registered thread",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")
		if threadHbID == "" {
			return fmt.Errorf("--thread is required")
		}
		upd := router.HeartbeatUpdate{}
		if threadHbStatus != "" {
			upd.Status = router.ThreadStatus(threadHbStatus)
		}
		if cmd.Flags().Changed("current-item") {
			upd.CurrentItem = &threadHbItem
		}
		if cmd.Flags().Changed("last-error") {
			upd.LastError = &threadHbError
		}
		thr, err := router.Heartbeat(routerRoot, threadHbID, upd)
		if err != nil {
			return err
		}
		if JsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(thr)
		}
		fmt.Printf("Heartbeat ok — %s (status=%s, last_seen=%s)\n",
			thr.ThreadID, thr.Status, thr.LastSeenAt.Format(time.RFC3339))
		return nil
	},
}

var threadCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "Mark a thread as closed",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")
		if threadCloseID == "" {
			return fmt.Errorf("--thread is required")
		}
		thr, err := router.CloseThread(routerRoot, threadCloseID)
		if err != nil {
			return err
		}
		fmt.Printf("Closed thread %s (agent=%s)\n", thr.ThreadID, thr.AgentID)
		return nil
	},
}

var threadListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered threads (active by default)",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := router.FindRepoRoot()
		if err != nil {
			return fmt.Errorf("no idea-router found: %w", err)
		}
		routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")
		// Sweep dead-PID threads to closed before reading.
		reapDeadPIDThreads(routerRoot)
		reg, err := router.LoadThreadRegistry(routerRoot)
		if err != nil {
			return err
		}
		stale := threadListStale
		if stale <= 0 {
			stale = router.DefaultThreadStaleAfter
		}
		now := time.Now().UTC()

		type row struct {
			thr   *router.Thread
			stale bool
		}
		var rows []row
		for _, t := range reg.SortedThreads() {
			if t.Status == router.ThreadStatusClosed && !threadListAll {
				continue
			}
			rows = append(rows, row{thr: t, stale: t.IsStale(now, stale)})
		}

		if JsonOutput {
			payload := make([]map[string]any, 0, len(rows))
			for _, r := range rows {
				payload = append(payload, map[string]any{
					"thread":       r.thr,
					"stale":        r.stale,
					"idle_seconds": now.Sub(r.thr.LastSeenAt).Seconds(),
				})
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(payload)
		}

		output.Header("CTR — Live Threads")
		fmt.Println()
		if len(rows) == 0 {
			fmt.Println("  No registered threads. Run `sirsi thread register --agent <id> --surface <surface>`.")
			return nil
		}
		for _, r := range rows {
			marker := "🟢"
			if r.stale {
				marker = "⚠️"
			}
			if r.thr.Status == router.ThreadStatusClosed {
				marker = "⚫"
			} else if r.thr.Status == router.ThreadStatusBlocked {
				marker = "⛔"
			} else if r.thr.Status == router.ThreadStatusIdle {
				marker = "💤"
			}
			fmt.Printf("  %s %s  agent=%s surface=%s status=%s\n",
				marker, r.thr.ThreadID, r.thr.AgentID, r.thr.Surface, r.thr.Status)
			fmt.Printf("      last_seen=%s (idle %.0fs)\n",
				r.thr.LastSeenAt.Format(time.RFC3339),
				now.Sub(r.thr.LastSeenAt).Seconds())
			if len(r.thr.Watches) > 0 {
				fmt.Printf("      watches=%s\n", strings.Join(r.thr.Watches, ","))
			}
			if r.thr.CurrentItem != "" {
				fmt.Printf("      current_item=%s\n", r.thr.CurrentItem)
			}
			if r.thr.LastError != "" {
				fmt.Printf("      last_error=%s\n", r.thr.LastError)
			}
		}
		return nil
	},
}

func init() {
	threadRegisterCmd.Flags().StringVar(&threadRegAgent, "agent", "", "Registered agent ID (e.g., claude-pantheon)")
	threadRegisterCmd.Flags().StringVar(&threadRegSurface, "surface", "", "Surface: claude|codex|gemini|gemma|qwen|mcp|api|webhook|worker")
	threadRegisterCmd.Flags().StringVar(&threadRegRepo, "repo", "", "Repository root (defaults to current router root)")
	threadRegisterCmd.Flags().StringVar(&threadRegWorkstream, "workstream", "", "Workstream name (optional)")
	threadRegisterCmd.Flags().StringSliceVar(&threadRegWatches, "watch", nil, "Inboxes this thread watches (defaults to --agent)")
	threadRegisterCmd.Flags().StringVar(&threadRegWake, "wake", "", "Wake mechanism (defaults to agent registry entry)")
	threadRegisterCmd.Flags().StringVar(&threadRegID, "thread", "", "Reuse a known thread_id instead of generating a new one")

	threadHeartbeatCmd.Flags().StringVar(&threadHbID, "thread", "", "Thread ID to heartbeat (required)")
	threadHeartbeatCmd.Flags().StringVar(&threadHbStatus, "status", "", "Set status: active|idle|blocked")
	threadHeartbeatCmd.Flags().StringVar(&threadHbItem, "current-item", "", "Currently active work item ID")
	threadHeartbeatCmd.Flags().StringVar(&threadHbError, "last-error", "", "Last error string")

	threadCloseCmd.Flags().StringVar(&threadCloseID, "thread", "", "Thread ID to close (required)")

	threadListCmd.Flags().BoolVar(&threadListAll, "all", false, "Include closed threads")
	threadListCmd.Flags().DurationVar(&threadListStale, "stale-after", router.DefaultThreadStaleAfter, "Stale threshold")

	threadCmd.AddCommand(threadRegisterCmd, threadHeartbeatCmd, threadCloseCmd, threadListCmd)
}
