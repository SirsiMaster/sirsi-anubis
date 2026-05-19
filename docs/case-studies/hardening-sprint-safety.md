# Case Study: Hardening Sprint — From Unsafe Deletion to Reversible Cleanup

**Date:** May 2026
**Category:** Safety
**Module:** Cleaner, Jackal

## Problem

Codex reviewed the Pantheon TUI v0.21.0 session and found critical safety violations:

1. `purge.go` and `installer.go` called `os.RemoveAll()` directly — bypassing all safety checks
2. `moveToTrash()` fell back to permanent deletion when macOS Finder/osascript failed
3. On Linux/Android/iOS (no trash support), cleanup silently permanently deleted files
4. No dry-run verification was available for the TUI cleanup path

## Solution

### DeleteFileReversible

A new API that refuses to permanently delete when the platform doesn't support trash:

```go
func DeleteFileReversible(path string, dryRun bool) (int64, error) {
    if !platform.Current().SupportsTrash() {
        return 0, fmt.Errorf("cannot safely delete: platform has no trash")
    }
    return size, platform.Current().MoveToTrash(path)
}
```

### Safety Gateway

All destructive TUI actions now route through a central `SafetyGateway` interface — one confirmation point for clean, purge, ghost exorcism, and installer removal.

## Evidence

- `moveToTrash()` dangerous fallback: removed entirely
- `os.RemoveAll` direct calls in jackal: replaced with `cleaner.DeleteFile()`
- Tests prove: `TestPurgeArtifacts_NoTrashPlatformSkips` — directory survives on trashless platforms
- Tests prove: `TestRemoveInstallers_NoTrashPlatformSkips` — same behavior for installers
