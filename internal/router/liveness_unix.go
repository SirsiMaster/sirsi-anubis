//go:build !windows

package router

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// defaultPIDState reports OS-truth liveness on unix-like systems.
//
// It asks `ps` for the process state code: a leading "Z" means defunct
// (zombie) — the process is dead but unreaped, and must NOT count as a live
// agent. Empty output means the PID is gone. If `ps` is unavailable we fall
// back to signal 0, which can only distinguish exists/gone (not defunct).
func defaultPIDState(pid int) PIDState {
	out, err := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(pid)).Output()
	state := strings.TrimSpace(string(out))
	if err != nil || state == "" {
		// ps failed or reported nothing — confirm existence via signal 0.
		if syscall.Kill(pid, 0) == nil {
			return PIDAlive
		}
		return PIDGone
	}
	if strings.HasPrefix(strings.ToUpper(state), "Z") {
		return PIDDefunct
	}
	return PIDAlive
}
