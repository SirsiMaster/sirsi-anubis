package horus

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndex_BuildAndQuery(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "manifest.gob")

	home, _ := os.UserHomeDir()

	// Build index of ~/Library (a large tree).
	start := time.Now()
	m, err := Index(IndexOptions{
		Roots:        []string{filepath.Join(home, "Library", "Caches")},
		MaxDepth:     5,
		CachePath:    cachePath,
		ForceRefresh: true,
	})
	buildTime := time.Since(start)
	if err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	t.Logf("BUILD: %d dirs, %d files in %s",
		m.Stats.DirsWalked, m.Stats.FilesIndexed, buildTime.Round(time.Millisecond))
	t.Logf("DIR SUMMARIES: %d entries (vs %d files — %.0fx reduction)",
		len(m.Dirs), m.Stats.FilesIndexed,
		float64(m.Stats.FilesIndexed)/float64(len(m.Dirs)))

	// Check cache file size.
	if info, statErr := os.Stat(cachePath); statErr == nil {
		t.Logf("CACHE SIZE: %.1f MB (gob)", float64(info.Size())/1024/1024)
	}

	// Load from cache — should be very fast.
	start = time.Now()
	m2, err := Index(IndexOptions{
		Roots:     []string{filepath.Join(home, "Library", "Caches")},
		MaxDepth:  5,
		CachePath: cachePath,
		TTL:       1 * time.Hour,
	})
	queryTime := time.Since(start)
	if err != nil {
		t.Fatalf("Cached load failed: %v", err)
	}

	t.Logf("CACHE LOAD: %d dir summaries in %s", len(m2.Dirs), queryTime.Round(time.Millisecond))

	// Query: DirSizeAndCount — O(1) lookup.
	cachesPath := filepath.Join(home, "Library", "Caches")
	start = time.Now()
	size, count := m.DirSizeAndCount(cachesPath)
	queryDur := time.Since(start)
	t.Logf("QUERY DirSizeAndCount(Library/Caches): %d bytes, %d files in %s", size, count, queryDur)

	if size == 0 {
		t.Error("Expected non-zero size for Library/Caches")
	}
}

func TestManifest_Exists(t *testing.T) {
	m := &Manifest{
		Entries: map[string]Entry{
			"/foo/bar": {Size: 100},
		},
		Dirs: map[string]DirSummary{
			"/foo": {TotalSize: 100, FileCount: 1},
		},
	}
	if !m.Exists("/foo/bar") {
		t.Error("Expected /foo/bar to exist")
	}
	if !m.Exists("/foo") {
		t.Error("Expected /foo to exist via Dirs")
	}
	if m.Exists("/foo/baz") {
		t.Error("Expected /foo/baz to not exist")
	}
}

func TestManifest_DirSizeAndCount(t *testing.T) {
	m := &Manifest{
		Dirs: map[string]DirSummary{
			"/data":     {TotalSize: 350, FileCount: 3, DirCount: 1},
			"/data/sub": {TotalSize: 50, FileCount: 1},
			"/other":    {TotalSize: 999, FileCount: 1},
		},
	}

	size, count := m.DirSizeAndCount("/data")
	if size != 350 {
		t.Errorf("Expected size 350, got %d", size)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Test O(1): subdir lookup.
	size2, count2 := m.DirSizeAndCount("/data/sub")
	if size2 != 50 || count2 != 1 {
		t.Errorf("Expected size 50/count 1, got %d/%d", size2, count2)
	}

	// Non-existent directory.
	size3, count3 := m.DirSizeAndCount("/nonexistent")
	if size3 != 0 || count3 != 0 {
		t.Errorf("Expected 0/0, got %d/%d", size3, count3)
	}
}
