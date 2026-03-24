package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ─── IDE Detection ───────────────────────────────────────────────────────────

// IDE represents a detected code editor or IDE.
type IDE struct {
	Name    string `json:"name"`    // Human-readable name (e.g., "VS Code", "GoLand")
	ID      string `json:"id"`      // Machine identifier (e.g., "vscode", "goland")
	Family  string `json:"family"`  // Vendor family (e.g., "jetbrains", "microsoft", "vim")
	Version string `json:"version"` // Version if detectable
	Running bool   `json:"running"` // Whether it's currently active
}

// DetectIDE detects the active IDE/editor environment.
// Uses environment variables, parent process inspection, and running process detection.
func DetectIDE() *IDE {
	// Priority 1: Environment variable signals (most reliable)
	if ide := detectIDEFromEnv(); ide != nil {
		return ide
	}

	// Priority 2: Terminal/editor-specific env vars
	if ide := detectIDEFromTerminalEnv(); ide != nil {
		return ide
	}

	// Priority 3: Running process detection
	if ide := detectIDEFromProcesses(); ide != nil {
		return ide
	}

	return &IDE{Name: "Unknown", ID: "unknown", Family: "unknown"}
}

// detectIDEFromEnv checks explicit IDE environment variables.
func detectIDEFromEnv() *IDE {
	// VS Code / Cursor / Windsurf (Code-based editors)
	if term := os.Getenv("TERM_PROGRAM"); term != "" {
		switch strings.ToLower(term) {
		case "vscode":
			return &IDE{
				Name:    "VS Code",
				ID:      "vscode",
				Family:  "microsoft",
				Version: os.Getenv("TERM_PROGRAM_VERSION"),
				Running: true,
			}
		}
	}

	// VS Code sets VSCODE_* env vars in its integrated terminal
	if os.Getenv("VSCODE_PID") != "" || os.Getenv("VSCODE_IPC_HOOK") != "" {
		return &IDE{
			Name:    "VS Code",
			ID:      "vscode",
			Family:  "microsoft",
			Version: os.Getenv("VSCODE_CLI_VERSION"),
			Running: true,
		}
	}

	// Cursor (VS Code fork)
	if os.Getenv("CURSOR_TRACE_DIR") != "" || os.Getenv("CURSOR_CHANNEL") != "" {
		return &IDE{
			Name:    "Cursor",
			ID:      "cursor",
			Family:  "cursor",
			Running: true,
		}
	}

	// JetBrains IDEs set TERMINAL_EMULATOR and JETBRAINS_IDE
	if jetbrainsIDE := os.Getenv("JETBRAINS_IDE"); jetbrainsIDE != "" {
		return mapJetBrainsIDE(jetbrainsIDE)
	}
	if termEmu := os.Getenv("TERMINAL_EMULATOR"); strings.Contains(strings.ToLower(termEmu), "jetbrains") {
		return &IDE{
			Name:    "JetBrains IDE",
			ID:      "jetbrains",
			Family:  "jetbrains",
			Running: true,
		}
	}

	// IntelliJ-based IDEs also set __INTELLIJ_COMMAND_HISTFILE__
	if os.Getenv("__INTELLIJ_COMMAND_HISTFILE__") != "" {
		return detectJetBrainsFromHistFile(os.Getenv("__INTELLIJ_COMMAND_HISTFILE__"))
	}

	return nil
}

// detectIDEFromTerminalEnv checks for editor-specific terminal environments.
func detectIDEFromTerminalEnv() *IDE {
	// Vim/Neovim — check VIM or NVIM env
	if os.Getenv("NVIM") != "" || os.Getenv("NVIM_LISTEN_ADDRESS") != "" {
		return &IDE{Name: "Neovim", ID: "neovim", Family: "vim", Running: true}
	}
	if os.Getenv("VIM") != "" || os.Getenv("VIMRUNTIME") != "" {
		return &IDE{Name: "Vim", ID: "vim", Family: "vim", Running: true}
	}

	// Emacs — check INSIDE_EMACS env
	if os.Getenv("INSIDE_EMACS") != "" {
		return &IDE{Name: "Emacs", ID: "emacs", Family: "emacs", Running: true}
	}

	// Zed editor
	if os.Getenv("ZED_TERM") != "" {
		return &IDE{Name: "Zed", ID: "zed", Family: "zed", Running: true}
	}

	return nil
}

