// Package router — executor.go
//
// Pluggable agent executor for Router v3. Resolves an agent ID from the
// registry, launches the agent's command with the work context, and verifies
// the agent wrote back to the router before marking dispatch successful.
// Ra owns execution; Thoth preserves continuity; Ma'at validates governance.
package router

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Executor launches registered agents and verifies writeback.
type Executor struct {
	registry  *Registry
	router    *Router
	workQueue *WorkQueue
	out       io.Writer
	timeout   time.Duration
}

// NewExecutor creates an executor backed by the given registry, router, and work queue.
func NewExecutor(reg *Registry, r *Router, wq *WorkQueue, out io.Writer) *Executor {
	return &Executor{
		registry:  reg,
		router:    r,
		workQueue: wq,
		out:       out,
		timeout:   5 * time.Minute,
	}
}

// Dispatch launches the target agent for a work item and verifies writeback.
// Returns nil if the agent completed and wrote back successfully.
func (e *Executor) Dispatch(ctx context.Context, item *WorkItem) error {
	// Resolve agent from registry
	cfg, err := e.registry.Lookup(item.TargetAgentID)
	if err != nil {
		e.workQueue.UpdateStatus(item.ID, StatusFailed, err.Error())
		return err
	}

	if err := cfg.Validate(); err != nil {
		e.workQueue.UpdateStatus(item.ID, StatusFailed, err.Error())
		return err
	}

	// Build the work prompt
	prompt := buildWorkPrompt(item, cfg)

	// Build the command: agent command + prompt as final arg
	args := make([]string, len(cfg.Command)-1)
	copy(args, cfg.Command[1:])
	args = append(args, prompt)

	// Record pre-dispatch state for writeback detection
	preState, _ := e.router.ReadState()
	preLastRead := ""
	if preState != nil {
		switch cfg.Type {
		case "claude":
			preLastRead = preState.LastClaudeRead
		case "codex":
			preLastRead = preState.LastCodexRead
		}
	}

	// Mark dispatched
	e.workQueue.UpdateStatus(item.ID, StatusDispatched, "")
	e.workQueue.Save()
	fmt.Fprintf(e.out, "  Dispatching to %s (%s)...\n", item.TargetAgentID, cfg.Type)

	// Launch with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(execCtx, cfg.Command[0], args...)
	cmd.Dir = cfg.Cwd
	cmd.Stdout = e.out
	cmd.Stderr = &stderr

	// Set env overrides
	if len(cfg.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range cfg.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Record attempt
	stderrStr := stderr.String()
	if len(stderrStr) > 4000 {
		stderrStr = stderrStr[:4000] + "...(truncated)"
	}
	e.workQueue.RecordAttempt(item.ID, exitCode, errString(err), stderrStr)

	// Check for writeback: did the agent update state.json?
	postState, _ := e.router.ReadState()
	writebackDetected := false
	if postState != nil {
		switch cfg.Type {
		case "claude":
			writebackDetected = postState.LastClaudeRead != preLastRead
		case "codex":
			writebackDetected = postState.LastCodexRead != preLastRead
		}
	}

	// Determine final status
	if err != nil {
		reason := fmt.Sprintf("agent exited with code %d: %s", exitCode, errString(err))
		e.workQueue.UpdateStatus(item.ID, StatusFailed, reason)
		e.workQueue.Save()
		fmt.Fprintf(e.out, "  Failed: %s\n", reason)
		return fmt.Errorf("dispatch to %s failed: %w", item.TargetAgentID, err)
	}

	if writebackDetected {
		e.workQueue.UpdateStatus(item.ID, StatusCompleted, "")
		e.workQueue.Save()
		fmt.Fprintf(e.out, "  Completed: %s wrote back to router\n", item.TargetAgentID)
		return nil
	}

	// Agent exited clean but no writeback detected
	reason := "agent exited successfully but no router writeback detected"
	e.workQueue.UpdateStatus(item.ID, StatusFailed, reason)
	e.workQueue.Save()
	fmt.Fprintf(e.out, "  Warning: %s\n", reason)
	return fmt.Errorf("%s", reason)
}

func buildWorkPrompt(item *WorkItem, cfg *AgentConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s (agent_id: %s), processing work from the Idea Router.\n\n", cfg.Type, cfg.ID))
	sb.WriteString(fmt.Sprintf("Repository: %s\n", cfg.Cwd))
	sb.WriteString(fmt.Sprintf("Work item: %s\n", item.DocID))
	if item.Topic != "" {
		sb.WriteString(fmt.Sprintf("Topic: %s\n", item.Topic))
	}
	if item.Goal != "" {
		sb.WriteString(fmt.Sprintf("/goal: %s\n", item.Goal))
	}
	sb.WriteString("\nRead the following files in order:\n")
	sb.WriteString("1. .agents/idea-router/state.json\n")
	sb.WriteString("2. .agents/idea-router/agents.json\n")
	sb.WriteString(fmt.Sprintf("3. The addressed work item document\n"))
	sb.WriteString("4. .agents/idea-router/DESIGN.md (for protocol rules)\n\n")
	sb.WriteString("Then act:\n")
	sb.WriteString("- Implement the work described in the router item.\n")
	sb.WriteString("- Keep work scoped to this repository only.\n")
	sb.WriteString("- Continue until the /goal is met, blocked by safety/user approval, or impossible.\n")
	sb.WriteString("- Run tests and record evidence.\n")
	sb.WriteString("- Write result to .agents/idea-router/reviews/ or decisions/.\n")
	sb.WriteString("- Update state.json: your last_read timestamp, pending items for the next agent.\n")
	sb.WriteString("- Do not leave vague next steps when you can complete the work yourself.\n")
	return sb.String()
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
