package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_OpenClose(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()
}

func TestStore_StoreAndGet(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	entry, err := s.Store("npm test", "test-output", "PASS: all 42 tests passed in 3.2s", 12)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if entry.ID == 0 {
		t.Error("expected non-zero entry ID")
	}
	if entry.Source != "npm test" {
		t.Errorf("source = %q, want %q", entry.Source, "npm test")
	}

	got, err := s.Get(entry.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Content != "PASS: all 42 tests passed in 3.2s" {
		t.Errorf("content = %q, want original", got.Content)
	}
	if got.Tokens != 12 {
		t.Errorf("tokens = %d, want 12", got.Tokens)
	}
}

func TestStore_Search(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Store("build", "logs", "ERROR: compilation failed in main.go line 42", 15)
	s.Store("build", "logs", "WARNING: unused import in utils.go", 10)
	s.Store("test", "results", "PASS: all tests passed", 8)

	result, err := s.Search("compilation failed", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if result.TotalHits == 0 {
		t.Error("expected at least one search hit")
	}
	if result.Entries[0].Source != "build" {
		t.Errorf("top result source = %q, want %q", result.Entries[0].Source, "build")
	}
}

func TestStore_Stats(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Store("a", "tag1", "content one", 5)
	s.Store("b", "tag2", "content two", 10)
	s.Store("c", "tag1", "content three", 7)

	stats, err := s.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalEntries != 3 {
		t.Errorf("totalEntries = %d, want 3", stats.TotalEntries)
	}
	if stats.TotalTokens != 22 {
		t.Errorf("totalTokens = %d, want 22", stats.TotalTokens)
	}
	if stats.TagCounts["tag1"] != 2 {
		t.Errorf("tag1 count = %d, want 2", stats.TagCounts["tag1"])
	}
}

func TestStore_Prune(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Store("old", "logs", "old entry", 5)
	// Prune with 0 duration should prune everything (since entries are "now").
	// Prune with 1 hour should prune nothing (entries are < 1 hour old).
	removed, err := s.Prune(1 * time.Hour)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 0 {
		t.Errorf("removed = %d, want 0 (entry is fresh)", removed)
	}
}

func TestCodeIndex_IndexAndSearch(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "code.db")
	ci, err := OpenCodeIndex(dbPath)
	if err != nil {
		t.Fatalf("OpenCodeIndex: %v", err)
	}
	defer ci.Close()

	goSrc := []byte(`package main

import "fmt"

// Greet returns a greeting message.
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(Greet("world"))
}
`)

	err = ci.IndexFile("main.go", goSrc)
	if err != nil {
		t.Fatalf("IndexFile: %v", err)
	}

	chunks, err := ci.Search("Greet", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected search results for 'Greet'")
	}

	found := false
	for _, c := range chunks {
		if c.Name == "Greet" || c.File == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find Greet function in search results")
	}
}

func TestCodeIndex_IndexDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create a small Go project.
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

func main() {}
`), 0o644)
	os.WriteFile(filepath.Join(dir, "util.go"), []byte(`package main

func add(a, b int) int { return a + b }
`), 0o644)

	dbPath := filepath.Join(t.TempDir(), "code.db")
	ci, err := OpenCodeIndex(dbPath)
	if err != nil {
		t.Fatalf("OpenCodeIndex: %v", err)
	}
	defer ci.Close()

	stats, err := ci.IndexDir(dir)
	if err != nil {
		t.Fatalf("IndexDir: %v", err)
	}
	if stats.FilesIndexed != 2 {
		t.Errorf("filesIndexed = %d, want 2", stats.FilesIndexed)
	}
	if stats.ChunksCreated == 0 {
		t.Error("expected chunks to be created")
	}
}

func TestGoChunker(t *testing.T) {
	t.Parallel()
	src := []byte(`package example

type Foo struct {
	Name string
}

func (f *Foo) Bar() string {
	return f.Name
}

func standalone() {}
`)

	c := &GoChunker{}
	chunks, err := c.Chunk("example.go", src)
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}

	kinds := make(map[string]int)
	for _, ch := range chunks {
		kinds[ch.Kind]++
	}

	if kinds["type"] != 1 {
		t.Errorf("type chunks = %d, want 1", kinds["type"])
	}
	if kinds["method"] != 1 {
		t.Errorf("method chunks = %d, want 1", kinds["method"])
	}
	if kinds["function"] != 1 {
		t.Errorf("function chunks = %d, want 1", kinds["function"])
	}
}

func TestGenericChunker(t *testing.T) {
	t.Parallel()
	// Create 100 lines of content.
	lines := make([]byte, 0, 1000)
	for i := 0; i < 100; i++ {
		lines = append(lines, []byte("line of code\n")...)
	}

	c := &GenericChunker{MaxChunkLines: 30, Overlap: 10}
	chunks, err := c.Chunk("test.py", lines)
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	if len(chunks) < 3 {
		t.Errorf("expected at least 3 chunks for 100 lines with window 30, got %d", len(chunks))
	}
	for _, ch := range chunks {
		if ch.Kind != "block" {
			t.Errorf("expected kind=block, got %q", ch.Kind)
		}
	}
}
