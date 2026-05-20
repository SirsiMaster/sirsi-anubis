package router

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadThreadRegistry_Empty(t *testing.T) {
	tmp := t.TempDir()
	reg, err := LoadThreadRegistry(tmp)
	if err != nil {
		t.Fatalf("LoadThreadRegistry: %v", err)
	}
	if len(reg.Threads) != 0 {
		t.Errorf("expected empty registry, got %d", len(reg.Threads))
	}
}

func TestRegisterAndHeartbeatThread(t *testing.T) {
	tmp := t.TempDir()
	thr, err := RegisterThread(tmp, &Thread{
		AgentID: "claude-pantheon",
		Surface: "claude",
		Repo:    "/repo",
		Watches: []string{"claude-pantheon"},
	})
	if err != nil {
		t.Fatalf("RegisterThread: %v", err)
	}
	if thr.ThreadID == "" {
		t.Fatal("expected generated thread_id")
	}
	if thr.Status != ThreadStatusActive {
		t.Errorf("expected active status, got %q", thr.Status)
	}
	if thr.StartedAt.IsZero() || thr.LastSeenAt.IsZero() {
		t.Errorf("timestamps not set")
	}

	// Heartbeat should advance last_seen_at
	time.Sleep(10 * time.Millisecond)
	item := "20260520-test"
	updated, err := Heartbeat(tmp, thr.ThreadID, HeartbeatUpdate{
		Status:      ThreadStatusIdle,
		CurrentItem: &item,
	})
	if err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	if !updated.LastSeenAt.After(thr.LastSeenAt) {
		t.Errorf("last_seen_at did not advance")
	}
	if updated.Status != ThreadStatusIdle {
		t.Errorf("status not updated")
	}
	if updated.CurrentItem != "20260520-test" {
		t.Errorf("current_item not updated")
	}

	// Persisted on disk
	if _, err := os.Stat(filepath.Join(tmp, "threads.json")); err != nil {
		t.Errorf("threads.json missing: %v", err)
	}
}

func TestRegisterThread_RequiresAgentAndSurface(t *testing.T) {
	tmp := t.TempDir()
	if _, err := RegisterThread(tmp, &Thread{Surface: "claude"}); err == nil {
		t.Error("expected error for missing agent_id")
	}
	if _, err := RegisterThread(tmp, &Thread{AgentID: "claude-pantheon"}); err == nil {
		t.Error("expected error for missing surface")
	}
}

func TestHeartbeat_UnknownThread(t *testing.T) {
	tmp := t.TempDir()
	if _, err := Heartbeat(tmp, "thr-missing", HeartbeatUpdate{}); err == nil {
		t.Error("expected error for unknown thread")
	}
}

func TestCloseThread(t *testing.T) {
	tmp := t.TempDir()
	thr, _ := RegisterThread(tmp, &Thread{AgentID: "claude-pantheon", Surface: "claude"})
	closed, err := CloseThread(tmp, thr.ThreadID)
	if err != nil {
		t.Fatalf("CloseThread: %v", err)
	}
	if closed.Status != ThreadStatusClosed {
		t.Errorf("expected closed status, got %q", closed.Status)
	}
	// Closed threads are not stale.
	if closed.IsStale(time.Now().Add(24*time.Hour), time.Minute) {
		t.Errorf("closed thread should not be considered stale")
	}
}

func TestIsStale(t *testing.T) {
	thr := &Thread{
		Status:     ThreadStatusActive,
		LastSeenAt: time.Now().Add(-10 * time.Minute),
	}
	if !thr.IsStale(time.Now(), 5*time.Minute) {
		t.Error("expected stale")
	}
	thr.LastSeenAt = time.Now()
	if thr.IsStale(time.Now(), 5*time.Minute) {
		t.Error("expected fresh")
	}
}

func TestSortedThreads_NewestFirst(t *testing.T) {
	tmp := t.TempDir()
	a, _ := RegisterThread(tmp, &Thread{AgentID: "claude-pantheon", Surface: "claude"})
	time.Sleep(10 * time.Millisecond)
	b, _ := RegisterThread(tmp, &Thread{AgentID: "codex-pantheon", Surface: "codex"})

	reg, _ := LoadThreadRegistry(tmp)
	sorted := reg.SortedThreads()
	if len(sorted) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(sorted))
	}
	if sorted[0].ThreadID != b.ThreadID {
		t.Errorf("expected newest first, got %s before %s", sorted[0].ThreadID, b.ThreadID)
	}
	_ = a
}

func TestPruneClosed(t *testing.T) {
	tmp := t.TempDir()
	thr, _ := RegisterThread(tmp, &Thread{AgentID: "claude-pantheon", Surface: "claude"})
	_, _ = CloseThread(tmp, thr.ThreadID)
	reg, _ := LoadThreadRegistry(tmp)
	// Set last_seen far in the past so it's prunable.
	reg.Threads[thr.ThreadID].LastSeenAt = time.Now().Add(-48 * time.Hour)
	removed := reg.PruneClosed(time.Now(), 24*time.Hour)
	if removed != 1 {
		t.Errorf("expected 1 pruned, got %d", removed)
	}
	if len(reg.Threads) != 0 {
		t.Errorf("expected empty after prune")
	}
}

func TestNewThreadID_Unique(t *testing.T) {
	a := NewThreadID()
	b := NewThreadID()
	if a == b {
		t.Errorf("expected unique IDs, got %s == %s", a, b)
	}
}
