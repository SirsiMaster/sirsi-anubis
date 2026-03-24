package platform

import (
	"os"
	"testing"
)

// clearIDEEnv saves and clears all IDE-related env vars so tests
// can set exactly the vars they need without interference from
// the real environment (e.g., running inside VS Code).
func clearIDEEnv(t *testing.T) {
	t.Helper()
	vscodeVars := []string{"TERM_PROGRAM", "TERM_PROGRAM_VERSION", "VSCODE_PID",
		"VSCODE_IPC_HOOK", "VSCODE_CLI_VERSION",
		"CURSOR_TRACE_DIR", "CURSOR_CHANNEL",
		"JETBRAINS_IDE", "TERMINAL_EMULATOR", "__INTELLIJ_COMMAND_HISTFILE__",
		"NVIM", "NVIM_LISTEN_ADDRESS", "VIM", "VIMRUNTIME",
		"INSIDE_EMACS", "ZED_TERM"}

	saved := make(map[string]string)
	for _, v := range vscodeVars {
		if val, ok := os.LookupEnv(v); ok {
			saved[v] = val
			os.Unsetenv(v)
		}
	}
	t.Cleanup(func() {
		for k, v := range saved {
			os.Setenv(k, v)
		}
	})
}

// ─── IDE Detection ───────────────────────────────────────────────────────────

func TestDetectIDE(t *testing.T) {
	ide := DetectIDE()
	if ide == nil {
		t.Fatal("DetectIDE returned nil")
	}
	t.Logf("Detected IDE: %s (id=%s, family=%s)", ide.Name, ide.ID, ide.Family)
}

func TestDetectIDE_VSCode(t *testing.T) {
	os.Setenv("VSCODE_PID", "12345")
	defer os.Unsetenv("VSCODE_PID")

	ide := detectIDEFromEnv()
	if ide == nil || ide.ID != "vscode" {
		t.Errorf("expected vscode, got %v", ide)
	}
}

func TestDetectIDE_VSCodeTermProgram(t *testing.T) {
	os.Setenv("TERM_PROGRAM", "vscode")
	os.Setenv("TERM_PROGRAM_VERSION", "1.90.0")
	defer os.Unsetenv("TERM_PROGRAM")
	defer os.Unsetenv("TERM_PROGRAM_VERSION")

	ide := detectIDEFromEnv()
	if ide == nil || ide.ID != "vscode" {
		t.Errorf("expected vscode, got %v", ide)
	}
	if ide != nil && ide.Version != "1.90.0" {
		t.Errorf("expected version 1.90.0, got %q", ide.Version)
	}
}

func TestDetectIDE_Cursor(t *testing.T) {
	clearIDEEnv(t)
	os.Setenv("CURSOR_TRACE_DIR", "/tmp/cursor")
	defer os.Unsetenv("CURSOR_TRACE_DIR")

	ide := detectIDEFromEnv()
	if ide == nil || ide.ID != "cursor" {
		t.Errorf("expected cursor, got %v", ide)
	}
}

func TestDetectIDE_JetBrains(t *testing.T) {
	clearIDEEnv(t)
	os.Setenv("JETBRAINS_IDE", "GoLand")
	defer os.Unsetenv("JETBRAINS_IDE")

	ide := detectIDEFromEnv()
	if ide == nil || ide.ID != "goland" {
		t.Errorf("expected goland, got %v", ide)
	}
}

func TestDetectIDE_JetBrainsTerminal(t *testing.T) {
	clearIDEEnv(t)
	os.Setenv("TERMINAL_EMULATOR", "JetBrains-JediTerm")
	defer os.Unsetenv("TERMINAL_EMULATOR")

	ide := detectIDEFromEnv()
	if ide == nil || ide.Family != "jetbrains" {
		t.Errorf("expected jetbrains family, got %v", ide)
	}
}

func TestDetectIDE_IntelliJHistFile(t *testing.T) {
	clearIDEEnv(t)
	os.Setenv("__INTELLIJ_COMMAND_HISTFILE__", "/Users/test/Library/Caches/JetBrains/GoLand2024.3")
	defer os.Unsetenv("__INTELLIJ_COMMAND_HISTFILE__")

	ide := detectIDEFromEnv()
	if ide == nil || ide.ID != "goland" {
		t.Errorf("expected goland from histfile, got %v", ide)
	}
}

func TestDetectIDE_Neovim(t *testing.T) {
	os.Setenv("NVIM", "true")
	defer os.Unsetenv("NVIM")

	ide := detectIDEFromTerminalEnv()
	if ide == nil || ide.ID != "neovim" {
		t.Errorf("expected neovim, got %v", ide)
	}
}