// detectIDEFromProcesses checks running processes for known IDE processes.
func detectIDEFromProcesses() *IDE {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return nil
	}

	out, err := exec.Command("ps", "-eo", "comm").Output()
	if err != nil {
		return nil
	}

	procs := strings.ToLower(string(out))

	// JetBrains family (check specific IDEs first)
	jetbrainsMap := map[string]*IDE{
		"goland":     {Name: "GoLand", ID: "goland", Family: "jetbrains", Running: true},
		"idea":       {Name: "IntelliJ IDEA", ID: "intellij", Family: "jetbrains", Running: true},
		"webstorm":   {Name: "WebStorm", ID: "webstorm", Family: "jetbrains", Running: true},
		"pycharm":    {Name: "PyCharm", ID: "pycharm", Family: "jetbrains", Running: true},
		"phpstorm":   {Name: "PhpStorm", ID: "phpstorm", Family: "jetbrains", Running: true},
		"rider":      {Name: "Rider", ID: "rider", Family: "jetbrains", Running: true},
		"rubymine":   {Name: "RubyMine", ID: "rubymine", Family: "jetbrains", Running: true},
		"clion":      {Name: "CLion", ID: "clion", Family: "jetbrains", Running: true},
		"datagrip":   {Name: "DataGrip", ID: "datagrip", Family: "jetbrains", Running: true},
		"fleet":      {Name: "Fleet", ID: "fleet", Family: "jetbrains", Running: true},
		"dataspell":  {Name: "DataSpell", ID: "dataspell", Family: "jetbrains", Running: true},
		"appcode":    {Name: "AppCode", ID: "appcode", Family: "jetbrains", Running: true},
		"aqua":       {Name: "Aqua", ID: "aqua", Family: "jetbrains", Running: true},
		"rustrover":  {Name: "RustRover", ID: "rustrover", Family: "jetbrains", Running: true},
		"writerside": {Name: "Writerside", ID: "writerside", Family: "jetbrains", Running: true},
	}
	for key, ide := range jetbrainsMap {
		if strings.Contains(procs, key) {
			return ide
		}
	}

	// Other IDEs
	if strings.Contains(procs, "cursor") {
		return &IDE{Name: "Cursor", ID: "cursor", Family: "cursor", Running: true}
	}
	if strings.Contains(procs, "windsurf") {
		return &IDE{Name: "Windsurf", ID: "windsurf", Family: "codeium", Running: true}
	}
	if strings.Contains(procs, "zed") {
		return &IDE{Name: "Zed", ID: "zed", Family: "zed", Running: true}
	}
	if strings.Contains(procs, "sublime_text") || strings.Contains(procs, "subl") {
		return &IDE{Name: "Sublime Text", ID: "sublime", Family: "sublime", Running: true}
	}
	if strings.Contains(procs, "atom") {
		return &IDE{Name: "Atom", ID: "atom", Family: "github", Running: true}
	}
	if strings.Contains(procs, "code") || strings.Contains(procs, "electron") {
		// Be careful — "code" is generic. Check for the VS Code helper.
		if strings.Contains(procs, "code helper") || strings.Contains(procs, "visual studio code") {
			return &IDE{Name: "VS Code", ID: "vscode", Family: "microsoft", Running: true}
		}
	}
	if strings.Contains(procs, "xcode") {
		return &IDE{Name: "Xcode", ID: "xcode", Family: "apple", Running: true}
	}
	if strings.Contains(procs, "android studio") {
		return &IDE{Name: "Android Studio", ID: "android-studio", Family: "google", Running: true}
	}

	return nil
}

