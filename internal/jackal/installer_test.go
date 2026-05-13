package jackal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsInstallerFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"dmg file", "Photoshop_2024.dmg", true},
		{"pkg file", "Illustrator_Setup.pkg", true},
		{"iso file", "ubuntu-22.04.iso", true},
		{"app.zip file", "MyApp.app.zip", true},
		{"zip file", "archive.zip", true},
		{"tar.gz file", "package.tar.gz", true},
		{"txt file", "readme.txt", false},
		{"go file", "main.go", false},
		{"empty name", "", false},
		{"uppercase DMG", "Setup.DMG", true},
		{"mixed case Pkg", "Install.Pkg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInstallerFile(tt.filename)
			if got != tt.want {
				t.Errorf("isInstallerFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestScanInstallersWithHome(t *testing.T) {
	// Create a temp directory structure simulating ~/Downloads
	tmpHome := t.TempDir()
	downloads := filepath.Join(tmpHome, "Downloads")
	desktop := filepath.Join(tmpHome, "Desktop")

	if err := os.MkdirAll(downloads, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(desktop, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a large .dmg file (> 10 MB threshold)
	largeDMG := filepath.Join(downloads, "BigApp.dmg")
	if err := createLargeFile(largeDMG, 11*1024*1024); err != nil {
		t.Fatal(err)
	}

	// Create a small .dmg file (< 10 MB, should be excluded)
	smallDMG := filepath.Join(downloads, "SmallApp.dmg")
	if err := createLargeFile(smallDMG, 5*1024*1024); err != nil {
		t.Fatal(err)
	}

	// Create a large .pkg on Desktop
	largePKG := filepath.Join(desktop, "Setup.pkg")
	if err := createLargeFile(largePKG, 15*1024*1024); err != nil {
		t.Fatal(err)
	}

	// Create a non-installer file (should be excluded)
	textFile := filepath.Join(downloads, "notes.txt")
	if err := os.WriteFile(textFile, make([]byte, 20*1024*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ScanInstallersWithHome(tmpHome)
	if err != nil {
		t.Fatalf("ScanInstallersWithHome() error = %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
		for _, f := range result.Files {
			t.Logf("  %s (%s) %d bytes", f.Name, f.Source, f.Size)
		}
	}

	// Verify sorted by size descending
	if len(result.Files) >= 2 {
		if result.Files[0].Size < result.Files[1].Size {
			t.Error("files not sorted by size descending")
		}
	}

	// Verify source labels
	for _, f := range result.Files {
		switch f.Name {
		case "BigApp.dmg":
			if f.Source != "Downloads" {
				t.Errorf("BigApp.dmg source = %q, want Downloads", f.Source)
			}
		case "Setup.pkg":
			if f.Source != "Desktop" {
				t.Errorf("Setup.pkg source = %q, want Desktop", f.Source)
			}
		}
	}

	// Verify total size
	expectedTotal := int64(11*1024*1024 + 15*1024*1024)
	if result.TotalSize != expectedTotal {
		t.Errorf("TotalSize = %d, want %d", result.TotalSize, expectedTotal)
	}

	// Verify scan time is recorded
	if result.ScanTime <= 0 {
		t.Error("ScanTime should be > 0")
	}
}

func TestScanInstallersEmptyHome(t *testing.T) {
	tmpHome := t.TempDir()
	// No Downloads/Desktop directories exist

	result, err := ScanInstallersWithHome(tmpHome)
	if err != nil {
		t.Fatalf("ScanInstallersWithHome() error = %v", err)
	}

	if len(result.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.Files))
	}
	if result.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0", result.TotalSize)
	}
}

func TestRemoveInstallers(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "test1.dmg")
	file2 := filepath.Join(tmpDir, "test2.pkg")
	if err := os.WriteFile(file1, make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, make([]byte, 200), 0o644); err != nil {
		t.Fatal(err)
	}

	files := []InstallerFile{
		{Name: "test1.dmg", Path: file1, Size: 100},
		{Name: "test2.pkg", Path: file2, Size: 200},
	}

	// Direct delete (not trash)
	result, err := RemoveInstallers(files, false)
	if err != nil {
		t.Fatalf("RemoveInstallers() error = %v", err)
	}

	if result.Cleaned != 2 {
		t.Errorf("Cleaned = %d, want 2", result.Cleaned)
	}
	if result.BytesFreed != 300 {
		t.Errorf("BytesFreed = %d, want 300", result.BytesFreed)
	}
	if result.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", result.Skipped)
	}

	// Verify files are gone
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should be deleted")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("file2 should be deleted")
	}
}

func TestRemoveInstallersNonexistent(t *testing.T) {
	files := []InstallerFile{
		{Name: "ghost.dmg", Path: "/tmp/nonexistent_installer_test.dmg", Size: 1000},
	}

	result, err := RemoveInstallers(files, false)
	if err != nil {
		t.Fatalf("RemoveInstallers() error = %v", err)
	}

	// cleaner.DeleteFile treats nonexistent files as a no-op (already gone).
	// This is the correct safety behavior — no phantom errors for missing files.
	if result.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0 (nonexistent = no-op)", result.Skipped)
	}
	if result.BytesFreed != 0 {
		t.Errorf("BytesFreed = %d, want 0", result.BytesFreed)
	}
}

func TestInstallerDirs(t *testing.T) {
	dirs := installerDirs("/Users/test")

	if len(dirs) != 4 {
		t.Fatalf("expected 4 directories, got %d", len(dirs))
	}

	// Verify labels
	labels := make(map[string]bool)
	for _, d := range dirs {
		labels[d.label] = true
	}
	for _, expected := range []string{"Downloads", "Desktop", "Homebrew", "iCloud"} {
		if !labels[expected] {
			t.Errorf("missing expected label %q", expected)
		}
	}
}

// createLargeFile creates a sparse file of the given size.
func createLargeFile(path string, size int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		return err
	}
	return nil
}
