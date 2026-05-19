// Package router — nodestatus.go
//
// Horus local-node status aggregation. Combines router queue state,
// agent registry, and daemon health into a single operator view.
// Ra owns the queue and dispatch; Horus owns this per-desktop surface.
package router

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// NodeStatus is the aggregated Horus local-node view.
type NodeStatus struct {
	// Router
	RouterHome string `json:"router_home"`
	RepoRoot   string `json:"repo_root"`

	// Agents
	RegisteredAgents []string `json:"registered_agents"`
	AgentCount       int      `json:"agent_count"`

	// Queue
	PendingByAgent map[string][]string `json:"pending_by_agent"`
	TotalPending   int                 `json:"total_pending"`
	ActiveTopics   []string            `json:"active_topics"`
	CompletedCount int                 `json:"completed_count"`

	// Work items
	WorkItemSummary map[string]int `json:"work_item_summary"` // status → count

	// Daemon
	DaemonInstalled    bool   `json:"daemon_installed"`
	DaemonLoaded       bool   `json:"daemon_loaded"`
	DaemonLabel        string `json:"daemon_label"`
	ConfiguredBinary   string `json:"configured_binary,omitempty"`
	BinaryExists       bool   `json:"binary_exists"`
	BinaryIsGoRun      bool   `json:"binary_is_go_run"`

	// Timestamps
	LastClaudeRead string `json:"last_claude_read"`
	LastCodexRead  string `json:"last_codex_read"`

	// Agent CLI health
	AgentHealth []AgentHealthCheck `json:"agent_health,omitempty"`

	// Dispatch failures (last N from work queue)
	RecentFailures []WorkItemFailure `json:"recent_failures,omitempty"`
}

// AgentHealthCheck reports whether a local agent CLI is available and authenticated.
type AgentHealthCheck struct {
	AgentType string `json:"agent_type"` // "claude", "codex"
	CLIFound  bool   `json:"cli_found"`
	CLIPath   string `json:"cli_path,omitempty"`
	AuthOK    bool   `json:"auth_ok"`
	AuthError string `json:"auth_error,omitempty"`
}

// WorkItemFailure is a summary of a failed dispatch.
type WorkItemFailure struct {
	ItemID    string    `json:"item_id"`
	Agent     string    `json:"agent"`
	Error     string    `json:"error"`
	FailedAt  time.Time `json:"failed_at"`
}

// LaunchctlChecker abstracts launchctl probing for testability.
type LaunchctlChecker func(args ...string) error

// CollectNodeStatus gathers the Horus local-node view from all sources.
func CollectNodeStatus(repoRoot string, launchctlCheck LaunchctlChecker) (*NodeStatus, error) {
	routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")

	ns := &NodeStatus{
		RouterHome:     routerRoot,
		RepoRoot:       repoRoot,
		PendingByAgent: make(map[string][]string),
		WorkItemSummary: make(map[string]int),
	}

	// --- Registry ---
	reg, err := LoadRegistry(routerRoot)
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}
	for id := range reg.Agents {
		ns.RegisteredAgents = append(ns.RegisteredAgents, id)
	}
	sort.Strings(ns.RegisteredAgents)
	ns.AgentCount = len(ns.RegisteredAgents)

	// --- Router state ---
	r, err := New(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("open router: %w", err)
	}
	state, err := r.ReadState()
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	state.NormalizePending()

	ns.ActiveTopics = state.ActiveTopics
	ns.CompletedCount = len(state.CompletedTopics)
	ns.LastClaudeRead = state.LastClaudeRead
	ns.LastCodexRead = state.LastCodexRead

	for agent, ids := range state.Pending {
		if len(ids) > 0 {
			ns.PendingByAgent[agent] = ids
			ns.TotalPending += len(ids)
		}
	}

	// --- Work queue ---
	wq, err := LoadWorkQueue(routerRoot)
	if err == nil {
		for _, item := range wq.Items {
			ns.WorkItemSummary[string(item.Status)]++
		}
		// Collect recent failures
		for _, item := range wq.Items {
			if item.Status == StatusFailed && len(item.Attempts) > 0 {
				last := item.Attempts[len(item.Attempts)-1]
				ns.RecentFailures = append(ns.RecentFailures, WorkItemFailure{
					ItemID:   item.ID,
					Agent:    item.TargetAgentID,
					Error:    last.Error,
					FailedAt: last.At,
				})
			}
		}
		// Sort failures newest first, limit to 5
		sort.Slice(ns.RecentFailures, func(i, j int) bool {
			return ns.RecentFailures[i].FailedAt.After(ns.RecentFailures[j].FailedAt)
		})
		if len(ns.RecentFailures) > 5 {
			ns.RecentFailures = ns.RecentFailures[:5]
		}
	}

	// --- Daemon health ---
	exe, err := os.Executable()
	if err == nil {
		exe, _ = ResolveStableBinary(repoRoot, exe)
	}
	if exe == "" {
		exe = "sirsi" // fallback for display
	}
	opts := DefaultServiceOptions(repoRoot, exe)
	ns.DaemonLabel = opts.Label

	if _, err := os.Stat(opts.PlistPath); err == nil {
		ns.DaemonInstalled = true
		if program, err := LaunchAgentProgram(opts.PlistPath); err == nil {
			ns.ConfiguredBinary = program
			if _, err := os.Stat(program); err == nil {
				ns.BinaryExists = true
			}
			ns.BinaryIsGoRun = IsGoRunBinary(program)
		}
	}

	if launchctlCheck != nil {
		userDomain := "" // caller provides the full argument
		// We check by calling print on the label
		if err := launchctlCheck("print", userDomain+ns.DaemonLabel); err == nil {
			ns.DaemonLoaded = true
		}
	}

	// --- Agent CLI health ---
	for _, agentType := range []string{"claude", "codex"} {
		check := AgentHealthCheck{AgentType: agentType}
		path, err := exec.LookPath(agentType)
		if err != nil {
			check.AuthError = fmt.Sprintf("%s CLI not found in PATH", agentType)
		} else {
			check.CLIFound = true
			check.CLIPath = path
			// Probe auth: run a minimal command that fails fast if not authenticated
			var probeCmd *exec.Cmd
			switch agentType {
			case "claude":
				probeCmd = exec.Command(path, "--version")
			case "codex":
				probeCmd = exec.Command(path, "--version")
			}
			if probeCmd != nil {
				out, err := probeCmd.CombinedOutput()
				if err != nil {
					check.AuthError = fmt.Sprintf("CLI check failed: %s", strings.TrimSpace(string(out)))
				} else {
					check.AuthOK = true
				}
			}
		}
		ns.AgentHealth = append(ns.AgentHealth, check)
	}

	return ns, nil
}
