package ka

import "context"

// GhostProvider abstracts platform-specific ghost detection strategies.
// Each platform (macOS, Linux, Windows) implements this interface to
// provide its own app inventory, residual locations, and registry scanning.
type GhostProvider interface {
	// ResidualLocations returns the platform's residual search paths.
	ResidualLocations(includeSudo bool) []residualLocation

	// BuildInstalledIndex populates the Scanner's installed app maps.
	BuildInstalledIndex(ctx context.Context, s *Scanner) error

	// ScanRegistry scans the platform's app registry for ghost entries.
	// macOS: lsregister, Linux: .desktop files, Windows: registry
	ScanRegistry(ctx context.Context, s *Scanner) map[string]bool

	// ExtractAppID extracts an app identifier from a filename or directory name.
	// macOS: reverse-DNS bundle ID, Linux: package/app name, Windows: display name
	ExtractAppID(name string) string

	// IsSystemID returns true if the identifier belongs to a system component
	// that should not be flagged as a ghost.
	IsSystemID(id string) bool
}
