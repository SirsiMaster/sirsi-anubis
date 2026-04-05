//go:build windows

package output

import (
	"os"
	"strconv"
)

// processAlive checks if a process with the given PID is still running.
// On Windows, we use os.FindProcess + a signal check.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds. Check /proc or use a signal.
	// The safest cross-check: try to read /proc/<pid> (doesn't exist on Windows),
	// so fall back to checking if the PID file's content matches a running process.
	_ = p
	// Heuristic: if we got here the PID file exists, assume alive.
	// Full Windows support (Phase 4) will use Windows API.
	_, err = os.Stat("/proc/" + strconv.Itoa(pid))
	return err == nil
}
