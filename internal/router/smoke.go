package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// SmokeResult holds the outcome of a single agent smoke probe.
type SmokeResult struct {
	Agent   string
	Passed  bool
	Detail  string
	Elapsed time.Duration
}

// SmokeOptions configures the smoke test.
type SmokeOptions struct {
	RepoRoot  string
	DryRun    bool
	AgentPair bool // full relay test: seed item → Claude → Codex → verify
	Out       io.Writer
	Timeout   time.Duration
}

// RunSmoke verifies that both agents can be launched and write back to the
// router directory. It creates a temporary probe file, asks each agent to
// write a token, and verifies the token exists.
//
// When AgentPair is true, it runs a full relay test: seed a router item,
// verify each agent reads it and writes back, then restore state.
func RunSmoke(ctx context.Context, opts SmokeOptions) ([]SmokeResult, error) {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Timeout == 0 {
		opts.Timeout = 90 * time.Second
	}

	if opts.AgentPair {
		return runAgentPairSmoke(ctx, opts)
	}

	routerDir := filepath.Join(opts.RepoRoot, ".agents", "idea-router")
	smokeDir := filepath.Join(routerDir, "smoke-test")
	if !opts.DryRun {
		if err := os.MkdirAll(smokeDir, 0o755); err != nil {
			return nil, fmt.Errorf("create smoke dir: %w", err)
		}
		defer os.RemoveAll(smokeDir)
	}

	var results []SmokeResult

	// Test Claude
	results = append(results, probeAgent(ctx, opts, smokeDir, "claude"))

	// Test Codex
	results = append(results, probeAgent(ctx, opts, smokeDir, "codex"))

	return results, nil
}

