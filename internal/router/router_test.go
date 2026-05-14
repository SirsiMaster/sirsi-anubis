package router

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestRouter(t *testing.T) (*Router, string) {
	t.Helper()
	tmp := t.TempDir()
	root := filepath.Join(tmp, ".agents", "idea-router")
	for _, dir := range []string{"proposals", "reviews", "decisions", "transcripts"} {
		os.MkdirAll(filepath.Join(root, dir), 0o755)
	}
	// Write initial state
	os.WriteFile(filepath.Join(root, "state.json"), []byte(`{
		"version": 1,
		"active_topics": ["safety-reset"],
		"completed_topics": [],
		"last_codex_read": "2026-05-13T00:00:00Z",
		"last_claude_read": null,
		"rules": {"no_feature_expansion": true}
	}`), 0o644)

	r, err := New(tmp)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return r, tmp
}

func TestNew_Valid(t *testing.T) {
	r, _ := setupTestRouter(t)
	if r == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestNew_MissingDir(t *testing.T) {
	_, err := New("/nonexistent/path")
	if err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestReadState(t *testing.T) {
	r, _ := setupTestRouter(t)
	state, err := r.ReadState()
	if err != nil {
		t.Fatalf("ReadState() error: %v", err)
	}
	if state.Version != 1 {
		t.Errorf("Version = %d, want 1", state.Version)
	}
	if len(state.ActiveTopics) != 1 || state.ActiveTopics[0] != "safety-reset" {
		t.Errorf("ActiveTopics = %v, want [safety-reset]", state.ActiveTopics)
	}
	if !state.Rules["no_feature_expansion"] {
		t.Error("expected no_feature_expansion rule to be true")
	}
}

func TestWriteState(t *testing.T) {
	r, _ := setupTestRouter(t)
	state := &State{
		Version:      1,
		ActiveTopics: []string{"new-topic"},
		LastClaudeRead: "2026-05-14T12:00:00Z",
	}
	if err := r.WriteState(state); err != nil {
		t.Fatalf("WriteState() error: %v", err)
	}

	// Re-read and verify
	got, err := r.ReadState()
	if err != nil {
		t.Fatalf("ReadState() after write: %v", err)
	}
	if len(got.ActiveTopics) != 1 || got.ActiveTopics[0] != "new-topic" {
		t.Errorf("ActiveTopics = %v, want [new-topic]", got.ActiveTopics)
	}
}

func TestSubmit_Proposal(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, err := r.Submit(DocProposal, "claude", "Safety Reset Plan", "# Proposal: Safety Reset\n\nauthor: claude\n\n## Problem\nUnsafe deletions.")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	// Verify file exists
	path := filepath.Join(r.root, "proposals", id+".md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("proposal file not created: %v", err)
	}
}

func TestSubmit_Review(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, err := r.Submit(DocReview, "codex", "Hardening Review", "# Review: Hardening\n\nreviewer: codex\nverdict: approve")
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	path := filepath.Join(r.root, "reviews", id+".md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("review file not created: %v", err)
	}
}

func TestSubmit_UpdatesState(t *testing.T) {
	r, _ := setupTestRouter(t)
	r.Submit(DocProposal, "claude", "Test", "content")

	state, _ := r.ReadState()
	if state.LastClaudeRead == "" || state.LastClaudeRead == "null" {
		t.Error("expected LastClaudeRead to be updated after claude submit")
	}
}

func TestList(t *testing.T) {
	r, _ := setupTestRouter(t)
	r.Submit(DocProposal, "claude", "Prop One", "# Proposal: Prop One\n\nauthor: claude\n\ncontent")
	r.Submit(DocReview, "codex", "Rev One", "# Review: Rev One\n\nreviewer: codex\n\ncontent")
	r.Submit(DocDecision, "claude", "Dec One", "# Decision: Dec One\n\nauthor: claude\n\ncontent")

	docs, err := r.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("expected 3 documents, got %d", len(docs))
	}

	// Should be sorted by ModTime descending (most recent first)
	for i := 1; i < len(docs); i++ {
		if docs[i].ModTime.After(docs[i-1].ModTime) {
			t.Errorf("docs not sorted: [%d] newer than [%d]", i, i-1)
		}
	}
}

func TestGet(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, _ := r.Submit(DocProposal, "claude", "Find Me", "# Proposal: Find Me\n\nauthor: claude\n\nBody text here.")

	doc, err := r.Get(id)
	if err != nil {
		t.Fatalf("Get(%q) error: %v", id, err)
	}
	if doc.Title != "Proposal: Find Me" {
		t.Errorf("Title = %q, want 'Proposal: Find Me'", doc.Title)
	}
	if doc.Author != "claude" {
		t.Errorf("Author = %q, want 'claude'", doc.Author)
	}
}

func TestGet_NotFound(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, err := r.Get("nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent document")
	}
}

