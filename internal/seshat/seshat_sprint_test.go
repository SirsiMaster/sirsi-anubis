package seshat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── Seshat Coverage Sprint Tests ────────────────────────────────────
// These tests run in ALL modes (including -short) to ensure Seshat
// coverage doesn't depend on full-mode execution.

func TestConversation_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	conv := Conversation{
		ID:           "test-123",
		Title:        "Test Conversation",
		StartedAt:    "2026-03-29T12:00:00Z",
		MessageCount: 2,
		Messages: []Message{
			{Role: "user", Content: "Hello", Timestamp: "2026-03-29T12:00:00Z"},
			{Role: "assistant", Content: "Hi there!", Timestamp: "2026-03-29T12:00:01Z"},
		},
		Metadata: ConversationMeta{
			SourceType:    "antigravity",
			Topics:        []string{"testing", "seshat"},
			SchemaVersion: SchemaVersion,
		},
	}

	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var loaded Conversation
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if loaded.ID != "test-123" {
		t.Errorf("ID = %q, want test-123", loaded.ID)
	}
	if loaded.MessageCount != 2 {
		t.Errorf("MessageCount = %d, want 2", loaded.MessageCount)
	}
	if len(loaded.Messages) != 2 {
		t.Errorf("Messages len = %d, want 2", len(loaded.Messages))
	}
	if loaded.Messages[0].Role != "user" {
		t.Errorf("Messages[0].Role = %q", loaded.Messages[0].Role)
	}
	if loaded.Metadata.SchemaVersion != SchemaVersion {
		t.Errorf("SchemaVersion = %q", loaded.Metadata.SchemaVersion)
	}
}

func TestExtractionResult_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	result := ExtractionResult{
		Source:            "gemini",
		ExtractedAt:       "2026-03-29T12:00:00Z",
		SchemaVersion:     SchemaVersion,
		ConversationCount: 1,
		Conversations:     []Conversation{{ID: "c1", Title: "Test"}},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var loaded ExtractionResult
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if loaded.Source != "gemini" {
		t.Errorf("Source = %q", loaded.Source)
	}
	if loaded.ConversationCount != 1 {
		t.Errorf("ConversationCount = %d", loaded.ConversationCount)
	}
}

func TestKnowledgeItem_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	ki := KnowledgeItem{
		Title:   "Pantheon Arch",
		Summary: "Core arch.",
		References: []KIReference{
			{Type: "file", Value: "docs/ARCH.md"},
			{Type: "conversation_id", Value: "abc-123"},
		},
	}

	data, err := json.Marshal(ki)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var loaded KnowledgeItem
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if loaded.Title != "Pantheon Arch" {
		t.Errorf("Title = %q", loaded.Title)
	}
	if len(loaded.References) != 2 {
		t.Errorf("References len = %d", len(loaded.References))
	}
}

func TestDefaultPaths_ShortMode(t *testing.T) {
	t.Parallel()

	paths := DefaultPaths()
	if paths.AntigravityDir == "" {
		t.Error("AntigravityDir empty")
	}
	if !strings.Contains(paths.KnowledgeDir, "knowledge") {
		t.Errorf("KnowledgeDir = %q", paths.KnowledgeDir)
	}
	if !strings.Contains(paths.BrainDir, "brain") {
		t.Errorf("BrainDir = %q", paths.BrainDir)
	}
	if !strings.Contains(paths.ConversationsDir, "conversations") {
		t.Errorf("ConversationsDir = %q", paths.ConversationsDir)
	}
}

func TestWriteKI_Short(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	paths := Paths{KnowledgeDir: filepath.Join(tmp, "knowledge")}
	ki := KnowledgeItem{Title: "Short Test", Summary: "Quick write test."}

	err := WriteKnowledgeItem(paths, "Short Test", ki, map[string]string{
		"overview.md": "# Test",
	})
	if err != nil {
		t.Fatalf("WriteKnowledgeItem: %v", err)
	}

	// Verify metadata
	metaPath := filepath.Join(paths.KnowledgeDir, "short_test", "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	var loaded KnowledgeItem
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if loaded.Title != "Short Test" {
		t.Errorf("Title = %q", loaded.Title)
	}

	// Verify artifact
	artPath := filepath.Join(paths.KnowledgeDir, "short_test", "artifacts", "overview.md")
	if _, err := os.Stat(artPath); err != nil {
		t.Errorf("artifact missing: %v", err)
	}
}

func TestWriteKI_LongName(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	paths := Paths{KnowledgeDir: filepath.Join(tmp, "knowledge")}
	longName := "Very Long Name That Far Exceeds The Sixty Character Maximum And Must Be Truncated"
	ki := KnowledgeItem{Title: longName}

	err := WriteKnowledgeItem(paths, longName, ki, nil)
	if err != nil {
		t.Fatalf("WriteKI(long): %v", err)
	}

	entries, _ := os.ReadDir(filepath.Join(tmp, "knowledge"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 dir, got %d", len(entries))
	}
	if len(entries[0].Name()) > 60 {
		t.Errorf("dir name %d chars, exceeds 60", len(entries[0].Name()))
	}
}

func TestListKI_Short(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	kiDir := filepath.Join(tmp, "knowledge")
	os.MkdirAll(filepath.Join(kiDir, "alpha"), 0o755)
	os.MkdirAll(filepath.Join(kiDir, "beta"), 0o755)
	os.WriteFile(filepath.Join(kiDir, "file.txt"), []byte("x"), 0o644)

	paths := Paths{KnowledgeDir: kiDir}
	items, err := ListKnowledgeItems(paths)
	if err != nil {
		t.Fatalf("ListKI: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("items = %d, want 2", len(items))
	}
}

func TestListKI_NonExistent(t *testing.T) {
	t.Parallel()
	paths := Paths{KnowledgeDir: "/no/such/path"}
	_, err := ListKnowledgeItems(paths)
	if err == nil {
		t.Error("should error on missing path")
	}
}

func TestReadKI_Short(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	paths := Paths{KnowledgeDir: filepath.Join(tmp, "knowledge")}
	ki := KnowledgeItem{Title: "Read Me", Summary: "Reading test."}
	WriteKnowledgeItem(paths, "Read Me", ki, nil)

	loaded, err := ReadKnowledgeItem(paths, "read_me")
	if err != nil {
		t.Fatalf("ReadKI: %v", err)
	}
	if loaded.Title != "Read Me" {
		t.Errorf("Title = %q", loaded.Title)
	}
}

func TestReadKI_NotFound(t *testing.T) {
	t.Parallel()
	paths := Paths{KnowledgeDir: "/no/such"}
	_, err := ReadKnowledgeItem(paths, "missing")
	if err == nil {
		t.Error("should error on missing KI")
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()
	if SchemaVersion != "1.0.0" {
		t.Errorf("SchemaVersion = %q", SchemaVersion)
	}
	if MaxSourcesPerNotebook != 50 {
		t.Errorf("MaxSourcesPerNotebook = %d", MaxSourcesPerNotebook)
	}
}