func TestDetectIDE_Vim(t *testing.T) {
	os.Setenv("VIMRUNTIME", "/usr/share/vim/vim90")
	defer os.Unsetenv("VIMRUNTIME")

	ide := detectIDEFromTerminalEnv()
	if ide == nil || ide.ID != "vim" {
		t.Errorf("expected vim, got %v", ide)
	}
}

func TestDetectIDE_Emacs(t *testing.T) {
	os.Setenv("INSIDE_EMACS", "29.1,comint")
	defer os.Unsetenv("INSIDE_EMACS")

	ide := detectIDEFromTerminalEnv()
	if ide == nil || ide.ID != "emacs" {
		t.Errorf("expected emacs, got %v", ide)
	}
}

func TestDetectIDE_Zed(t *testing.T) {
	os.Setenv("ZED_TERM", "true")
	defer os.Unsetenv("ZED_TERM")

	ide := detectIDEFromTerminalEnv()
	if ide == nil || ide.ID != "zed" {
		t.Errorf("expected zed, got %v", ide)
	}
}

func TestDetectIDEFromProcesses(t *testing.T) {
	// Just verify it doesn't panic — actual detection depends on running apps
	ide := detectIDEFromProcesses()
	if ide != nil {
		t.Logf("Detected IDE from processes: %s", ide.Name)
	}
}

// ─── mapJetBrainsIDE ────────────────────────────────────────────────────────

func TestMapJetBrainsIDE(t *testing.T) {
	tests := []struct {
		input    string
		wantID   string
		wantName string
	}{
		{"GoLand", "goland", "GoLand"},
		{"IntelliJIdea", "intellij", "IntelliJ IDEA"},
		{"WebStorm", "webstorm", "WebStorm"},
		{"PyCharm", "pycharm", "PyCharm"},
		{"PhpStorm", "phpstorm", "PhpStorm"},
		{"Rider", "rider", "Rider"},
		{"CLion", "clion", "CLion"},
		{"RustRover", "rustrover", "RustRover"},
		{"Fleet", "fleet", "Fleet"},
		{"SomeNewIDE", "jetbrains", "JetBrains IDE"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ide := mapJetBrainsIDE(tt.input)
			if ide.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", ide.ID, tt.wantID)
			}
			if ide.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", ide.Name, tt.wantName)
			}
		})
	}
}

// ─── detectJetBrainsFromHistFile ────────────────────────────────────────────

func TestDetectJetBrainsFromHistFile(t *testing.T) {
	tests := []struct {
		path   string
		wantID string
	}{
		{"/Users/test/Library/Caches/JetBrains/GoLand2024.3/terminal/history", "goland"},
		{"/Users/test/.cache/JetBrains/IntelliJIdea2024.2/terminal/history", "intellij"},
		{"/Users/test/.cache/JetBrains/WebStorm/terminal/history", "webstorm"},
		{"/Users/test/.cache/JetBrains/PyCharm/terminal/history", "pycharm"},
		{"/Users/test/.cache/JetBrains/Unknown/terminal/history", "jetbrains"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.wantID, func(t *testing.T) {
			ide := detectJetBrainsFromHistFile(tt.path)
			if ide.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", ide.ID, tt.wantID)
			}
		})
	}
}

// ─── CI/CD Detection ────────────────────────────────────────────────────────

func TestDetectCI_None(t *testing.T) {
	// In a local dev environment, CI should be nil
	ci := DetectCI()
	if ci != nil {
		t.Logf("CI detected (might be running in CI): %s", ci.Name)
	}
}

