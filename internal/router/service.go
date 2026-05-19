package router

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ServiceOptions describes the per-repo launchd service.
type ServiceOptions struct {
	RepoRoot   string
	BinaryPath string
	Label      string
	PlistPath  string
	LogPath    string
	ErrPath    string
	PathEnv    string
}

// DefaultServiceOptions builds the launchd paths for a repo-local autorouter.
func DefaultServiceOptions(repoRoot, binaryPath string) ServiceOptions {
	label := "com.sirsi.router." + serviceSlug(repoRoot)
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(repoRoot, ".agents", "idea-router", "logs")
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		pathEnv = "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Applications/Codex.app/Contents/Resources"
	}
	return ServiceOptions{
		RepoRoot:   repoRoot,
		BinaryPath: binaryPath,
		Label:      label,
		PlistPath:  filepath.Join(home, "Library", "LaunchAgents", label+".plist"),
		LogPath:    filepath.Join(logDir, "autorouter.out.log"),
		ErrPath:    filepath.Join(logDir, "autorouter.err.log"),
		PathEnv:    pathEnv,
	}
}

// RenderLaunchAgentPlist returns a launchd plist that starts the router daemon.
func RenderLaunchAgentPlist(opts ServiceOptions) string {
	args := []string{opts.BinaryPath, "router", "daemon", "--target", "all"}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>`)
	buf.WriteString(xmlEscape(opts.Label))
	buf.WriteString(`</string>
  <key>WorkingDirectory</key>
  <string>`)
	buf.WriteString(xmlEscape(opts.RepoRoot))
	buf.WriteString(`</string>
  <key>ProgramArguments</key>
  <array>
`)
	for _, arg := range args {
		buf.WriteString("    <string>")
		buf.WriteString(xmlEscape(arg))
		buf.WriteString("</string>\n")
	}
	buf.WriteString(`  </array>
  <key>EnvironmentVariables</key>
  <dict>
    <key>SIRSI_ROUTER_NOTIFY</key>
    <string>1</string>
    <key>PATH</key>
    <string>`)
	buf.WriteString(xmlEscape(opts.PathEnv))
	buf.WriteString(`</string>
  </dict>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>`)
	buf.WriteString(xmlEscape(opts.LogPath))
	buf.WriteString(`</string>
  <key>StandardErrorPath</key>
  <string>`)
	buf.WriteString(xmlEscape(opts.ErrPath))
	buf.WriteString(`</string>
</dict>
</plist>
`)
	return buf.String()
}

// InstallLaunchAgent writes the launchd plist. Loading is left to the caller.
func InstallLaunchAgent(opts ServiceOptions) error {
	if err := os.MkdirAll(filepath.Dir(opts.PlistPath), 0o755); err != nil {
		return fmt.Errorf("create launch agent dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(opts.LogPath), 0o755); err != nil {
		return fmt.Errorf("create router log dir: %w", err)
	}
	if err := os.WriteFile(opts.PlistPath, []byte(RenderLaunchAgentPlist(opts)), 0o644); err != nil {
		return fmt.Errorf("write launch agent plist: %w", err)
	}
	return nil
}

// Launchctl runs launchctl with the supplied arguments.
func Launchctl(args ...string) error {
	cmd := exec.Command("launchctl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl %s failed: %w\n%s", strings.Join(args, " "), err, string(out))
	}
	return nil
}

func serviceSlug(repoRoot string) string {
	cleaned := strings.Trim(filepath.Base(repoRoot), ".")
	if cleaned == "" || cleaned == string(filepath.Separator) {
		cleaned = "repo"
	}
	var b strings.Builder
	for _, r := range strings.ToLower(cleaned) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('-')
	}
	return strings.Trim(b.String(), "-")
}

func xmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}
