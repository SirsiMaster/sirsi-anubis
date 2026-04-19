package mobile

import (
	"encoding/json"
	"testing"
)

func TestVersion(t *testing.T) {
	v := Version()
	if v == "" {
		t.Fatal("Version() returned empty string")
	}
	if v != "0.16.0-ios" {
		t.Errorf("Version() = %q, want %q", v, "0.16.0-ios")
	}
}

func TestSuccessJSON(t *testing.T) {
	result := successJSON(map[string]string{"hello": "world"})

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.Error != "" {
		t.Errorf("expected no error, got %q", resp.Error)
	}
}

func TestErrorJSON(t *testing.T) {
	result := errorJSON("something broke")

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false")
	}
	if resp.Error != "something broke" {
		t.Errorf("error = %q, want %q", resp.Error, "something broke")
	}
}
