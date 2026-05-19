package router

import (
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
