package mobile

import (
	"encoding/json"
	"testing"
)

func TestAnubisCategories(t *testing.T) {
	result := AnubisCategories()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	var cats []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	}
	if err := json.Unmarshal(resp.Data, &cats); err != nil {
		t.Fatalf("failed to parse categories: %v", err)
	}

	if len(cats) != 7 {
		t.Errorf("expected 7 categories, got %d", len(cats))
	}

	// Verify known categories exist.
	ids := make(map[string]bool)
	for _, c := range cats {
		ids[c.ID] = true
	}
	for _, want := range []string{"general", "dev", "ai", "ides", "cloud", "storage"} {
		if !ids[want] {
			t.Errorf("missing category %q", want)
		}
	}
}

func TestAnubisScan_InvalidOptions(t *testing.T) {
	result := AnubisScan("/nonexistent", "{invalid json")

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false for invalid JSON options")
	}
	if resp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestAnubisScan_EmptyOptions(t *testing.T) {
	// Scan with empty options should succeed (scans home dir).
	result := AnubisScan("", "")

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	// Should succeed — even if no findings, the envelope is ok.
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}
}
