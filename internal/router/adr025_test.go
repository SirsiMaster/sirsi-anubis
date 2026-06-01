package router

import (
	"testing"
	"time"
)

func mustRegister(t *testing.T, root, agent string, pid int) *Thread {
	t.Helper()
	thr, err := RegisterThread(root, &Thread{AgentID: agent, Surface: "claude", PID: pid})
	if err != nil {
		t.Fatalf("RegisterThread: %v", err)
	}
	return thr
}

// ADR-025 test 1: suspend is idempotent and snapshots the payload.
func TestSuspendThread_IdempotentWithPayload(t *testing.T) {
	root := t.TempDir()
	thr := mustRegister(t, root, "claude-pantheon", 4242)

	payload := &SuspendPayload{ThothRef: "stele-abc", OwnedOpenItems: []string{"item-1"}, ResumePrompt: "Pantheon Native App"}
	s1, err := SuspendThread(root, thr.ThreadID, payload)
	if err != nil {
		t.Fatalf("SuspendThread: %v", err)
	}
	if s1.Status != ThreadStatusSuspended {
		t.Errorf("status = %q, want suspended", s1.Status)
	}
	if s1.SuspendPayload == nil || s1.SuspendPayload.ThothRef != "stele-abc" || s1.SuspendPayload.SuspendedAt.IsZero() {
		t.Errorf("payload not snapshotted: %+v", s1.SuspendPayload)
	}

	// Idempotent: suspending again returns it unchanged, no error.
	s2, err := SuspendThread(root, thr.ThreadID, nil)
	if err != nil {
		t.Fatalf("second SuspendThread: %v", err)
	}
	if s2.Status != ThreadStatusSuspended || s2.SuspendPayload == nil || s2.SuspendPayload.ThothRef != "stele-abc" {
		t.Error("second suspend must be a no-op preserving the original payload")
	}
}

// ADR-025 test 2: resume restores payload to the caller and flips to active.
func TestResumeThread_RestoresPayloadAndActivates(t *testing.T) {
	root := t.TempDir()
	thr := mustRegister(t, root, "claude-pantheon", 4243)
	_, _ = SuspendThread(root, thr.ThreadID, &SuspendPayload{OwnedOpenItems: []string{"i1", "i2"}, ResumePrompt: "resume me"})

	resumed, err := ResumeThread(root, thr.ThreadID)
	if err != nil {
		t.Fatalf("ResumeThread: %v", err)
	}
	if resumed.Status != ThreadStatusActive {
		t.Errorf("status = %q, want active", resumed.Status)
	}
	if resumed.SuspendPayload == nil || len(resumed.SuspendPayload.OwnedOpenItems) != 2 {
		t.Error("resume must return the payload so the caller can re-surface owned items")
	}
	// Persisted record has the payload cleared.
	reg, _ := LoadThreadRegistry(root)
	if reg.Threads[thr.ThreadID].SuspendPayload != nil {
		t.Error("persisted suspend_payload must be cleared after resume")
	}
	// Resuming a non-suspended thread errors.
	if _, err := ResumeThread(root, thr.ThreadID); err == nil {
		t.Error("resuming an active thread must error")
	}
}

// ADR-025 test 3a: heartbeat rejects suspended (no revive, no last_seen refresh).
func TestHeartbeat_RejectsSuspended(t *testing.T) {
	root := t.TempDir()
	thr := mustRegister(t, root, "claude-pantheon", 4244)
	suspended, _ := SuspendThread(root, thr.ThreadID, nil)
	before := suspended.LastSeenAt

	if _, err := Heartbeat(root, thr.ThreadID, HeartbeatUpdate{}); err == nil {
		t.Fatal("heartbeat to a suspended thread must be rejected")
	}
	reg, _ := LoadThreadRegistry(root)
	got := reg.Threads[thr.ThreadID]
	if got.Status != ThreadStatusSuspended {
		t.Errorf("status after rejected heartbeat = %q, want suspended (no revive)", got.Status)
	}
	if !got.LastSeenAt.Equal(before) {
		t.Error("rejected heartbeat must NOT refresh last_seen_at")
	}
}

// ADR-025 test 3b: register bypasses the live fast-path for a suspended record.
func TestRegisterThread_BypassesSuspendedFastPath(t *testing.T) {
	root := t.TempDir()
	thr := mustRegister(t, root, "claude-pantheon", 4245)
	_, _ = SuspendThread(root, thr.ThreadID, nil)

	// Re-register same (agent, pid) with no pinned ThreadID: must NOT reuse the
	// suspended record (that would silently reactivate without restoring state).
	fresh, err := RegisterThread(root, &Thread{AgentID: "claude-pantheon", Surface: "claude", PID: 4245})
	if err != nil {
		t.Fatalf("RegisterThread: %v", err)
	}
	if fresh.ThreadID == thr.ThreadID {
		t.Error("register must mint a fresh thread, not revive the suspended one via fast-path")
	}
	// The suspended record is untouched.
	reg, _ := LoadThreadRegistry(root)
	if reg.Threads[thr.ThreadID].Status != ThreadStatusSuspended {
		t.Error("the suspended record must remain suspended")
	}
}

// ADR-025 test 7: prune never removes suspended (it is non-terminal).
func TestPruneClosed_NeverRemovesSuspended(t *testing.T) {
	root := t.TempDir()
	thr := mustRegister(t, root, "claude-pantheon", 4246)
	_, _ = SuspendThread(root, thr.ThreadID, nil)

	reg, _ := LoadThreadRegistry(root)
	// Backdate well past any maxAge.
	reg.Threads[thr.ThreadID].LastSeenAt = time.Now().Add(-720 * time.Hour)
	removed := reg.PruneClosed(time.Now(), time.Hour)
	if removed != 0 {
		t.Errorf("prune removed %d suspended thread(s), want 0", removed)
	}
	if _, ok := reg.Threads[thr.ThreadID]; !ok {
		t.Error("suspended thread must survive prune")
	}
}

// ADR-025 reaper guard: a suspended thread (dead PID by design) must NOT be
// reaped to terminal — that would destroy its recoverable state.
func TestReapDeadThreads_SkipsSuspended(t *testing.T) {
	root := t.TempDir()
	host := "test-host"
	// PID 1 is alive; use a PID that's almost certainly dead to prove the point
	// only matters via the suspended guard. Register with this host.
	thr, err := RegisterThread(root, &Thread{AgentID: "claude-pantheon", Surface: "claude", PID: 999999, Host: host})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = SuspendThread(root, thr.ThreadID, nil)

	reaped, err := ReapDeadThreads(root, host)
	if err != nil {
		t.Fatalf("ReapDeadThreads: %v", err)
	}
	for _, r := range reaped {
		if r.ThreadID == thr.ThreadID {
			t.Fatal("a suspended thread must never be reaped")
		}
	}
	reg, _ := LoadThreadRegistry(root)
	if reg.Threads[thr.ThreadID].Status != ThreadStatusSuspended {
		t.Errorf("status = %q, want suspended (reaper must skip it)", reg.Threads[thr.ThreadID].Status)
	}
}
