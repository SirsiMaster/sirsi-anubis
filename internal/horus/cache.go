package horus

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// Cache stores parsed SymbolGraphs keyed by project root.
// Uses GOB encoding for fast serialization. Invalidation is manual.
type Cache struct {
	dir string
}

// NewCache creates a cache in the default directory.
func NewCache() *Cache {
	home, _ := os.UserHomeDir()
	return &Cache{dir: filepath.Join(home, ".config", "sirsi", "horus")}
}

// Get loads a cached graph for the given root. Returns nil if not cached.
func (c *Cache) Get(root string) (*SymbolGraph, bool) {
	path := c.cachePath(root)
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	var g SymbolGraph
	if err := gob.NewDecoder(f).Decode(&g); err != nil {
		return nil, false
	}
	return &g, true
}

// Put stores a graph in the cache.
func (c *Cache) Put(root string, g *SymbolGraph) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	path := c.cachePath(root)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(g)
}

// Invalidate removes the cached graph for the given root.
func (c *Cache) Invalidate(root string) {
	os.Remove(c.cachePath(root))
}

func (c *Cache) cachePath(root string) string {
	// Use a safe filename derived from the root path.
	safe := filepath.Base(root)
	if safe == "" || safe == "." || safe == "/" {
		safe = "default"
	}
	return filepath.Join(c.dir, safe+".gob")
}
