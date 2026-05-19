package main_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// testBinary holds the path to the compiled sirsi binary, built once in TestMain.
var testBinary string

// repoRoot is the absolute path to the repository root.
var repoRoot string

func TestMain(m *testing.M) {
	// Determine the repo root (two levels up from cmd/sirsi/).
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot determine working directory: %v\n", err)
		os.Exit(1)
	}
	repoRoot = filepath.Join(wd, "..", "..")

	// Build the binary once into a temp directory.
	tmpDir, err := os.MkdirTemp("", "sirsi-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	testBinary = filepath.Join(tmpDir, "sirsi")

	buildCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	build := exec.CommandContext(buildCtx, "go", "build", "-o", testBinary, "./cmd/sirsi/")
	build.Dir = repoRoot
	build.Env = append(os.Environ(), "CGO_ENABLED=1")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to build sirsi binary:\n%s\n%v\n", out, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runSirsi executes the test binary with the given args and a timeout.
// It returns stdout, stderr, and any error (including non-zero exit).
func runSirsi(t *testing.T, timeout time.Duration, args ...string) (stdout, stderr string, err error) {
	return runSirsiWithEnv(t, timeout, nil, args...)
}

func runSirsiWithEnv(t *testing.T, timeout time.Duration, env []string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, testBinary, args...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), env...)
	// Prevent interactive prompts by closing stdin.
	cmd.Stdin = nil

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// --- Table-Driven Deity Command Tests ---

// deityTest defines a single integration test case for a CLI command.
type deityTest struct {
	name           string
	args           []string
	timeout        time.Duration
	wantExit0      bool
	outputContains []string // substrings expected in combined stdout+stderr
	skipShort      bool     // skip when -short flag is set
	skipReason     string   // reason for skip (displayed with t.Skip)
}

func TestVersion(t *testing.T) {
	t.Parallel()

	stdout, _, err := runSirsi(t, 10*time.Second, "version")
	if err != nil {
		t.Fatalf("sirsi version failed: %v", err)
	}

	combined := stdout
	if !strings.Contains(combined, "0.21.0") {
		t.Errorf("version output missing '0.21.0', got:\n%s", combined)
	}
	if !strings.Contains(combined, "Pantheon") {
		t.Errorf("version output missing 'Pantheon', got:\n%s", combined)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 10*time.Second, "--help")
	if err != nil {
		t.Fatalf("sirsi --help failed: %v", err)
	}

	combined := stdout + stderr
	if !strings.Contains(combined, "Pantheon") {
		t.Errorf("help output missing 'Pantheon', got:\n%s", combined)
	}
	if !strings.Contains(combined, "sirsi") {
		t.Errorf("help output missing 'sirsi' command references, got:\n%s", combined)
	}
}

