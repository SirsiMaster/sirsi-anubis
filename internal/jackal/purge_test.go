package jackal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
)

func TestScanArtifacts_FindsNodeModules(t *testing.T) {
	tmp := t.TempDir()

	// Create a project with node_modules
	projDir := filepath.Join(tmp, "myapp")
	nm := filepath.Join(projDir, "node_modules")
	os.MkdirAll(nm, 0o755)
	os.WriteFile(filepath.Join(nm, "dep.js"), make([]byte, 4096), 0o644)
	os.WriteFile(filepath.Join(projDir, "package.json"), []byte(`{}`), 0o644)

	result, err := ScanArtifacts([]string{tmp})
	if err != nil {
		t.Fatalf("ScanArtifacts() error = %v", err)
	}

	if len(result.Artifacts) == 0 {
		t.Fatal("expected at least 1 artifact, got 0")
	}

	found := false
	for _, a := range result.Artifacts {
		if a.Type == ArtifactNodeModules {
			found = true
			if a.Size == 0 {
				t.Error("node_modules size should be > 0")
			}
			if a.ProjectName != "myapp" {
				t.Errorf("ProjectName = %q, want 'myapp'", a.ProjectName)
			}
		}
	}
	if !found {
		t.Error("node_modules artifact not found")
	}
}

func TestScanArtifacts_FindsRustTarget(t *testing.T) {
	tmp := t.TempDir()

	projDir := filepath.Join(tmp, "myrust")
	target := filepath.Join(projDir, "target")
	os.MkdirAll(target, 0o755)
	os.WriteFile(filepath.Join(target, "build.o"), make([]byte, 2048), 0o644)
	// Rust target requires Cargo.toml confirmer
	os.WriteFile(filepath.Join(projDir, "Cargo.toml"), []byte("[package]"), 0o644)

	result, err := ScanArtifacts([]string{tmp})
	if err != nil {
		t.Fatalf("ScanArtifacts() error = %v", err)
	}

	found := false
	for _, a := range result.Artifacts {
		if a.Type == ArtifactTarget {
			found = true
		}
	}
	if !found {
		t.Error("Rust target artifact not found")
	}
}

func TestScanArtifacts_SkipsNonexistentRoots(t *testing.T) {
	result, err := ScanArtifacts([]string{"/nonexistent/path/xyzzy"})
	if err != nil {
		t.Fatalf("ScanArtifacts() error = %v", err)
	}
	if len(result.Artifacts) != 0 {
		t.Errorf("expected 0 artifacts for nonexistent root, got %d", len(result.Artifacts))
	}
}

func TestScanArtifacts_EmptyRoots(t *testing.T) {
	result, err := ScanArtifacts(nil)
	if err != nil {
		t.Fatalf("ScanArtifacts() error = %v", err)
	}
	if len(result.Artifacts) != 0 {
		t.Errorf("expected 0 artifacts, got %d", len(result.Artifacts))
	}
}

func TestScanArtifacts_SortedBySize(t *testing.T) {
	tmp := t.TempDir()

	// Create two projects with different-sized artifacts
	small := filepath.Join(tmp, "small", "node_modules")
	large := filepath.Join(tmp, "large", "node_modules")
	os.MkdirAll(small, 0o755)
	os.MkdirAll(large, 0o755)
	os.WriteFile(filepath.Join(tmp, "small", "package.json"), []byte(`{}`), 0o644)
	os.WriteFile(filepath.Join(tmp, "large", "package.json"), []byte(`{}`), 0o644)
	os.WriteFile(filepath.Join(small, "s.js"), make([]byte, 100), 0o644)
	os.WriteFile(filepath.Join(large, "l.js"), make([]byte, 10000), 0o644)

	result, err := ScanArtifacts([]string{tmp})
	if err != nil {
		t.Fatalf("ScanArtifacts() error = %v", err)
	}

	if len(result.Artifacts) < 2 {
		t.Fatalf("expected at least 2 artifacts, got %d", len(result.Artifacts))
	}

	// Should be sorted descending
	for i := 1; i < len(result.Artifacts); i++ {
		if result.Artifacts[i].Size > result.Artifacts[i-1].Size {
			t.Errorf("artifacts not sorted by size: [%d]=%d > [%d]=%d",
				i, result.Artifacts[i].Size, i-1, result.Artifacts[i-1].Size)
		}
	}
}

