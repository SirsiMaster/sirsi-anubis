// Package jackal — analyze.go
//
// Directory size analyzer with drill-down support. Walks a directory
// one level deep, sums sizes recursively, and returns entries sorted
// by size descending. Children are populated lazily on drill-down.
package jackal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// DirEntry represents a single file or directory with its aggregate size.
type DirEntry struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	ModTime  time.Time
	Children []DirEntry // populated on drill-down
}

// AnalyzeResult holds the result of a directory analysis.
type AnalyzeResult struct {
	Path      string
	TotalSize int64
	Entries   []DirEntry // sorted by size descending
	ScanTime  time.Duration
}

// maxAnalyzeEntries is the cap per level.
const maxAnalyzeEntries = 20

// Analyze scans a directory and returns entries sorted by size.
// Only reads one level deep; children are populated on drill-down.
// maxDepth controls how deep the recursive size calculation goes
// (0 = unlimited).
func Analyze(path string, maxDepth int) (*AnalyzeResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return AnalyzeCtx(ctx, path, maxDepth)
}

// AnalyzeCtx is the context-aware version of Analyze.
func AnalyzeCtx(ctx context.Context, path string, maxDepth int) (*AnalyzeResult, error) {
	start := time.Now()

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("analyze: stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("analyze: %s is not a directory", path)
	}

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("analyze: read %s: %w", path, err)
	}

	// Scan entries in parallel for speed
	type entryResult struct {
		entry DirEntry
		err   error
	}

	var (
		wg      sync.WaitGroup
		results = make([]entryResult, len(dirEntries))
	)

	// Limit concurrency to avoid fd exhaustion
	sem := make(chan struct{}, 8)

	for i, de := range dirEntries {
		// Skip hidden dirs by default
		name := de.Name()
		if name[0] == '.' {
			continue
		}

		wg.Add(1)
		go func(idx int, de os.DirEntry) {
			defer wg.Done()

			// Check context before starting
			select {
			case <-ctx.Done():
				results[idx] = entryResult{err: ctx.Err()}
				return
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			entryPath := filepath.Join(path, de.Name())
			fi, err := de.Info()
			if err != nil {
				results[idx] = entryResult{err: err}
				return
			}

			entry := DirEntry{
				Name:    de.Name(),
				Path:    entryPath,
				IsDir:   de.IsDir(),
				ModTime: fi.ModTime(),
			}

			if de.IsDir() {
				entry.Size = dirSize(ctx, entryPath, maxDepth, 1)
			} else {
				entry.Size = fi.Size()
			}

			results[idx] = entryResult{entry: entry}
		}(i, de)
	}

	wg.Wait()

	var entries []DirEntry
	var totalSize int64
	for _, r := range results {
		if r.err != nil || r.entry.Name == "" {
			continue
		}
		entries = append(entries, r.entry)
		totalSize += r.entry.Size
	}

	// Sort by size descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	// Cap entries
	if len(entries) > maxAnalyzeEntries {
		entries = entries[:maxAnalyzeEntries]
	}

	return &AnalyzeResult{
		Path:      path,
		TotalSize: totalSize,
		Entries:   entries,
		ScanTime:  time.Since(start),
	}, nil
}

// dirSize recursively calculates the total size of a directory.
// Uses atomic counter and goroutine-per-subdirectory for speed.
func dirSize(ctx context.Context, path string, maxDepth, currentDepth int) int64 {
	if maxDepth > 0 && currentDepth > maxDepth {
		return 0
	}

	select {
	case <-ctx.Done():
		return 0
	default:
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	var total atomic.Int64

	// For shallow directories, just iterate; for deep ones, parallelize
	if len(entries) < 50 || currentDepth > 3 {
		for _, e := range entries {
			select {
			case <-ctx.Done():
				return total.Load()
			default:
			}

			ep := filepath.Join(path, e.Name())
			if e.IsDir() {
				total.Add(dirSize(ctx, ep, maxDepth, currentDepth+1))
			} else {
				if fi, err := e.Info(); err == nil {
					total.Add(fi.Size())
				}
			}
		}
	} else {
		var wg sync.WaitGroup
		sem := make(chan struct{}, 4)
		for _, e := range entries {
			ep := filepath.Join(path, e.Name())
			if e.IsDir() {
				wg.Add(1)
				go func(p string) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()
					total.Add(dirSize(ctx, p, maxDepth, currentDepth+1))
				}(ep)
			} else {
				if fi, err := e.Info(); err == nil {
					total.Add(fi.Size())
				}
			}
		}
		wg.Wait()
	}

	return total.Load()
}
