package thoth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompact_AppendsSessionDecisions(t *testing.T) {
	tmp := t.TempDir()
	thothDir := filepath.Join(tmp, ".thoth")
	os.MkdirAll(thothDir, 0755)

	// Create minimal memory.yaml
	memPath := filepath.Join(thothDir, "memory.yaml")
	os.WriteFile(memPath, []byte("project: test\nversion: 1.0\n"), 0644)

	// Create minimal journal
	journalPath := filepath.Join(thothDir, "journal.md")
	os.WriteFile(journalPath, []byte("# Journal\n---\n"), 0644)

	err := Compact(CompactOptions{
		RepoRoot: tmp,
		Summary:  "Use interface-based providers\nKeep backward compat",
	})
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}

	// Check memory.yaml has session decisions
	data, _ := os.ReadFile(memPath)
	content := string(data)
	if !strings.Contains(content, "## Session Decisions") {
		t.Error("expected Session Decisions section in memory.yaml")
	}
	if !strings.Contains(content, "Use interface-based providers") {
		t.Error("expected decision text in memory.yaml")
	}
}

func TestCompact_CreatesJournalEntry(t *testing.T) {
	tmp := t.TempDir()
	thothDir := filepath.Join(tmp, ".thoth")
	os.MkdirAll(thothDir, 0755)
	os.WriteFile(filepath.Join(thothDir, "memory.yaml"), []byte("project: test\n"), 0644)
	os.WriteFile(filepath.Join(thothDir, "journal.md"), []byte("# Journal\n---\n"), 0644)

	err := Compact(CompactOptions{RepoRoot: tmp, Summary: "Decision one"})
	if err != nil {
		t.Fatalf("Compact: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(thothDir, "journal.md"))
	content := string(data)
	if !strings.Contains(content, "(COMPACT)") {
		t.Error("expected (COMPACT) marker in journal entry")
	}
	if !strings.Contains(content, "Decision one") {
		t.Error("expected decision text in journal entry")
	}
}

func TestCompact_IdempotentSection(t *testing.T) {
	tmp := t.TempDir()
	thothDir := filepath.Join(tmp, ".thoth")
	os.MkdirAll(thothDir, 0755)
	os.WriteFile(filepath.Join(thothDir, "memory.yaml"), []byte("project: test\n"), 0644)
	os.WriteFile(filepath.Join(thothDir, "journal.md"), []byte("# Journal\n---\n"), 0644)

	// Compact twice
	Compact(CompactOptions{RepoRoot: tmp, Summary: "First decision"})
	Compact(CompactOptions{RepoRoot: tmp, Summary: "Second decision"})

	data, _ := os.ReadFile(filepath.Join(thothDir, "memory.yaml"))
	content := string(data)
	count := strings.Count(content, "## Session Decisions")
	if count != 1 {
		t.Errorf("expected exactly 1 Session Decisions header, got %d", count)
	}
	if !strings.Contains(content, "First decision") {
		t.Error("expected first decision")
	}
	if !strings.Contains(content, "Second decision") {
		t.Error("expected second decision")
	}
}

func TestCompact_EmptySummary(t *testing.T) {
	tmp := t.TempDir()
	thothDir := filepath.Join(tmp, ".thoth")
	os.MkdirAll(thothDir, 0755)
	os.WriteFile(filepath.Join(thothDir, "memory.yaml"), []byte("project: test\n"), 0644)

	err := Compact(CompactOptions{RepoRoot: tmp, Summary: ""})
	if err == nil {
		t.Error("expected error for empty summary")
	}
}

func TestCompact_MissingThothDir(t *testing.T) {
	err := Compact(CompactOptions{RepoRoot: t.TempDir(), Summary: "test"})
	if err == nil {
		t.Error("expected error for missing .thoth dir")
	}
}

func TestPruneJournal_MaxKeep(t *testing.T) {
	tmp := t.TempDir()
	thothDir := filepath.Join(tmp, ".thoth")
	os.MkdirAll(thothDir, 0755)

	journal := `# Journal
---

## Entry 001 — 2026-01-01 — Old entry

Content 1

---

## Entry 002 — 2026-02-01 — Middle entry

Content 2

---

## Entry 003 — 2026-03-01 — Recent entry

Content 3

---
`
	os.WriteFile(filepath.Join(thothDir, "journal.md"), []byte(journal), 0644)

	removed, err := PruneJournal(PruneOptions{RepoRoot: tmp, MaxKeep: 2})
	if err != nil {
		t.Fatalf("PruneJournal: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	data, _ := os.ReadFile(filepath.Join(thothDir, "journal.md"))
	content := string(data)
	if strings.Contains(content, "Entry 001") {
		t.Error("entry 001 should have been pruned")
	}
	if !strings.Contains(content, "Entry 002") {
		t.Error("entry 002 should be kept")
	}
	if !strings.Contains(content, "Entry 003") {
		t.Error("entry 003 should be kept")
	}
}

func TestPruneJournal_PreservesHeader(t *testing.T) {
	tmp := t.TempDir()
	thothDir := filepath.Join(tmp, ".thoth")
	os.MkdirAll(thothDir, 0755)

	journal := `# Engineering Journal
# Description line
---

## Entry 001 — 2026-01-01 — Only entry

Content

---
`
	os.WriteFile(filepath.Join(thothDir, "journal.md"), []byte(journal), 0644)

	removed, err := PruneJournal(PruneOptions{RepoRoot: tmp, MaxKeep: 1})
	if err != nil {
		t.Fatalf("PruneJournal: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed (only 1 entry, keep 1), got %d", removed)
	}

	data, _ := os.ReadFile(filepath.Join(thothDir, "journal.md"))
	if !strings.Contains(string(data), "# Engineering Journal") {
		t.Error("header should be preserved")
	}
}

func TestPruneJournal_NoEntries(t *testing.T) {
	tmp := t.TempDir()
	thothDir := filepath.Join(tmp, ".thoth")
	os.MkdirAll(thothDir, 0755)
	os.WriteFile(filepath.Join(thothDir, "journal.md"), []byte("# Journal\n---\n"), 0644)

	removed, err := PruneJournal(PruneOptions{RepoRoot: tmp, MaxKeep: 5})
	if err != nil {
		t.Fatalf("PruneJournal: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}
