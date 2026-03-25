# ADR-009: Injectable System Providers for Deterministic Testing

## Status
Proposed (2026-03-24)

## Context
As Pantheon approaches 99% test coverage (current: 90% weighted average), we hit a "hard ceiling" where remaining code paths involved non-deterministic system calls (signals, process termination, `exec.Command`, `os.RemoveAll`, `os.UserHomeDir`).

Testing these paths previously required:
1.  **System Mutation**: Actually killing processes or deleting files on the host (risky and flakey).
2.  **Skipping**: Using `if runtime.GOOS != "darwin"` or `t.Skip()` which left critical safety logic (e.g., `isProtectedProcess`) untested.

## Decision
We will standardize on **Interface Injection** (or Function Types) for all system-level side effects.

### Implementation Patterns

#### 1. Function Types (Lightweight)
For single-purpose actions like process killing or command execution.

```go
// internal/guard/slayer.go
type ProcessKiller func(pid int, signal os.Signal) error

func Slay(pid int) error {
    return SlayWith(pid, os.Kill, killProcess)
}

func SlayWith(pid int, sig os.Signal, killer ProcessKiller) error {
    // Logic here uses killer(pid, sig)
}
```

#### 2. Interface Wrappers (Robust)
For complex providers like the Antigravity IPC Bridge or Ma'at CI Runners.

```go
// internal/maat/pipeline.go
type Runner interface {
    Run(args ...string) (string, error)
}
```

### Constraints
1.  **Zero API Change**: Every module MUST export its original simple function (e.g., `Fix(dryRun)`) which delegates to an internal or unexported "With" variant (e.g., `FixWith(dryRun, runner)`).
2.  **Default Provider**: The original function MUST use the "Real" system provider by default.
3.  **No Global State**: Avoid package-level mock variables; use dependency injection through the "With" variants.

## Consequences
- **Coverage**: Achieved 91.0% in `guard` (up from 89%) and 88.0% in `maat` (up from 80%) in a single session.
- **Safety**: We can now test "What happens if `kill` fails for a root process?" without actually failing a root kill.
- **Complexity**: Codebase gained ~5% more lines due to "With" variants and interface definitions.
- **Speed**: Tests run faster as they no longer wait for real `exec.Command` timeouts or `lsregister` dumps.
