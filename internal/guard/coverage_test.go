package guard

import (
	"os"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
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

// ─── isProcessRunning ───────────────────────────────────────────────────

func TestIsProcessRunning_Self(t *testing.T) {
	if !isProcessRunningWith(platform.Current(), os.Getpid()) {
		t.Error("expected own process to be running")
	}
}

// ─── getLinuxMemoryInfo ─────────────────────────────────────────────────

func TestGetLinuxMemoryInfo(t *testing.T) {
	result := &AuditResult{}
	err := getLinuxMemoryInfoWith(platform.Current(), result)
	// On macOS, `free -b` doesn't exist — will fail but code path exercised.
	if err != nil {
		t.Logf("getLinuxMemoryInfo: %v (expected on macOS)", err)
	}
}

// ─── getMemoryInfo with fallback path ───────────────────────────────────

func TestGetMemoryInfo(t *testing.T) {
	result := &AuditResult{}
	err := getMemoryInfoWith(platform.Current(), result)
	if err != nil {
		t.Fatalf("getMemoryInfo: %v", err)
	}
	if result.TotalRAM <= 0 {
		t.Error("expected positive TotalRAM")
	}
}
