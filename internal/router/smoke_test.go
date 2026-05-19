package router

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSmokeAgentPairDryRunDoesNotMutateRouter(t *testing.T) {
	r, tmp := setupTestRouter(t)
	installFakeCLI(t, "claude")
	installFakeCLI(t, "codex")

	before, err := os.ReadFile(filepath.Join(tmp, ".agents", "idea-router", "state.json"))
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	results, err := RunSmoke(context.Background(), SmokeOptions{
		RepoRoot:  tmp,
		DryRun:    true,
		AgentPair: true,
		Out:       &out,
	})
	if err != nil {
		t.Fatalf("RunSmoke() error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("results len = %d, want 3", len(results))
	}
	for _, result := range results {
		if !result.Passed {
			t.Fatalf("%s failed: %s", result.Agent, result.Detail)
		}
		if !strings.Contains(result.Detail, "dry-run") {
			t.Fatalf("%s detail missing dry-run marker: %s", result.Agent, result.Detail)
		}
	}

	after, err := os.ReadFile(filepath.Join(tmp, ".agents", "idea-router", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("dry-run mutated state.json\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if _, err := os.Stat(filepath.Join(r.root, "smoke-test")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create smoke-test dir, stat err: %v", err)
	}
}

func TestRunSmokeDryRunDoesNotCreateSmokeDir(t *testing.T) {
	r, tmp := setupTestRouter(t)
	installFakeCLI(t, "claude")
	installFakeCLI(t, "codex")

	results, err := RunSmoke(context.Background(), SmokeOptions{
		RepoRoot: tmp,
		DryRun:   true,
		Out:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("RunSmoke() error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results len = %d, want 2", len(results))
	}
	for _, result := range results {
		if !result.Passed {
			t.Fatalf("%s failed: %s", result.Agent, result.Detail)
		}
	}
	if _, err := os.Stat(filepath.Join(r.root, "smoke-test")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create smoke-test dir, stat err: %v", err)
	}
}

func installFakeCLI(t *testing.T, name string) {
	t.Helper()
	binDir := t.TempDir()
	path := filepath.Join(binDir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
