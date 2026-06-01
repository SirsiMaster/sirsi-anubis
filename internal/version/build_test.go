package version

import (
	"runtime"
	"testing"
)

func TestCurrent_ReportsBinaryAndRuntime(t *testing.T) {
	info := Current("sirsi")
	if info.Binary != "sirsi" {
		t.Errorf("Binary = %q, want sirsi", info.Binary)
	}
	if info.GoVer != runtime.Version() {
		t.Errorf("GoVer = %q, want %q", info.GoVer, runtime.Version())
	}
	if info.Version == "" {
		t.Error("Version must never be empty")
	}
	if info.Path == "" {
		t.Error("Path should resolve to the running test binary")
	}
}

func TestCurrent_FallsBackToVCSWhenUnstamped(t *testing.T) {
	// Default build (no ldflags) → Version stays the placeholder, and Commit
	// must be honest: either the ldflags value or a VCS-derived one, never empty.
	info := Current("sirsi")
	if info.Commit == "" {
		t.Error("Commit must never be empty (ldflags or VCS fallback)")
	}
}
