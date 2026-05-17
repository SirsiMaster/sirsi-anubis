package router

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// CommandRunner abstracts exec.Command for testability (Rule A16).
type CommandRunner func(name string, args ...string) *exec.Cmd

var (
	runnerMu sync.RWMutex
	runner   CommandRunner = exec.Command
)

// SetRunner replaces the command runner (for tests).
func SetRunner(fn CommandRunner) {
	runnerMu.Lock()
	defer runnerMu.Unlock()
	runner = fn
}

func getRunner() CommandRunner {
	runnerMu.RLock()
	defer runnerMu.RUnlock()
	return runner
}

// NotifyAgent attempts to notify the other agent about a new router document.
// Target must be "codex" or "claude". Returns an error if the target CLI
// is not installed or cannot be started.
func NotifyAgent(target, docType, docID, repoRoot string) error {
	if err := ValidateAuthor(target); err != nil {
		return fmt.Errorf("invalid notification target: %w", err)
	}

	switch target {
	case "codex":
		return notifyCodex(docType, docID, repoRoot)
	case "claude":
		return notifyClaude(docType, docID, repoRoot)
	default:
		return fmt.Errorf("unknown agent: %s", target)
	}
}

func notifyCodex(docType, docID, repoRoot string) error {
	if _, err := exec.LookPath("codex"); err != nil {
		return fmt.Errorf("codex CLI not found in PATH — install with: npm i -g @openai/codex")
	}

	prompt := buildRouterPrompt("codex", docType, docID)
	run := getRunner()

	cmd := run("codex",
		"exec",
		"--ask-for-approval", "on-request",
		"--sandbox", "workspace-write",
		prompt,
	)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func notifyClaude(docType, docID, repoRoot string) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found in PATH — install from https://claude.ai/code")
	}

	prompt := buildRouterPrompt("claude", docType, docID)
	run := getRunner()

	cmd := run("claude",
		"--print",
		"--permission-mode", "auto",
		prompt,
	)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func buildRouterPrompt(agent, docType, docID string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, processing pending Idea Router work.\n\n", agent))
	sb.WriteString("Read the following files in order:\n")
	sb.WriteString("1. .agents/idea-router/state.json\n")
	sb.WriteString(fmt.Sprintf("2. .agents/idea-router/%ss/%s.md\n", docType, docID))
	sb.WriteString("3. .agents/idea-router/README.md\n")
	sb.WriteString("4. .agents/idea-router/DESIGN.md (for protocol rules)\n\n")
	sb.WriteString("Then act, not merely acknowledge:\n")
	sb.WriteString("- If the router item assigns implementation work in this repo, implement it.\n")
	sb.WriteString("- Keep work scoped to this repository unless the plan explicitly designates a super-agent cross-repo mandate.\n")
	sb.WriteString("- Continue until the item /goal is met, blocked by safety/user approval, or impossible with a precise reason.\n")
	sb.WriteString("- Run the relevant tests or verification commands and record the exact evidence.\n")
	sb.WriteString("- Write the result to .agents/idea-router/reviews/ or .agents/idea-router/decisions/ as appropriate.\n")
	sb.WriteString("- Update state.json: acknowledge your pending item only after reading and acting, update your last_read timestamp, and add the other agent to the pending inbox only when further review or work is required.\n")
	sb.WriteString("- Do not leave a vague next step when you can complete the work yourself.\n")
	return sb.String()
}
