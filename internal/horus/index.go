// Package horus implements the All-Seeing Eye — a shared filesystem index
// that all deities query instead of independently walking the filesystem.
//
// Named after Horus, the falcon-headed god whose eye sees all.
//
// Architecture (ADR-008):
//   - Walk once: parallel goroutine tree traversal builds the index
//   - Share many: all deities (Jackal, Ka, Seba) query the index
//   - Cache: index persists to disk with a configurable TTL
//   - Phase 2: pre-aggregated directory summaries + gob encoding
//
// Performance: 856K file entries → ~50K directory summaries.
// DirSizeAndCount: O(n) scan → O(1) hash lookup.
// Cache: 110MB JSON/936ms → ~3MB gob/<50ms.
package horus

import (
	"bytes"
	"encoding/gob"
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
// Phase 2: stores pre-aggregated directory summaries instead of every file.
type Manifest struct {
	Version   string                `json:"version"`
	Platform  string                `json:"platform"`
	Timestamp time.Time             `json:"timestamp"`
	Roots     []string              `json:"roots"`
	Dirs      map[string]DirSummary `json:"dirs"`    // directory path → summary
	Entries   map[string]Entry      `json:"entries"` // legacy: file entries (only if needed)
	Stats     WalkStats             `json:"stats"`
}

// DirSummary is the pre-aggregated summary of a directory.
// This is the key Phase 2 optimization: O(1) lookup instead of O(n) scan.
type DirSummary struct {
	TotalSize int64 `json:"ts"` // total size of all files recursively
	FileCount int   `json:"fc"` // number of files recursively
	DirCount  int   `json:"dc"` // number of subdirectories
}

// Entry is a single filesystem entry in the index (kept for Exists/Glob queries).
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
	return filepath.Join(home, ".config", "pantheon", "horus", "manifest.gob")
}

// DefaultTTL is how long a cached manifest is considered fresh.
const DefaultTTL = 5 * time.Minute

// DefaultMaxDepth prevents runaway indexing.
const DefaultMaxDepth = 6

// IndexOptions configures the indexing behavior.
type IndexOptions struct {
	Roots        []string
	MaxDepth     int
	CachePath    string
	TTL          time.Duration
	ForceRefresh bool
	Parallelism  int
}

// DefaultRoots returns the standard filesystem roots that deities care about.
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

