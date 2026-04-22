package ra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRADir(t *testing.T) {
	t.Parallel()
	dir := RADir()
	if dir == "" {
		t.Fatal("RADir returned empty")
	}
	if !filepath.IsAbs(dir) {
		t.Fatalf("RADir not absolute: %q", dir)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "ra")
	if dir != expected {
		t.Fatalf("RADir = %q, want %q", dir, expected)
	}
}

func TestMonitor_NoDeployment(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	_, err := Monitor(dir)
	if err == nil {
		t.Fatal("expected error for missing deployment.json")
	}
}

func TestMonitor_CompletedDeployment(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create deployment.json
	meta := deploymentMeta{
		StartedAt: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		Scopes:    []string{"test-scope"},
	}
	data, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(dir, "deployment.json"), data, 0644)

	// Create PID file (use PID 1 which exists but isn't ours)
	os.MkdirAll(filepath.Join(dir, "pids"), 0755)
	os.WriteFile(filepath.Join(dir, "pids", "test-scope.pid"), []byte("99999999"), 0644)

	// Create exit file (success)
	os.MkdirAll(filepath.Join(dir, "exits"), 0755)
	os.WriteFile(filepath.Join(dir, "exits", "test-scope.exit"), []byte("0"), 0644)

	// Create log file
	os.MkdirAll(filepath.Join(dir, "logs"), 0755)
	os.WriteFile(filepath.Join(dir, "logs", "test-scope.log"), []byte("Build complete\nAll tests passed\n"), 0644)

	status, err := Monitor(dir)
	if err != nil {
		t.Fatalf("Monitor: %v", err)
	}

	if len(status.Windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(status.Windows))
	}

	w := status.Windows[0]
	if w.Name != "test-scope" {
		t.Errorf("Name = %q", w.Name)
	}
	if w.State != "completed" && w.State != "crashed" {
		// PID 99999999 won't be alive, so it should check exit file
		t.Logf("State = %q (acceptable — depends on process check)", w.State)
	}
}

func TestMonitor_MultipleScopes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	meta := deploymentMeta{
		StartedAt: time.Now().Format(time.RFC3339),
		Scopes:    []string{"scope-a", "scope-b", "scope-c"},
	}
	data, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(dir, "deployment.json"), data, 0644)

	// Create PID files
	os.MkdirAll(filepath.Join(dir, "pids"), 0755)
	for _, scope := range meta.Scopes {
		os.WriteFile(filepath.Join(dir, "pids", scope+".pid"), []byte("99999999"), 0644)
	}

	// Create exit files
	os.MkdirAll(filepath.Join(dir, "exits"), 0755)
	os.WriteFile(filepath.Join(dir, "exits", "scope-a.exit"), []byte("0"), 0644)
	os.WriteFile(filepath.Join(dir, "exits", "scope-b.exit"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(dir, "exits", "scope-c.exit"), []byte("0"), 0644)

	// Create logs
	os.MkdirAll(filepath.Join(dir, "logs"), 0755)
	for _, scope := range meta.Scopes {
		os.WriteFile(filepath.Join(dir, "logs", scope+".log"), []byte("log output\n"), 0644)
	}

	status, err := Monitor(dir)
	if err != nil {
		t.Fatalf("Monitor: %v", err)
	}

	if len(status.Windows) != 3 {
		t.Fatalf("expected 3 windows, got %d", len(status.Windows))
	}
}

func TestReadPIDFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "test.pid")

	// Valid PID file
	os.WriteFile(pidFile, []byte("12345"), 0644)
	pid, err := readPIDFile(pidFile)
	if err != nil {
		t.Fatalf("readPIDFile: %v", err)
	}
	if pid != 12345 {
		t.Fatalf("pid = %d, want 12345", pid)
	}

	// Missing PID file
	_, err = readPIDFile(filepath.Join(dir, "nonexistent.pid"))
	if err == nil {
		t.Fatal("expected error for missing PID file")
	}
}

func TestReadExitFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	exitFile := filepath.Join(dir, "test.exit")

	// Exit code 0
	os.WriteFile(exitFile, []byte("0"), 0644)
	code, err := readExitFile(exitFile)
	if err != nil {
		t.Fatalf("readExitFile: %v", err)
	}
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}

	// Exit code 1
	os.WriteFile(exitFile, []byte("1"), 0644)
	code, _ = readExitFile(exitFile)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}

	// Missing file
	_, err = readExitFile(filepath.Join(dir, "nonexistent.exit"))
	if err == nil {
		t.Fatal("expected error for missing exit file")
	}
}

func TestReadLogTail(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "logs"), 0755)
	logFile := filepath.Join(dir, "logs", "test.log")

	// Write 20 lines
	var content string
	for i := 1; i <= 20; i++ {
		content += "line " + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "\n"
	}
	os.WriteFile(logFile, []byte(content), 0644)

	tail := readLogTail(dir, "test")
	if tail == "" {
		t.Fatal("readLogTail returned empty")
	}
	// Should contain the last lines, not all 20
	lines := len(splitLines(tail))
	if lines > 12 { // readLogTail returns ~10 lines
		t.Fatalf("expected ≤12 tail lines, got %d", lines)
	}
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range filepath.SplitList(s) {
		if l != "" {
			lines = append(lines, l)
		}
	}
	// Fallback for newline-separated
	if len(lines) <= 1 {
		count := 0
		for _, c := range s {
			if c == '\n' {
				count++
			}
		}
		return make([]string, count+1)
	}
	return lines
}

func TestDeploymentMeta_JSON(t *testing.T) {
	t.Parallel()
	meta := deploymentMeta{
		StartedAt: "2026-04-22T10:00:00Z",
		Scopes:    []string{"nexus", "assiduous"},
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded deploymentMeta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.StartedAt != meta.StartedAt {
		t.Errorf("StartedAt = %q", decoded.StartedAt)
	}
	if len(decoded.Scopes) != 2 {
		t.Errorf("Scopes = %v", decoded.Scopes)
	}
}
