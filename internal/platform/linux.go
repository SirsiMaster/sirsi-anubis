package platform

import (
	"fmt"
	"os/exec"
)

// Linux implements Platform for Linux distributions.
type Linux struct{}

func (l *Linux) Name() string { return "linux" }

func (l *Linux) SupportsTrash() bool { return false } // TODO: freedesktop.org trash spec

func (l *Linux) MoveToTrash(path string) error {
	// TODO: Implement freedesktop.org trash spec
	// For now, use gio trash if available
	cmd := exec.Command("gio", "trash", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("trash not available on this Linux system (install glib2): %w", err)
	}
	return nil
}

func (l *Linux) ProtectedPrefixes() []string {
	return []string{
		"/boot/",
		"/etc/",
		"/usr/",
		"/bin/",
		"/sbin/",
		"/lib/",
		"/lib64/",
		"/proc/",
		"/sys/",
		"/dev/",
		"/var/lib/dpkg/",
		"/var/lib/rpm/",
	}
}

func (l *Linux) PickFolder() (string, error) {
	// Try zenity first, then kdialog
	cmd := exec.Command("zenity", "--file-selection", "--directory",
		"--title=Select a folder to scan")
	out, err := cmd.Output()
	if err != nil {
		// Fallback to kdialog
		cmd = exec.Command("kdialog", "--getexistingdirectory", ".")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("no folder picker available (install zenity or kdialog): %w", err)
		}
	}
	return string(out), nil
}

func (l *Linux) OpenBrowser(url string) error {
	return exec.Command("xdg-open", url).Start()
}
