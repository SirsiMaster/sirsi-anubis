package router

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestRunnerDryRunDoesNotAck(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, err := r.SubmitAddressed(DocReview, "claude", "Needs Codex", "# Review: Needs Codex\n\nreviewer: claude", "codex")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rr := NewRunner(r, RunnerOptions{Agent: "all", DryRun: true, Once: true, Out: &buf})
	if err := rr.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), id) {
		t.Fatalf("dry-run output missing id: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Error("dry-run output should contain [dry-run] prefix")
	}

	// Inbox should NOT be cleared
	pending, _ := r.PollInbox("codex")
	if len(pending) != 1 || pending[0] != id {
		t.Fatalf("runner auto-acked inbox: %v", pending)
	}
}

func TestRunnerNotifyCalledOncePerSession(t *testing.T) {
	r, _ := setupTestRouter(t)
	id, err := r.SubmitAddressed(DocReview, "claude", "Needs Codex", "# Review: Needs Codex\n\nreviewer: claude", "codex")
	if err != nil {
		t.Fatal(err)
	}

	calls := 0
	notify := func(target, docType, docID, repoRoot string) error {
		calls++
		if target != "codex" || docID != id {
			t.Fatalf("bad notify: %s %s", target, docID)
		}
		return nil
	}

	rr := NewRunner(r, RunnerOptions{Agent: "all", Out: io.Discard, Notify: notify})
	if err := rr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Second tick should NOT re-dispatch (repeat suppression)
	if err := rr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("notify calls = %d, want 1 (repeat suppression)", calls)
	}
}

func TestRunnerNoPendingItems(t *testing.T) {
	r, _ := setupTestRouter(t)

	var buf bytes.Buffer
	rr := NewRunner(r, RunnerOptions{Agent: "all", DryRun: true, Once: true, Out: &buf})
	if err := rr.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "No pending dispatches") {
		t.Errorf("expected no-pending message, got: %s", buf.String())
	}
}

func TestRunnerFilterByAgent(t *testing.T) {
	r, _ := setupTestRouter(t)
	r.SubmitAddressed(DocReview, "claude", "For Codex", "content", "codex")
	r.SubmitAddressed(DocReview, "codex", "For Claude", "content", "claude")

	// Filter to codex only
	var buf bytes.Buffer
	rr := NewRunner(r, RunnerOptions{Agent: "codex", DryRun: true, Once: true, Out: &buf})
	if err := rr.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "codex") {
		t.Error("expected codex dispatch in output")
	}
	if strings.Contains(output, "claude") && !strings.Contains(output, "codex") {
		t.Error("should not dispatch to claude when agent=codex")
	}
}

func TestRunnerMissingDocSkipped(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Manually inject a pending ID that doesn't exist as a file
	state, _ := r.ReadState()
	state.AddToInbox("codex", "nonexistent-doc-id")
	r.WriteState(state)

	var buf bytes.Buffer
	rr := NewRunner(r, RunnerOptions{Agent: "all", DryRun: true, Once: true, Out: &buf})
	if err := rr.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "Skipping") {
		t.Errorf("expected skip message for missing doc, got: %s", buf.String())
	}
}

func TestRunnerNotifyFailureDoesNotMarkDispatched(t *testing.T) {
	r, _ := setupTestRouter(t)
	r.SubmitAddressed(DocReview, "claude", "Fails", "content", "codex")

	calls := 0
	failNotify := func(target, docType, docID, repoRoot string) error {
		calls++
		return io.ErrUnexpectedEOF // simulate failure
	}

	var buf bytes.Buffer
	rr := NewRunner(r, RunnerOptions{Agent: "all", Out: &buf, Notify: failNotify})

	// First tick: notify fails, should not mark dispatched
	rr.Tick(context.Background())
	// Second tick: should retry
	rr.Tick(context.Background())

	if calls != 2 {
		t.Fatalf("notify calls = %d, want 2 (retry after failure)", calls)
	}
}