func TestDetectCI_GitHubActions(t *testing.T) {
	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("GITHUB_RUN_ID", "12345")
	os.Setenv("GITHUB_REF_NAME", "main")
	os.Setenv("GITHUB_SHA", "abc123")
	os.Setenv("GITHUB_SERVER_URL", "https://github.com")
	os.Setenv("GITHUB_REPOSITORY", "SirsiMaster/sirsi-pantheon")
	defer func() {
		os.Unsetenv("GITHUB_ACTIONS")
		os.Unsetenv("GITHUB_RUN_ID")
		os.Unsetenv("GITHUB_REF_NAME")
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("GITHUB_SERVER_URL")
		os.Unsetenv("GITHUB_REPOSITORY")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "github_actions" {
		t.Errorf("expected github_actions, got %v", ci)
	}
	if ci != nil && ci.BuildID != "12345" {
		t.Errorf("BuildID = %q, want 12345", ci.BuildID)
	}
}

func TestDetectCI_Jenkins(t *testing.T) {
	os.Setenv("JENKINS_URL", "https://jenkins.example.com")
	os.Setenv("BUILD_NUMBER", "42")
	os.Setenv("GIT_BRANCH", "main")
	os.Setenv("GIT_COMMIT", "def456")
	defer func() {
		os.Unsetenv("JENKINS_URL")
		os.Unsetenv("BUILD_NUMBER")
		os.Unsetenv("GIT_BRANCH")
		os.Unsetenv("GIT_COMMIT")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "jenkins" {
		t.Errorf("expected jenkins, got %v", ci)
	}
}

func TestDetectCI_TeamCity(t *testing.T) {
	os.Setenv("TEAMCITY_VERSION", "2024.1")
	os.Setenv("BUILD_VCS_NUMBER", "abc123")
	defer func() {
		os.Unsetenv("TEAMCITY_VERSION")
		os.Unsetenv("BUILD_VCS_NUMBER")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "teamcity" {
		t.Errorf("expected teamcity, got %v", ci)
	}
	if ci != nil && ci.Family != "jetbrains" {
		t.Errorf("expected family=jetbrains, got %q", ci.Family)
	}
}

func TestDetectCI_JetBrainsSpace(t *testing.T) {
	os.Setenv("JB_SPACE_EXECUTION_NUMBER", "7")
	os.Setenv("JB_SPACE_GIT_BRANCH", "feature/test")
	defer func() {
		os.Unsetenv("JB_SPACE_EXECUTION_NUMBER")
		os.Unsetenv("JB_SPACE_GIT_BRANCH")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "jetbrains_space" {
		t.Errorf("expected jetbrains_space, got %v", ci)
	}
}

func TestDetectCI_GitLab(t *testing.T) {
	os.Setenv("GITLAB_CI", "true")
	os.Setenv("CI_JOB_ID", "9999")
	defer func() {
		os.Unsetenv("GITLAB_CI")
		os.Unsetenv("CI_JOB_ID")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "gitlab_ci" {
		t.Errorf("expected gitlab_ci, got %v", ci)
	}
}

func TestDetectCI_CircleCI(t *testing.T) {
	os.Setenv("CIRCLECI", "true")
	os.Setenv("CIRCLE_BUILD_NUM", "100")
	defer func() {
		os.Unsetenv("CIRCLECI")
		os.Unsetenv("CIRCLE_BUILD_NUM")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "circleci" {
		t.Errorf("expected circleci, got %v", ci)
	}
}

func TestDetectCI_Travis(t *testing.T) {
	os.Setenv("TRAVIS", "true")
	os.Setenv("TRAVIS_BUILD_ID", "555")
	defer func() {
		os.Unsetenv("TRAVIS")
		os.Unsetenv("TRAVIS_BUILD_ID")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "travis" {
		t.Errorf("expected travis, got %v", ci)
	}
}

func TestDetectCI_Bitbucket(t *testing.T) {
	os.Setenv("BITBUCKET_PIPELINE_UUID", "{uuid}")
	os.Setenv("BITBUCKET_BUILD_NUMBER", "10")
	defer func() {
		os.Unsetenv("BITBUCKET_PIPELINE_UUID")
		os.Unsetenv("BITBUCKET_BUILD_NUMBER")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "bitbucket" {
		t.Errorf("expected bitbucket, got %v", ci)
	}
}

func TestDetectCI_Azure(t *testing.T) {
	os.Setenv("TF_BUILD", "True")
	os.Setenv("BUILD_BUILDID", "999")
	defer func() {
		os.Unsetenv("TF_BUILD")
		os.Unsetenv("BUILD_BUILDID")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "azure_pipelines" {
		t.Errorf("expected azure_pipelines, got %v", ci)
	}
}

func TestDetectCI_Buildkite(t *testing.T) {
	os.Setenv("BUILDKITE", "true")
	os.Setenv("BUILDKITE_BUILD_NUMBER", "77")
	defer func() {
		os.Unsetenv("BUILDKITE")
		os.Unsetenv("BUILDKITE_BUILD_NUMBER")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "buildkite" {
		t.Errorf("expected buildkite, got %v", ci)
	}
}

func TestDetectCI_Drone(t *testing.T) {
	os.Setenv("DRONE", "true")
	os.Setenv("DRONE_BUILD_NUMBER", "33")
	defer func() {
		os.Unsetenv("DRONE")
		os.Unsetenv("DRONE_BUILD_NUMBER")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "drone" {
		t.Errorf("expected drone, got %v", ci)
	}
}

func TestDetectCI_AWSCodeBuild(t *testing.T) {
	os.Setenv("CODEBUILD_BUILD_ID", "build:123")
	defer os.Unsetenv("CODEBUILD_BUILD_ID")

	ci := DetectCI()
	if ci == nil || ci.ID != "aws_codebuild" {
		t.Errorf("expected aws_codebuild, got %v", ci)
	}
}

func TestDetectCI_Semaphore(t *testing.T) {
	os.Setenv("SEMAPHORE", "true")
	os.Setenv("SEMAPHORE_WORKFLOW_ID", "wf-123")
	defer func() {
		os.Unsetenv("SEMAPHORE")
		os.Unsetenv("SEMAPHORE_WORKFLOW_ID")
	}()

	ci := DetectCI()
	if ci == nil || ci.ID != "semaphore" {
		t.Errorf("expected semaphore, got %v", ci)
	}
}

func TestDetectCI_Harness(t *testing.T) {
	os.Setenv("HARNESS_BUILD_ID", "h-456")
	defer os.Unsetenv("HARNESS_BUILD_ID")

	ci := DetectCI()
	if ci == nil || ci.ID != "harness" {
		t.Errorf("expected harness, got %v", ci)
	}
}

// ─── AI Agent Detection ────────────────────────────────────────────────────

func TestDetectAIAgent(t *testing.T) {
	agent := DetectAIAgent()
	if agent != nil {
		t.Logf("Detected AI agent: %s (family=%s)", agent.Name, agent.Family)
	} else {
		t.Log("No AI agent detected")
	}
}

func TestDetectAIAgent_Antigravity(t *testing.T) {
	os.Setenv("ANTIGRAVITY_SESSION_ID", "sess-123")
	os.Setenv("ANTIGRAVITY_VERSION", "1.0.0")
	defer func() {
		os.Unsetenv("ANTIGRAVITY_SESSION_ID")
		os.Unsetenv("ANTIGRAVITY_VERSION")
	}()

	agent := DetectAIAgent()
	if agent == nil || agent.ID != "antigravity" {
		t.Errorf("expected antigravity, got %v", agent)
	}
	if agent != nil && agent.Family != "google" {
		t.Errorf("expected family=google, got %q", agent.Family)
	}
}

func TestDetectAIAgent_ClaudeCode(t *testing.T) {
	os.Setenv("CLAUDE_CODE", "true")
	defer os.Unsetenv("CLAUDE_CODE")

	agent := DetectAIAgent()
	if agent == nil || agent.ID != "claude_code" {
		t.Errorf("expected claude_code, got %v", agent)
	}
}

func TestDetectAIAgent_Copilot(t *testing.T) {
	os.Setenv("GITHUB_COPILOT", "true")
	defer os.Unsetenv("GITHUB_COPILOT")

	agent := DetectAIAgent()
	if agent == nil || agent.ID != "copilot" {
		t.Errorf("expected copilot, got %v", agent)
	}
}

func TestDetectAIAgent_GeminiCLI(t *testing.T) {
	os.Setenv("GEMINI_CLI", "true")
	defer os.Unsetenv("GEMINI_CLI")

	agent := DetectAIAgent()
	if agent == nil || agent.ID != "gemini_cli" {
		t.Errorf("expected gemini_cli, got %v", agent)
	}
}

func TestDetectAIAgent_Aider(t *testing.T) {
	os.Setenv("AIDER_MODEL", "gpt-4o")
	defer os.Unsetenv("AIDER_MODEL")

	agent := DetectAIAgent()
	if agent == nil || agent.ID != "aider" {
		t.Errorf("expected aider, got %v", agent)
	}
}

// ─── IsCI ───────────────────────────────────────────────────────────────────

func TestIsCI_Local(t *testing.T) {
	// Should return false on a dev machine
	result := IsCI()
	t.Logf("IsCI() = %v", result)
}

// ─── Full Environment ───────────────────────────────────────────────────────

func TestDetectEnvironment(t *testing.T) {
	env := DetectEnvironment()
	if env == nil {
		t.Fatal("DetectEnvironment returned nil")
	}
	if env.OS == "" {
		t.Error("OS should not be empty")
	}
	if env.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if env.IDE == nil {
		t.Error("IDE should not be nil")
	}
	if env.HomeDir == "" {
		t.Error("HomeDir should not be empty")
	}
	if env.Shell == "" {
		t.Log("Shell is empty (may be expected in CI)")
	}

	t.Logf("Environment: OS=%s, Arch=%s, IDE=%s, CI=%v, Shell=%s",
		env.OS, env.Arch, env.IDE.Name, env.IsCI, env.Shell)
}

// ─── coalesce helper ────────────────────────────────────────────────────────

func TestCoalesce(t *testing.T) {
	if coalesce("", "", "c") != "c" {
		t.Error("expected 'c'")
	}
	if coalesce("a", "b") != "a" {
		t.Error("expected 'a'")
	}
	if coalesce("", "") != "" {
		t.Error("expected empty")
	}
}
