// Package oplog provides operation logging for destructive actions.
//
// Every file deletion, cleanup, or purge is logged to
// ~/Library/Logs/sirsi/operations.log with timestamp, action, path,
// and bytes affected. Disable with SIRSI_NO_OPLOG=1.
package oplog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	logFile *os.File
)

// Log records a destructive operation.
func Log(action, path string, bytes int64) {
	if os.Getenv("SIRSI_NO_OPLOG") == "1" {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if logFile == nil {
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, "Library", "Logs", "sirsi")
		_ = os.MkdirAll(dir, 0o755)
		f, err := os.OpenFile(
			filepath.Join(dir, "operations.log"),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644,
		)
		if err != nil {
			return
		}
		logFile = f
	}

	ts := time.Now().Format("2006-01-02T15:04:05")
	sizeStr := ""
	if bytes > 0 {
		sizeStr = fmt.Sprintf(" (%d bytes)", bytes)
	}
	_, _ = fmt.Fprintf(logFile, "%s  %s  %s%s\n", ts, action, path, sizeStr)
}

// Close flushes and closes the log file.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}
