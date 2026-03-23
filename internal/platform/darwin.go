package platform

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// Darwin implements Platform for macOS.
type Darwin struct{}

func (d *Darwin) Name() string { return "darwin" }

func (d *Darwin) SupportsTrash() bool { return true }

func (d *Darwin) MoveToTrash(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path for trash: %w", err)
	}
	script := fmt.Sprintf(
		`tell application "Finder" to delete POSIX file %q`,
		absPath,
	)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func (d *Darwin) ProtectedPrefixes() []string {
	return []string{
		"/System/",
		"/usr/",
		"/bin/",
		"/sbin/",
		"/private/var/db/",
		"/Library/Extensions/",
		"/Library/Frameworks/",
	}
}

func (d *Darwin) PickFolder() (string, error) {
	cmd := exec.Command("osascript", "-e",
		`POSIX path of (choose folder with prompt "Select a folder to scan:")`)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("folder picker canceled or failed: %w", err)
	}
	return filepath.Clean(string(out)), nil
}

func (d *Darwin) OpenBrowser(url string) error {
	return exec.Command("open", url).Start()
}