// runAgentPairSmoke tests the full relay:
//  1. Seed a temporary router review addressed to Claude.
//  2. Launch Claude with the router work prompt and verify it writes an artifact.
//  3. Seed a temporary router review addressed to Codex.
//  4. Launch Codex with the router work prompt and verify it writes an artifact.
//  5. Clean up all temporary artifacts and restore state.json.
func runAgentPairSmoke(ctx context.Context, opts SmokeOptions) ([]SmokeResult, error) {
	var results []SmokeResult

	if opts.DryRun {
		routerDir := filepath.Join(opts.RepoRoot, ".agents", "idea-router")
		smokeDir := filepath.Join(routerDir, "smoke-test")
		results = append(results, probeAgent(ctx, opts, smokeDir, "claude"))
		results = append(results, probeAgent(ctx, opts, smokeDir, "codex"))
		results = append(results, SmokeResult{
			Agent:   "relay:pair",
			Passed:  true,
			Detail:  "[dry-run] CLIs exist; live relay writeback was not tested",
			Elapsed: 0,
		})
		return results, nil
	}

	r, err := New(opts.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("open router: %w", err)
	}

	// Save original state for restoration
	origState, err := r.ReadState()
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	origStateJSON, _ := json.Marshal(origState)

	routerDir := filepath.Join(opts.RepoRoot, ".agents", "idea-router")
	smokeDir := filepath.Join(routerDir, "smoke-test")
	if err := os.MkdirAll(smokeDir, 0o755); err != nil {
		return nil, fmt.Errorf("create smoke dir: %w", err)
	}

	// Cleanup function: restore original state and remove temp files
	cleanup := func(artifacts []string) {
		for _, a := range artifacts {
			os.Remove(a)
		}
		os.RemoveAll(smokeDir)
		// Restore original state
		var restored State
		_ = json.Unmarshal(origStateJSON, &restored)
		_ = r.WriteState(&restored)
	}

	var artifacts []string

	// --- Step 1: Probe Claude write access ---
	claudeResult := probeAgent(ctx, opts, smokeDir, "claude")
	results = append(results, claudeResult)
	if !claudeResult.Passed {
		cleanup(artifacts)
		return results, nil
	}

	// --- Step 2: Probe Codex write access ---
	codexResult := probeAgent(ctx, opts, smokeDir, "codex")
	results = append(results, codexResult)
	if !codexResult.Passed {
		cleanup(artifacts)
		return results, nil
	}

	// --- Step 3: Test router relay — seed item for Claude ---
	// The live smoke uses the real router prompt (buildRouterPrompt) and verifies
	// real protocol side effects: artifact creation, state.json timestamp update,
	// and inbox acknowledgment — not just token-file writes.
	start := time.Now()
	smokeReviewPath := filepath.Join(routerDir, "reviews", "smoke-pair-test-claude.md")
	smokeContent := "# Review: Smoke Pair Test\n\n" +
		"- reviewer: codex\n" +
		"- addressed_to: claude\n" +
		"- related_topics: autorouter-daemon-v2\n" +
		"- verdict: approve\n" +
		"- created_at: " + time.Now().UTC().Format(time.RFC3339) + "\n\n" +
		"## Summary\n\n" +
		"This is an automated smoke test verifying the router relay protocol.\n" +
		"Claude should read this item, write a response review/decision artifact\n" +
		"to `.agents/idea-router/reviews/` or `.agents/idea-router/decisions/`,\n" +
		"update `state.json` (acknowledge the inbox item, update `last_claude_read`),\n" +
		"and exit. No implementation work required — just confirm the item and update state.\n"
	if err := os.WriteFile(smokeReviewPath, []byte(smokeContent), 0o644); err != nil {
		cleanup(artifacts)
		return nil, fmt.Errorf("write smoke review: %w", err)
	}
	artifacts = append(artifacts, smokeReviewPath)

	// Capture pre-launch state for verification
	preState, _ := r.ReadState()
	preClaudeRead := preState.LastClaudeRead
	preState.AddToInbox("claude", "smoke-pair-test-claude")
	_ = r.WriteState(preState)

	fmt.Fprintln(opts.Out, "  Seeded router item for Claude, launching...")

	prompt := buildRouterPrompt("claude", "review", "smoke-pair-test-claude")
	timeoutCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	claudeCmd := exec.CommandContext(timeoutCtx, "claude",
		"--print",
		"--permission-mode", "auto",
		prompt,
	)
	claudeCmd.Dir = opts.RepoRoot
	claudeCmd.Stdout = io.Discard
	claudeCmd.Stderr = io.Discard

	relayErr := claudeCmd.Run()
	cancel()

	if relayErr != nil {
		results = append(results, SmokeResult{
			Agent:   "relay:claude",
			Passed:  false,
			Detail:  fmt.Sprintf("Claude relay launch failed: %v", relayErr),
			Elapsed: time.Since(start),
		})
		cleanup(artifacts)
		return results, nil
	}

	// Verify router protocol side effects, not just token files:
	// 1. state.json last_claude_read timestamp advanced
	// 2. Inbox item acknowledged (removed from pending_for_claude)
	// 3. A response artifact exists in reviews/ or decisions/
	claudeRelayResult := verifyProtocolSideEffects(r, routerDir, "claude", preClaudeRead, "smoke-pair-test-claude", time.Since(start))
	results = append(results, claudeRelayResult)
	if !claudeRelayResult.Passed {
		cleanup(artifacts)
		return results, nil
	}
	// Track any artifacts Claude created for cleanup
	artifacts = append(artifacts, findSmokeArtifacts(routerDir, "smoke-pair-test")...)

	// --- Step 4: Test router relay — seed item for Codex ---
	start = time.Now()
	smokeReviewPathCodex := filepath.Join(routerDir, "reviews", "smoke-pair-test-codex.md")
	smokeContentCodex := "# Review: Smoke Pair Test\n\n" +
		"- reviewer: claude\n" +
		"- addressed_to: codex\n" +
		"- related_topics: autorouter-daemon-v2\n" +
		"- verdict: approve\n" +
		"- created_at: " + time.Now().UTC().Format(time.RFC3339) + "\n\n" +
		"## Summary\n\n" +
		"This is an automated smoke test verifying the router relay protocol.\n" +
		"Codex should read this item, write a response review/decision artifact\n" +
		"to `.agents/idea-router/reviews/` or `.agents/idea-router/decisions/`,\n" +
		"update `state.json` (acknowledge the inbox item, update `last_codex_read`),\n" +
		"and exit. No implementation work required — just confirm the item and update state.\n"
	if err := os.WriteFile(smokeReviewPathCodex, []byte(smokeContentCodex), 0o644); err != nil {
		cleanup(artifacts)
		return nil, fmt.Errorf("write smoke review: %w", err)
	}
	artifacts = append(artifacts, smokeReviewPathCodex)

	postState, _ := r.ReadState()
	preCodexRead := postState.LastCodexRead
	postState.AddToInbox("codex", "smoke-pair-test-codex")
	_ = r.WriteState(postState)

	fmt.Fprintln(opts.Out, "  Seeded router item for Codex, launching...")

	codexPrompt := buildRouterPrompt("codex", "review", "smoke-pair-test-codex")
	timeoutCtx2, cancel2 := context.WithTimeout(ctx, opts.Timeout)
	codexCmd := exec.CommandContext(timeoutCtx2, "codex",
		"exec",
		"-C", opts.RepoRoot,
		"--sandbox", "workspace-write",
		codexPrompt,
	)
	codexCmd.Dir = opts.RepoRoot
	codexCmd.Stdout = io.Discard
	codexCmd.Stderr = io.Discard

	relayErr = codexCmd.Run()
	cancel2()

	if relayErr != nil {
		results = append(results, SmokeResult{
			Agent:   "relay:codex",
			Passed:  false,
			Detail:  fmt.Sprintf("Codex relay launch failed: %v", relayErr),
			Elapsed: time.Since(start),
		})
		cleanup(artifacts)
		return results, nil
	}

	codexRelayResult := verifyProtocolSideEffects(r, routerDir, "codex", preCodexRead, "smoke-pair-test-codex", time.Since(start))
	results = append(results, codexRelayResult)
	artifacts = append(artifacts, findSmokeArtifacts(routerDir, "smoke-pair-test")...)

	// --- Step 5: Full relay passed (only if both agents passed) ---
	if codexRelayResult.Passed {
		results = append(results, SmokeResult{
			Agent:   "relay:pair",
			Passed:  true,
			Detail:  "Both agents processed router items: artifacts created, state.json updated, inboxes advanced",
			Elapsed: time.Since(start),
		})
	}

	cleanup(artifacts)
	return results, nil
}

