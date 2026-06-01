package router

import (
	"testing"
	"time"
)

func TestClassifyProcessRole(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want ProcessRole
	}{
		{"claude", "claude", ProcessRoleAgent},
		{"Terminal", "/System/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal", ProcessRoleTerminal},
		{"Code Helper", "/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper", ProcessRoleIDE},
		{"launchd", "/sbin/launchd", ProcessRoleSystem},
		{"Safari", "/Applications/Safari.app/Contents/MacOS/Safari", ProcessRoleProcess},
	}
	for _, tt := range tests {
		if got := ClassifyProcessRole(tt.name, tt.cmd); got != tt.want {
			t.Fatalf("ClassifyProcessRole(%q, %q) = %s, want %s", tt.name, tt.cmd, got, tt.want)
		}
	}
}

func TestReconcileProcessRegistry_PreservesFirstSeenAndMarksGone(t *testing.T) {
	t0 := time.Date(2026, 5, 31, 20, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	prev := &ProcessRegistry{Processes: map[string]*ProcessRecord{
		"host:10": {PID: 10, Host: "host", FirstSeen: t0, LastSeen: t0, Status: "visible"},
		"host:11": {PID: 11, Host: "host", FirstSeen: t0, LastSeen: t0, Status: "visible"},
	}}

	next := ReconcileProcessRegistry(prev, []ProcessRecord{
		{PID: 10, Name: "claude", Command: "claude", RSS: 42},
		{PID: 12, Name: "zsh", Command: "-zsh", RSS: 7},
	}, "host", t1)

	if next.Processes["host:10"].FirstSeen != t0 {
		t.Fatalf("expected first_seen preserved for pid 10")
	}
	if next.Processes["host:10"].LastSeen != t1 || next.Processes["host:10"].Status != "visible" {
		t.Fatalf("expected pid 10 refreshed visible")
	}
	if next.Processes["host:11"].Status != "gone" {
		t.Fatalf("expected missing pid 11 marked gone")
	}
	if next.Processes["host:12"].FirstSeen != t1 || next.Processes["host:12"].Role != ProcessRoleTerminal {
		t.Fatalf("expected new pid 12 registered as terminal")
	}
}
