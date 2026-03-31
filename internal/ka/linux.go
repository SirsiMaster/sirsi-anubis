package ka

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/SirsiMaster/sirsi-pantheon/internal/logging"
)

// LinuxProvider implements GhostProvider for Linux.
// It scans XDG directories for residuals, reads .desktop files for
// installed app detection, and checks dpkg for package status.
type LinuxProvider struct{}

// linuxUserResidualDirs are XDG-standard directories where app data hides.
var linuxUserResidualDirs = []residualLocation{
	{ResidualPreferences, "~/.config", false},
	{ResidualAppSupport, "~/.local/share", false},
	{ResidualCaches, "~/.cache", false},
	{ResidualSavedState, "~/.local/state", false},
}

// linuxSystemResidualDirs require root access.
var linuxSystemResidualDirs = []residualLocation{
	{ResidualPreferences, "/etc", true},
	{ResidualLogs, "/var/log", true},
	{ResidualAppSupport, "/opt", true},
}

func (l *LinuxProvider) ResidualLocations(includeSudo bool) []residualLocation {
	locations := make([]residualLocation, len(linuxUserResidualDirs))
	copy(locations, linuxUserResidualDirs)
	if includeSudo {
		locations = append(locations, linuxSystemResidualDirs...)
	}
	return locations
}

func (l *LinuxProvider) BuildInstalledIndex(ctx context.Context, s *Scanner) error {
	// Parse .desktop files for installed GUI apps
	desktopDirs := []string{
		"/usr/share/applications",
		expandPath("~/.local/share/applications", s.homeDir),
	}

	for _, dir := range desktopDirs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entries, err := s.DirReader(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			desktopPath := filepath.Join(dir, entry.Name())
			appName := l.parseDesktopName(desktopPath)
			if appName != "" {
				s.installedNames[strings.ToLower(appName)] = true
			}

			// Also index by desktop file stem (e.g., "firefox" from "firefox.desktop")
			stem := strings.TrimSuffix(entry.Name(), ".desktop")
			s.installedNames[strings.ToLower(stem)] = true
		}
	}

	// Index dpkg packages if available
	l.indexDpkg(ctx, s)

	return nil
}

// parseDesktopName reads the Name= field from a .desktop file.
func (l *LinuxProvider) parseDesktopName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name=") {
			return strings.TrimPrefix(line, "Name=")
		}
	}
	return ""
}

// indexDpkg adds dpkg-installed packages to the installed names index.
func (l *LinuxProvider) indexDpkg(ctx context.Context, s *Scanner) {
	cmd := s.ExecCommand(ctx, "dpkg", "--get-selections")
	output, err := cmd.Output()
	if err != nil {
		logging.Debug("dpkg not available or failed", "err", err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == "install" {
			pkgName := fields[0]
			// Strip architecture suffix (e.g., "firefox:amd64" → "firefox")
			if idx := strings.Index(pkgName, ":"); idx > 0 {
				pkgName = pkgName[:idx]
			}
			s.installedNames[strings.ToLower(pkgName)] = true
		}
	}
}

func (l *LinuxProvider) ScanRegistry(ctx context.Context, s *Scanner) map[string]bool {
	ghosts := make(map[string]bool)

	// Check .desktop files where the Exec binary no longer exists
	desktopDirs := []string{
		"/usr/share/applications",
		expandPath("~/.local/share/applications", s.homeDir),
	}

	for _, dir := range desktopDirs {
		entries, err := s.DirReader(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			desktopPath := filepath.Join(dir, entry.Name())
			execPath := l.parseDesktopExec(desktopPath)
			if execPath == "" {
				continue
			}

			// Check if the executable still exists
			if _, err := os.Stat(execPath); os.IsNotExist(err) {
				stem := strings.TrimSuffix(entry.Name(), ".desktop")
				if !s.isInstalled(stem, entry.Name()) {
					ghosts[stem] = true
				}
			}
		}
	}

	logging.Debug("linux: scanned desktop file registry", "ghosts", len(ghosts))
	return ghosts
}

// parseDesktopExec reads the Exec= field from a .desktop file and returns the binary path.
func (l *LinuxProvider) parseDesktopExec(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Exec=") {
			execLine := strings.TrimPrefix(line, "Exec=")
			// Exec line may have arguments: "/usr/bin/firefox %u"
			fields := strings.Fields(execLine)
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}
	return ""
}

func (l *LinuxProvider) ExtractAppID(name string) string {
	// On Linux, directory names in ~/.config etc. are typically
	// the app name itself (e.g., "firefox", "google-chrome", "vlc")
	// Strip common extensions
	name = strings.TrimSuffix(name, ".conf")
	name = strings.TrimSuffix(name, ".cfg")

	// Must not be hidden
	if strings.HasPrefix(name, ".") {
		return ""
	}

	// Require at least 2 characters
	if len(name) < 2 {
		return ""
	}

	return strings.ToLower(name)
}

func (l *LinuxProvider) IsSystemID(id string) bool {
	systemNames := []string{
		"systemd", "dbus", "networkmanager", "network-manager",
		"dconf", "glib-2.0", "gtk-2.0", "gtk-3.0", "gtk-4.0",
		"fontconfig", "ibus", "pulse", "pipewire", "alsa",
		"xdg-desktop-portal", "xfce4", "gnome-shell", "cinnamon",
		"kde", "plasma", "kdeconnect", "baloo",
		"apt", "dpkg", "snap", "flatpak",
		"bash", "zsh", "fish",
		"man", "groff", "info",
		"mime", "shared-mime-info",
		"gvfs", "udisks2", "polkit-1",
	}

	idLower := strings.ToLower(id)
	for _, sys := range systemNames {
		if idLower == sys || strings.HasPrefix(idLower, sys+"-") {
			return true
		}
	}

	return false
}
