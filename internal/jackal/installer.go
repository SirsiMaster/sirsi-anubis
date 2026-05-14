// Package jackal — installer file scanner.
//
// ScanInstallers searches common macOS directories for large installer files
// (.dmg, .pkg, .iso, .zip, .tar.gz, .app.zip) and reports them sorted by size.
// RemoveInstallers deletes or trashes the selected files.
package jackal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/cleaner"
)

// installerExtensions are the file extensions we consider installer files.
var installerExtensions = []string{
	".dmg", ".pkg", ".iso", ".app.zip",
}

// archiveExtensions are checked separately since .zip and .tar.gz
// are common for non-installer files too — we only include large ones.
var archiveExtensions = []string{
	".zip", ".tar.gz",
}

// minInstallerSize is the minimum file size to include (10 MB).
const minInstallerSize = 10 * 1024 * 1024

// installerTimeout is the maximum time for scanning.
const installerTimeout = 30 * time.Second

// InstallerFile represents a single installer file found on disk.
type InstallerFile struct {
	Name    string
	Path    string
	Size    int64
	Source  string // "Downloads", "Desktop", "Homebrew", "iCloud", "Mail"
	ModTime time.Time
}

// InstallerResult holds the aggregated results of an installer scan.
type InstallerResult struct {
	Files     []InstallerFile // sorted by size descending
	TotalSize int64
	ScanTime  time.Duration
}

// installerDir maps a directory path (relative to home) to a source label.
type installerDir struct {
	path  string // relative to home, or absolute
	label string
}

// installerDirs returns the directories to scan for installer files.
func installerDirs(home string) []installerDir {
	return []installerDir{
		{filepath.Join(home, "Downloads"), "Downloads"},
		{filepath.Join(home, "Desktop"), "Desktop"},
		{filepath.Join(home, "Library", "Caches", "Homebrew", "downloads"), "Homebrew"},
		{filepath.Join(home, "Library", "Mobile Documents", "com~apple~CloudDocs", "Downloads"), "iCloud"},
	}
}

// ScanInstallers searches common directories for large installer files.
// Files must be > 10 MB and match known installer extensions.
// Results are sorted by size descending.
func ScanInstallers() (*InstallerResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("installer scan: %w", err)
	}
	return ScanInstallersWithHome(home)
}

// ScanInstallersWithHome is the testable variant that accepts a home directory.
func ScanInstallersWithHome(home string) (*InstallerResult, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), installerTimeout)
	defer cancel()

	dirs := installerDirs(home)

	var (
		mu    sync.Mutex
		files []InstallerFile
		wg    sync.WaitGroup
	)

	for _, dir := range dirs {
		info, err := os.Stat(dir.path)
		if err != nil || !info.IsDir() {
			continue
		}
		wg.Add(1)
		go func(dir installerDir) {
			defer wg.Done()
			found := scanDirForInstallers(ctx, dir.path, dir.label)
			mu.Lock()
			files = append(files, found...)
			mu.Unlock()
		}(dir)
	}

	wg.Wait()

	// Sort by size descending
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	var total int64
	for _, f := range files {
		total += f.Size
	}

	return &InstallerResult{
		Files:     files,
		TotalSize: total,
		ScanTime:  time.Since(start),
	}, nil
}

// scanDirForInstallers walks a directory tree looking for installer files.
func scanDirForInstallers(ctx context.Context, root, label string) []InstallerFile {
	var results []InstallerFile

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // skip unreadable entries
		}

		// Skip hidden directories (but not root itself)
		if info.IsDir() && path != root && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Check size threshold
		if info.Size() < minInstallerSize {
			return nil
		}

		// Check if the file matches an installer extension
		if !isInstallerFile(info.Name()) {
			return nil
		}

		results = append(results, InstallerFile{
			Name:    info.Name(),
			Path:    path,
			Size:    info.Size(),
			Source:  label,
			ModTime: info.ModTime(),
		})

		return nil
	})

	return results
}

// isInstallerFile checks whether a filename has a known installer extension.
func isInstallerFile(name string) bool {
	lower := strings.ToLower(name)

	// Check compound extensions first (.app.zip, .tar.gz)
	for _, ext := range installerExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	for _, ext := range archiveExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}

	return false
}

// RemoveInstallers deletes the given installer files via the cleaner safety
// layer. Every path is validated against protected-path rules before deletion.
// If useTrash is true, uses DeleteFileReversible which refuses to permanently
// delete when the platform has no trash support. Pass useTrash=false only for
// explicit permanent deletion with user confirmation.
// Returns a CleanResult compatible with the existing rendering system.
func RemoveInstallers(files []InstallerFile, useTrash bool) (*CleanResult, error) {
	result := &CleanResult{}

	for _, f := range files {
		var freed int64
		var err error
		if useTrash {
			freed, err = cleaner.DeleteFileReversible(f.Path, false)
		} else {
			freed, err = cleaner.DeleteFile(f.Path, false, false)
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("remove %s: %w", f.Path, err))
			result.Skipped++
			continue
		}
		result.BytesFreed += freed
		result.Cleaned++
	}

	return result, nil
}
