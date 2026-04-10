package mobile

import (
	"encoding/json"
	"testing"
)

func TestSeshatListSources(t *testing.T) {
	result := SeshatListSources()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	var sources []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(resp.Data, &sources); err != nil {
		t.Fatalf("failed to parse sources: %v", err)
	}

	if len(sources) == 0 {
		t.Error("expected at least one knowledge source")
	}

	for _, s := range sources {
		if s.Name == "" {
			t.Error("source has empty name")
		}
		if s.Description == "" {
			t.Errorf("source %q has empty description", s.Name)
		}
	}
}

func TestSeshatListTargets(t *testing.T) {
	result := SeshatListTargets()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}
}

func TestSeshatListKnowledgeItems(t *testing.T) {
	result := SeshatListKnowledgeItems()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	// May fail if default paths don't exist — that's acceptable.
	// We're testing the bridge, not the filesystem.
}

func TestSeshatIngest_InvalidOptions(t *testing.T) {
	result := SeshatIngest("{not valid json")

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false for invalid JSON options")
	}
}
