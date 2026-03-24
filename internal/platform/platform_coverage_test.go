package platform

import (
	"runtime"
	"testing"
)

// ─── Darwin ──────────────────────────────────────────────────────────────────
// These methods call OS-level commands. We exercise the code paths but expect
// failures in non-interactive environments.

func TestDarwin_MoveToTrash(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("osascript only available on macOS")
	}
	d := &Darwin{}
	// MoveToTrash on a nonexistent file — osascript will fail but code path exercised.
	err := d.MoveToTrash("/nonexistent/file/for/test")
	if err == nil {
		t.Log("MoveToTrash succeeded (unexpected on nonexistent file)")
	}
}

func TestDarwin_PickFolder(t *testing.T) {
	t.Skip("PickFolder launches interactive Finder dialog — skip in automated tests")
}

func TestDarwin_OpenBrowser(t *testing.T) {
	t.Skip("OpenBrowser launches real browser — skip in automated tests")
}

// ─── Linux ──────────────────────────────────────────────────────────────────

func TestLinux_MoveToTrash(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("gio/trash-cli only available on Linux")
	}
	l := &Linux{}
	err := l.MoveToTrash("/nonexistent/file/for/test")
	if err == nil {
		t.Log("MoveToTrash succeeded (unexpected)")
	}
}

func TestLinux_PickFolder(t *testing.T) {
	t.Skip("PickFolder launches interactive dialog — skip in automated tests")
}

func TestLinux_OpenBrowser(t *testing.T) {
	t.Skip("OpenBrowser launches real browser — skip in automated tests")
}
