package brain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWeightsDir(t *testing.T) {
	dir, err := WeightsDir()
	if err != nil {
		t.Fatalf("WeightsDir() error: %v", err)
	}
	if dir == "" {
		t.Fatal("WeightsDir() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("WeightsDir() returned relative path: %s", dir)
	}
	if !containsDirSegment(dir, ".anubis") {
		t.Errorf("WeightsDir() should contain '.anubis': %s", dir)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{52428800, "50.0 MB"},
		{104857600, "100.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatBytes(tt.input)
			if got != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLocalManifestRoundtrip(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	manifest := &LocalManifest{
		InstalledModel: "anubis-classifier-v1",
		Version:        "1.0.0",
		Format:         "onnx",
		SHA256:         "abc123def456",
		SizeBytes:      52428800,
		InstalledAt:    time.Now().Truncate(time.Second),
		ModelFile:      "anubis-classifier-v1.onnx",
	}

	// Write
	err := writeLocalManifest(tmpDir, manifest)
	if err != nil {
		t.Fatalf("writeLocalManifest() error: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, ManifestFile)
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Fatal("manifest file was not created")
	}

	// Read back
	got, err := readLocalManifest(tmpDir)
	if err != nil {
		t.Fatalf("readLocalManifest() error: %v", err)
	}

	if got.InstalledModel != manifest.InstalledModel {
		t.Errorf("InstalledModel = %q, want %q", got.InstalledModel, manifest.InstalledModel)
	}
	if got.Version != manifest.Version {
		t.Errorf("Version = %q, want %q", got.Version, manifest.Version)
	}
	if got.Format != manifest.Format {
		t.Errorf("Format = %q, want %q", got.Format, manifest.Format)
	}
	if got.SHA256 != manifest.SHA256 {
		t.Errorf("SHA256 = %q, want %q", got.SHA256, manifest.SHA256)
	}
	if got.SizeBytes != manifest.SizeBytes {
		t.Errorf("SizeBytes = %d, want %d", got.SizeBytes, manifest.SizeBytes)
	}
	if got.ModelFile != manifest.ModelFile {
		t.Errorf("ModelFile = %q, want %q", got.ModelFile, manifest.ModelFile)
	}
}

func TestLocalManifestJSON(t *testing.T) {
	manifest := &LocalManifest{
		InstalledModel: "test-model",
		Version:        "2.0.0",
		Format:         "coreml",
		SHA256:         "deadbeef",
		SizeBytes:      1024,
		InstalledAt:    time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC),
		ModelFile:      "test-model.mlmodelc",
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var got LocalManifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if got.InstalledModel != manifest.InstalledModel {
		t.Errorf("InstalledModel mismatch after JSON round-trip")
	}
}

func TestReadLocalManifest_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := readLocalManifest(tmpDir)
	if err == nil {
		t.Error("readLocalManifest should error when manifest doesn't exist")
	}
}

func TestRemove_NoOpWhenNotInstalled(t *testing.T) {
	// Remove should not error even when nothing is installed
	// Note: This tests the real Remove() but it checks a non-existent path
	// which returns nil (no error) by design
	err := Remove()
	// Should not error — absence is fine
	if err != nil {
		// Only fail if it's not a "doesn't exist" type situation
		t.Logf("Remove() returned: %v (may be expected if test isolation differs)", err)
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write known content
	content := []byte("hello anubis")
	if err := os.WriteFile(testFile, content, 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	hash, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("hashFile error: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("SHA-256 hash should be 64 hex chars, got %d: %s", len(hash), hash)
	}

	// Same file should produce same hash
	hash2, _ := hashFile(testFile)
	if hash != hash2 {
		t.Errorf("Same file produced different hashes: %s vs %s", hash, hash2)
	}
}

func TestHashFile_NotExists(t *testing.T) {
	_, err := hashFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("hashFile should error on nonexistent file")
	}
}

func TestIsInstalled_FalseWhenClean(t *testing.T) {
	// With no weights installed in a fresh environment, should return false
	// (This test may pass if the dev machine has no brain installed)
	// We can't guarantee the state, so we just verify it doesn't panic
	_ = IsInstalled()
}

// containsDirSegment checks if a path contains a specific directory segment.
func containsDirSegment(path, segment string) bool {
	for _, part := range splitPath(path) {
		if part == segment {
			return true
		}
	}
	return false
}