func TestScanArtifacts_Deduplication(t *testing.T) {
	tmp := t.TempDir()

	// Create nested node_modules (parent and child)
	outer := filepath.Join(tmp, "app", "node_modules")
	inner := filepath.Join(outer, "some-pkg", "node_modules")
	os.MkdirAll(inner, 0o755)
	os.WriteFile(filepath.Join(outer, "pkg.js"), make([]byte, 100), 0o644)
	os.WriteFile(filepath.Join(inner, "dep.js"), make([]byte, 100), 0o644)
	os.WriteFile(filepath.Join(tmp, "app", "package.json"), []byte(`{}`), 0o644)

	result, err := ScanArtifacts([]string{tmp})
	if err != nil {
		t.Fatalf("ScanArtifacts() error = %v", err)
	}

	// Deduplication should keep only the outermost
	nmCount := 0
	for _, a := range result.Artifacts {
		if a.Type == ArtifactNodeModules {
			nmCount++
		}
	}
	if nmCount > 1 {
		t.Errorf("expected 1 node_modules after dedup, got %d", nmCount)
	}
}

func TestDefaultPurgeRoots(t *testing.T) {
	roots := DefaultPurgeRoots()
	// Should return at least 1 root on a real system (the ones that exist)
	// All returned roots should be directories that exist
	for _, r := range roots {
		info, err := os.Stat(r)
		if err != nil {
			t.Errorf("root %q does not exist: %v", r, err)
		} else if !info.IsDir() {
			t.Errorf("root %q is not a directory", r)
		}
	}
}

func TestDeduplicateArtifacts(t *testing.T) {
	artifacts := []ProjectArtifact{
		{ArtifactDir: "/a/b/node_modules", Size: 1000},
		{ArtifactDir: "/a/b/node_modules/pkg/node_modules", Size: 200},
		{ArtifactDir: "/c/d/target", Size: 500},
	}

	result := deduplicateArtifacts(artifacts)

	if len(result) != 2 {
		t.Errorf("expected 2 after dedup, got %d", len(result))
	}
	for _, a := range result {
		if a.ArtifactDir == "/a/b/node_modules/pkg/node_modules" {
			t.Error("nested artifact should have been removed")
		}
	}
}

func TestPurgeArtifacts_NoTrashPlatformSkips(t *testing.T) {
	// Simulate a platform without trash support (Linux, Android, iOS).
	// PurgeArtifacts with useTrash=true should skip/error, NOT permanently delete.
	old := platform.Current()
	platform.Set(&platform.Mock{NoTrash: true})
	defer platform.Set(old)

	tmp := t.TempDir()
	dir := filepath.Join(tmp, "node_modules")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "pkg.js"), make([]byte, 100), 0o644)

	artifacts := []ProjectArtifact{
		{ArtifactDir: dir, Size: 100, ProjectName: "test"},
	}

	result, err := PurgeArtifacts(artifacts, true)
	if err != nil {
		t.Fatalf("PurgeArtifacts() error = %v", err)
	}

	// Should skip, not delete — platform has no trash
	if result.Cleaned != 0 {
		t.Errorf("Cleaned = %d, want 0 (no trash platform should skip)", result.Cleaned)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}

	// Directory should still exist — NOT deleted
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("artifact directory was permanently deleted on no-trash platform — SAFETY VIOLATION")
	}
}

func TestPurgeArtifacts_WithTrashPlatformDeletes(t *testing.T) {
	// Simulate a platform WITH trash support — should succeed.
	old := platform.Current()
	mock := &platform.Mock{}
	platform.Set(mock)
	defer platform.Set(old)

	tmp := t.TempDir()
	dir := filepath.Join(tmp, "node_modules")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "pkg.js"), make([]byte, 100), 0o644)

	artifacts := []ProjectArtifact{
		{ArtifactDir: dir, Size: 100, ProjectName: "test"},
	}

	result, err := PurgeArtifacts(artifacts, true)
	if err != nil {
		t.Fatalf("PurgeArtifacts() error = %v", err)
	}

	if result.Cleaned != 1 {
		t.Errorf("Cleaned = %d, want 1", result.Cleaned)
	}
	// Mock records trash calls
	if len(mock.TrashCalls) != 1 {
		t.Errorf("TrashCalls = %d, want 1", len(mock.TrashCalls))
	}
}
