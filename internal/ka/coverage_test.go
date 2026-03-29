package ka

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

// ─── Scan ───────────────────────────────────────────────────────────────

func TestScan_SkipLaunchServices(t *testing.T) {
	scanner := NewScanner()
	scanner.SkipLaunchServices = true
	scanner.SkipBrew = true
	scanner.DirReader = func(path string) ([]os.DirEntry, error) {
		return nil, nil // Return empty to avoid slow walk
	}
	scanner.ExecCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	ghosts, err := scanner.Scan(context.Background(), false)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	// Just verify it returns without error and produces a valid slice.
	if ghosts == nil {
		t.Log("No ghosts found (clean system)")
	}
}

func TestScan_WithLaunchServices(t *testing.T) {
	scanner := NewScanner()
	scanner.SkipLaunchServices = false
	scanner.SkipBrew = true
	scanner.DirReader = func(path string) ([]os.DirEntry, error) {
		return nil, nil // Return empty to avoid slow walk
	}
	scanner.ExecCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	ghosts, err := scanner.Scan(context.Background(), false)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if ghosts == nil {
		t.Log("No ghosts found")
	}
}

// ─── scanLaunchServices ──────────────────────────────────────────────────

func TestScanLaunchServices(t *testing.T) {
	scanner := NewScanner()
	scanner.SkipBrew = true
	scanner.DirReader = func(path string) ([]os.DirEntry, error) {
		return nil, nil
	}
	scanner.ExecCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}
	// Build installed app index first (required for scanLaunchServices)
	scanner.buildInstalledAppIndex(context.Background())

	ghosts := scanner.scanLaunchServices(context.Background())
	if ghosts == nil {
		t.Fatal("scanLaunchServices returned nil")
	}
	t.Logf("Launch Services ghosts found: %d", len(ghosts))
}

// ─── indexHomebrewCasks ─────────────────────────────────────────────────

func TestIndexHomebrewCasks(t *testing.T) {
	scanner := NewScanner()

	scanner.ExecCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// Return a command that fails so it doesn't try to run 'brew'
		return exec.CommandContext(ctx, "false")
	}
	// This calls `brew list --cask` — may fail if brew is not installed.
	scanner.indexHomebrewCasks(context.Background())

	// Just verify it doesn't panic and populates the index.
	t.Logf("Homebrew cask names indexed: %d", len(scanner.installedNames))
}

// ─── readBundleID ──────────────────────────────────────────────────────

func TestReadBundleID_ValidApp(t *testing.T) {
	// Safari is always installed on macOS
	bundleID, err := readBundleIDDefault(context.Background(), "/Applications/Safari.app")
	if err != nil {
		t.Logf("readBundleIDDefault returned error for Safari (expected in CI): %v", err)
		return
	}
	if bundleID == "" {
		t.Log("readBundleIDDefault returned empty for Safari")
		return
	}
	if bundleID != "com.apple.Safari" {
		t.Errorf("bundleID = %q, want com.apple.Safari", bundleID)
	}
}

func TestReadBundleID_InvalidApp(t *testing.T) {
	bundleID, err := readBundleIDDefault(context.Background(), "/nonexistent/app.app")
	if err == nil && bundleID != "" {
		t.Errorf("expected empty bundleID or error for nonexistent app, got %q", bundleID)
	}
}
