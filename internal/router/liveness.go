// Package router — liveness.go
//
// OS-truth PID liveness for the CTR thread registry. A registered thread
// records the PID of its long-lived agent process; this is the single source
// that answers "is that process actually running right now?" against the live
// OS process table — not against a heartbeat timestamp.
//
// Why this exists: the registry classified threads as live/stale purely on
// heartbeat recency, and the original reaper used `kill -0` (signal 0), which
// a DEFUNCT (zombie, state Z) process answers successfully because it still
// occupies a slot in the table until its parent reaps it. The result was dead
// and zombie PIDs presenting as `active` forever (the CTR false-active bug).
// This primitive distinguishes alive / gone / defunct so the reaper and Horus
// agree with `ps`.
package router

import "sync"

// PIDState is the OS-truth liveness of a recorded PID.
type PIDState string

const (
	// PIDAlive: the process exists and is not defunct — a real live agent.
	PIDAlive PIDState = "alive"
	// PIDGone: no such process in the table.
	PIDGone PIDState = "gone"
	// PIDDefunct: the process exists but is a zombie (state Z) awaiting reaping
	// by its parent. It answers `kill -0` but is NOT a live agent.
	PIDDefunct PIDState = "defunct"
	// PIDUnknown: pid not recorded (<=0). Cannot verify — never treat as dead.
	PIDUnknown PIDState = "unknown"
)

// pidStateFn is the injectable liveness prober (Rule A16). Guarded by a
// RWMutex (Rule A21) so tests can install gone/defunct stubs without racing
// any concurrent reader.
var (
	pidStateMu sync.RWMutex
	pidStateFn = defaultPIDState
)

func getPIDStateFn() func(int) PIDState {
	pidStateMu.RLock()
	defer pidStateMu.RUnlock()
	return pidStateFn
}

// setPIDStateFn installs a prober (nil restores the platform default).
func setPIDStateFn(fn func(int) PIDState) {
	pidStateMu.Lock()
	defer pidStateMu.Unlock()
	if fn == nil {
		fn = defaultPIDState
	}
	pidStateFn = fn
}

// PIDStateOf returns the OS-truth liveness of pid using the active prober.
// A non-positive pid is unverifiable and reported as PIDUnknown.
func PIDStateOf(pid int) PIDState {
	if pid <= 0 {
		return PIDUnknown
	}
	return getPIDStateFn()(pid)
}

// DeadByOSTruth reports whether a recorded PID is confirmed not running:
// gone entirely, or defunct (Z) awaiting reaping. PIDUnknown is NOT dead —
// an unverifiable PID must never be reaped.
func DeadByOSTruth(state PIDState) bool {
	return state == PIDGone || state == PIDDefunct
}
