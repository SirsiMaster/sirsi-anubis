package router

import (
	"strings"
	"testing"
)

func TestWatcherFor_Claude(t *testing.T) {
	s := WatcherFor("claude", "claude-pantheon", "thr-abc123")
	if s.Type != "loop-monitor" {
		t.Errorf("claude Type = %q, want loop-monitor", s.Type)
	}
	if !s.WatchesInbox {
		t.Error("claude must watch the inbox")
	}
	if s.Resident {
		t.Error("claude is not a resident surface")
	}
	if s.HeartbeatIntervalS != 60 {
		t.Errorf("claude heartbeat = %d, want 60", s.HeartbeatIntervalS)
	}
	// claude-deck's correction: the idempotency signature MUST key on the
	// thread_id, never the shared DIR= loop body, never TaskList.
	if !strings.Contains(s.ArmInstruction, "pgrep -f thr-abc123") {
		t.Errorf("arm_instruction must key on thread_id signature; got: %s", s.ArmInstruction)
	}
	if !strings.Contains(s.ArmInstruction, "NOT TaskList") {
		t.Error("arm_instruction must forbid keying on TaskList")
	}
}

func TestWatcherFor_MenubarResident(t *testing.T) {
	s := WatcherFor("menubar", "sirsi-menubar", "thr-x")
	if s.Type != "native-runloop" {
		t.Errorf("menubar Type = %q, want native-runloop", s.Type)
	}
	if s.WatchesInbox {
		t.Error("resident menubar must NOT be an inbox worker (preserves ADR-020)")
	}
	if !s.Resident {
		t.Error("menubar must be marked resident")
	}
	if s.HeartbeatIntervalS < 60 {
		t.Errorf("menubar heartbeat = %d, want >=60", s.HeartbeatIntervalS)
	}
}

func TestWatcherFor_Idempotent(t *testing.T) {
	a := WatcherFor("claude", "claude-pantheon", "thr-1")
	b := WatcherFor("claude", "claude-pantheon", "thr-1")
	if a != b {
		t.Error("same (surface, agent, thread) must yield an identical spec")
	}
}

func TestWatcherFor_UnknownFallsBackToDaemon(t *testing.T) {
	s := WatcherFor("nonsense", "x", "thr-1")
	if s.Type != "daemon" {
		t.Errorf("unknown surface Type = %q, want daemon fallback", s.Type)
	}
	if !s.WatchesInbox {
		t.Error("daemon fallback must watch the inbox")
	}
}

func TestWatcherFor_ResidentSurfacesNotInboxWorkers(t *testing.T) {
	for _, sfc := range []string{"menubar", "tui", "vscode", "jetbrains", "cursor", "macapp"} {
		s := WatcherFor(sfc, "x", "thr-1")
		if !s.Resident || s.WatchesInbox {
			t.Errorf("surface %q: want resident && !watches_inbox, got resident=%v watches=%v", sfc, s.Resident, s.WatchesInbox)
		}
	}
}