// mapJetBrainsIDE maps a JETBRAINS_IDE env value to a specific IDE.
func mapJetBrainsIDE(value string) *IDE {
	v := strings.ToLower(value)
	switch {
	case strings.Contains(v, "goland"):
		return &IDE{Name: "GoLand", ID: "goland", Family: "jetbrains", Running: true}
	case strings.Contains(v, "idea"):
		return &IDE{Name: "IntelliJ IDEA", ID: "intellij", Family: "jetbrains", Running: true}
	case strings.Contains(v, "webstorm"):
		return &IDE{Name: "WebStorm", ID: "webstorm", Family: "jetbrains", Running: true}
	case strings.Contains(v, "pycharm"):
		return &IDE{Name: "PyCharm", ID: "pycharm", Family: "jetbrains", Running: true}
	case strings.Contains(v, "phpstorm"):
		return &IDE{Name: "PhpStorm", ID: "phpstorm", Family: "jetbrains", Running: true}
	case strings.Contains(v, "rider"):
		return &IDE{Name: "Rider", ID: "rider", Family: "jetbrains", Running: true}
	case strings.Contains(v, "clion"):
		return &IDE{Name: "CLion", ID: "clion", Family: "jetbrains", Running: true}
	case strings.Contains(v, "rustrover"):
		return &IDE{Name: "RustRover", ID: "rustrover", Family: "jetbrains", Running: true}
	case strings.Contains(v, "fleet"):
		return &IDE{Name: "Fleet", ID: "fleet", Family: "jetbrains", Running: true}
	default:
		return &IDE{Name: "JetBrains IDE", ID: "jetbrains", Family: "jetbrains", Running: true}
	}
}

// detectJetBrainsFromHistFile extracts the IDE name from the history file path.
func detectJetBrainsFromHistFile(path string) *IDE {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "goland"):
		return &IDE{Name: "GoLand", ID: "goland", Family: "jetbrains", Running: true}
	case strings.Contains(lower, "intellijidea") || strings.Contains(lower, "idea"):
		return &IDE{Name: "IntelliJ IDEA", ID: "intellij", Family: "jetbrains", Running: true}
	case strings.Contains(lower, "webstorm"):
		return &IDE{Name: "WebStorm", ID: "webstorm", Family: "jetbrains", Running: true}
	case strings.Contains(lower, "pycharm"):
		return &IDE{Name: "PyCharm", ID: "pycharm", Family: "jetbrains", Running: true}
	default:
		return &IDE{Name: "JetBrains IDE", ID: "jetbrains", Family: "jetbrains", Running: true}
	}
}

// ─── CI/CD Pipeline Detection ────────────────────────────────────────────────

// CIPipeline represents a detected CI/CD environment.
type CIPipeline struct {
	Name     string `json:"name"`      // Human-readable (e.g., "GitHub Actions", "Jenkins")
	ID       string `json:"id"`        // Machine identifier (e.g., "github_actions", "jenkins")
	Family   string `json:"family"`    // Vendor family
	BuildID  string `json:"build_id"`  // Current build/run ID
	Branch   string `json:"branch"`    // Source branch
	Commit   string `json:"commit"`    // Commit SHA
	RepoURL  string `json:"repo_url"`  // Repository URL
	BuildURL string `json:"build_url"` // Link to the build/run
}

