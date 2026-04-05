//go:build !windows

package output

import "syscall"

// processAlive checks if a process with the given PID is still running.
func processAlive(pid int) bool {
	return pid > 0 && syscall.Kill(pid, 0) == nil
}
