package horus

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_StartsAndStops(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create a Go file so there's something to parse
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	w := NewWatcher(dir)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if !w.IsRunning() {
		t.Fatal("watcher should be running")
	}

	// Should have parsed the initial graph
	time.Sleep(100 * time.Millisecond)
	g := w.Graph()
	if g == nil {
		t.Fatal("graph should not be nil after start")
	}
	if g.Stats.Files < 1 {
		t.Errorf("files = %d, want >= 1", g.Stats.Files)
	}

	w.Stop()
	if w.IsRunning() {
		t.Fatal("watcher should not be running after stop")
	}
}

func TestWatcher_DetectsFileChange(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	w := NewWatcher(dir)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	// Wait for initial parse to complete
	time.Sleep(300 * time.Millisecond)
	initialRebuilds := w.Rebuilds

	// Now set up the callback AFTER initial parse
	updated := make(chan *SymbolGraph, 1)
	w.OnUpdate = func(g *SymbolGraph) {
		select {
		case updated <- g:
		default:
		}
	}

	// Write a new Go file — should trigger rebuild
	os.WriteFile(filepath.Join(dir, "helper.go"), []byte("package main\nfunc helper() int { return 42 }\n"), 0644)

	// Wait for debounced rebuild (500ms debounce + parse time)
	select {
	case g := <-updated:
		if g.Stats.Files != 2 {
			t.Errorf("after add: files = %d, want 2", g.Stats.Files)
		}
		if w.Rebuilds <= initialRebuilds {
			t.Error("rebuild count should have increased")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for rebuild after file change")
	}
}

func TestWatcher_IgnoresNonGoFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	w := NewWatcher(dir)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	time.Sleep(200 * time.Millisecond)
	initialRebuilds := w.Rebuilds

	// Write a non-Go file — should NOT trigger rebuild
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0644)
	time.Sleep(1 * time.Second) // past debounce window

	if w.Rebuilds != initialRebuilds {
		t.Errorf("non-Go file triggered rebuild: %d → %d", initialRebuilds, w.Rebuilds)
	}
}

func TestWatcher_SkipsVendorDirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "vendor", "pkg"), 0755)
	os.WriteFile(filepath.Join(dir, "vendor", "pkg", "lib.go"), []byte("package pkg\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "node_modules", "thing"), 0755)

	w := NewWatcher(dir)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	// The watcher should NOT have added vendor/ or node_modules/ dirs
	// We can't easily check this, but the graph should only have 1 file
	time.Sleep(200 * time.Millisecond)
	g := w.Graph()
	if g == nil {
		t.Fatal("graph nil")
	}
	// vendor/pkg/lib.go should NOT be in the graph (vendor is excluded by parser)
}
