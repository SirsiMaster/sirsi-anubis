package platform

import (
	"testing"
)

// ─── IDE Detection ───────────────────────────────────────────────────────────

func TestDetectIDEWith_MockVSCode(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"VSCODE_PID": "12345",
		},
	}
	ide := DetectIDEWith(m)
	if ide == nil || ide.ID != "vscode" {
		t.Errorf("expected vscode, got %v", ide)
	}
}

func TestDetectIDEWith_MockCursor(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"CURSOR_TRACE_DIR": "/tmp/cursor",
		},
	}
	ide := DetectIDEWith(m)
	if ide == nil || ide.ID != "cursor" {
		t.Errorf("expected cursor, got %v", ide)
	}
}

func TestDetectIDEWith_MockJetBrains(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"JETBRAINS_IDE": "GoLand",
		},
	}
	ide := DetectIDEWith(m)
	if ide == nil || ide.ID != "goland" {
		t.Errorf("expected goland, got %v", ide)
	}
}

func TestDetectIDEWith_MockNeovim(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"NVIM": "true",
		},
	}
	ide := DetectIDEWith(m)
	if ide == nil || ide.ID != "neovim" {
		t.Errorf("expected neovim, got %v", ide)
	}
}

func TestDetectIDEWith_MockEmacs(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"INSIDE_EMACS": "29.1,comint",
		},
	}
	ide := DetectIDEWith(m)
	if ide == nil || ide.ID != "emacs" {
		t.Errorf("expected emacs, got %v", ide)
	}
}

func TestDetectIDEWith_MockProcesses(t *testing.T) {
	m := &Mock{
		NameStr:     "darwin",
		ProcessList: []string{"Visual Studio Code", "System Events"},
	}
	ide := DetectIDEWith(m)
	if ide == nil || ide.ID != "vscode" {
		t.Errorf("expected vscode from processes, got %v", ide)
	}
}

// ─── CI/CD Detection ────────────────────────────────────────────────────────

func TestDetectCIWith_MockGitHubActions(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"GITHUB_ACTIONS": "true",
			"GITHUB_RUN_ID":  "12345",
		},
	}
	ci := DetectCIWith(m)
	if ci == nil || ci.ID != "github_actions" {
		t.Errorf("expected github_actions, got %v", ci)
	}
	if ci != nil && ci.BuildID != "12345" {
		t.Errorf("BuildID = %q, want 12345", ci.BuildID)
	}
}

func TestDetectCIWith_MockJenkins(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"JENKINS_URL":  "https://jenkins.example.com",
			"BUILD_NUMBER": "42",
		},
	}
	ci := DetectCIWith(m)
	if ci == nil || ci.ID != "jenkins" {
		t.Errorf("expected jenkins, got %v", ci)
	}
}

// ─── AI Agent Detection ────────────────────────────────────────────────────

func TestDetectAIAgentWith_MockAntigravity(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"ANTIGRAVITY_SESSION_ID": "sess-123",
			"ANTIGRAVITY_VERSION":    "1.0.0",
		},
	}
	agent := DetectAIAgentWith(m)
	if agent == nil || agent.ID != "antigravity" {
		t.Errorf("expected antigravity, got %v", agent)
	}
	if agent != nil && agent.Family != "google" {
		t.Errorf("expected family=google, got %q", agent.Family)
	}
}

func TestDetectAIAgentWith_MockClaudeCode(t *testing.T) {
	m := &Mock{
		Env: map[string]string{
			"CLAUDE_CODE": "true",
		},
	}
	agent := DetectAIAgentWith(m)
	if agent == nil || agent.ID != "claude_code" {
		t.Errorf("expected claude_code, got %v", agent)
	}
}

func TestDetectAIAgentWith_MockProcesses(t *testing.T) {
	m := &Mock{
		NameStr:     "darwin",
		ProcessList: []string{"claude-code", "zsh"},
	}
	agent := DetectAIAgentWith(m)
	if agent == nil || agent.ID != "claude_code" {
		t.Errorf("expected claude_code from processes, got %v", agent)
	}
}

// ─── Full Environment ───────────────────────────────────────────────────────

func TestDetectEnvironmentWith_Mock(t *testing.T) {
	m := &Mock{
		NameStr: "darwin",
		Env: map[string]string{
			"VSCODE_PID": "12345",
			"SHELL":      "/bin/zsh",
		},
		HomeDir: "/users/mock",
		WorkDir: "/work/mock",
	}

	env := DetectEnvironmentWith(m)
	if env == nil {
		t.Fatal("DetectEnvironmentWith returned nil")
	}
	if env.IDE == nil || env.IDE.ID != "vscode" {
		t.Errorf("expected vscode, got %v", env.IDE)
	}
	if env.HomeDir != "/users/mock" {
		t.Errorf("HomeDir = %q, want %q", env.HomeDir, "/users/mock")
	}
	if env.Shell != "zsh" {
		t.Errorf("Shell = %q, want %q", env.Shell, "zsh")
	}
}

// ─── coalesce helper ────────────────────────────────────────────────────────

func TestCoalesce(t *testing.T) {
	if coalesce("", "", "c") != "c" {
		t.Error("expected 'c'")
	}
	if coalesce("a", "b") != "a" {
		t.Error("expected 'a'")
	}
}
