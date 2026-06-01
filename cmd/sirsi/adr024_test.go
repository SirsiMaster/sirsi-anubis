package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// adr024RegisterOutput mirrors the JSON shape `thread register --json` emits
// after ADR-024: the thread fields plus the prescribed watcher block.
type adr024RegisterOutput struct {
	ThreadID string `json:"thread_id"`
	Surface  string `json:"surface"`
	Watcher  struct {
		Type               string `json:"type"`
		Mechanism          string `json:"mechanism"`
		ArmInstruction     string `json:"arm_instruction"`
		HeartbeatIntervalS int    `json:"heartbeat_interval_s"`
		WatchesInbox       bool   `json:"watches_inbox"`
		Resident           bool   `json:"resident"`
	} `json:"watcher"`
}

func setupTempRouter(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "idea-router"), 0o755); err != nil {
		t.Fatal(err)
	}
	return tmp
}

func registerThread(t *testing.T, dir, agent, surface string) adr024RegisterOutput {
	t.Helper()
	stdout, stderr, err := runSirsiInDir(t, dir, 10*time.Second,
		"thread", "register",
		"--agent", agent, "--surface", surface, "--repo", dir,
		"--anchor-pid", strconv.Itoa(os.Getpid()), "--json")
	if err != nil {
		t.Fatalf("register failed: %v\nstdout:%s\nstderr:%s", err, stdout, stderr)
	}
	var out adr024RegisterOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("decode register --json: %v\nstdout:%s", err, stdout)
	}
	return out
}

// Acceptance test 1: register --surface claude --json returns a watcher block
// and does NOT spawn an fs-watcher (no /tmp/sirsi-router-watch-*.pid).
func TestADR024_RegisterClaude_NoSpawn_ReturnsSpec(t *testing.T) {
	t.Parallel()
	tmp := setupTempRouter(t)
	out := registerThread(t, tmp, "test-adr024-claude", "claude")

	if out.Watcher.Type != "loop-monitor" {
		t.Errorf("watcher.type = %q, want loop-monitor", out.Watcher.Type)
	}
	if !out.Watcher.WatchesInbox {
		t.Error("claude watcher must watch the inbox")
	}
	if !strings.Contains(out.Watcher.ArmInstruction, "pgrep -f "+out.ThreadID) {
		t.Errorf("arm_instruction must key on thread_id %q; got: %s", out.ThreadID, out.Watcher.ArmInstruction)
	}
	// The defining behavior: register no longer spawns the fs-watcher.
	pidfile := "/tmp/sirsi-router-watch-" + out.ThreadID + ".pid"
	if _, err := os.Stat(pidfile); err == nil {
		os.Remove(pidfile)
		t.Errorf("register spawned an fs-watcher pidfile %s — ADR-024 forbids this", pidfile)
	}
}

// Acceptance test 2: register --surface menubar --json returns a native-runloop
// heartbeat spec, interval >=60s, and no inbox-worker requirement.
func TestADR024_RegisterMenubar_ResidentSpec(t *testing.T) {
	t.Parallel()
	tmp := setupTempRouter(t)
	out := registerThread(t, tmp, "test-adr024-menubar", "menubar")

	if out.Watcher.Type != "native-runloop" {
		t.Errorf("watcher.type = %q, want native-runloop", out.Watcher.Type)
	}
	if out.Watcher.WatchesInbox {
		t.Error("resident menubar must NOT be an inbox worker")
	}
	if !out.Watcher.Resident {
		t.Error("menubar watcher must be marked resident")
	}
	if out.Watcher.HeartbeatIntervalS < 60 {
		t.Errorf("menubar heartbeat = %d, want >=60", out.Watcher.HeartbeatIntervalS)
	}
}

// Acceptance test 3: repeated register on the same (agent_id, pid) returns the
// same thread and the same watcher spec (idempotent).
func TestADR024_RegisterIdempotent_SameThreadAndSpec(t *testing.T) {
	t.Parallel()
	tmp := setupTempRouter(t)
	first := registerThread(t, tmp, "test-adr024-idem", "claude")
	second := registerThread(t, tmp, "test-adr024-idem", "claude")

	if first.ThreadID != second.ThreadID {
		t.Errorf("idempotent register must reuse thread: %q != %q", first.ThreadID, second.ThreadID)
	}
	if first.Watcher != second.Watcher {
		t.Errorf("idempotent register must return identical watcher spec")
	}
}
