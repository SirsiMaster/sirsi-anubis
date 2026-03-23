// Package platform provides an abstraction layer over OS-specific operations.
//
// The Platform interface enables:
// 1. Cross-platform support (macOS, Linux, Windows)
// 2. Testability (mock platform operations in unit tests)
// 3. Clean separation between business logic and OS integration
//
// Architecture:
//   - Current() returns the platform for the running OS
//   - Mock() returns a test-friendly implementation
//   - All OS-specific code lives behind this interface
package platform

// Platform abstracts OS-specific operations.
// All system interactions that differ across platforms go through this interface.
type Platform interface {
	// MoveToTrash moves a file to the OS-native trash/recycle bin.
	// Returns an error if trash is not supported (falls back to permanent delete).
	MoveToTrash(path string) error

	// ProtectedPrefixes returns the list of path prefixes that must
	// never be modified or deleted on this platform.
	ProtectedPrefixes() []string

	// PickFolder opens a native folder-picker dialog and returns
	// the selected absolute path. Returns empty string if canceled.
	PickFolder() (string, error)

	// OpenBrowser opens a URL in the system default browser.
	OpenBrowser(url string) error

	// Name returns the platform identifier ("darwin", "linux", "windows", "mock").
	Name() string

	// SupportsTrash returns true if the platform supports reversible trash.
	SupportsTrash() bool
}
