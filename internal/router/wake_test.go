package router

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWakeCLI_Success(t *testing.T) {
	r, tmp := setupTestRouter(t)
	cfg := &AgentConfig{
		ID:      "echo-agent",
		Type:    "test",
		Command: []string{"echo", "hello"},
		Cwd:     tmp,
	}
	reg := &Registry{Agents: map[string]AgentConfig{cfg.ID: *cfg}}
	wq, _ := LoadWorkQueue(filepath.Join(tmp, ".agents", "idea-router"))
	exec := NewExecutor(reg, r, wq, io.Discard)

	item := &WorkItem{ID: "test", DocID: "doc-1", TargetAgentID: cfg.ID}
	err := exec.wakeCLI(context.Background(), item, cfg, "test prompt")
	if err != nil {
		t.Fatalf("wakeCLI() error: %v", err)
	}
}

func TestWakeCLI_Failure(t *testing.T) {
	r, tmp := setupTestRouter(t)
	cfg := &AgentConfig{
		ID:      "fail-agent",
		Type:    "test",
		Command: []string{"false"},
		Cwd:     tmp,
	}
	reg := &Registry{Agents: map[string]AgentConfig{cfg.ID: *cfg}}
	wq, _ := LoadWorkQueue(filepath.Join(tmp, ".agents", "idea-router"))
	exec := NewExecutor(reg, r, wq, io.Discard)

	item := &WorkItem{ID: "test", DocID: "doc-1", TargetAgentID: cfg.ID}
	err := exec.wakeCLI(context.Background(), item, cfg, "test prompt")
	if err == nil {
		t.Error("expected error from failing command")
	}
}

func TestWakeAPI_Success(t *testing.T) {
	received := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Get("Authorization")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := &AgentConfig{
		ID:   "api-agent",
		Type: "gemini",
		Wake: WakeConfig{
			Mechanism: "api-call",
			Endpoint:  server.URL,
			Auth:      "bearer:test-token",
		},
	}

	r, tmp := setupTestRouter(t)
	wq, _ := LoadWorkQueue(filepath.Join(tmp, ".agents", "idea-router"))
	exec := NewExecutor(&Registry{Agents: map[string]AgentConfig{cfg.ID: *cfg}}, r, wq, io.Discard)
	exec.wakeHTTPClient = server.Client()

	item := &WorkItem{ID: "test", DocID: "doc-1", TargetAgentID: cfg.ID}
	err := exec.wakeAPI(context.Background(), item, cfg, "test prompt")
	if err != nil {
		t.Fatalf("wakeAPI() error: %v", err)
	}

	auth := <-received
	if auth != "Bearer test-token" {
		t.Errorf("auth = %q, want 'Bearer test-token'", auth)
	}
}

func TestWakeAPI_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	cfg := &AgentConfig{
		ID:   "api-agent",
		Wake: WakeConfig{Mechanism: "api-call", Endpoint: server.URL},
	}

	r, tmp := setupTestRouter(t)
	wq, _ := LoadWorkQueue(filepath.Join(tmp, ".agents", "idea-router"))
	exec := NewExecutor(&Registry{Agents: map[string]AgentConfig{cfg.ID: *cfg}}, r, wq, io.Discard)
	exec.wakeHTTPClient = server.Client()

	item := &WorkItem{ID: "test", DocID: "doc-1", TargetAgentID: cfg.ID}
	err := exec.wakeAPI(context.Background(), item, cfg, "test prompt")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestWakeMCP_WritesNotification(t *testing.T) {
	r, tmp := setupTestRouter(t)
	wq, _ := LoadWorkQueue(filepath.Join(tmp, ".agents", "idea-router"))

	cfg := &AgentConfig{
		ID:   "mcp-agent",
		Type: "cursor",
		Wake: WakeConfig{Mechanism: "mcp-notification", MCPServer: "sirsi"},
		Cwd:  tmp,
	}

	exec := NewExecutor(&Registry{Agents: map[string]AgentConfig{cfg.ID: *cfg}}, r, wq, io.Discard)
	item := &WorkItem{ID: "test", DocID: "doc-1", TargetAgentID: cfg.ID, Topic: "test-topic"}

	err := exec.wakeMCP(context.Background(), item, cfg, "test prompt")
	if err != nil {
		t.Fatalf("wakeMCP() error: %v", err)
	}

	notifPath := filepath.Join(tmp, ".agents", "idea-router", mcpNotificationFile)
	data, err := os.ReadFile(notifPath)
	if err != nil {
		t.Fatalf("notification file not created: %v", err)
	}
	if len(data) == 0 {
		t.Error("notification file is empty")
	}
}

func TestResolveWakeAuth(t *testing.T) {
	tests := []struct {
		spec    string
		wantVal string
		wantOK  bool
	}{
		{"", "", false},
		{"bearer:my-token", "my-token", true},
		{"raw-string", "raw-string", true},
	}
	for _, tt := range tests {
		val, ok := resolveWakeAuth(tt.spec)
		if val != tt.wantVal || ok != tt.wantOK {
			t.Errorf("resolveWakeAuth(%q) = (%q, %v), want (%q, %v)", tt.spec, val, ok, tt.wantVal, tt.wantOK)
		}
	}
}

func TestResolveWakeAuth_Env(t *testing.T) {
	t.Setenv("TEST_WAKE_TOKEN", "env-token-123")
	val, ok := resolveWakeAuth("env:TEST_WAKE_TOKEN")
	if !ok || val != "env-token-123" {
		t.Errorf("env auth = (%q, %v), want ('env-token-123', true)", val, ok)
	}
}

func TestWakeMechanism_Default(t *testing.T) {
	cfg := AgentConfig{ID: "test"}
	// Default is empty string, which wake() dispatches to cli-spawn
	mech := cfg.WakeMechanism()
	if mech != "" && mech != WakeCLISpawn {
		t.Errorf("default mechanism = %q, want empty or cli-spawn", mech)
	}
}

func TestWakeMechanism_Explicit(t *testing.T) {
	cfg := AgentConfig{ID: "test", Wake: WakeConfig{Mechanism: "api-call"}}
	if cfg.WakeMechanism() != WakeAPICall {
		t.Errorf("mechanism = %q, want api-call", cfg.WakeMechanism())
	}
}
