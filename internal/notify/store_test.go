package notify

import (
	"path/filepath"
	"testing"
	"time"
)

func init() {
	// Disable actual toasts during tests.
	setToastExecFn(func(_ string) error { return nil })
	// Disable rate limiting for tests.
	minToastGap = 0
}

func TestStore_OpenClose(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()
}

func TestStore_RecordAndRecent(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	err = s.Record(Notification{
		Source:   "anubis",
		Action:   "scan",
		Severity: SeveritySuccess,
		Summary:  "Found 12 items (3.2 GB)",
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	results, err := s.Recent(10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Source != "anubis" {
		t.Errorf("source = %q, want anubis", results[0].Source)
	}
	if results[0].Summary != "Found 12 items (3.2 GB)" {
		t.Errorf("summary = %q, want original", results[0].Summary)
	}
}

func TestStore_BySource(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Record(Notification{Source: "anubis", Action: "scan", Severity: SeveritySuccess, Summary: "scan done"})
	s.Record(Notification{Source: "ka", Action: "ghost_hunt", Severity: SeveritySuccess, Summary: "3 ghosts"})
	s.Record(Notification{Source: "anubis", Action: "judge", Severity: SeverityInfo, Summary: "judged"})

	results, err := s.BySource("anubis", 10)
	if err != nil {
		t.Fatalf("BySource: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 anubis results, got %d", len(results))
	}
}

func TestStore_BySeverity(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Record(Notification{Source: "a", Action: "x", Severity: SeveritySuccess, Summary: "ok"})
	s.Record(Notification{Source: "b", Action: "y", Severity: SeverityError, Summary: "fail"})
	s.Record(Notification{Source: "c", Action: "z", Severity: SeveritySuccess, Summary: "ok2"})

	results, err := s.BySeverity(SeverityError, 10)
	if err != nil {
		t.Fatalf("BySeverity: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 error result, got %d", len(results))
	}
}

func TestStore_Since(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	past := time.Now().Add(-1 * time.Hour)
	s.Record(Notification{Source: "a", Action: "x", Severity: SeverityInfo, Summary: "old", Timestamp: past})
	s.Record(Notification{Source: "b", Action: "y", Severity: SeverityInfo, Summary: "new"})

	results, err := s.Since(past.Add(30*time.Minute), 10)
	if err != nil {
		t.Fatalf("Since: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result since 30min ago, got %d", len(results))
	}
}

func TestStore_Prune(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Record(Notification{Source: "a", Action: "x", Severity: SeverityInfo, Summary: "fresh"})

	removed, err := s.Prune(1 * time.Hour)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if removed != 0 {
		t.Errorf("removed = %d, want 0 (entry is fresh)", removed)
	}
}

func TestStore_Clear(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Record(Notification{Source: "a", Action: "x", Severity: SeverityInfo, Summary: "one"})
	s.Record(Notification{Source: "b", Action: "y", Severity: SeverityInfo, Summary: "two"})

	removed, err := s.Clear()
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}

	count, _ := s.Count()
	if count != 0 {
		t.Errorf("count after clear = %d, want 0", count)
	}
}

func TestStore_Count(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	count, _ := s.Count()
	if count != 0 {
		t.Errorf("empty count = %d", count)
	}

	s.Record(Notification{Source: "a", Action: "x", Severity: SeverityInfo, Summary: "one"})
	s.Record(Notification{Source: "b", Action: "y", Severity: SeverityInfo, Summary: "two"})

	count, _ = s.Count()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestStore_RecordWithDetails(t *testing.T) {
	t.Parallel()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	s.Record(Notification{
		Source:     "anubis",
		Action:     "scan",
		Severity:   SeveritySuccess,
		Summary:    "12 findings",
		Details:    "Full output here\nLine 2\nLine 3",
		DurationMs: 1500,
	})

	results, _ := s.Recent(1)
	if len(results) != 1 {
		t.Fatal("expected 1 result")
	}
	if results[0].Details != "Full output here\nLine 2\nLine 3" {
		t.Errorf("details not preserved: %q", results[0].Details)
	}
	if results[0].DurationMs != 1500 {
		t.Errorf("duration_ms = %d, want 1500", results[0].DurationMs)
	}
}
