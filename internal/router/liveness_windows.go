//go:build windows

package router

import "os"

// defaultPIDState reports OS-truth liveness on Windows. Windows has no zombie
// concept, so liveness collapses to exists/gone: os.FindProcess opens a handle
// to the PID and fails if no such process exists.
func defaultPIDState(pid int) PIDState {
	if _, err := os.FindProcess(pid); err != nil {
		return PIDGone
	}
	return PIDAlive
}