// DetectCI detects the current CI/CD pipeline from environment variables.
// Returns nil if not running in a CI environment.
func DetectCI() *CIPipeline {
	// Detection works by checking pipeline-specific env vars in priority order.
	// No generic "CI" env var is checked — each pipeline is detected explicitly.

	// ─── GitHub Actions ─────────────────────────────────────────────
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return &CIPipeline{
			Name:     "GitHub Actions",
			ID:       "github_actions",
			Family:   "github",
			BuildID:  os.Getenv("GITHUB_RUN_ID"),
			Branch:   os.Getenv("GITHUB_REF_NAME"),
			Commit:   os.Getenv("GITHUB_SHA"),
			RepoURL:  os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY"),
			BuildURL: os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/actions/runs/" + os.Getenv("GITHUB_RUN_ID"),
		}
	}

	// ─── Jenkins ────────────────────────────────────────────────────
	if os.Getenv("JENKINS_URL") != "" || os.Getenv("JENKINS_HOME") != "" {
		return &CIPipeline{
			Name:     "Jenkins",
			ID:       "jenkins",
			Family:   "jenkins",
			BuildID:  os.Getenv("BUILD_NUMBER"),
			Branch:   coalesce(os.Getenv("GIT_BRANCH"), os.Getenv("BRANCH_NAME")),
			Commit:   os.Getenv("GIT_COMMIT"),
			RepoURL:  os.Getenv("GIT_URL"),
			BuildURL: os.Getenv("BUILD_URL"),
		}
	}

	// ─── JetBrains TeamCity ─────────────────────────────────────────
	if os.Getenv("TEAMCITY_VERSION") != "" {
		return &CIPipeline{
			Name:     "TeamCity",
			ID:       "teamcity",
			Family:   "jetbrains",
			BuildID:  os.Getenv("BUILD_NUMBER"),
			Branch:   os.Getenv("BRANCH_NAME"),
			Commit:   os.Getenv("BUILD_VCS_NUMBER"),
			BuildURL: os.Getenv("BUILD_URL"),
		}
	}

	// ─── JetBrains Space ────────────────────────────────────────────
	if os.Getenv("JB_SPACE_EXECUTION_NUMBER") != "" {
		return &CIPipeline{
			Name:    "JetBrains Space",
			ID:      "jetbrains_space",
			Family:  "jetbrains",
			BuildID: os.Getenv("JB_SPACE_EXECUTION_NUMBER"),
			Branch:  os.Getenv("JB_SPACE_GIT_BRANCH"),
			Commit:  os.Getenv("JB_SPACE_GIT_REVISION"),
		}
	}

	// ─── GitLab CI ──────────────────────────────────────────────────
	if os.Getenv("GITLAB_CI") == "true" {
		return &CIPipeline{
			Name:     "GitLab CI",
			ID:       "gitlab_ci",
			Family:   "gitlab",
			BuildID:  os.Getenv("CI_JOB_ID"),
			Branch:   os.Getenv("CI_COMMIT_BRANCH"),
			Commit:   os.Getenv("CI_COMMIT_SHA"),
			RepoURL:  os.Getenv("CI_PROJECT_URL"),
			BuildURL: os.Getenv("CI_JOB_URL"),
		}
	}

	// ─── CircleCI ───────────────────────────────────────────────────
	if os.Getenv("CIRCLECI") == "true" {
		return &CIPipeline{
			Name:     "CircleCI",
			ID:       "circleci",
			Family:   "circleci",
			BuildID:  os.Getenv("CIRCLE_BUILD_NUM"),
			Branch:   os.Getenv("CIRCLE_BRANCH"),
			Commit:   os.Getenv("CIRCLE_SHA1"),
			RepoURL:  os.Getenv("CIRCLE_REPOSITORY_URL"),
			BuildURL: os.Getenv("CIRCLE_BUILD_URL"),
		}
	}

	// ─── Travis CI ──────────────────────────────────────────────────
	if os.Getenv("TRAVIS") == "true" {
		return &CIPipeline{
			Name:     "Travis CI",
			ID:       "travis",
			Family:   "travis",
			BuildID:  os.Getenv("TRAVIS_BUILD_ID"),
			Branch:   os.Getenv("TRAVIS_BRANCH"),
			Commit:   os.Getenv("TRAVIS_COMMIT"),
			RepoURL:  "https://github.com/" + os.Getenv("TRAVIS_REPO_SLUG"),
			BuildURL: os.Getenv("TRAVIS_BUILD_WEB_URL"),
		}
	}

	// ─── Bitbucket Pipelines ────────────────────────────────────────
	if os.Getenv("BITBUCKET_PIPELINE_UUID") != "" {
		return &CIPipeline{
			Name:     "Bitbucket Pipelines",
			ID:       "bitbucket",
			Family:   "atlassian",
			BuildID:  os.Getenv("BITBUCKET_BUILD_NUMBER"),
			Branch:   os.Getenv("BITBUCKET_BRANCH"),
			Commit:   os.Getenv("BITBUCKET_COMMIT"),
			RepoURL:  os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN"),
			BuildURL: os.Getenv("BITBUCKET_PIPELINE_UUID"),
		}
	}

	// ─── Azure DevOps (Pipelines) ───────────────────────────────────
	if os.Getenv("TF_BUILD") == "True" || os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") != "" {
		return &CIPipeline{
			Name:     "Azure Pipelines",
			ID:       "azure_pipelines",
			Family:   "microsoft",
			BuildID:  os.Getenv("BUILD_BUILDID"),
			Branch:   os.Getenv("BUILD_SOURCEBRANCH"),
			Commit:   os.Getenv("BUILD_SOURCEVERSION"),
			RepoURL:  os.Getenv("BUILD_REPOSITORY_URI"),
			BuildURL: os.Getenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI") + os.Getenv("SYSTEM_TEAMPROJECT") + "/_build/results?buildId=" + os.Getenv("BUILD_BUILDID"),
		}
	}

	// ─── AWS CodeBuild ──────────────────────────────────────────────
	if os.Getenv("CODEBUILD_BUILD_ID") != "" {
		return &CIPipeline{
			Name:    "AWS CodeBuild",
			ID:      "aws_codebuild",
			Family:  "aws",
			BuildID: os.Getenv("CODEBUILD_BUILD_ID"),
			Branch:  os.Getenv("CODEBUILD_WEBHOOK_HEAD_REF"),
			Commit:  os.Getenv("CODEBUILD_RESOLVED_SOURCE_VERSION"),
		}
	}

	// ─── Google Cloud Build ─────────────────────────────────────────
	if os.Getenv("BUILDER_OUTPUT") != "" || os.Getenv("BUILD_ID") != "" && os.Getenv("PROJECT_ID") != "" {
		return &CIPipeline{
			Name:    "Google Cloud Build",
			ID:      "cloud_build",
			Family:  "google",
			BuildID: os.Getenv("BUILD_ID"),
			Branch:  os.Getenv("BRANCH_NAME"),
			Commit:  os.Getenv("COMMIT_SHA"),
		}
	}

	// ─── Buildkite ──────────────────────────────────────────────────
	if os.Getenv("BUILDKITE") == "true" {
		return &CIPipeline{
			Name:     "Buildkite",
			ID:       "buildkite",
			Family:   "buildkite",
			BuildID:  os.Getenv("BUILDKITE_BUILD_NUMBER"),
			Branch:   os.Getenv("BUILDKITE_BRANCH"),
			Commit:   os.Getenv("BUILDKITE_COMMIT"),
			RepoURL:  os.Getenv("BUILDKITE_REPO"),
			BuildURL: os.Getenv("BUILDKITE_BUILD_URL"),
		}
	}

	// ─── Drone CI ───────────────────────────────────────────────────
	if os.Getenv("DRONE") == "true" {
		return &CIPipeline{
			Name:     "Drone CI",
			ID:       "drone",
			Family:   "drone",
			BuildID:  os.Getenv("DRONE_BUILD_NUMBER"),
			Branch:   os.Getenv("DRONE_BRANCH"),
			Commit:   os.Getenv("DRONE_COMMIT_SHA"),
			RepoURL:  os.Getenv("DRONE_REPO_LINK"),
			BuildURL: os.Getenv("DRONE_BUILD_LINK"),
		}
	}

	// ─── Woodpecker CI (Drone fork) ─────────────────────────────────
	if os.Getenv("CI_PIPELINE_NUMBER") != "" && os.Getenv("CI_REPO") != "" {
		return &CIPipeline{
			Name:     "Woodpecker CI",
			ID:       "woodpecker",
			Family:   "woodpecker",
			BuildID:  os.Getenv("CI_PIPELINE_NUMBER"),
			Branch:   os.Getenv("CI_COMMIT_BRANCH"),
			Commit:   os.Getenv("CI_COMMIT_SHA"),
			RepoURL:  os.Getenv("CI_REPO_LINK"),
			BuildURL: os.Getenv("CI_PIPELINE_URL"),
		}
	}

	// ─── Semaphore CI ───────────────────────────────────────────────
	if os.Getenv("SEMAPHORE") == "true" {
		return &CIPipeline{
			Name:    "Semaphore CI",
			ID:      "semaphore",
			Family:  "semaphore",
			BuildID: os.Getenv("SEMAPHORE_WORKFLOW_ID"),
			Branch:  os.Getenv("SEMAPHORE_GIT_BRANCH"),
			Commit:  os.Getenv("SEMAPHORE_GIT_SHA"),
		}
	}

	// ─── Harness CI ─────────────────────────────────────────────────
	if os.Getenv("HARNESS_BUILD_ID") != "" {
		return &CIPipeline{
			Name:    "Harness CI",
			ID:      "harness",
			Family:  "harness",
			BuildID: os.Getenv("HARNESS_BUILD_ID"),
			Branch:  os.Getenv("HARNESS_GIT_BRANCH"),
			Commit:  os.Getenv("HARNESS_GIT_COMMIT"),
		}
	}

	return nil
}

