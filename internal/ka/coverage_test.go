package ka

import (
	"testing"
)

// ─── Scan ───────────────────────────────────────────────────────────────

func TestScan_SkipLaunchServices(t *testing.T) {
	scanner := NewScanner()
	scanner.SkipLaunchServices = true
	scanner.skipBrew = true

	ghosts, err := scanner.Scan(false)
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
	scanner.skipBrew = true

	ghosts, err := scanner.Scan(false)
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
	scanner.skipBrew = true
	// Build installed app index first (required for scanLaunchServices)
	scanner.buildInstalledAppIndex()

	ghosts := scanner.scanLaunchServices()
	if ghosts == nil {
		t.Fatal("scanLaunchServices returned nil")
	}
	t.Logf("Launch Services ghosts found: %d", len(ghosts))
}

// ─── indexHomebrewCasks ─────────────────────────────────────────────────

func TestIndexHomebrewCasks(t *testing.T) {
	scanner := NewScanner()

	// This calls `brew list --cask` — may fail if brew is not installed.
	scanner.indexHomebrewCasks()

	// Just verify it doesn't panic and populates the index.
	t.Logf("Homebrew cask names indexed: %d", len(scanner.installedNames))
}

// ─── readBundleID ──────────────────────────────────────────────────────

func TestReadBundleID_ValidApp(t *testing.T) {
	// Safari is always installed on macOS
	bundleID := readBundleID("/Applications/Safari.app")
	if bundleID == "" {
		t.Log("readBundleID returned empty for Safari (expected in CI)")
		return
	}
	if bundleID != "com.apple.Safari" {
		t.Errorf("bundleID = %q, want com.apple.Safari", bundleID)
	}
}

func TestReadBundleID_InvalidApp(t *testing.T) {
	bundleID := readBundleID("/nonexistent/app.app")
	if bundleID != "" {
		t.Errorf("expected empty bundleID for nonexistent app, got %q", bundleID)
	}
}
