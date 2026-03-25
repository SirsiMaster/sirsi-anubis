package platform

import (
	"runtime"
	"testing"
)

func TestCurrent_ReturnsCorrectPlatform(t *testing.T) {
	p := Current()
	if p == nil {
		t.Fatal("Current() returned nil")
	}
	if p.Name() != runtime.GOOS {
		t.Errorf("Name() = %q, want %q", p.Name(), runtime.GOOS)
	}
}

func TestSet_OverridesPlatform(t *testing.T) {
	mock := &Mock{PickFolderPath: "/test/path"}
	Set(mock)
	defer Reset()

	p := Current()
	if p.Name() != "mock" {
		t.Errorf("after Set(mock), Name() = %q, want %q", p.Name(), "mock")
	}
}

func TestReset_RestoresDetected(t *testing.T) {
	Set(&Mock{})
	Reset()

	p := Current()
	if p.Name() != runtime.GOOS {
		t.Errorf("after Reset(), Name() = %q, want %q", p.Name(), runtime.GOOS)
	}
}

func TestMock_RecordsTrashCalls(t *testing.T) {
	m := &Mock{}
	_ = m.MoveToTrash("/tmp/file1")
	_ = m.MoveToTrash("/tmp/file2")

	if len(m.TrashCalls) != 2 {
		t.Errorf("TrashCalls = %d, want 2", len(m.TrashCalls))
	}
	if m.TrashCalls[0] != "/tmp/file1" {
		t.Errorf("TrashCalls[0] = %q, want %q", m.TrashCalls[0], "/tmp/file1")
	}
}

func TestMock_PickFolder(t *testing.T) {
	m := &Mock{PickFolderPath: "/chosen/folder"}
	path, err := m.PickFolder()
	if err != nil {
		t.Fatalf("PickFolder() error: %v", err)
	}
	if path != "/chosen/folder" {
		t.Errorf("PickFolder() = %q, want %q", path, "/chosen/folder")
	}
}

func TestMock_OpenBrowser(t *testing.T) {
	m := &Mock{}
	_ = m.OpenBrowser("https://example.com")
	if m.OpenBrowserURL != "https://example.com" {
		t.Errorf("OpenBrowserURL = %q, want %q", m.OpenBrowserURL, "https://example.com")
	}
}

func TestMock_SupportsTrash(t *testing.T) {
	m := &Mock{}
	if !m.SupportsTrash() {
		t.Error("Mock.SupportsTrash() should return true")
	}
}

func TestMock_ProtectedPrefixes(t *testing.T) {
	m := &Mock{}
	prefixes := m.ProtectedPrefixes()
	if len(prefixes) < 3 {
		t.Errorf("expected at least 3 protected prefixes, got %d", len(prefixes))
	}
}

func TestDarwin_ProtectedPrefixes(t *testing.T) {
	d := &Darwin{}
	prefixes := d.ProtectedPrefixes()
	mustContain := []string{"/System/", "/usr/", "/bin/"}
	for _, want := range mustContain {
		found := false
		for _, p := range prefixes {
			if p == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Darwin.ProtectedPrefixes() missing %q", want)
		}
	}
}

func TestLinux_ProtectedPrefixes(t *testing.T) {
	l := &Linux{}
	prefixes := l.ProtectedPrefixes()
	mustContain := []string{"/boot/", "/etc/", "/proc/", "/sys/"}
	for _, want := range mustContain {
		found := false
		for _, p := range prefixes {
			if p == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Linux.ProtectedPrefixes() missing %q", want)
		}
	}
}

func TestDarwin_SupportsTrash(t *testing.T) {
	d := &Darwin{}
	if !d.SupportsTrash() {
		t.Error("Darwin should support trash")
	}
}

func TestDetectFor(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{"darwin", "darwin"},
		{"linux", "linux"},
		{"windows", "linux"}, // Fallback
		{"unknown", "linux"}, // Fallback
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			p := detectFor(tt.goos)
			if p.Name() != tt.want {
				t.Errorf("detectFor(%q).Name() = %q, want %q", tt.goos, p.Name(), tt.want)
			}
		})
	}
}

func TestDarwin_Name(t *testing.T) {
	d := &Darwin{}
	if d.Name() != "darwin" {
		t.Errorf("Darwin.Name() = %q, want %q", d.Name(), "darwin")
	}
}

func TestLinux_Name(t *testing.T) {
	l := &Linux{}
	if l.Name() != "linux" {
		t.Errorf("Linux.Name() = %q, want %q", l.Name(), "linux")
	}
}

func TestLinux_SupportsTrash(t *testing.T) {
	l := &Linux{}
	if l.SupportsTrash() {
		t.Error("Linux should NOT claim to support trash yet")
	}
}

func TestMock_Name(t *testing.T) {
	m := &Mock{}
	if m.Name() != "mock" {
		t.Errorf("Mock.Name() = %q, want %q", m.Name(), "mock")
	}
}

func TestMock_Getenv(t *testing.T) {
	m := &Mock{Env: map[string]string{"FOO": "BAR"}}
	if m.Getenv("FOO") != "BAR" {
		t.Errorf("Getenv(FOO) = %q, want %q", m.Getenv("FOO"), "BAR")
	}
	if m.Getenv("BAZ") != "" {
		t.Errorf("Getenv(BAZ) = %q, want empty", m.Getenv("BAZ"))
	}
}

func TestMock_UserHomeDir(t *testing.T) {
	m := &Mock{HomeDir: "/users/mock"}
	home, err := m.UserHomeDir()
	if err != nil || home != "/users/mock" {
		t.Errorf("UserHomeDir() = %q, err = %v", home, err)
	}
}

func TestMock_Getwd(t *testing.T) {
	m := &Mock{WorkDir: "/work/mock"}
	wd, err := m.Getwd()
	if err != nil || wd != "/work/mock" {
		t.Errorf("Getwd() = %q, err = %v", wd, err)
	}
}

func TestMock_Command(t *testing.T) {
	m := &Mock{
		CommandResults: map[string]string{
			"echo hello": "hello\n",
		},
	}
	out, err := m.Command("echo", "hello")
	if err != nil || string(out) != "hello\n" {
		t.Errorf("Command(echo hello) = %q, err = %v", string(out), err)
	}

	// Test fallback
	out, err = m.Command("ls", "-la")
	if err != nil || string(out) != "" {
		t.Errorf("Command(ls -la) should return empty string, got %q", string(out))
	}
}

func TestMock_Processes(t *testing.T) {
	m := &Mock{ProcessList: []string{"go", "grep", "ls"}}
	procs, err := m.Processes()
	if err != nil || len(procs) != 3 {
		t.Errorf("Processes() = %v, err = %v", procs, err)
	}
}