// IsCI returns true if the current environment is a CI/CD pipeline.
func IsCI() bool {
	return DetectCI() != nil
}

// ─── AI Agent Detection ──────────────────────────────────────────────────────

// AIAgent represents a detected AI coding assistant.
type AIAgent struct {
	Name    string `json:"name"`    // Human-readable name
	ID      string `json:"id"`      // Machine identifier
	Family  string `json:"family"`  // Vendor family
	Version string `json:"version"` // Version if detectable
}

// DetectAIAgent detects which AI coding agent is driving the current session.
func DetectAIAgent() *AIAgent {
	// Antigravity (VS Code extension — Google DeepMind)
	if os.Getenv("ANTIGRAVITY_SESSION_ID") != "" || os.Getenv("ANTIGRAVITY_VERSION") != "" {
		return &AIAgent{
			Name:    "Antigravity",
			ID:      "antigravity",
			Family:  "google",
			Version: os.Getenv("ANTIGRAVITY_VERSION"),
		}
	}

	// Claude Code (Anthropic)
	if os.Getenv("CLAUDE_CODE") != "" || os.Getenv("CLAUDE_CODE_SESSION") != "" {
		return &AIAgent{
			Name:   "Claude Code",
			ID:     "claude_code",
			Family: "anthropic",
		}
	}

	// Gemini CLI (Google)
	if os.Getenv("GEMINI_CLI") != "" {
		return &AIAgent{
			Name:   "Gemini CLI",
			ID:     "gemini_cli",
			Family: "google",
		}
	}

	// GitHub Copilot
	if os.Getenv("GITHUB_COPILOT") != "" || os.Getenv("GH_COPILOT_CHAT") != "" {
		return &AIAgent{
			Name:   "GitHub Copilot",
			ID:     "copilot",
			Family: "github",
		}
	}

	// Cursor AI
	if os.Getenv("CURSOR_TRACE_DIR") != "" || os.Getenv("CURSOR_CHANNEL") != "" {
		return &AIAgent{
			Name:   "Cursor AI",
			ID:     "cursor",
			Family: "cursor",
		}
	}

	// Windsurf (Codeium)
	if os.Getenv("WINDSURF_SESSION") != "" {
		return &AIAgent{
			Name:   "Windsurf",
			ID:     "windsurf",
			Family: "codeium",
		}
	}

	// Aider
	if os.Getenv("AIDER_MODEL") != "" {
		return &AIAgent{
			Name:   "Aider",
			ID:     "aider",
			Family: "aider",
		}
	}

	// Continue.dev
	if os.Getenv("CONTINUE_SESSION") != "" {
		return &AIAgent{
			Name:   "Continue",
			ID:     "continue",
			Family: "continue",
		}
	}

	// Check running processes for AI agent CLIs
	return detectAIAgentFromProcesses()
}

