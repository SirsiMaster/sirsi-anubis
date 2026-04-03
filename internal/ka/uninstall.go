// uninstall.go implements complete app removal for macOS.
// Ka is the spirit that persists after death — this module releases those spirits.
//
// SAFETY: All deletions go through cleaner.ValidatePath (Rule A1).
// Default behavior is UseTrash: true (move to Trash, never hard-delete).
// DryRun MUST be run before any actual removal.
package ka

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SirsiMaster/sirsi-pantheon/internal/cleaner"
	"github.com/SirsiMaster/sirsi-pantheon/internal/logging"
)

// UninstallOptions controls how an app is removed.
type UninstallOptions struct {
	AppPath  string // path to the .app bundle
	BundleID string // com.example.app — used to find residuals
	AppName  string // human-readable name
	Complete bool   // true = remove all residuals, false = just .app
	DryRun   bool   // true = preview only, no deletion
	UseTrash bool   // true = move to Trash instead of permanent delete
}

// UninstallResult reports what was (or would be) removed.
type UninstallResult struct {
	AppRemoved     bool     `json:"app_removed"`
	FilesRemoved   int      `json:"files_removed"`
	BytesReclaimed int64    `json:"bytes_reclaimed"`
	Residuals      []string `json:"residuals"` // paths of cleaned residuals
	Errors         []string `json:"errors,omitempty"`
}

// Uninstall removes an application and optionally all its residuals.
// SAFETY: Defaults to UseTrash=true. DryRun MUST be run first.
func Uninstall(opts UninstallOptions) (*UninstallResult, error) {
	if opts.AppPath == "" && opts.BundleID == "" && opts.AppName == "" {
		return nil, fmt.Errorf("ka uninstall: at least one of AppPath, BundleID, or AppName is required")
	}

	result := &UninstallResult{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ka uninstall: cannot resolve home dir: %w", err)
	}

	// Step 1: Remove the .app bundle itself
	if opts.AppPath != "" {
		if _, statErr := os.Stat(opts.AppPath); statErr == nil {
			freed, err := cleaner.DeleteFile(opts.AppPath, opts.DryRun, opts.UseTrash)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("app bundle: %s", err))
			} else {
				result.AppRemoved = true
				result.BytesReclaimed += freed
				result.FilesRemoved++
				result.Residuals = append(result.Residuals, opts.AppPath)
				logging.Info("Ka: removed app bundle", "path", opts.AppPath, "dryRun", opts.DryRun)
			}
		}
	}

	// Step 2: If complete uninstall, remove all residuals
	if opts.Complete {
		residualPaths := buildResidualPaths(homeDir, opts.BundleID, opts.AppName)

		for _, pattern := range residualPaths {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				continue
			}

			for _, match := range matches {
				// Safety: validate before removal
				if err := cleaner.ValidatePath(match); err != nil {
					logging.Debug("ka uninstall: skipping protected path", "path", match, "reason", err)
					continue
				}

				freed, err := cleaner.DeleteFile(match, opts.DryRun, opts.UseTrash)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", match, err))
					continue
				}

				result.FilesRemoved++
				result.BytesReclaimed += freed
				result.Residuals = append(result.Residuals, match)
				logging.Debug("ka uninstall: removed residual", "path", match, "dryRun", opts.DryRun)
			}
		}
	}

	return result, nil
}

// buildResidualPaths generates all glob patterns where app residuals may hide.
// Uses bundle ID and app name to locate files across ~/Library and /private/var.
func buildResidualPaths(homeDir, bundleID, appName string) []string {
	var paths []string

	if bundleID != "" {
		lib := filepath.Join(homeDir, "Library")

		// Preferences
		paths = append(paths, filepath.Join(lib, "Preferences", bundleID+".plist"))
		paths = append(paths, filepath.Join(lib, "Preferences", bundleID+".*"))

		// Caches
		paths = append(paths, filepath.Join(lib, "Caches", bundleID))
		paths = append(paths, filepath.Join(lib, "Caches", bundleID+".*"))

		// Containers
		paths = append(paths, filepath.Join(lib, "Containers", bundleID))

		// Group Containers (e.g., "group.com.docker")
		paths = append(paths, filepath.Join(lib, "Group Containers", "*."+bundleID))

		// Saved Application State
		paths = append(paths, filepath.Join(lib, "Saved Application State", bundleID+".savedState"))

		// HTTP Storages
		paths = append(paths, filepath.Join(lib, "HTTPStorages", bundleID))

		// WebKit
		paths = append(paths, filepath.Join(lib, "WebKit", bundleID))

		// Launch Agents
		paths = append(paths, filepath.Join(lib, "LaunchAgents", "*"+bundleID+"*"))

		// Cookies
		paths = append(paths, filepath.Join(lib, "Cookies", bundleID+".*"))

		// Application Scripts
		paths = append(paths, filepath.Join(lib, "Application Scripts", bundleID))

		// /private/var/folders temp data
		paths = append(paths, filepath.Join("/private/var/folders", "*", "*", bundleID))
	}

	if appName != "" {
		lib := filepath.Join(homeDir, "Library")

		// Application Support (by app name)
		paths = append(paths, filepath.Join(lib, "Application Support", appName))

		// Logs
		paths = append(paths, filepath.Join(lib, "Logs", appName))

		// Also try case-insensitive common patterns
		nameLower := strings.ToLower(appName)
		if nameLower != appName {
			paths = append(paths, filepath.Join(lib, "Application Support", nameLower))
			paths = append(paths, filepath.Join(lib, "Logs", nameLower))
		}
	}

	return paths
}
