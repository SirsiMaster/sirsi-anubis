package mobile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestThothDetectProject(t *testing.T) {
	// Create a temp directory with a go.mod to simulate a project.
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := ThothDetectProject(tmp)

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}
}

func TestThothInit(t *testing.T) {
	tmp := t.TempDir()

	// Write a minimal project marker.
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := ThothInit(tmp)

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	// Verify .thoth/ was created.
	if _, err := os.Stat(filepath.Join(tmp, ".thoth")); os.IsNotExist(err) {
		t.Error("expected .thoth/ directory to be created")
	}
}

func TestThothSync_InvalidOptions(t *testing.T) {
	result := ThothSync("{invalid")

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false for invalid JSON")
	}
}

func TestThothCompact_InvalidOptions(t *testing.T) {
	result := ThothCompact("{invalid")

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false for invalid JSON")
	}
}
