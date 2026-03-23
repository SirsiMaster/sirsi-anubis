package horus

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndex_BuildAndQuery(t *testing.T) {
	// Use a temp cache to avoid polluting real cache.
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "manifest.json")

	home, _ := os.UserHomeDir()

	// Build index of just ~/Library (a large tree).
	start := time.Now()
	m, err := Index(IndexOptions{
		Roots:        []string{filepath.Join(home, "Library")},
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

	// Now query from cache — should be instant.
	start = time.Now()
	m2, err := Index(IndexOptions{
		Roots:     []string{filepath.Join(home, "Library")},
		MaxDepth:  5,
		CachePath: cachePath,
		TTL:       1 * time.Hour,
	})
	queryTime := time.Since(start)
	if err != nil {
		t.Fatalf("Cached load failed: %v", err)
	}

	t.Logf("CACHE LOAD: %d entries in %s", len(m2.Entries), queryTime.Round(time.Millisecond))

	// Query: DirSize of Library/Caches
	start = time.Now()
	cachesSize := m.DirSize(filepath.Join(home, "Library", "Caches"))
	cachesDur := time.Since(start)
	t.Logf("QUERY DirSize(Library/Caches): %d bytes in %s", cachesSize, cachesDur)

	// Query: DirSizeAndCount
	start = time.Now()
	size, count := m.DirSizeAndCount(filepath.Join(home, "Library", "Caches"))
	queryDur := time.Since(start)
	t.Logf("QUERY DirSizeAndCount(Library/Caches): %d bytes, %d files in %s", size, count, queryDur)

	// Verify cache was saved.
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("Cache file was not created")
	}
}

func TestManifest_Exists(t *testing.T) {
	m := &Manifest{
		Entries: map[string]Entry{
			"/foo/bar": {Size: 100},
		},
	}
	if !m.Exists("/foo/bar") {
		t.Error("Expected /foo/bar to exist")
	}
	if m.Exists("/foo/baz") {
		t.Error("Expected /foo/baz to not exist")
	}
}

func TestManifest_DirSizeAndCount(t *testing.T) {
	m := &Manifest{
		Entries: map[string]Entry{
			"/data":          {IsDir: true},
			"/data/a.txt":    {Size: 100},
			"/data/b.txt":    {Size: 200},
			"/data/sub":      {IsDir: true},
			"/data/sub/c.go": {Size: 50},
			"/other/d.txt":   {Size: 999},
		},
	}

	size, count := m.DirSizeAndCount("/data")
	if size != 350 {
		t.Errorf("Expected size 350, got %d", size)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}
