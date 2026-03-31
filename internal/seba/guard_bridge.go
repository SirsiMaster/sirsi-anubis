// Package seba — guard_bridge.go
//
// Breaks the seba → guard import cycle by using function injection.
// The main package (or any wiring package) registers guard.Audit via
// RegisterGuardAudit() at init time, so seba never imports guard directly.
package seba

import "fmt"

// ── Mirror types from guard to avoid import cycle ───────────────────

// ProcessInfo mirrors guard.ProcessInfo for runtime diagram generation.
type ProcessInfo struct {
	PID        int
	Name       string
	Command    string
	RSS        int64
	VSZ        int64
	User       string
	CPUPercent float64
	Group      string
}

// ProcessGroup mirrors guard.ProcessGroup for runtime diagram generation.
type ProcessGroup struct {
	Name       string
	Processes  []ProcessInfo
	TotalRSS   int64
	TotalCount int
}

// AuditResult mirrors guard.AuditResult for runtime diagram generation.
type AuditResult struct {
	TotalRAM     int64
	UsedRAM      int64
	FreeRAM      int64
	Groups       []ProcessGroup
	Orphans      []ProcessInfo
	TotalOrphans int
	OrphanRSS    int64
}

// guardAuditFn is the injected guard.Audit function.
// Registered via RegisterGuardAudit at startup.
var guardAuditFn func() (*AuditResult, error)

// RegisterGuardAudit sets the guard audit function to break the import cycle.
// Must be called before any runtime diagram generation.
func RegisterGuardAudit(fn func() (*AuditResult, error)) {
	guardAuditFn = fn
}

// guardAudit calls the registered guard audit function.
func guardAudit() (*AuditResult, error) {
	if guardAuditFn == nil {
		return nil, fmt.Errorf("guard audit not registered: call seba.RegisterGuardAudit()")
	}
	return guardAuditFn()
}
