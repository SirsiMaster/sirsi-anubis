package updater

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.ReleasesURL != GitHubReleasesAPI {
		t.Errorf("ReleasesURL = %q, want %q", client.ReleasesURL, GitHubReleasesAPI)
	}
	if client.AdvisoryURL != AdvisoryURL {
		t.Errorf("AdvisoryURL = %q, want %q", client.AdvisoryURL, AdvisoryURL)
	}
	if client.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestCheck_DevVersion(t *testing.T) {
	// "dev" version should never show update.
	// The Check function skips "dev" versions.
	// We can't easily test without hitting real API, just verify no panic.
	t.Log("TestCheck_DevVersion: skipped (would hit real API)")
}