func probeAgent(ctx context.Context, opts SmokeOptions, smokeDir, agent string) SmokeResult {
	start := time.Now()

	// Check CLI exists
	cliName := agent
	if _, err := exec.LookPath(cliName); err != nil {
		return SmokeResult{
			Agent:   agent,
			Passed:  false,
			Detail:  fmt.Sprintf("%s CLI not found in PATH", agent),
			Elapsed: time.Since(start),
		}
	}

	if opts.DryRun {
		return SmokeResult{
			Agent:   agent,
			Passed:  true,
			Detail:  fmt.Sprintf("[dry-run] %s CLI found, would probe write access", agent),
			Elapsed: time.Since(start),
		}
	}

	// Create token path the agent must write
	tokenPath := filepath.Join(smokeDir, agent+"-write-token")
	os.Remove(tokenPath) // clean slate

	// Build a prompt that just writes the token file
	prompt := fmt.Sprintf(
		"Write the exact text 'SMOKE_OK' to the file %s — no other output, no explanation. Just write that file and exit.",
		tokenPath,
	)

	timeoutCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	var cmd *exec.Cmd
	switch agent {
	case "claude":
		cmd = exec.CommandContext(timeoutCtx, "claude",
			"--print",
			"--permission-mode", "auto",
			prompt,
		)
	case "codex":
		// workspace-write covers the entire -C directory; no --add-dir needed
		// for paths inside the workspace (see: codex-router-automation-blocker).
		cmd = exec.CommandContext(timeoutCtx, "codex",
			"exec",
			"-C", opts.RepoRoot,
			"--sandbox", "workspace-write",
			prompt,
		)
	}
	cmd.Dir = opts.RepoRoot
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return SmokeResult{
			Agent:   agent,
			Passed:  false,
			Detail:  fmt.Sprintf("launch failed: %v", err),
			Elapsed: time.Since(start),
		}
	}

	// Verify the token was written
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return SmokeResult{
			Agent:   agent,
			Passed:  false,
			Detail:  "agent ran but did not write the token file (write access blocked)",
			Elapsed: time.Since(start),
		}
	}

	if len(data) == 0 {
		return SmokeResult{
			Agent:   agent,
			Passed:  false,
			Detail:  "agent wrote an empty token file",
			Elapsed: time.Since(start),
		}
	}

	return SmokeResult{
		Agent:   agent,
		Passed:  true,
		Detail:  "launched and wrote token successfully",
		Elapsed: time.Since(start),
	}
}

