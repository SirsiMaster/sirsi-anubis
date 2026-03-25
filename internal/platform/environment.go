package platform

import (
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
func DetectIDE() *IDE {
	return DetectIDEWith(Current())
}

// DetectIDEWith is the injectable version of DetectIDE.
func DetectIDEWith(p Platform) *IDE {
	// Priority 1: Environment variable signals (most reliable)
	if ide := detectIDEFromEnv(p); ide != nil {
		return ide
	}

	// Priority 2: Terminal/editor-specific env vars
	if ide := detectIDEFromTerminalEnv(p); ide != nil {
		return ide
	}

	// Priority 3: Running process detection
	if ide := detectIDEFromProcesses(p); ide != nil {
		return ide
	}

	return &IDE{Name: "Unknown", ID: "unknown", Family: "unknown"}
}

func detectIDEFromEnv(p Platform) *IDE {
	// VS Code / Cursor / Windsurf (Code-based editors)
	if term := p.Getenv("TERM_PROGRAM"); term != "" {
		switch strings.ToLower(term) {
		case "vscode":
			return &IDE{
				Name:    "VS Code",
				ID:      "vscode",
				Family:  "microsoft",
				Version: p.Getenv("TERM_PROGRAM_VERSION"),
				Running: true,
			}
		}
	}

	if p.Getenv("VSCODE_PID") != "" || p.Getenv("VSCODE_IPC_HOOK") != "" {
		return &IDE{
			Name:    "VS Code",
			ID:      "vscode",
			Family:  "microsoft",
			Version: p.Getenv("VSCODE_CLI_VERSION"),
			Running: true,
		}
	}

	if p.Getenv("CURSOR_TRACE_DIR") != "" || p.Getenv("CURSOR_CHANNEL") != "" {
		return &IDE{
			Name:    "Cursor",
			ID:      "cursor",
			Family:  "cursor",
			Running: true,
		}
	}

	if jetbrainsIDE := p.Getenv("JETBRAINS_IDE"); jetbrainsIDE != "" {
		return mapJetBrainsIDE(jetbrainsIDE)
	}
	if termEmu := p.Getenv("TERMINAL_EMULATOR"); strings.Contains(strings.ToLower(termEmu), "jetbrains") {
		return &IDE{
			Name:    "JetBrains IDE",
			ID:      "jetbrains",
			Family:  "jetbrains",
			Running: true,
		}
	}

	if hist := p.Getenv("__INTELLIJ_COMMAND_HISTFILE__"); hist != "" {
		return detectJetBrainsFromHistFile(hist)
	}

	return nil
}

func detectIDEFromTerminalEnv(p Platform) *IDE {
	if p.Getenv("NVIM") != "" || p.Getenv("NVIM_LISTEN_ADDRESS") != "" {
		return &IDE{Name: "Neovim", ID: "neovim", Family: "vim", Running: true}
	}
	if p.Getenv("VIM") != "" || p.Getenv("VIMRUNTIME") != "" {
		return &IDE{Name: "Vim", ID: "vim", Family: "vim", Running: true}
	}
	if p.Getenv("INSIDE_EMACS") != "" {
		return &IDE{Name: "Emacs", ID: "emacs", Family: "emacs", Running: true}
	}
	if p.Getenv("ZED_TERM") != "" {
		return &IDE{Name: "Zed", ID: "zed", Family: "zed", Running: true}
	}
	return nil
}

func detectIDEFromProcesses(p Platform) *IDE {
	procs, err := p.Processes()
	if err != nil {
		return nil
	}

	allProcs := strings.ToLower(strings.Join(procs, "\n"))

	jetbrainsMap := map[string]*IDE{
		"goland":    {Name: "GoLand", ID: "goland", Family: "jetbrains", Running: true},
		"idea":      {Name: "IntelliJ IDEA", ID: "intellij", Family: "jetbrains", Running: true},
		"webstorm":  {Name: "WebStorm", ID: "webstorm", Family: "jetbrains", Running: true},
		"pycharm":   {Name: "PyCharm", ID: "pycharm", Family: "jetbrains", Running: true},
		"phpstorm":  {Name: "PhpStorm", ID: "phpstorm", Family: "jetbrains", Running: true},
		"rider":     {Name: "Rider", ID: "rider", Family: "jetbrains", Running: true},
		"clion":     {Name: "CLion", ID: "clion", Family: "jetbrains", Running: true},
		"rustrover": {Name: "RustRover", ID: "rustrover", Family: "jetbrains", Running: true},
	}
	for key, ide := range jetbrainsMap {
		if strings.Contains(allProcs, key) {
			return ide
		}
	}

	if strings.Contains(allProcs, "cursor") {
		return &IDE{Name: "Cursor", ID: "cursor", Family: "cursor", Running: true}
	}
	if strings.Contains(allProcs, "windsurf") {
		return &IDE{Name: "Windsurf", ID: "windsurf", Family: "codeium", Running: true}
	}
	if strings.Contains(allProcs, "zed") {
		return &IDE{Name: "Zed", ID: "zed", Family: "zed", Running: true}
	}
	if strings.Contains(allProcs, "code helper") || strings.Contains(allProcs, "visual studio code") {
		return &IDE{Name: "VS Code", ID: "vscode", Family: "microsoft", Running: true}
	}
	if strings.Contains(allProcs, "xcode") {
		return &IDE{Name: "Xcode", ID: "xcode", Family: "apple", Running: true}
	}

	return nil
}

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
	default:
		return &IDE{Name: "JetBrains IDE", ID: "jetbrains", Family: "jetbrains", Running: true}
	}
}

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
	Name     string `json:"name"`      // Human-readable
	ID       string `json:"id"`        // Machine identifier
	Family   string `json:"family"`    // Vendor family
	BuildID  string `json:"build_id"`  // Current build/run ID
	Branch   string `json:"branch"`    // Source branch
	Commit   string `json:"commit"`    // Commit SHA
	RepoURL  string `json:"repo_url"`  // Repository URL
	BuildURL string `json:"build_url"` // Link to the build/run
}