func TestAnubisWeigh(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping scan in short mode (may take several seconds)")
	}

	stdout, stderr, err := runSirsi(t, 60*time.Second, "anubis", "weigh", "--json")
	if err != nil {
		t.Fatalf("sirsi anubis weigh --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// JSON mode outputs structured data to stdout.
	if len(stdout) == 0 {
		t.Error("expected non-empty JSON output from anubis weigh")
	}
}

func TestAnubisWeighTerminal(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping scan in short mode")
	}

	stdout, stderr, err := runSirsi(t, 60*time.Second, "scan")
	if err != nil {
		t.Fatalf("sirsi scan failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	combined := stdout + stderr
	// Terminal mode should contain either "Waste Found" in dashboard or "Completed in" in footer.
	if !strings.Contains(combined, "Waste Found") && !strings.Contains(combined, "Completed in") {
		t.Errorf("scan output missing expected patterns, got:\n%s", combined)
	}
}

func TestAnubisKa(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping ghost scan in short mode")
	}

	stdout, stderr, err := runSirsi(t, 30*time.Second, "ghosts")
	if err != nil {
		t.Fatalf("sirsi ghosts failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	combined := stdout + stderr
	// Ghost scan should produce dashboard output with "Ghosts" count.
	if !strings.Contains(combined, "Ghost apps") && !strings.Contains(combined, "ghost") && !strings.Contains(combined, "Completed in") {
		t.Errorf("ghost scan output missing expected patterns, got:\n%s", combined)
	}
}

func TestDoctor(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 30*time.Second, "doctor", "--json")
	if err != nil {
		t.Fatalf("sirsi doctor --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if len(stdout) == 0 {
		t.Error("expected non-empty JSON output from doctor")
	}

	// Also test terminal mode for Health Score.
	stdout2, stderr2, err := runSirsi(t, 30*time.Second, "doctor")
	if err != nil {
		t.Fatalf("sirsi doctor failed: %v\nstdout: %s\nstderr: %s", err, stdout2, stderr2)
	}

	combined := stdout2 + stderr2
	if !strings.Contains(combined, "Health Score") && !strings.Contains(combined, "Completed in") {
		t.Errorf("doctor output missing 'Health Score' or 'Completed in', got:\n%s", combined)
	}
}

func TestIsisNetwork(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 30*time.Second, "isis", "network", "--json")
	if err != nil {
		t.Fatalf("sirsi isis network --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if len(stdout) == 0 {
		t.Error("expected non-empty JSON output from isis network")
	}

	// Also test terminal mode for Security Score.
	stdout2, stderr2, err := runSirsi(t, 30*time.Second, "network")
	if err != nil {
		t.Fatalf("sirsi network failed: %v\nstdout: %s\nstderr: %s", err, stdout2, stderr2)
	}

	combined := stdout2 + stderr2
	if !strings.Contains(combined, "Security Score") && !strings.Contains(combined, "Completed in") {
		t.Errorf("network output missing 'Security Score' or 'Completed in', got:\n%s", combined)
	}
}

func TestSebaHardware(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 15*time.Second, "hardware", "--json")
	if err != nil {
		t.Fatalf("sirsi hardware --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if len(stdout) == 0 {
		t.Error("expected non-empty JSON output from hardware")
	}

	// Terminal mode should show hardware details.
	stdout2, stderr2, err := runSirsi(t, 15*time.Second, "hardware")
	if err != nil {
		t.Fatalf("sirsi hardware failed: %v\nstdout: %s\nstderr: %s", err, stdout2, stderr2)
	}

	combined := stdout2 + stderr2
	if !strings.Contains(combined, "CPU") && !strings.Contains(combined, "SEBA") {
		t.Errorf("hardware output missing expected content, got:\n%s", combined)
	}
}

func TestHorusStats(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 30*time.Second, "horus", "scan", ".")
	if err != nil {
		t.Fatalf("sirsi horus scan failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	combined := stdout + stderr
	if !strings.Contains(combined, "Files") {
		t.Errorf("horus scan output missing 'Files', got:\n%s", combined)
	}
}

func TestOsirisStatus(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 15*time.Second, "osiris", "status", "--json")
	if err != nil {
		t.Fatalf("sirsi osiris status --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if len(stdout) == 0 {
		t.Error("expected non-empty JSON output from osiris status")
	}
}

func TestOsirisAssess(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 15*time.Second, "osiris", "assess", "--json")
	if err != nil {
		t.Fatalf("sirsi osiris assess --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if len(stdout) == 0 {
		t.Error("expected non-empty JSON output from osiris assess")
	}

	// Terminal mode should show risk information.
	stdout2, stderr2, err := runSirsi(t, 15*time.Second, "osiris", "assess")
	if err != nil {
		t.Fatalf("sirsi osiris assess failed: %v\nstdout: %s\nstderr: %s", err, stdout2, stderr2)
	}

	combined := stdout2 + stderr2
	if !strings.Contains(combined, "Risk") && !strings.Contains(combined, "OSIRIS") {
		t.Errorf("osiris assess output missing expected content, got:\n%s", combined)
	}
}

func TestMaatPulse(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping maat pulse in short mode (runs go test internally)")
	}

	stdout, stderr, err := runSirsi(t, 5*time.Minute, "maat", "pulse", "--skip-test", "--json")
	if err != nil {
		t.Fatalf("sirsi maat pulse --skip-test --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if len(stdout) == 0 {
		t.Error("expected non-empty JSON output from maat pulse")
	}
}

// TestDeityCommands is the master table-driven test that covers all deity
// commands with exit code and output pattern verification.
func TestDeityCommands(t *testing.T) {
	tests := []deityTest{
		{
			name:           "version",
			args:           []string{"version"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"0.21.0", "Pantheon"},
		},
		{
			name:           "help",
			args:           []string{"--help"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Pantheon", "sirsi"},
		},
		{
			name:           "anubis_help",
			args:           []string{"anubis", "--help"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Anubis"},
		},
		{
			name:           "maat_help",
			args:           []string{"maat", "--help"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Ma'at"},
		},
		{
			name:           "seba_help",
			args:           []string{"seba", "--help"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Seba"},
		},
		{
			name:           "osiris_help",
			args:           []string{"osiris", "--help"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Osiris"},
		},
		{
			name:           "isis_help",
			args:           []string{"isis", "--help"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Isis"},
		},
		{
			name:           "horus_help",
			args:           []string{"horus", "--help"},
			timeout:        10 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Horus"},
		},
		{
			name:           "doctor_json",
			args:           []string{"doctor", "--json"},
			timeout:        30 * time.Second,
			wantExit0:      true,
			outputContains: []string{"{"},
		},
		{
			name:           "hardware_json",
			args:           []string{"hardware", "--json"},
			timeout:        15 * time.Second,
			wantExit0:      true,
			outputContains: []string{"{"},
		},
		{
			name:           "network_json",
			args:           []string{"network", "--json"},
			timeout:        30 * time.Second,
			wantExit0:      true,
			outputContains: []string{"{"},
		},
		{
			name:           "osiris_status_json",
			args:           []string{"osiris", "status", "--json"},
			timeout:        15 * time.Second,
			wantExit0:      true,
			outputContains: []string{"{"},
		},
		{
			name:           "osiris_assess_json",
			args:           []string{"osiris", "assess", "--json"},
			timeout:        15 * time.Second,
			wantExit0:      true,
			outputContains: []string{"{"},
		},
		{
			name:           "horus_scan",
			args:           []string{"horus", "scan", "."},
			timeout:        30 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Files"},
		},
		{
			name:           "scan_json",
			args:           []string{"scan", "--json"},
			timeout:        60 * time.Second,
			wantExit0:      true,
			outputContains: []string{"{"},
			skipShort:      true,
			skipReason:     "full scan may take several seconds",
		},
		{
			name:           "ghosts",
			args:           []string{"ghosts"},
			timeout:        30 * time.Second,
			wantExit0:      true,
			outputContains: []string{"Ghost apps"},
			skipShort:      true,
			skipReason:     "ghost scan may take several seconds",
		},
		{
			name:           "maat_pulse_skip_test",
			args:           []string{"maat", "pulse", "--skip-test", "--json"},
			timeout:        5 * time.Minute,
			wantExit0:      true,
			outputContains: []string{"{"},
			skipShort:      true,
			skipReason:     "pulse runs real measurements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.skipShort && testing.Short() {
				t.Skip(tt.skipReason)
			}

			stdout, stderr, err := runSirsi(t, tt.timeout, tt.args...)
			combined := stdout + stderr

			if tt.wantExit0 && err != nil {
				t.Fatalf("command %v failed (wanted exit 0): %v\noutput:\n%s", tt.args, err, combined)
			}

			for _, want := range tt.outputContains {
				if !strings.Contains(combined, want) {
					t.Errorf("output missing %q for command %v\noutput:\n%s", want, tt.args, combined)
				}
			}
		})
	}
}

// TestNextStepsPresent is a table-driven test verifying that commands which
// produce NextSteps suggestions include "sirsi" in their output (as a proxy
// for the suggestion containing a follow-up command).
func TestNextStepsPresent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping NextSteps verification in short mode")
	}

	tests := []struct {
		name string
		args []string
	}{
		{"scan_next_steps", []string{"scan"}},
		{"ghosts_next_steps", []string{"ghosts"}},
		{"doctor_next_steps", []string{"doctor"}},
		{"network_next_steps", []string{"network"}},
		{"hardware_next_steps", []string{"hardware"}},
		{"osiris_assess_next_steps", []string{"osiris", "assess"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, err := runSirsi(t, 60*time.Second, tt.args...)
			if err != nil {
				t.Fatalf("command %v failed: %v", tt.args, err)
			}

			combined := stdout + stderr
			// NextSteps suggestions reference follow-up sirsi commands.
			if !strings.Contains(combined, "sirsi") {
				t.Errorf("output for %v missing 'sirsi' (expected NextSteps suggestion)\noutput:\n%s",
					tt.args, combined)
			}
		})
	}
}

// TestBinaryExists verifies the test binary was built successfully.
func TestBinaryExists(t *testing.T) {
	t.Parallel()

	info, err := os.Stat(testBinary)
	if err != nil {
		t.Fatalf("test binary does not exist at %s: %v", testBinary, err)
	}
	if info.Size() == 0 {
		t.Fatal("test binary has zero size")
	}
	// Verify it is executable.
	if info.Mode()&0111 == 0 {
		t.Fatal("test binary is not executable")
	}
}

// TestUXContract_JSONClean verifies that --json commands emit only valid JSON
// to stdout with no styled UI framing (banner, header, progress text).
// This directly addresses Codex review blocking finding #3.
func TestUXContract_JSONClean(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping UX contract tests in short mode")
	}

	tests := []struct {
		name string
		args []string
	}{
		{"audit_json", []string{"maat", "audit", "--skip-test", "--json"}},
		{"risk_json", []string{"risk", "--json"}},
		{"status_json", []string{"status", "--json"}},
		{"network_json", []string{"network", "--json"}},
		{"diagnose_json", []string{"diagnose", "--json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout, _, err := runSirsi(t, 2*time.Minute, tt.args...)
			if err != nil {
				t.Fatalf("command %v failed: %v", tt.args, err)
			}

			// stdout must start with '{' or '[' (valid JSON)
			trimmed := strings.TrimSpace(stdout)
			if len(trimmed) == 0 {
				t.Fatalf("command %v produced empty stdout", tt.args)
			}
			if trimmed[0] != '{' && trimmed[0] != '[' {
				t.Errorf("command %v stdout is not clean JSON — starts with %q\nfirst 200 chars:\n%s",
					tt.args, string(trimmed[0]), trimmed[:min(200, len(trimmed))])
			}

			// stdout must NOT contain ANSI escape codes or banner text
			if strings.Contains(stdout, "P A N T H E O N") {
				t.Errorf("command %v stdout contains banner text (should be JSON only)", tt.args)
			}
			if strings.Contains(stdout, "\033[") {
				t.Errorf("command %v stdout contains ANSI escape codes", tt.args)
			}
		})
	}
}

// TestUXContract_WhatsNext verifies that normal-mode commands emit
// "What's Next" section with suggested follow-up commands.
// This directly addresses Codex review blocking finding #4.
func TestUXContract_WhatsNext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping UX contract tests in short mode")
	}

	tests := []struct {
		name    string
		args    []string
		timeout time.Duration
	}{
		{"scan", []string{"scan"}, 3 * time.Minute},
		{"ghosts", []string{"ghosts"}, 3 * time.Minute},
		{"diagnose", []string{"diagnose"}, 60 * time.Second},
		{"network", []string{"network"}, 60 * time.Second},
		{"risk", []string{"risk"}, 30 * time.Second},
		{"status", []string{"status"}, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			homeDir := t.TempDir()
			env := []string{
				"HOME=" + homeDir,
				"XDG_CONFIG_HOME=" + filepath.Join(homeDir, ".config"),
				"XDG_CACHE_HOME=" + filepath.Join(homeDir, ".cache"),
			}

			stdout, stderr, err := runSirsiWithEnv(t, tt.timeout, env, tt.args...)
			if err != nil {
				t.Fatalf("command %v failed: %v", tt.args, err)
			}

			combined := stdout + stderr
			if !strings.Contains(combined, "What's Next") {
				t.Errorf("command %v missing 'What's Next' section\noutput:\n%s",
					tt.args, combined[:min(500, len(combined))])
			}
		})
	}
}

// TestUXContract_NoDeityVocab verifies that normal-mode output does not
// expose internal deity/module names to users.
// This directly addresses Codex review requirement: deity vocabulary hidden.
func TestUXContract_NoDeityVocab(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping UX contract tests in short mode")
	}

	// These deity names should NOT appear in user-facing output (normal mode)
	deityNames := []string{"𓆄", "𓁹", "𓁐", "Anubis", "Osiris", "Isis", "Jackal", "Scarab"}

	tests := []struct {
		name string
		args []string
	}{
		{"risk", []string{"risk"}},
		{"status", []string{"status"}},
		{"diagnose", []string{"diagnose"}},
		{"network", []string{"network"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout, stderr, err := runSirsi(t, 60*time.Second, tt.args...)
			if err != nil {
				t.Fatalf("command %v failed: %v", tt.args, err)
			}

			combined := stdout + stderr
			for _, deity := range deityNames {
				if strings.Contains(combined, deity) {
					t.Errorf("command %v exposes deity vocabulary %q in user-facing output",
						tt.args, deity)
				}
			}
		})
	}
}

// TestUXContract_StatusCLI verifies the new status non-interactive mode.
// This directly addresses Codex review blocking finding #2.
func TestUXContract_StatusCLI(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := runSirsi(t, 30*time.Second, "status")
	if err != nil {
		t.Fatalf("sirsi status failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	combined := stdout + stderr
	// Must show health score
	if !strings.Contains(combined, "health") && !strings.Contains(combined, "Health") {
		t.Errorf("status output missing health info\noutput:\n%s", combined)
	}
	// Must show next actions
	if !strings.Contains(combined, "What's Next") {
		t.Errorf("status output missing 'What's Next' section\noutput:\n%s", combined)
	}
	// Must suggest --live for TUI
	if !strings.Contains(combined, "--live") {
		t.Errorf("status output missing '--live' suggestion\noutput:\n%s", combined)
	}
}

// TestSubcommandHelp verifies every registered subcommand's --help exits 0.
func TestSubcommandHelp(t *testing.T) {
	t.Parallel()

	subcommands := []string{
		"scan", "ghosts", "dedup", "guard", "doctor", "judge", "clean",
		"network", "hardware", "quality", "diagram",
		"anubis", "seba", "osiris", "isis", "maat",
		"thoth", "seshat", "horus", "rtk", "vault",
		"version", "mcp",
	}

	for _, sub := range subcommands {
		t.Run(sub+"_help", func(t *testing.T) {
			t.Parallel()

			_, _, err := runSirsi(t, 10*time.Second, sub, "--help")
			if err != nil {
				t.Errorf("sirsi %s --help failed: %v", sub, err)
			}
		})
	}
}
