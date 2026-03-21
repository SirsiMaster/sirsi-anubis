// Package sight provides Spotlight and Launch Services management.
// Extracts ghost-detection logic from Ka into a dedicated module
// that can also rebuild the Launch Services database.
package sight

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GhostRegistration represents an app registered in Launch Services
// whose .app bundle no longer exists on disk.
type GhostRegistration struct {
	BundleID string
	Path     string // Original .app path
	Name     string
}

// SightResult contains the results of a Spotlight/Launch Services audit.
type SightResult struct {
	GhostRegistrations []GhostRegistration
	TotalGhosts        int
	LaunchServicesSize int64
	CanFix             bool
}

// Scan queries Launch Services for ghost app registrations.
func Scan() (*SightResult, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("sight: only supported on macOS")
	}

	result := &SightResult{CanFix: true}

	// Dump Launch Services database
	out, err := exec.Command(
		"/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister",
		"-dump",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("sight: lsregister dump failed: %w", err)
	}

	// Parse registrations for missing .app bundles
	ghosts := parseLSRegisterDump(string(out))
	result.GhostRegistrations = ghosts
	result.TotalGhosts = len(ghosts)

	return result, nil
}

// Fix rebuilds the Launch Services database, removing ghost registrations.
// This is a DESTRUCTIVE operation — it resets all file associations.
func Fix(dryRun bool) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("sight: only supported on macOS")
	}

	if dryRun {
		return nil // Dry run — don't actually rebuild
	}

	lsregister := "/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister"

	// Kill the Launch Services database and rebuild
	cmd := exec.Command(lsregister, "-kill", "-r", "-domain", "local", "-domain", "system", "-domain", "user")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sight: lsregister rebuild failed: %w", err)
	}

	// Restart Finder to pick up changes
	cmd = exec.Command("killall", "Finder")
	_ = cmd.Run() // Non-fatal if Finder restart fails

	return nil
}

// ReindexSpotlight triggers a Spotlight re-index for the boot volume.
func ReindexSpotlight(dryRun bool) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("sight: only supported on macOS")
	}
	if dryRun {
		return nil
	}

	cmd := exec.Command("mdutil", "-E", "/")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sight: Spotlight reindex failed (may need sudo): %w", err)
	}
	return nil
}

// parseLSRegisterDump extracts ghost registrations from lsregister output.
func parseLSRegisterDump(dump string) []GhostRegistration {
	var ghosts []GhostRegistration
	seen := make(map[string]bool)

	blocks := strings.Split(dump, "--------------------------------------------------------------------------------")

	for _, block := range blocks {
		if !strings.Contains(block, "bundle id:") {
			continue
		}
		if !strings.Contains(block, ".app") {
			continue
		}

		var bundleID, path, name string

		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "bundle id:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					bundleID = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "path:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					path = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "name:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					name = strings.TrimSpace(parts[1])
				}
			}
		}

		if bundleID == "" || path == "" {
			continue
		}

		// Skip Apple system apps
		if strings.HasPrefix(bundleID, "com.apple.") {
			continue
		}

		// Check if the .app exists
		if !strings.Contains(path, ".app") {
			continue
		}

		// Extract .app path
		appPath := path
		if idx := strings.Index(path, ".app"); idx > 0 {
			appPath = path[:idx+4]
		}

		// Check if already seen
		if seen[bundleID] {
			continue
		}

		// Check if .app exists on disk
		cmd := exec.Command("test", "-d", appPath)
		if cmd.Run() == nil {
			continue // App exists — not a ghost
		}

		seen[bundleID] = true
		if name == "" {
			name = bundleID
		}
		ghosts = append(ghosts, GhostRegistration{
			BundleID: bundleID,
			Path:     appPath,
			Name:     name,
		})
	}

	return ghosts
}
