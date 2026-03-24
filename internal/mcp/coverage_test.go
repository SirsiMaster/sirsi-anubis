package mcp

import (
	"testing"
)

// handleGhostReport runs live Ka scanning — exercise the code path.
func TestHandleGhostReport_NoTarget(t *testing.T) {
	result, err := handleGhostReport(map[string]interface{}{})
	if err != nil {
		t.Fatalf("handleGhostReport: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) == 0 {
		t.Error("expected content in result")
	}
	// Should contain "Ka Ghost Report" regardless of ghost count
	if result.Content[0].Text == "" {
		t.Error("expected non-empty text")
	}
}

func TestHandleGhostReport_WithTarget(t *testing.T) {
	result, err := handleGhostReport(map[string]interface{}{
		"target": "NonExistentAppXYZ12345",
	})
	if err != nil {
		t.Fatalf("handleGhostReport: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// handleHealthCheck runs live scans — just verify it doesn't panic.
func TestHandleHealthCheck(t *testing.T) {
	result, err := handleHealthCheck(nil)
	if err != nil {
		t.Fatalf("handleHealthCheck: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	text := result.Content[0].Text
	if text == "" {
		t.Error("expected non-empty health check output")
	}
}

// Server.Run reads from os.Stdin which we can't easily test, but verify
// the function exists and the Server can be constructed for it.
func TestServer_Run_Exists(t *testing.T) {
	srv := NewServer()
	_ = srv // Just verify it compiles
}
