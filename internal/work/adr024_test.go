package work

import (
	"os"
	"path/filepath"
	"testing"
)

// Acceptance test F3 (ADR-024 §5): a reply addressed to an agent is surfaced by
// the inbox reader from items/ ONLY; a sibling reviews/ entry is never polled.
func TestADR024_F3_InboxReadsItemsOnly(t *testing.T) {
	root := t.TempDir()

	// A real reply lands in items/ via the normal sender path.
	if _, err := Send(root, "codex-pantheon", "claude-home", "review: ADR-024", "verdict: approve"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	// A stray reply parked in reviews/ (the old channel) MUST be ignored.
	reviewsDir := filepath.Join(root, "reviews")
	if err := os.MkdirAll(reviewsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stray := "---\nfrom: codex-pantheon\nto: claude-home\nstatus: open\n---\n\n## Instructions\n\nstray review in reviews/\n"
	if err := os.WriteFile(filepath.Join(reviewsDir, "stray-review.md"), []byte(stray), 0o644); err != nil {
		t.Fatal(err)
	}

	inbox, err := ListInbox(root, "claude-home")
	if err != nil {
		t.Fatalf("ListInbox: %v", err)
	}
	if len(inbox) != 1 {
		t.Fatalf("inbox = %d items, want 1 (reviews/ must not be polled)", len(inbox))
	}
	if inbox[0].From != "codex-pantheon" || inbox[0].Title != "review: ADR-024" {
		t.Errorf("unexpected inbox item: %+v", inbox[0])
	}
}

// SendTyped writes a type: frontmatter field that round-trips through parse,
// so review/decision messages live as addressed items/ entries (Decision 5).
func TestADR024_SendTyped_RoundTrips(t *testing.T) {
	root := t.TempDir()
	id, err := SendTyped(root, "claude-pantheon", "codex-pantheon", "freeze review", "review", "PASS")
	if err != nil {
		t.Fatalf("SendTyped: %v", err)
	}
	it, err := Get(root, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if it.Type != "review" {
		t.Errorf("Type = %q, want review", it.Type)
	}
	// A plain Send leaves Type empty (no type: line emitted).
	id2, _ := Send(root, "a", "b", "plain item", "do thing")
	it2, _ := Get(root, id2)
	if it2.Type != "" {
		t.Errorf("plain Send Type = %q, want empty", it2.Type)
	}
}
