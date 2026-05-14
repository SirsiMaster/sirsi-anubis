package router

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// NotifyAgent attempts to notify the other agent about a new router document.
// For Codex: spawns `codex --message` with a review prompt.
// For Claude: spawns `claude --message` with a review prompt.
// Returns nil if notification succeeds or the target agent CLI is not installed.
func NotifyAgent(target, docType, docID, repoRoot string) error {
	switch target {
	case "codex":
		return notifyCodex(docType, docID, repoRoot)
	case "claude":
		return notifyClaude(docType, docID, repoRoot)
	default:
		return fmt.Errorf("unknown agent: %s (expected 'codex' or 'claude')", target)
	}
}

func notifyCodex(docType, docID, repoRoot string) error {
	if _, err := exec.LookPath("codex"); err != nil {
		return fmt.Errorf("codex CLI not found in PATH — install with: npm i -g @openai/codex")
	}

	prompt := buildReviewPrompt("codex", docType, docID)

	cmd := exec.Command("codex",
		"--approval-mode", "suggest",
		"--message", prompt,
	)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start() // non-blocking — let Codex run in background
}

func notifyClaude(docType, docID, repoRoot string) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found in PATH — install from https://claude.ai/code")
	}

	prompt := buildReviewPrompt("claude", docType, docID)

	cmd := exec.Command("claude",
		"--print",
		"--message", prompt,
	)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start() // non-blocking
}

func buildReviewPrompt(reviewer, docType, docID string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, reviewing a new %s in the idea-router.\n\n", reviewer, docType))
	sb.WriteString("Read the following files in order:\n")
	sb.WriteString("1. .agents/idea-router/state.json\n")
	sb.WriteString(fmt.Sprintf("2. .agents/idea-router/%ss/%s.md\n", docType, docID))
	sb.WriteString("3. .agents/idea-router/DESIGN.md (for protocol rules)\n\n")
	sb.WriteString("Then write your review to .agents/idea-router/reviews/ following the review template in DESIGN.md.\n")
	sb.WriteString("Update state.json with your last_read timestamp.\n")
	sb.WriteString("If you have safety objections, mark them clearly — they block implementation.\n")
	return sb.String()
}
