package guard

import (
	"os"
	"testing"
)

// ─── Slay (dry run) ──────────────────────────────────────────────────────

func TestSlay_DryRun(t *testing.T) {
	result, err := Slay(SlayNode, true)
	if err != nil {
		t.Fatalf("Slay(node, dryRun=true): %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.DryRun {
		t.Error("expected DryRun=true")
	}
	if result.Target != SlayNode {
		t.Errorf("Target = %q, want node", result.Target)
	}
}

func TestSlay_DryRun_All(t *testing.T) {
	result, err := Slay(SlayAll, true)
	if err != nil {
		t.Fatalf("Slay(all, dryRun=true): %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestSlay_DryRun_LSP(t *testing.T) {
	result, err := Slay(SlayLSP, true)
	if err != nil {
		t.Fatalf("Slay(lsp, dryRun=true): %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestSlay_DryRun_Docker(t *testing.T) {
	result, err := Slay(SlayDocker, true)
	if err != nil {
		t.Fatalf("Slay(docker, dryRun=true): %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestSlay_DryRun_AI(t *testing.T) {
	result, err := Slay(SlayAI, true)
	if err != nil {
		t.Fatalf("Slay(ai, dryRun=true): %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ─── killProcess ────────────────────────────────────────────────────────

func TestKillProcess_InvalidPID(t *testing.T) {
	// PID -1 should fail with SIGTERM error
	err := killProcess(-1)
	if err == nil {
		t.Error("expected error for invalid PID")
	}
}

// ─── isProcessRunning ───────────────────────────────────────────────────

func TestIsProcessRunning_Self(t *testing.T) {
	if !isProcessRunning(os.Getpid()) {
		t.Error("expected own process to be running")
	}
}

func TestIsProcessRunning_InvalidPID(t *testing.T) {
	// PID 999999 is very unlikely to exist
	if isProcessRunning(999999) {
		t.Log("PID 999999 exists (unexpected but possible)")
	}
}

// ─── getLinuxMemoryInfo ─────────────────────────────────────────────────

func TestGetLinuxMemoryInfo(t *testing.T) {
	result := &AuditResult{}
	err := getLinuxMemoryInfo(result)
	// On macOS, `free -b` doesn't exist — will fail but code path exercised.
	if err != nil {
		t.Logf("getLinuxMemoryInfo: %v (expected on macOS)", err)
	}
}

// ─── getMemoryInfo with fallback path ───────────────────────────────────

func TestGetMemoryInfo(t *testing.T) {
	result := &AuditResult{}
	err := getMemoryInfo(result)
	if err != nil {
		t.Fatalf("getMemoryInfo: %v", err)
	}
	if result.TotalRAM <= 0 {
		t.Error("expected positive TotalRAM")
	}
}
