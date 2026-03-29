package seshat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── Seshat Syncer Sprint Tests ──────────────────────────────────────
// Tests for syncer.go — no testing.Short() guards.

func TestExportKIToMarkdown(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	paths := Paths{KnowledgeDir: filepath.Join(tmp, "knowledge")}
	ki := KnowledgeItem{Title: "Export Test", Summary: "Testing export."}
	WriteKnowledgeItem(paths, "export_test", ki, map[string]string{
		"overview.md": "# Overview\nContent here.",
		"details.md":  "# Details\nMore here.",
	})

	md, err := ExportKIToMarkdown(paths, "export_test")
	if err != nil {
		t.Fatalf("ExportKIToMarkdown: %v", err)
	}

	if !strings.Contains(md, "# Export Test") {
		t.Error("missing title in export")
	}
	if !strings.Contains(md, "Testing export.") {
		t.Error("missing summary in export")
	}
	if !strings.Contains(md, "## Artifact: overview.md") {
		t.Error("missing overview artifact")
	}
	if !strings.Contains(md, "Content here.") {
		t.Error("missing overview content")
	}
}

func TestExportKIToMarkdown_NotFound(t *testing.T) {
	t.Parallel()
	paths := Paths{KnowledgeDir: "/nonexistent"}
	_, err := ExportKIToMarkdown(paths, "missing")
	if err == nil {
		t.Error("should error on missing KI")
	}
}

func TestExportAllKIsToMarkdown(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	paths := Paths{KnowledgeDir: filepath.Join(tmp, "knowledge")}

	// Create 2 KIs
	WriteKnowledgeItem(paths, "ki_alpha", KnowledgeItem{Title: "Alpha", Summary: "First."}, map[string]string{"a.md": "# A"})
	WriteKnowledgeItem(paths, "ki_beta", KnowledgeItem{Title: "Beta", Summary: "Second."}, map[string]string{"b.md": "# B"})

	outputDir := filepath.Join(tmp, "output")
	exported, err := ExportAllKIsToMarkdown(paths, outputDir)
	if err != nil {
		t.Fatalf("ExportAll: %v", err)
	}

	if len(exported) != 2 {
		t.Errorf("exported %d files, want 2", len(exported))
	}

	// Verify files exist
	for _, f := range exported {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("exported file missing: %s", f)
		}
	}
}

func TestExportAllKIsToMarkdown_Empty(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	kiDir := filepath.Join(tmp, "knowledge")
	os.MkdirAll(kiDir, 0o755)
	paths := Paths{KnowledgeDir: kiDir}

	outputDir := filepath.Join(tmp, "output")
	exported, err := ExportAllKIsToMarkdown(paths, outputDir)
	if err != nil {
		t.Fatalf("ExportAll(empty): %v", err)
	}
	if len(exported) != 0 {
		t.Errorf("exported %d, want 0", len(exported))
	}
}

func TestSyncKIToGeminiMD_Sprint(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	paths := Paths{KnowledgeDir: filepath.Join(tmp, "knowledge")}
	WriteKnowledgeItem(paths, "sync_ki", KnowledgeItem{Title: "Sync KI", Summary: "Sync test."}, nil)

	targetFile := filepath.Join(tmp, "GEMINI.md")

	// First sync — should create file
	err := SyncKIToGeminiMD(paths, "sync_ki", targetFile)
	if err != nil {
		t.Fatalf("SyncKI: %v", err)
	}

	content, _ := os.ReadFile(targetFile)
	s := string(content)
	if !strings.Contains(s, "<!-- KI:sync_ki:START -->") {
		t.Error("missing start marker")
	}
	if !strings.Contains(s, "<!-- KI:sync_ki:END -->") {
		t.Error("missing end marker")
	}
	if !strings.Contains(s, "Sync KI") {
		t.Error("missing title")
	}

	// Second sync — should replace, not duplicate
	err = SyncKIToGeminiMD(paths, "sync_ki", targetFile)
	if err != nil {
		t.Fatalf("SyncKI (2nd): %v", err)
	}

	content, _ = os.ReadFile(targetFile)
	hits := strings.Count(string(content), "<!-- KI:sync_ki:START -->")
	if hits != 1 {
		t.Errorf("start marker count = %d, want 1 (idempotent)", hits)
	}
}

func TestSyncKIToGeminiMD_NotFound(t *testing.T) {
	t.Parallel()
	paths := Paths{KnowledgeDir: "/nonexistent"}
	err := SyncKIToGeminiMD(paths, "missing", "/tmp/test.md")
	if err == nil {
		t.Error("should error on missing KI")
	}
}

func TestListBrainConversations(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	brainDir := filepath.Join(tmp, "brain")
	os.MkdirAll(filepath.Join(brainDir, "conv-001"), 0o755)
	os.MkdirAll(filepath.Join(brainDir, "conv-002"), 0o755)
	os.MkdirAll(filepath.Join(brainDir, "tempmediaStorage"), 0o755)       // should be skipped
	os.WriteFile(filepath.Join(brainDir, "file.txt"), []byte("x"), 0o644) // should be skipped

	paths := Paths{BrainDir: brainDir}
	ids, err := ListBrainConversations(paths, 0)
	if err != nil {
		t.Fatalf("ListBrain: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 conversations, got %d", len(ids))
	}
}

func TestListBrainConversations_LastN(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	brainDir := filepath.Join(tmp, "brain")
	for _, name := range []string{"c1", "c2", "c3", "c4", "c5"} {
		os.MkdirAll(filepath.Join(brainDir, name), 0o755)
	}

	paths := Paths{BrainDir: brainDir}
	ids, err := ListBrainConversations(paths, 3)
	if err != nil {
		t.Fatalf("ListBrain(3): %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 conversations (lastN=3), got %d", len(ids))
	}
}

func TestListBrainConversations_NonExistent(t *testing.T) {
	t.Parallel()
	paths := Paths{BrainDir: "/nonexistent"}
	_, err := ListBrainConversations(paths, 0)
	if err == nil {
		t.Error("should error on missing brain dir")
	}
}

func TestSaveExtractionResult(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	result := &ExtractionResult{
		Conversations: []Conversation{
			{ID: "c1", Title: "Test Conv"},
		},
	}

	outputDir := filepath.Join(tmp, "extracted")
	path, err := SaveExtractionResult(result, outputDir)
	if err != nil {
		t.Fatalf("SaveExtractionResult: %v", err)
	}

	if path == "" {
		t.Error("output path should not be empty")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("output file missing: %v", err)
	}

	// Verify JSON
	data, _ := os.ReadFile(path)
	var loaded ExtractionResult
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if loaded.ConversationCount != 1 {
		t.Errorf("ConversationCount = %d, want 1", loaded.ConversationCount)
	}
	if loaded.Source != "gemini-bridge" {
		t.Errorf("Source = %q, want gemini-bridge", loaded.Source)
	}
	if loaded.SchemaVersion != SchemaVersion {
		t.Errorf("SchemaVersion = %q", loaded.SchemaVersion)
	}
}
