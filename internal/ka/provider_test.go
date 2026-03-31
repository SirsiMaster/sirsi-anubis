package ka

import (
	"context"
	"testing"
)

// TestMockProvider is a configurable GhostProvider for testing.
type TestMockProvider struct {
	Locations      []residualLocation
	BuildIndexFn   func(ctx context.Context, s *Scanner) error
	ScanRegistryFn func(ctx context.Context, s *Scanner) map[string]bool
	ExtractAppIDFn func(name string) string
	IsSystemIDFn   func(id string) bool
}

func (m *TestMockProvider) ResidualLocations(includeSudo bool) []residualLocation {
	return m.Locations
}

func (m *TestMockProvider) BuildInstalledIndex(ctx context.Context, s *Scanner) error {
	if m.BuildIndexFn != nil {
		return m.BuildIndexFn(ctx, s)
	}
	return nil
}

func (m *TestMockProvider) ScanRegistry(ctx context.Context, s *Scanner) map[string]bool {
	if m.ScanRegistryFn != nil {
		return m.ScanRegistryFn(ctx, s)
	}
	return make(map[string]bool)
}

func (m *TestMockProvider) ExtractAppID(name string) string {
	if m.ExtractAppIDFn != nil {
		return m.ExtractAppIDFn(name)
	}
	// Default: delegate to Darwin extractBundleID for backward compat
	return extractBundleID(name)
}

func (m *TestMockProvider) IsSystemID(id string) bool {
	if m.IsSystemIDFn != nil {
		return m.IsSystemIDFn(id)
	}
	return isSystemBundleID(id)
}

// --- Provider interface tests ---

func TestDarwinProvider_ResidualLocations(t *testing.T) {
	p := &DarwinProvider{}
	user := p.ResidualLocations(false)
	if len(user) != 12 {
		t.Errorf("expected 12 user locations, got %d", len(user))
	}
	all := p.ResidualLocations(true)
	if len(all) != 17 {
		t.Errorf("expected 17 total locations (12 user + 5 system), got %d", len(all))
	}
}

func TestDarwinProvider_ExtractAppID(t *testing.T) {
	p := &DarwinProvider{}
	if id := p.ExtractAppID("com.test.app.plist"); id != "com.test.app" {
		t.Errorf("expected com.test.app, got %q", id)
	}
	if id := p.ExtractAppID("NotABundleID"); id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestDarwinProvider_IsSystemID(t *testing.T) {
	p := &DarwinProvider{}
	if !p.IsSystemID("com.apple.Safari") {
		t.Error("com.apple.Safari should be system")
	}
	if p.IsSystemID("com.parallels.desktop") {
		t.Error("com.parallels.desktop should NOT be system")
	}
}

func TestLinuxProvider_ResidualLocations(t *testing.T) {
	p := &LinuxProvider{}
	user := p.ResidualLocations(false)
	if len(user) != 4 {
		t.Errorf("expected 4 user locations, got %d", len(user))
	}
	all := p.ResidualLocations(true)
	if len(all) != 7 {
		t.Errorf("expected 7 total locations, got %d", len(all))
	}
}

func TestLinuxProvider_ExtractAppID(t *testing.T) {
	p := &LinuxProvider{}
	if id := p.ExtractAppID("firefox"); id != "firefox" {
		t.Errorf("expected firefox, got %q", id)
	}
	if id := p.ExtractAppID(".hidden"); id != "" {
		t.Errorf("expected empty for hidden, got %q", id)
	}
	if id := p.ExtractAppID("x"); id != "" {
		t.Errorf("expected empty for single char, got %q", id)
	}
}

func TestLinuxProvider_IsSystemID(t *testing.T) {
	p := &LinuxProvider{}
	if !p.IsSystemID("systemd") {
		t.Error("systemd should be system")
	}
	if !p.IsSystemID("dbus") {
		t.Error("dbus should be system")
	}
	if p.IsSystemID("firefox") {
		t.Error("firefox should NOT be system")
	}
}

func TestWindowsProvider_ResidualLocations(t *testing.T) {
	p := &WindowsProvider{}
	user := p.ResidualLocations(false)
	if len(user) != 3 {
		t.Errorf("expected 3 user locations, got %d", len(user))
	}
	all := p.ResidualLocations(true)
	if len(all) != 4 {
		t.Errorf("expected 4 total locations, got %d", len(all))
	}
}

func TestWindowsProvider_ExtractAppID(t *testing.T) {
	p := &WindowsProvider{}
	if id := p.ExtractAppID("SomeApp"); id != "someapp" {
		t.Errorf("expected someapp, got %q", id)
	}
	if id := p.ExtractAppID("."); id != "" {
		t.Errorf("expected empty for hidden, got %q", id)
	}
}

func TestWindowsProvider_IsSystemID(t *testing.T) {
	p := &WindowsProvider{}
	if !p.IsSystemID("Microsoft") {
		t.Error("Microsoft should be system")
	}
	if !p.IsSystemID("Windows") {
		t.Error("Windows should be system")
	}
	if p.IsSystemID("Firefox") {
		t.Error("Firefox should NOT be system")
	}
}

func TestWindowsProvider_BuildInstalledIndex_Stub(t *testing.T) {
	p := &WindowsProvider{}
	s := NewScanner()
	err := p.BuildInstalledIndex(context.Background(), s)
	if err != nil {
		t.Errorf("stub should not error: %v", err)
	}
}

func TestWindowsProvider_ScanRegistry_Stub(t *testing.T) {
	p := &WindowsProvider{}
	s := NewScanner()
	ghosts := p.ScanRegistry(context.Background(), s)
	if len(ghosts) != 0 {
		t.Errorf("stub should return empty, got %d", len(ghosts))
	}
}