// Index builds or loads the shared filesystem manifest.
func Index(opts IndexOptions) (*Manifest, error) {
	cachePath := opts.CachePath
	if cachePath == "" {
		cachePath = DefaultCachePath()
	}

	ttl := opts.TTL
	if ttl == 0 {
		ttl = DefaultTTL
	}

	// Try gob cache first.
	if !opts.ForceRefresh {
		if m, err := LoadManifest(cachePath); err == nil {
			if time.Since(m.Timestamp) < ttl {
				return m, nil
			}
		}
		// Try legacy JSON cache as fallback.
		jsonPath := strings.TrimSuffix(cachePath, ".gob") + ".json"
		if m, err := loadJSONManifest(jsonPath); err == nil {
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

	// Persist to cache (gob format).
	_ = SaveManifest(cachePath, m)

	return m, nil
}

// buildIndex performs parallel filesystem traversal with pre-aggregation.
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

	// Collect all files per root, then aggregate.
	type fileEntry struct {
		dir  string
		size int64
	}

	allDirs := make(map[string]bool)
	allFiles := make([]fileEntry, 0, 100000)
	var mu sync.Mutex
	var totalDirs, totalFiles int

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

			localDirs := make(map[string]bool)
			localFiles := make([]fileEntry, 0, 10000)
			localDirCount := 0
			localFileCount := 0

			_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}

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

				if d.IsDir() {
					localDirs[path] = true
					localDirCount++
				} else {
					info, err := d.Info()
					if err != nil {
						return nil
					}
					dir := filepath.Dir(path)
					localFiles = append(localFiles, fileEntry{dir: dir, size: info.Size()})
					localFileCount++
				}

				return nil
			})

			mu.Lock()
			for k := range localDirs {
				allDirs[k] = true
			}
			allFiles = append(allFiles, localFiles...)
			totalDirs += localDirCount
			totalFiles += localFileCount
			mu.Unlock()
		}(root)
	}

	wg.Wait()

	// Phase 2: Pre-aggregate directory summaries.
	// For each directory, sum the sizes and counts of all files beneath it.
	dirSummaries := make(map[string]DirSummary, len(allDirs))

	// Initialize all directories.
	for dir := range allDirs {
		dirSummaries[dir] = DirSummary{}
	}

	// Aggregate: for each file, add its size to ALL ancestor directories.
	for _, f := range allFiles {
		dir := f.dir
		for {
			s := dirSummaries[dir]
			s.TotalSize += f.size
			s.FileCount++
			dirSummaries[dir] = s

			parent := filepath.Dir(dir)
			if parent == dir {
				break // reached root
			}
			if _, exists := allDirs[parent]; !exists {
				break // parent not in our indexed roots
			}
			dir = parent
		}
	}

	// Count subdirectories per directory.
	for dir := range allDirs {
		parent := filepath.Dir(dir)
		if parent != dir {
			if s, ok := dirSummaries[parent]; ok {
				s.DirCount++
				dirSummaries[parent] = s
			}
		}
	}

	// Also store a minimal entries map for Exists/Glob queries (directories only).
	entries := make(map[string]Entry, len(allDirs))
	for dir := range allDirs {
		entries[dir] = Entry{IsDir: true}
	}

	return &Manifest{
		Version:   "2.0.0",
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Timestamp: time.Now(),
		Roots:     roots,
		Dirs:      dirSummaries,
		Entries:   entries,
		Stats: WalkStats{
			DirsWalked:   totalDirs,
			FilesIndexed: totalFiles,
			WalkDuration: time.Since(start),
			Parallelism:  parallelism,
		},
	}, nil
}

// ─── Query API ───────────────────────────────────────────────────────────

// DirSize returns the total size of all files under a directory path.
// Phase 2: O(1) hash lookup instead of O(n) scan.
func (m *Manifest) DirSize(dir string) int64 {
	if s, ok := m.Dirs[dir]; ok {
		return s.TotalSize
	}
	return 0
}

// DirCount returns the number of files under a directory path.
func (m *Manifest) DirCount(dir string) int {
	if s, ok := m.Dirs[dir]; ok {
		return s.FileCount
	}
	return 0
}

// DirSizeAndCount returns both size and count in O(1).
func (m *Manifest) DirSizeAndCount(dir string) (int64, int) {
	if s, ok := m.Dirs[dir]; ok {
		return s.TotalSize, s.FileCount
	}
	return 0, 0
}

// Exists checks if a path exists in the index.
func (m *Manifest) Exists(path string) bool {
	_, ok := m.Entries[path]
	if ok {
		return true
	}
	_, ok = m.Dirs[path]
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

// EntriesUnder returns all directory entries under a prefix.
func (m *Manifest) EntriesUnder(dir string) map[string]DirSummary {
	prefix := dir + string(filepath.Separator)
	result := make(map[string]DirSummary)
	for path, summary := range m.Dirs {
		if strings.HasPrefix(path, prefix) {
			result[path] = summary
		}
	}
	return result
}

// ─── Persistence ─────────────────────────────────────────────────────────

// SaveManifest persists the manifest to disk using gob encoding.
// Phase 2: gob is ~5× faster to decode than JSON and ~20× smaller.
func SaveManifest(path string, m *Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		return err
	}

	// Atomic write: temp file + rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadManifest reads a cached manifest from disk (gob format).
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// loadJSONManifest reads a legacy JSON manifest (Phase 1 format).
func loadJSONManifest(path string) (*Manifest, error) {
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
		"👁️ Horus v%s: %d dirs, %d files indexed in %s (%d goroutines)",
		m.Version,
		m.Stats.DirsWalked,
		m.Stats.FilesIndexed,
		m.Stats.WalkDuration.Round(time.Millisecond),
		m.Stats.Parallelism,
	)
}
