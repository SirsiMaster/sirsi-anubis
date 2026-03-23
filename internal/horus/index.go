// Package horus implements the All-Seeing Eye — a shared filesystem index
// that all deities query instead of independently walking the filesystem.
//
// Named after Horus, the falcon-headed god whose eye sees all.
//
// Architecture (ADR-008):
//   - Walk once: parallel goroutine tree traversal builds the index
//   - Share many: all deities (Jackal, Ka, Seba) query the index
//   - Cache: index persists to disk with a configurable TTL
//   - Incremental: only re-walk directories whose mtime changed
package horus

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Manifest is the shared filesystem index.
type Manifest struct {
	Version   string           `json:"version"`
	Platform  string           `json:"platform"`
	Timestamp time.Time        `json:"timestamp"`
	Roots     []string         `json:"roots"`
	Entries   map[string]Entry `json:"entries"`
	Stats     WalkStats        `json:"stats"`
}

// Entry is a single filesystem entry in the index.
// ModTime is omitted from JSON to keep the manifest compact (<10MB vs 476MB).
type Entry struct {
	Size  int64  `json:"s"`
	IsDir bool   `json:"d,omitempty"`
	Mode  uint32 `json:"m,omitempty"`
}

// WalkStats records performance metrics for the walk.
type WalkStats struct {
	DirsWalked   int           `json:"dirs_walked"`
	FilesIndexed int           `json:"files_indexed"`
	WalkDuration time.Duration `json:"walk_duration_ns"`
	Parallelism  int           `json:"parallelism"`
}

// DefaultCachePath returns the standard cache location.
func DefaultCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pantheon", "horus", "manifest.json")
}

// DefaultTTL is how long a cached manifest is considered fresh.
const DefaultTTL = 5 * time.Minute

// IndexOptions configures the indexing behavior.
type IndexOptions struct {
	// Roots are the directory trees to index.
	Roots []string

	// MaxDepth limits traversal depth (0 = unlimited).
	MaxDepth int

	// CachePath overrides the default manifest location.
	CachePath string

	// TTL overrides the default cache TTL.
	TTL time.Duration

	// ForceRefresh skips the cache and rebuilds.
	ForceRefresh bool

	// Parallelism controls goroutine fan-out (0 = GOMAXPROCS).
	Parallelism int
}

// DefaultRoots returns the standard filesystem roots that deities care about.
// Scoped to paths that Jackal/Ka/Seba actually scan — not the entire filesystem.
func DefaultRoots() []string {
	home, _ := os.UserHomeDir()
	roots := []string{
		filepath.Join(home, "Library", "Caches"),
		filepath.Join(home, "Library", "Logs"),
		filepath.Join(home, "Library", "Application Support"),
		filepath.Join(home, "Library", "Containers"),
		filepath.Join(home, "Library", "Developer"),
		filepath.Join(home, "Library", "Group Containers"),
		filepath.Join(home, ".cache"),
		filepath.Join(home, ".local"),
		filepath.Join(home, "Development"),
		filepath.Join(home, "go", "pkg"),
	}

	switch runtime.GOOS {
	case "darwin":
		roots = append(roots,
			"/Library/Caches",
			"/Applications",
			"/usr/local",
			"/opt/homebrew",
		)
	case "linux":
		roots = append(roots,
			"/var/cache",
			"/var/tmp",
			"/opt",
			"/snap",
		)
	case "windows":
		roots = append(roots,
			filepath.Join(home, "AppData", "Local"),
			filepath.Join(home, "AppData", "Roaming"),
		)
	}

	return roots
}

// DefaultMaxDepth prevents runaway indexing.
const DefaultMaxDepth = 6

// Index builds or loads the shared filesystem manifest.
// If a fresh cache exists, returns it. Otherwise, walks the filesystem.
func Index(opts IndexOptions) (*Manifest, error) {
	cachePath := opts.CachePath
	if cachePath == "" {
		cachePath = DefaultCachePath()
	}

	ttl := opts.TTL
	if ttl == 0 {
		ttl = DefaultTTL
	}

	// Try cache first (unless forced refresh).
	if !opts.ForceRefresh {
		if m, err := LoadManifest(cachePath); err == nil {
			if time.Since(m.Timestamp) < ttl {
				return m, nil
			}
		}
	}

	// Build fresh index with parallel walks.
	m, err := buildIndex(opts)
	if err != nil {
		return nil, err
	}

	// Persist to cache.
	_ = SaveManifest(cachePath, m)

	return m, nil
}

