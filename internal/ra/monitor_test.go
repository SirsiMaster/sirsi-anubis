package ra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupRaDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	raDir := filepath.Join(tmp, ".ra")
	os.MkdirAll(filepath.Join(raDir, "logs"), 0o755)
	return raDir
}

func TestReadPIDFile_Valid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "pid")
	os.WriteFile(path, []byte("12345\n"), 0o644)

	pid, err := readPIDFile(path)
	if err != nil {
		t.Fatalf("readPIDFile() error = %v", err)
	}
	if pid != 12345 {
		t.Errorf("pid = %d, want 12345", pid)
	}
}

func TestReadPIDFile_Invalid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "pid")
	os.WriteFile(path, []byte("notanumber"), 0o644)

	_, err := readPIDFile(path)
	if err == nil {
		t.Error("expected error for invalid PID")
	}
}

func TestReadPIDFile_Missing(t *testing.T) {
	_, err := readPIDFile("/nonexistent/pid")
	if err == nil {
		t.Error("expected error for missing PID file")
	}
}

func TestReadExitFile_Valid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "exit")
	os.WriteFile(path, []byte("0"), 0o644)

	code, err := readExitFile(path)
	if err != nil {
		t.Fatalf("readExitFile() error = %v", err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestReadExitFile_NonZero(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "exit")
	os.WriteFile(path, []byte("1"), 0o644)

	code, err := readExitFile(path)
	if err != nil {
		t.Fatalf("readExitFile() error = %v", err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestReadLogTail_UnderTenLines(t *testing.T) {
	raDir := setupRaDir(t)
	logPath := filepath.Join(raDir, "logs", "test-scope.log")
	os.WriteFile(logPath, []byte("line1\nline2\nline3"), 0o644)

	got := readLogTail(raDir, "test-scope")
	if got != "line1\nline2\nline3" {
		t.Errorf("readLogTail = %q, want all 3 lines", got)
	}
}

func TestReadLogTail_OverTenLines(t *testing.T) {
	raDir := setupRaDir(t)
	logPath := filepath.Join(raDir, "logs", "big.log")
	var lines string
	for i := 1; i <= 15; i++ {
		lines += "line" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "\n"
	}
	os.WriteFile(logPath, []byte(lines), 0o644)

	got := readLogTail(raDir, "big")
	gotLines := len(splitNonEmpty(got))
	if gotLines > 10 {
		t.Errorf("readLogTail should return ≤10 lines, got %d", gotLines)
	}
}

func splitNonEmpty(s string) []string {
	var result []string
	for _, line := range filepath.SplitList(s) {
		if line != "" {
			result = append(result, line)
		}
	}
	// Use simple split instead
	parts := make([]string, 0)
	for _, p := range filepath.SplitList(s) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func TestReadLogTail_MissingFile(t *testing.T) {
	got := readLogTail("/nonexistent", "nope")
	if got != "" {
		t.Errorf("readLogTail should return empty for missing file, got %q", got)
	}
}

func TestReadDeploymentMeta_Valid(t *testing.T) {
	raDir := setupRaDir(t)
	meta := deploymentMeta{
		Scopes:    []string{"test-scope"},
		StartedAt: "2026-05-14T12:00:00Z",
	}
	data, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(raDir, "deployment.json"), data, 0o644)

	got, err := readDeploymentMeta(raDir)
	if err != nil {
		t.Fatalf("readDeploymentMeta() error = %v", err)
	}
	if len(got.Scopes) != 1 || got.Scopes[0] != "test-scope" {
		t.Errorf("Scopes = %v, want [test-scope]", got.Scopes)
	}
}

func TestReadDeploymentMeta_Missing(t *testing.T) {
	_, err := readDeploymentMeta("/nonexistent")
	if err == nil {
		t.Error("expected error for missing deployment.json")
	}
}

func TestReadDeploymentMeta_MalformedJSON(t *testing.T) {
	raDir := setupRaDir(t)
	os.WriteFile(filepath.Join(raDir, "deployment.json"), []byte("{bad json"), 0o644)

	_, err := readDeploymentMeta(raDir)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{"~/Documents", home + "/Documents"},
		{"/usr/local", "/usr/local"},
		{"relative", "relative"},
		{"", ""},
	}
	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSyncStatus(t *testing.T) {
	synced := syncStatus(true)
	if synced == "" {
		t.Error("syncStatus(true) should not be empty")
	}
	skipped := syncStatus(false)
	if skipped == "" {
		t.Error("syncStatus(false) should not be empty")
	}
	if synced == skipped {
		t.Error("syncStatus(true) and syncStatus(false) should differ")
	}
}

func TestTryParseJSON_ValidArray(t *testing.T) {
	input := `[{"title":"Test","content":"Hello"}]`
	items := tryParseJSON(input)
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestTryParseJSON_InvalidJSON(t *testing.T) {
	items := tryParseJSON("not json at all")
	if len(items) != 0 {
		t.Errorf("expected 0 items for invalid JSON, got %d", len(items))
	}
}

func TestTryParseJSON_EmptyArray(t *testing.T) {
	items := tryParseJSON("[]")
	if len(items) != 0 {
		t.Errorf("expected 0 items for empty array, got %d", len(items))
	}
}
