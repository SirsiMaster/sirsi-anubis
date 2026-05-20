package router

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestExecutor_DispatchAPICallWake(t *testing.T) {
	r, reg, wq, _ := setupExecutorTest(t)
	var got WakePayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", req.Method)
		}
		if req.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("authorization header = %q", req.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
			t.Errorf("decode wake payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	t.Setenv("WAKE_TOKEN", "secret-token")

	reg.Agents["api-agent"] = AgentConfig{
		ID:   "api-agent",
		Type: "gemini",
		Wake: WakeConfig{Mechanism: WakeAPICall, Endpoint: server.URL, Auth: "env:WAKE_TOKEN"},
	}
	item := wq.AddItem("test-doc", "api-agent", "codex-pantheon", "api wake")
	exec := NewExecutor(reg, r, wq, io.Discard)

	if err := exec.Dispatch(context.Background(), item); err != nil {
		t.Fatalf("Dispatch() error: %v", err)
	}
	if got.AgentID != "api-agent" || got.DocID != "test-doc" {
		t.Fatalf("wake payload = %+v", got)
	}
	if wq.Items[0].Status != StatusDispatched {
		t.Errorf("status = %s, want dispatched", wq.Items[0].Status)
	}
}

func TestExecutor_RegisterAndDispatchWebhookAgent(t *testing.T) {
	r, _, wq, tmp := setupExecutorTest(t)
	routerRoot := filepath.Join(tmp, ".agents", "idea-router")

	received := make(chan WakePayload, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", req.Method)
		}
		var payload WakePayload
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Errorf("decode wake payload: %v", err)
		}
		received <- payload
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	cfg := AgentConfig{
		ID:   "test-webhook",
		Type: "webhook",
		Wake: WakeConfig{Mechanism: WakeAPICall, Endpoint: server.URL},
	}
	if err := RegisterAgent(routerRoot, cfg); err != nil {
		t.Fatalf("RegisterAgent() error: %v", err)
	}

	reg, err := LoadRegistry(routerRoot)
	if err != nil {
		t.Fatalf("LoadRegistry() error: %v", err)
	}
	registered, err := reg.Lookup("test-webhook")
	if err != nil {
		t.Fatalf("Lookup(test-webhook) error: %v", err)
	}
	if registered.Type == "claude" || registered.Type == "codex" {
		t.Fatalf("registered non-Claude/Codex proof has type %q", registered.Type)
	}

	item := wq.AddItem("test-doc", "test-webhook", "codex-pantheon", "webhook wake")
	exec := NewExecutor(reg, r, wq, io.Discard)
	if err := exec.Dispatch(context.Background(), item); err != nil {
		t.Fatalf("Dispatch() error: %v", err)
	}

	select {
	case payload := <-received:
		if payload.AgentID != "test-webhook" || payload.Type != "webhook" || payload.DocID != "test-doc" {
			t.Fatalf("wake payload = %+v", payload)
		}
	case <-time.After(time.Second):
		t.Fatal("webhook agent did not receive wake payload")
	}
	if wq.Items[0].Status != StatusDispatched {
		t.Errorf("status = %s, want dispatched", wq.Items[0].Status)
	}
}

func TestExecutor_DispatchMCPNotificationWake(t *testing.T) {
	r, reg, wq, tmp := setupExecutorTest(t)
	reg.Agents["mcp-agent"] = AgentConfig{
		ID:   "mcp-agent",
		Type: "ide-extension",
		Wake: WakeConfig{Mechanism: WakeMCPNotification, MCPServer: "sirsi"},
	}
	item := wq.AddItem("test-doc", "mcp-agent", "codex-pantheon", "mcp wake")
	exec := NewExecutor(reg, r, wq, io.Discard)

	if err := exec.Dispatch(context.Background(), item); err != nil {
		t.Fatalf("Dispatch() error: %v", err)
	}
	path := filepath.Join(tmp, ".agents", "idea-router", mcpNotificationFile)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read MCP notification outbox: %v", err)
	}
	if !contains(string(data), `"mcp_server":"sirsi"`) || !contains(string(data), `"agent_id":"mcp-agent"`) {
		t.Fatalf("notification outbox missing payload: %s", string(data))
	}
	if wq.Items[0].Status != StatusDispatched {
		t.Errorf("status = %s, want dispatched", wq.Items[0].Status)
	}
}

func TestExecutor_DirectAgentCLIAuthFailureBlocksDispatch(t *testing.T) {
	r, reg, wq, tmp := setupExecutorTest(t)
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cliPath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(cliPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	reg.Agents["direct-claude"] = AgentConfig{
		ID:      "direct-claude",
		Type:    "claude",
		Command: []string{"claude", "--print"},
		Cwd:     tmp,
	}

	item := wq.AddItem("test-doc", "direct-claude", "codex-pantheon", "test")
	exec := NewExecutor(reg, r, wq, io.Discard)
	exec.SetAuthProbe(func(cliPath, agentType string) (bool, bool, string) {
		return false, true, "Not logged in"
	})

	err := exec.Dispatch(context.Background(), item)
	if err == nil {
		t.Fatal("expected auth failure to block dispatch")
	}
	if wq.Items[0].Status != StatusBlocked {
		t.Errorf("status = %s, want blocked", wq.Items[0].Status)
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

func TestResolveExecutorTimeout(t *testing.T) {
	t.Run("default when unset", func(t *testing.T) {
		t.Setenv("SIRSI_ROUTER_EXECUTOR_TIMEOUT", "")
		if got := resolveExecutorTimeout(); got != DefaultExecutorTimeout {
			t.Errorf("got %s, want %s", got, DefaultExecutorTimeout)
		}
	})
	t.Run("override via env", func(t *testing.T) {
		t.Setenv("SIRSI_ROUTER_EXECUTOR_TIMEOUT", "45m")
		if got := resolveExecutorTimeout(); got != 45*time.Minute {
			t.Errorf("got %s, want 45m", got)
		}
	})
	t.Run("invalid env falls back to default", func(t *testing.T) {
		t.Setenv("SIRSI_ROUTER_EXECUTOR_TIMEOUT", "not-a-duration")
		if got := resolveExecutorTimeout(); got != DefaultExecutorTimeout {
			t.Errorf("got %s, want default %s", got, DefaultExecutorTimeout)
		}
	})
	t.Run("zero/negative falls back", func(t *testing.T) {
		t.Setenv("SIRSI_ROUTER_EXECUTOR_TIMEOUT", "0s")
		if got := resolveExecutorTimeout(); got != DefaultExecutorTimeout {
			t.Errorf("got %s, want default %s", got, DefaultExecutorTimeout)
		}
	})
}

func TestExecutor_SetTimeout(t *testing.T) {
	r, reg, wq, _ := setupExecutorTest(t)
	exec := NewExecutor(reg, r, wq, io.Discard)
	exec.SetTimeout(2 * time.Second)
	if exec.timeout != 2*time.Second {
		t.Errorf("timeout = %s, want 2s", exec.timeout)
	}
	exec.SetTimeout(0) // ignored
	if exec.timeout != 2*time.Second {
		t.Errorf("zero override should be ignored, got %s", exec.timeout)
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
