package router

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderLaunchAgentPlistStartsDaemonWithNotify(t *testing.T) {
	opts := ServiceOptions{
		RepoRoot:   "/tmp/sirsi-pantheon",
		BinaryPath: "/usr/local/bin/sirsi",
		Label:      "com.sirsi.router.sirsi-pantheon",
		LogPath:    "/tmp/sirsi-pantheon/.agents/idea-router/logs/out.log",
		ErrPath:    "/tmp/sirsi-pantheon/.agents/idea-router/logs/err.log",
		PathEnv:    "/opt/homebrew/bin:/usr/bin:/bin",
	}
	plist := RenderLaunchAgentPlist(opts)
	for _, want := range []string{
		"<string>/usr/local/bin/sirsi</string>",
		"<string>router</string>",
		"<string>daemon</string>",
		"<key>SIRSI_ROUTER_NOTIFY</key>",
		"<key>PATH</key>",
		"/opt/homebrew/bin:/usr/bin:/bin",
		"<string>1</string>",
		"<key>RunAtLoad</key>",
		"<key>KeepAlive</key>",
	} {
		if !strings.Contains(plist, want) {
			t.Fatalf("plist missing %q:\n%s", want, plist)
		}
	}
}

func TestDefaultServiceOptionsAreRepoSpecific(t *testing.T) {
	opts := DefaultServiceOptions("/tmp/Sirsi Pantheon!", "/bin/sirsi")
	if !strings.Contains(opts.Label, "sirsi-pantheon") {
		t.Fatalf("label = %q, want repo slug", opts.Label)
	}
	if !strings.Contains(opts.PlistPath, opts.Label+".plist") {
		t.Fatalf("plist path = %q, want label plist", opts.PlistPath)
	}
}

func TestIsGoRunBinaryDetectsTemporaryExecutable(t *testing.T) {
	path := filepath.Join(os.TempDir(), "go-build123", "b001", "exe", "sirsi")
	if !IsGoRunBinary(path) {
		t.Fatalf("expected %q to be detected as go-run binary", path)
	}
	cachePath := filepath.Join(os.TempDir(), "sirsi-go-cache", "ad", "ad34eab1485676f0ffa3732e9201413cee6d5300d011f99f677bf74bef13aa7d-d", "sirsi")
	t.Setenv("GOCACHE", filepath.Join(os.TempDir(), "sirsi-go-cache"))
	if !IsGoRunBinary(cachePath) {
		t.Fatalf("expected %q to be detected as go-run cache binary", cachePath)
	}
	if IsGoRunBinary("/usr/local/bin/sirsi") {
		t.Fatal("stable binary reported as go-run binary")
	}
}

func TestLaunchAgentProgramReadsConfiguredBinary(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "agent.plist")
	plist := RenderLaunchAgentPlist(ServiceOptions{
		RepoRoot:   "/tmp/sirsi-pantheon",
		BinaryPath: "/tmp/with spaces/sirsi",
		Label:      "com.sirsi.router.test",
		LogPath:    "/tmp/out.log",
		ErrPath:    "/tmp/err.log",
		PathEnv:    "/usr/bin:/bin",
	})
	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LaunchAgentProgram(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/with spaces/sirsi" {
		t.Fatalf("program = %q, want configured binary", got)
	}
}
