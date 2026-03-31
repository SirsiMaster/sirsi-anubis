package ka

import (
	"context"
	"strings"

	"github.com/SirsiMaster/sirsi-pantheon/internal/logging"
)

// WindowsProvider implements GhostProvider for Windows.
// This is a minimal stub — real implementation deferred to post-rc1.
type WindowsProvider struct{}

func (w *WindowsProvider) ResidualLocations(includeSudo bool) []residualLocation {
	locations := []residualLocation{
		{ResidualPreferences, "~/AppData/Roaming", false},
		{ResidualCaches, "~/AppData/Local", false},
		{ResidualAppSupport, "~/AppData/Local/Programs", false},
	}
	if includeSudo {
		locations = append(locations,
			residualLocation{ResidualAppSupport, "C:/ProgramData", true},
		)
	}
	return locations
}

func (w *WindowsProvider) BuildInstalledIndex(ctx context.Context, s *Scanner) error {
	// TODO: Query Windows registry HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall
	// via "reg query" command using s.ExecCommand
	logging.Debug("windows: BuildInstalledIndex is a stub")
	return nil
}

func (w *WindowsProvider) ScanRegistry(ctx context.Context, s *Scanner) map[string]bool {
	// TODO: Scan Start Menu shortcuts for dead links
	// TODO: Check registry uninstall entries for missing paths
	logging.Debug("windows: ScanRegistry is a stub")
	return make(map[string]bool)
}

func (w *WindowsProvider) ExtractAppID(name string) string {
	// On Windows, directory names in AppData are typically the app/vendor name
	name = strings.TrimSuffix(name, ".cfg")
	name = strings.TrimSuffix(name, ".ini")

	if strings.HasPrefix(name, ".") {
		return ""
	}
	if len(name) < 2 {
		return ""
	}

	return strings.ToLower(name)
}

func (w *WindowsProvider) IsSystemID(id string) bool {
	systemNames := []string{
		"microsoft", "windows", "windowsapps",
		"packages", "connecteddevicesplatform",
		"d3dscache", "comms",
	}

	idLower := strings.ToLower(id)
	for _, sys := range systemNames {
		if idLower == sys || strings.HasPrefix(idLower, sys) {
			return true
		}
	}

	return false
}
