package thoth

import (
	"fmt"
	"testing"
	"time"
)

// ── delegate.go Tests ──────────────────────────────────────────────
// Uses injectable mocks per Rule A16 + A21 (mutex-protected function pointers).

func TestBinaryAvailable_Found(t *testing.T) {
	t.Parallel()

	old := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})
	defer func() { setLookPathFn(old) }()

	if !BinaryAvailable("thoth-init") {
		t.Error("BinaryAvailable should return true when binary exists")
	}
}

func TestBinaryAvailable_NotFound(t *testing.T) {
	t.Parallel()

	old := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(old)
	}()

	if BinaryAvailable("thoth-init") {
		t.Error("BinaryAvailable should return false when binary missing")
	}
}

func TestTryDelegateInit_BinaryMissing(t *testing.T) {
	t.Parallel()

	old := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(old)
	}()

	delegated, err := TryDelegateInit(InitOptions{RepoRoot: "/tmp", Yes: true})
	if delegated {
		t.Error("should not delegate when binary missing")
	}
	if err != nil {
		t.Errorf("should not error when binary missing, got: %v", err)
	}
}

func TestTryDelegateInit_BinarySuccess(t *testing.T) {
	t.Parallel()

	oldLook := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(oldLook)
	}()

	oldRun := getRunCmdFn()
	var capturedArgs []string
	setRunCmdFn(func(name string, args ...string) (string, string, error) {
		capturedArgs = append([]string{name}, args...)
		return "init done\n", "", nil
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setRunCmdFn(oldRun)
	}()

	delegated, err := TryDelegateInit(InitOptions{
		RepoRoot: "/tmp/project",
		Name:     "myproj",
		Yes:      true,
	})
	if !delegated {
		t.Error("should delegate when binary available")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify args passed correctly
	if len(capturedArgs) < 1 || capturedArgs[0] != "thoth-init" {
		t.Errorf("expected thoth-init command, got %v", capturedArgs)
	}
}

func TestTryDelegateInit_BinaryFails(t *testing.T) {
	t.Parallel()

	oldLook := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(oldLook)
	}()

	oldRun := getRunCmdFn()
	setRunCmdFn(func(name string, args ...string) (string, string, error) {
		return "", "segfault", fmt.Errorf("exit 1")
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setRunCmdFn(oldRun)
	}()

	delegated, err := TryDelegateInit(InitOptions{Yes: true})
	if !delegated {
		t.Error("should report delegated=true even on failure")
	}
	if err == nil {
		t.Error("should return error on binary failure")
	}
}

func TestTryDelegateSync_BinaryMissing(t *testing.T) {
	t.Parallel()

	old := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(old)
	}()

	delegated, err := TryDelegateSync(SyncOptions{RepoRoot: "/tmp"})
	if delegated {
		t.Error("should not delegate when binary missing")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTryDelegateSync_BinarySuccess(t *testing.T) {
	t.Parallel()

	oldLook := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(oldLook)
	}()

	oldRun := getRunCmdFn()
	setRunCmdFn(func(name string, args ...string) (string, string, error) {
		return "", "", nil
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setRunCmdFn(oldRun)
	}()

	delegated, err := TryDelegateSync(SyncOptions{RepoRoot: "/tmp", UpdateDate: true})
	if !delegated {
		t.Error("should delegate when binary available")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTryDelegateCompact_BinaryMissing(t *testing.T) {
	t.Parallel()

	old := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "", fmt.Errorf("not found")
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(old)
	}()

	delegated, err := TryDelegateCompact(CompactOptions{RepoRoot: "/tmp", Summary: "test"})
	if delegated {
		t.Error("should not delegate when binary missing")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTryDelegateCompact_BinarySuccess(t *testing.T) {
	t.Parallel()

	oldLook := getLookPathFn()
	setLookPathFn(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setLookPathFn(oldLook)
	}()

	oldRun := getRunCmdFn()
	var capturedArgs []string
	setRunCmdFn(func(name string, args ...string) (string, string, error) {
		capturedArgs = append([]string{name}, args...)
		return "", "", nil
	})
	defer func() {
		time.Sleep(10 * time.Millisecond)
		setRunCmdFn(oldRun)
	}()

	delegated, err := TryDelegateCompact(CompactOptions{
		RepoRoot: "/tmp",
		Summary:  "Use interfaces",
		MaxAge:   30,
		MaxKeep:  10,
	})
	if !delegated {
		t.Error("should delegate when binary available")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify args
	if len(capturedArgs) == 0 || capturedArgs[0] != "thoth-compact" {
		t.Errorf("expected thoth-compact, got %v", capturedArgs)
	}
}