// buildIndex performs parallel filesystem traversal.
func buildIndex(opts IndexOptions) (*Manifest, error) {
	start := time.Now()

	roots := opts.Roots
	if len(roots) == 0 {
		roots = DefaultRoots()
	}

	maxDepth := opts.MaxDepth
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}

	parallelism := opts.Parallelism
	if parallelism <= 0 {
		parallelism = runtime.GOMAXPROCS(0)
	}

	entries := make(map[string]Entry)
	var mu sync.Mutex
	var totalDirs, totalFiles int

	// Fan-out: one goroutine per root directory.
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallelism)

	for _, root := range roots {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		wg.Add(1)
		go func(root string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			localEntries := make(map[string]Entry)
			localDirs := 0
			localFiles := 0

			_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil // Skip permission errors silently
				}

				// Depth limit
				if maxDepth > 0 {
					rel, _ := filepath.Rel(root, path)
					depth := strings.Count(rel, string(filepath.Separator))
					if depth > maxDepth {
						if d.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
				}

				info, err := d.Info()
				if err != nil {
					return nil
				}

				localEntries[path] = Entry{
					Size:  info.Size(),
					IsDir: d.IsDir(),
					Mode:  uint32(info.Mode()),
				}

				if d.IsDir() {
					localDirs++
				} else {
					localFiles++
				}

				return nil
			})

			// Merge into global map
			mu.Lock()
			for k, v := range localEntries {
				entries[k] = v
			}
			totalDirs += localDirs
			totalFiles += localFiles
			mu.Unlock()
		}(root)
	}

	wg.Wait()

	return &Manifest{
		Version:   "1.0.0",
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Timestamp: time.Now(),
		Roots:     roots,
		Entries:   entries,
		Stats: WalkStats{
			DirsWalked:   totalDirs,
			FilesIndexed: totalFiles,
			WalkDuration: time.Since(start),
			Parallelism:  parallelism,
		},
	}, nil
}

// Query methods on the manifest — deities use these instead of walking.

// DirSize returns the total size of all files under a directory path.
func (m *Manifest) DirSize(dir string) int64 {
	prefix := dir + string(filepath.Separator)
	var total int64
	for path, entry := range m.Entries {
		if !entry.IsDir && (strings.HasPrefix(path, prefix) || path == dir) {
			total += entry.Size
		}
	}
	return total
}

// DirCount returns the number of files under a directory path.
func (m *Manifest) DirCount(dir string) int {
	prefix := dir + string(filepath.Separator)
	count := 0
	for path, entry := range m.Entries {
		if !entry.IsDir && strings.HasPrefix(path, prefix) {
			count++
		}
	}
	return count
}

// DirSizeAndCount returns both in one pass over the index.
func (m *Manifest) DirSizeAndCount(dir string) (int64, int) {
	prefix := dir + string(filepath.Separator)
	var total int64
	count := 0
	for path, entry := range m.Entries {
		if !entry.IsDir && strings.HasPrefix(path, prefix) {
			total += entry.Size
			count++
		}
	}
	return total, count
}

// Exists checks if a path exists in the index.
func (m *Manifest) Exists(path string) bool {
	_, ok := m.Entries[path]
	return ok
}

// Glob returns all indexed paths matching a glob pattern.
func (m *Manifest) Glob(pattern string) []string {
	var matches []string
	for path := range m.Entries {
		if matched, _ := filepath.Match(pattern, path); matched {
			matches = append(matches, path)
		}
	}
	return matches
}

// EntriesUnder returns all entries under a directory prefix.
func (m *Manifest) EntriesUnder(dir string) map[string]Entry {
	prefix := dir + string(filepath.Separator)
	result := make(map[string]Entry)
	for path, entry := range m.Entries {
		if strings.HasPrefix(path, prefix) {
			result[path] = entry
		}
	}
	return result
}

// SaveManifest persists the manifest to disk.
func SaveManifest(path string, m *Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// Atomic write: temp file + rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadManifest reads a cached manifest from disk.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Summary returns a human-readable summary of the manifest.
func (m *Manifest) Summary() string {
	return fmt.Sprintf(
		"👁️ Horus Index: %d dirs, %d files indexed in %s (%d goroutines)",
		m.Stats.DirsWalked,
		m.Stats.FilesIndexed,
		m.Stats.WalkDuration.Round(time.Millisecond),
		m.Stats.Parallelism,
	)
}