func TestPollSince(t *testing.T) {
	r, _ := setupTestRouter(t)
	before := time.Now().Add(-time.Second)
	r.Submit(DocProposal, "claude", "Recent", "# Proposal: Recent\n\nauthor: claude")

	docs, err := r.PollSince(before, 10)
	if err != nil {
		t.Fatalf("PollSince() error: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("expected 1 recent doc, got %d", len(docs))
	}
}

func TestPollSince_FutureReturnsNone(t *testing.T) {
	r, _ := setupTestRouter(t)
	r.Submit(DocProposal, "claude", "Old", "content")

	docs, err := r.PollSince(time.Now().Add(time.Hour), 10)
	if err != nil {
		t.Fatalf("PollSince() error: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("expected 0 docs for future timestamp, got %d", len(docs))
	}
}

func TestValidateAuthor(t *testing.T) {
	tests := []struct {
		author  string
		wantErr bool
	}{
		{"claude", false},
		{"codex", false},
		{"", true},
		{"mallory", true},
		{"../etc", true},
		{"claude/../../etc", true},
		{"codex\x00evil", true},
	}
	for _, tt := range tests {
		err := ValidateAuthor(tt.author)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateAuthor(%q) error = %v, wantErr = %v", tt.author, err, tt.wantErr)
		}
	}
}

func TestSubmit_RejectsInvalidAuthor(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, err := r.Submit(DocProposal, "../escape", "Evil", "content")
	if err == nil {
		t.Error("expected error for path-traversal author")
	}
}

func TestSubmit_RejectsUnknownAuthor(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, err := r.Submit(DocProposal, "mallory", "Evil", "content")
	if err == nil {
		t.Error("expected error for unknown author")
	}
}

func TestNotifyAgent_UnknownTarget(t *testing.T) {
	err := NotifyAgent("mallory", "proposal", "test-id", "/tmp")
	if err == nil {
		t.Error("expected error for unknown notification target")
	}
}

func TestNotifyAgent_DisabledByDefault(t *testing.T) {
	// NotifyAgent validates the target but doesn't check env — the MCP handler does.
	// Test that invalid targets are rejected at the router level.
	err := NotifyAgent("evil", "proposal", "test-id", "/tmp")
	if err == nil {
		t.Error("expected error for invalid target")
	}
}

func TestSubmitAddressed_AddsToInbox(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, err := r.SubmitAddressed(DocReview, "claude", "Review for Codex", "# Review\n\nauthor: claude\ncontent", "codex")
	if err != nil {
		t.Fatalf("SubmitAddressed() error: %v", err)
	}

	state, _ := r.ReadState()
	if len(state.PendingForCodex) != 1 || state.PendingForCodex[0] != id {
		t.Errorf("PendingForCodex = %v, want [%s]", state.PendingForCodex, id)
	}
}

func TestPollInbox_PeeksWithoutClearing(t *testing.T) {
	r, _ := setupTestRouter(t)
	r.SubmitAddressed(DocReview, "claude", "Review 1", "content", "codex")
	r.SubmitAddressed(DocProposal, "claude", "Proposal 1", "content", "codex")

	pending, err := r.PollInbox("codex")
	if err != nil {
		t.Fatalf("PollInbox() error: %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("expected 2 pending items, got %d", len(pending))
	}

	// Second poll should still return the same items (peek doesn't clear)
	pending2, _ := r.PollInbox("codex")
	if len(pending2) != 2 {
		t.Errorf("expected 2 pending after peek (no ack), got %d", len(pending2))
	}
}

func TestAckInbox_ClearsItems(t *testing.T) {
	r, _ := setupTestRouter(t)
	id1, _ := r.SubmitAddressed(DocReview, "claude", "Review 1", "content", "codex")
	r.SubmitAddressed(DocProposal, "claude", "Proposal 1", "content", "codex")

	// Ack only the first item
	err := r.AckInbox("codex", []string{id1})
	if err != nil {
		t.Fatalf("AckInbox() error: %v", err)
	}

	// Poll should show only the unacknowledged item
	pending, _ := r.PollInbox("codex")
	if len(pending) != 1 {
		t.Errorf("expected 1 pending after partial ack, got %d", len(pending))
	}
}

func TestSubmitAddressed_InvalidTarget(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, err := r.SubmitAddressed(DocReview, "claude", "Bad Target", "content", "mallory")
	if err == nil {
		t.Error("expected error for invalid addressed_to")
	}
}

func TestPollInbox_EmptyInbox(t *testing.T) {
	r, _ := setupTestRouter(t)
	pending, err := r.PollInbox("claude")
	if err != nil {
		t.Fatalf("PollInbox() error: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("expected empty inbox, got %d items", len(pending))
	}
}

func TestInboxFor_InvalidAgent(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, err := r.PollInbox("mallory")
	if err == nil {
		t.Error("expected error for invalid agent")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Safety Reset Plan", "safety-reset-plan"},
		{"Ma'at Credibility Fix", "maat-credibility-fix"},
		{"  Spaces  Everywhere  ", "spaces-everywhere"},
		{"UPPER_CASE", "upper-case"},
		{"", ""},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