// DetectCI detects the current CI/CD pipeline from environment variables.
func DetectCI() *CIPipeline {
	return DetectCIWith(Current())
}

// DetectCIWith is the injectable version of DetectCI.
func DetectCIWith(p Platform) *CIPipeline {
	if p.Getenv("GITHUB_ACTIONS") == "true" {
		return &CIPipeline{
			Name:     "GitHub Actions",
			ID:       "github_actions",
			Family:   "github",
			BuildID:  p.Getenv("GITHUB_RUN_ID"),
			Branch:   p.Getenv("GITHUB_REF_NAME"),
			Commit:   p.Getenv("GITHUB_SHA"),
			RepoURL:  p.Getenv("GITHUB_SERVER_URL") + "/" + p.Getenv("GITHUB_REPOSITORY"),
			BuildURL: p.Getenv("GITHUB_SERVER_URL") + "/" + p.Getenv("GITHUB_REPOSITORY") + "/actions/runs/" + p.Getenv("GITHUB_RUN_ID"),
		}
	}

	if p.Getenv("JENKINS_URL") != "" {
		return &CIPipeline{
			Name:     "Jenkins",
			ID:       "jenkins",
			Family:   "jenkins",
			BuildID:  p.Getenv("BUILD_NUMBER"),
			Branch:   p.Getenv("GIT_BRANCH"),
			Commit:   p.Getenv("GIT_COMMIT"),
			RepoURL:  p.Getenv("GIT_URL"),
			BuildURL: p.Getenv("BUILD_URL"),
		}
	}

	if p.Getenv("TEAMCITY_VERSION") != "" {
		return &CIPipeline{
			Name:    "TeamCity",
			ID:      "teamcity",
			Family:  "jetbrains",
			BuildID: p.Getenv("BUILD_NUMBER"),
			Commit:  p.Getenv("BUILD_VCS_NUMBER"),
		}
	}

	if p.Getenv("GITLAB_CI") == "true" {
		return &CIPipeline{
			Name:     "GitLab CI",
			ID:       "gitlab_ci",
			Family:   "gitlab",
			BuildID:  p.Getenv("CI_JOB_ID"),
			Branch:   p.Getenv("CI_COMMIT_BRANCH"),
			Commit:   p.Getenv("CI_COMMIT_SHA"),
			RepoURL:  p.Getenv("CI_PROJECT_URL"),
			BuildURL: p.Getenv("CI_JOB_URL"),
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
	return DetectAIAgentWith(Current())
}

// DetectAIAgentWith is the injectable version of DetectAIAgent.
func DetectAIAgentWith(p Platform) *AIAgent {
	if p.Getenv("ANTIGRAVITY_SESSION_ID") != "" || p.Getenv("ANTIGRAVITY_VERSION") != "" {
		return &AIAgent{
			Name:    "Antigravity",
			ID:      "antigravity",
			Family:  "google",
			Version: p.Getenv("ANTIGRAVITY_VERSION"),
		}
	}

	if p.Getenv("CLAUDE_CODE") != "" || p.Getenv("CLAUDE_CODE_SESSION") != "" {
		return &AIAgent{Name: "Claude Code", ID: "claude_code", Family: "anthropic"}
	}

	if p.Getenv("GEMINI_CLI") != "" {
		return &AIAgent{Name: "Gemini CLI", ID: "gemini_cli", Family: "google"}
	}

	if p.Getenv("GITHUB_COPILOT") != "" || p.Getenv("GH_COPILOT_CHAT") != "" {
		return &AIAgent{Name: "GitHub Copilot", ID: "copilot", Family: "github"}
	}

	if p.Getenv("CURSOR_TRACE_DIR") != "" || p.Getenv("CURSOR_CHANNEL") != "" {
		return &AIAgent{Name: "Cursor AI", ID: "cursor", Family: "cursor"}
	}

	return detectAIAgentFromProcesses(p)
}

func detectAIAgentFromProcesses(p Platform) *AIAgent {
	procs, err := p.Processes()
	if err != nil {
		return nil
	}

	allProcs := strings.ToLower(strings.Join(procs, "\n"))
	if strings.Contains(allProcs, "claude-code") {
		return &AIAgent{Name: "Claude Code", ID: "claude_code", Family: "anthropic"}
	}
	if strings.Contains(allProcs, "antigravity") {
		return &AIAgent{Name: "Antigravity", ID: "antigravity", Family: "google"}
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
	return DetectEnvironmentWith(Current())
}

// DetectEnvironmentWith is the injectable version of DetectEnvironment.
func DetectEnvironmentWith(p Platform) *Environment {
	home, _ := p.UserHomeDir()
	cwd, _ := p.Getwd()

	shell := p.Getenv("SHELL")
	if shell != "" {
		shell = filepath.Base(shell)
	}

	env := &Environment{
		Platform: p,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		IDE:      DetectIDEWith(p),
		CI:       DetectCIWith(p),
		Agent:    DetectAIAgentWith(p),
		HomeDir:  home,
		WorkDir:  cwd,
		Shell:    shell,
	}
	env.IsCI = env.CI != nil

	return env
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