// detectAIAgentFromProcesses scans for running AI agent processes.
func detectAIAgentFromProcesses() *AIAgent {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return nil
	}

	out, err := exec.Command("ps", "-eo", "comm").Output()
	if err != nil {
		return nil
	}

	procs := strings.ToLower(string(out))

	if strings.Contains(procs, "claude") {
		return &AIAgent{Name: "Claude Code", ID: "claude_code", Family: "anthropic"}
	}
	if strings.Contains(procs, "gemini") {
		return &AIAgent{Name: "Gemini CLI", ID: "gemini_cli", Family: "google"}
	}
	if strings.Contains(procs, "aider") {
		return &AIAgent{Name: "Aider", ID: "aider", Family: "aider"}
	}

	return nil
}

// ─── Full Environment Snapshot ───────────────────────────────────────────────

// Environment is the complete runtime environment snapshot.
type Environment struct {
	Platform Platform    `json:"-"`               // OS platform
	OS       string      `json:"os"`              // "darwin", "linux", "windows"
	Arch     string      `json:"arch"`            // "arm64", "amd64"
	IDE      *IDE        `json:"ide,omitempty"`   // Detected IDE/editor
	CI       *CIPipeline `json:"ci,omitempty"`    // Detected CI pipeline (nil if local)
	Agent    *AIAgent    `json:"agent,omitempty"` // Detected AI agent (nil if none)
	IsCI     bool        `json:"is_ci"`           // Quick check: are we in CI?
	HomeDir  string      `json:"home_dir"`        // User home directory
	WorkDir  string      `json:"work_dir"`        // Current working directory
	Shell    string      `json:"shell"`           // Active shell
}

// DetectEnvironment builds a full environment snapshot.
func DetectEnvironment() *Environment {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	env := &Environment{
		Platform: Current(),
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		IDE:      DetectIDE(),
		CI:       DetectCI(),
		Agent:    DetectAIAgent(),
		HomeDir:  home,
		WorkDir:  cwd,
		Shell:    filepath.Base(os.Getenv("SHELL")),
	}
	env.IsCI = env.CI != nil

	return env
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// coalesce returns the first non-empty string.
func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
