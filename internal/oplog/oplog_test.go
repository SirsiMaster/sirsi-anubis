package oplog

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func resetState() {
	mu.Lock()
	if logFile != nil {
		logFile.Close()
	}
	logFile = nil
	mu.Unlock()
}

func TestLog_WritesEntry(t *testing.T) {
	resetState()

	tmp := t.TempDir()
	// Point home to temp so the log goes to a predictable location
	t.Setenv("HOME", tmp)

	Log("purge", "/tmp/test/node_modules", 1024)
	Close()

	logPath := filepath.Join(tmp, "Library", "Logs", "sirsi", "operations.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "purge") {
		t.Errorf("log missing action 'purge': %s", content)
	}
	if !strings.Contains(content, "/tmp/test/node_modules") {
		t.Errorf("log missing path: %s", content)
	}
	if !strings.Contains(content, "(1024 bytes)") {
		t.Errorf("log missing byte count: %s", content)
	}
}

func TestLog_ZeroBytesOmitsSize(t *testing.T) {
	resetState()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	Log("clean", "/tmp/empty", 0)
	Close()

	logPath := filepath.Join(tmp, "Library", "Logs", "sirsi", "operations.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "bytes") {
		t.Errorf("zero-byte entry should not contain size: %s", content)
	}
}

func TestLog_NoOplog(t *testing.T) {
	resetState()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("SIRSI_NO_OPLOG", "1")

	Log("purge", "/tmp/test", 500)

	logPath := filepath.Join(tmp, "Library", "Logs", "sirsi", "operations.log")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("log file should not be created when SIRSI_NO_OPLOG=1")
	}
}

func TestLog_Concurrent(t *testing.T) {
	resetState()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			Log("clean", "/tmp/concurrent", int64(n*100))
		}(i)
	}
	wg.Wait()
	Close()

	logPath := filepath.Join(tmp, "Library", "Logs", "sirsi", "operations.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 20 {
		t.Errorf("expected 20 log lines, got %d", len(lines))
	}
}

func TestClose_NilFile(t *testing.T) {
	resetState()
	// Close without any Log calls should not panic
	Close()
}

func TestClose_DoubleClose(t *testing.T) {
	resetState()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	Log("test", "/tmp/double", 1)
	Close()
	Close() // should not panic
}
