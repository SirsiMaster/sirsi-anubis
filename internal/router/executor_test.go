package router

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupExecutorTest(t *testing.T) (*Router, *Registry, *WorkQueue, string) {
	t.Helper()
	r, tmp := setupTestRouter(t)

	// Create agents.json with a fake agent that writes back
	reg := &Registry{Agents: map[string]AgentConfig{
		"fake-agent": {
			ID:      "fake-agent",
			Type:    "claude",
			Command: []string{"bash", "-c", "echo done"},
			Cwd:     tmp,
		},
		"crash-agent": {
			ID:      "crash-agent",
			Type:    "codex",
			Command: []string{"bash", "-c", "exit 1"},
			Cwd:     tmp,
		},
	}}
	SaveRegistry(filepath.Join(tmp, ".agents", "idea-router"), reg)

	wq, _ := LoadWorkQueue(filepath.Join(tmp, ".agents", "idea-router"))
	return r, reg, wq, tmp
}

func TestExecutor_DispatchSuccess(t *testing.T) {
	r, reg, wq, tmp := setupExecutorTest(t)

	// Create a writeback-simulating agent: writes state.json update
	writebackScript := `cd "` + tmp + `" && python3 -c "
import json, os
path = '.agents/idea-router/state.json'
with open(path) as f: s = json.load(f)
s['last_claude_read'] = '2099-01-01T00:00:00Z'
with open(path, 'w') as f: json.dump(s, f)
"`
	reg.Agents["writeback-agent"] = AgentConfig{
		ID:      "writeback-agent",
		Type:    "claude",
		Command: []string{"bash", "-c", writebackScript},
		Cwd:     tmp,
	}

	item := wq.AddItem("test-doc", "writeback-agent", "codex-pantheon", "test")
	exec := NewExecutor(reg, r, wq, io.Discard)
	exec.timeout = 10 * time.Second

	err := exec.Dispatch(context.Background(), item)
	if err != nil {
		t.Fatalf("Dispatch() error: %v", err)
	}

	if wq.Items[0].Status != StatusCompleted {
		t.Errorf("status = %s, want completed", wq.Items[0].Status)
	}
}

func TestExecutor_DispatchCrash(t *testing.T) {
	_, reg, wq, _ := setupExecutorTest(t)

	r, _ := setupTestRouter(t)
	item := wq.AddItem("test-doc", "crash-agent", "claude-pantheon", "test")
	exec := NewExecutor(reg, r, wq, io.Discard)
	exec.timeout = 5 * time.Second

	err := exec.Dispatch(context.Background(), item)
	if err == nil {
		t.Error("expected error for crashing agent")
	}

	if wq.Items[0].Status != StatusFailed {
		t.Errorf("status = %s, want failed", wq.Items[0].Status)
	}
	if len(wq.Items[0].Attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", len(wq.Items[0].Attempts))
	}
	if wq.Items[0].Attempts[0].ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", wq.Items[0].Attempts[0].ExitCode)
	}
}

func TestExecutor_UnregisteredAgent(t *testing.T) {
	r, reg, wq, _ := setupExecutorTest(t)

	item := wq.AddItem("test-doc", "nonexistent-agent", "claude-pantheon", "test")
	exec := NewExecutor(reg, r, wq, io.Discard)

	err := exec.Dispatch(context.Background(), item)
	if err == nil {
		t.Error("expected error for unregistered agent")
	}
	if wq.Items[0].Status != StatusFailed {
		t.Errorf("status = %s, want failed", wq.Items[0].Status)
	}
}

func TestExecutor_NoWriteback(t *testing.T) {
	r, reg, wq, _ := setupExecutorTest(t)

	// fake-agent exits clean but doesn't write to state.json
	item := wq.AddItem("test-doc", "fake-agent", "codex-pantheon", "test")
	exec := NewExecutor(reg, r, wq, io.Discard)
	exec.timeout = 5 * time.Second

	err := exec.Dispatch(context.Background(), item)
	if err == nil {
		t.Error("expected error for no writeback")
	}
	if wq.Items[0].Status != StatusFailed {
		t.Errorf("status = %s, want failed (no writeback)", wq.Items[0].Status)
	}
}

func TestExecutor_Timeout(t *testing.T) {
	r, reg, wq, tmp := setupExecutorTest(t)

	reg.Agents["slow-agent"] = AgentConfig{
		ID:      "slow-agent",
		Type:    "claude",
		Command: []string{"sleep", "60"},
		Cwd:     tmp,
	}

	item := wq.AddItem("test-doc", "slow-agent", "codex-pantheon", "test")
	exec := NewExecutor(reg, r, wq, io.Discard)
	exec.timeout = 1 * time.Second

	err := exec.Dispatch(context.Background(), item)
	if err == nil {
		t.Error("expected error for timeout")
	}
	if wq.Items[0].Status != StatusFailed {
		t.Errorf("status = %s, want failed (timeout)", wq.Items[0].Status)
	}
}

func TestBuildWorkPrompt(t *testing.T) {
	item := &WorkItem{
		DocID:         "test-doc",
		TargetAgentID: "claude-pantheon",
		Topic:         "test-topic",
		Goal:          "make it work",
	}
	cfg := &AgentConfig{
		ID:   "claude-pantheon",
		Type: "claude",
		Cwd:  "/tmp/test",
	}

	prompt := buildWorkPrompt(item, cfg)
	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}

	checks := []string{"claude-pantheon", "test-doc", "test-topic", "make it work", "state.json", "agents.json"}
	for _, check := range checks {
		if !contains(prompt, check) {
			t.Errorf("prompt missing %q", check)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && os.Getenv("_") != "impossible" && // avoid unused import
		len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
