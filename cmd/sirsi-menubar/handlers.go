// Package main — sirsi-menubar
//
// handlers.go — Menu click handlers that dispatch to Sirsi CLI subcommands.
//
// Each handler spawns the corresponding `sirsi` CLI command in the background.
// When the command completes, a macOS toast notification is fired and the result
// is stored in the persistent notification history (internal/notify).
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
)

// Handler wraps a menu action with name and execution logic.
type Handler struct {
	Name    string
	Command string
	Args    []string
}

// SirsiHandlers returns the set of available menu actions.
func SirsiHandlers() []Handler {
	sirsiBin := findSirsiBinary()
	return []Handler{
		{Name: "Scan (Weigh)", Command: sirsiBin, Args: []string{"weigh"}},
		{Name: "Judge", Command: sirsiBin, Args: []string{"judge"}},
		{Name: "Guard", Command: sirsiBin, Args: []string{"guard"}},
		{Name: "Ka (Ghost Hunt)", Command: sirsiBin, Args: []string{"ka"}},
		{Name: "Mirror (Dedup)", Command: sirsiBin, Args: []string{"mirror"}},
		{Name: "Ma'at (QA)", Command: sirsiBin, Args: []string{"maat"}},
	}
}

// RaHandlers returns Ra orchestration menu actions.
func RaHandlers() []Handler {
	sirsiBin := findSirsiBinary()
	return []Handler{
		{Name: "𓇶 Ra Deploy", Command: sirsiBin, Args: []string{"ra", "deploy"}},
		{Name: "𓇶 Ra Kill All", Command: sirsiBin, Args: []string{"ra", "kill"}},
		{Name: "𓇶 Ra Collect", Command: sirsiBin, Args: []string{"ra", "collect"}},
		{Name: "𓇶 Ra Status", Command: sirsiBin, Args: []string{"ra", "status"}},
	}
}

// QuickActions returns quick-access menu actions.
func QuickActions() []Handler {
	sirsiBin := findSirsiBinary()
	return []Handler{
		{Name: "Start Watchdog", Command: sirsiBin, Args: []string{"guard", "--watch"}},
	}
}

// Execute runs the handler command in the background (legacy, no feedback).
func (h *Handler) Execute() error {
	cmd := exec.Command(h.Command, h.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// ExecuteWithNotify starts the command in the background and records
// the result in the notification store when it completes.
// The caller (event loop) returns immediately — never blocks.
func (h *Handler) ExecuteWithNotify(store *notify.Store) {
	cmd := exec.Command(h.Command, h.Args...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	start := time.Now()
	if err := cmd.Start(); err != nil {
		if store != nil {
			_ = store.Record(notify.Notification{
				Source:   h.source(),
				Action:   h.action(),
				Severity: notify.SeverityError,
				Summary:  fmt.Sprintf("%s failed to start: %v", h.Name, err),
			})
		}
		return
	}

	go func() {
		err := cmd.Wait()
		elapsed := time.Since(start)
		output := buf.String()

		n := notify.Notification{
			Source:     h.source(),
			Action:     h.action(),
			DurationMs: elapsed.Milliseconds(),
			Details:    output,
		}

		if err != nil {
			n.Severity = notify.SeverityError
			n.Summary = fmt.Sprintf("%s failed (%s)", h.Name, elapsed.Truncate(time.Second))
		} else {
			n.Severity = notify.SeveritySuccess
			n.Summary = parseSummary(h.Name, output, elapsed)
		}

		if store != nil {
			_ = store.Record(n)
		}
	}()
}

// source maps handler args to the deity name for notification display.
func (h *Handler) source() string {
	if len(h.Args) > 0 && h.Args[0] == "ra" {
		return "ra"
	}
	sourceMap := map[string]string{
		"weigh": "anubis", "judge": "anubis",
		"guard": "isis", "ka": "anubis",
		"mirror": "anubis", "maat": "maat",
	}
	if len(h.Args) > 0 {
		if s, ok := sourceMap[h.Args[0]]; ok {
			return s
		}
	}
	return "sirsi"
}

// action extracts the action verb from the handler args.
func (h *Handler) action() string {
	if len(h.Args) > 0 {
		return h.Args[len(h.Args)-1]
	}
	return "unknown"
}

// ansiRe strips ANSI escape sequences from CLI output.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// parseSummary extracts a meaningful one-liner from CLI output.
func parseSummary(name, output string, elapsed time.Duration) string {
	clean := ansiRe.ReplaceAllString(output, "")
	lines := strings.Split(strings.TrimSpace(clean), "\n")

	// Walk backwards to find the last non-empty, meaningful line.
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || line == "---" || strings.HasPrefix(line, "━") {
			continue
		}
		if len(line) > 120 {
			line = line[:117] + "..."
		}
		return line
	}
	return fmt.Sprintf("%s completed (%s)", name, elapsed.Truncate(time.Second))
}

// OpenBuildLog opens the build log in the default browser.
func OpenBuildLog() error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	path := filepath.Join(root, "docs", "build-log.html")
	return exec.Command("open", path).Start()
}

// OpenCaseStudies opens the case studies in the default browser.
func OpenCaseStudies() error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	path := filepath.Join(root, "docs", "case-studies.html")
	return exec.Command("open", path).Start()
}

// OpenCommandCenter launches `sirsi ra watch` in a new terminal.
func OpenCommandCenter() error {
	sirsiBin := findSirsiBinary()
	cmd := exec.Command(sirsiBin, "ra", "watch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// OpenScopeLog opens the log file for a given Ra scope.
func OpenScopeLog(scopeName string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	logPath := filepath.Join(home, ".config", "ra", "logs", scopeName+".log")
	return exec.Command("open", "-a", "Console", logPath).Start()
}

// findSirsiBinary locates the sirsi binary.
func findSirsiBinary() string {
	// Check PATH first
	if p, err := exec.LookPath("sirsi"); err == nil {
		return p
	}
	// Check Homebrew location
	if p, err := exec.LookPath("/opt/homebrew/bin/sirsi"); err == nil {
		return p
	}
	// Check local bin
	if p, err := exec.LookPath("./bin/sirsi"); err == nil {
		return p
	}
	// Fallback to just "sirsi" and hope it's in PATH
	return "sirsi"
}

// findRepoRoot locates the sirsi-pantheon repo root.
func findRepoRoot() (string, error) {
	// Try git rev-parse first
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	// Fallback to known location
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home dir: %w", err)
	}
	candidate := filepath.Join(home, "Development", "sirsi-pantheon")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", fmt.Errorf("cannot find sirsi-pantheon repo root")
}