// verifyProtocolSideEffects checks that an agent actually performed real router
// protocol work after being launched: updated last_*_read in state.json,
// acknowledged the inbox item, and wrote a response artifact.
func verifyProtocolSideEffects(r *Router, routerDir, agent, preLastRead, seededID string, elapsed time.Duration) SmokeResult {
	agentLabel := "relay:" + agent

	// Re-read state.json to see what the agent changed
	postState, err := r.ReadState()
	if err != nil {
		return SmokeResult{Agent: agentLabel, Passed: false, Detail: fmt.Sprintf("cannot read state.json after launch: %v", err), Elapsed: elapsed}
	}

	// Check 1: last_*_read timestamp advanced
	var postLastRead string
	switch agent {
	case "claude":
		postLastRead = postState.LastClaudeRead
	case "codex":
		postLastRead = postState.LastCodexRead
	}

	timestampAdvanced := postLastRead != preLastRead && postLastRead != ""

	// Check 2: inbox item acknowledged (no longer in pending list)
	inbox := postState.InboxFor(agent)
	inboxCleared := true
	for _, id := range inbox {
		if id == seededID {
			inboxCleared = false
			break
		}
	}

	// Check 3: response artifact exists (any file matching smoke-pair-test* in reviews/ or decisions/)
	artifactFound := len(findSmokeArtifacts(routerDir, "smoke-pair-test")) > 0

	// Build detailed result
	var failures []string
	if !timestampAdvanced {
		failures = append(failures, fmt.Sprintf("last_%s_read not updated (was %q, still %q)", agent, preLastRead, postLastRead))
	}
	if !inboxCleared {
		failures = append(failures, fmt.Sprintf("%q still in %s inbox", seededID, agent))
	}
	if !artifactFound {
		failures = append(failures, "no response artifact written to reviews/ or decisions/")
	}

	if len(failures) > 0 {
		// Partial success: agent launched but didn't complete protocol.
		// Report what passed and what failed.
		detail := fmt.Sprintf("%s launched but protocol incomplete: %s", agent, joinFailures(failures))
		return SmokeResult{Agent: agentLabel, Passed: false, Detail: detail, Elapsed: elapsed}
	}

	return SmokeResult{
		Agent:   agentLabel,
		Passed:  true,
		Detail:  fmt.Sprintf("%s processed router item: artifact written, state.json updated, inbox cleared", agent),
		Elapsed: elapsed,
	}
}

// findSmokeArtifacts returns paths to any files matching the given prefix in
// reviews/ and decisions/ subdirectories. Used for cleanup and verification.
func findSmokeArtifacts(routerDir, prefix string) []string {
	var found []string
	for _, subdir := range []string{"reviews", "decisions"} {
		dir := filepath.Join(routerDir, subdir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && len(e.Name()) >= len(prefix) && e.Name()[:len(prefix)] == prefix {
				found = append(found, filepath.Join(dir, e.Name()))
			}
		}
	}
	return found
}

func joinFailures(failures []string) string {
	result := ""
	for i, f := range failures {
		if i > 0 {
			result += "; "
		}
		result += f
	}
	return result
}
