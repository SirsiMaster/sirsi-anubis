package dashboard

import (
	"fmt"
	"testing"
	"time"
)

// newTestGuard returns a guard with a controllable clock so expiry is
// deterministic (no wall-clock flakiness).
func newTestGuard(now *time.Time) *ConfirmGuard {
	g := NewConfirmGuard()
	g.now = func() time.Time { return *now }
	return g
}

func TestConfirmGuard_PrepareIsDryRun(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)

	prep, err := g.Prepare("clean", "/tmp/cache", map[string]string{"force": "true"},
		"would delete 3 files", []string{"/tmp/cache/a", "/tmp/cache/b", "/tmp/cache/c"}, "3 files, ~12 MB")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if !prep.DryRun {
		t.Error("prepare response must have DryRun=true")
	}
	if prep.ConfirmToken == "" || prep.ActionHash == "" {
		t.Error("prepare must return a token and hash")
	}
	if !prep.ExpiresAt.Equal(clock.Add(confirmTokenTTL)) {
		t.Errorf("ExpiresAt = %v, want %v", prep.ExpiresAt, clock.Add(confirmTokenTTL))
	}
	if len(prep.AffectedResources) != 3 {
		t.Errorf("AffectedResources = %d, want 3", len(prep.AffectedResources))
	}
}

func TestConfirmGuard_HappyPathCommit(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	params := map[string]string{"force": "true"}

	prep, err := g.Prepare("clean", "/tmp/cache", params, "", nil, "")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := g.Validate(prep.ConfirmToken, "clean", "/tmp/cache", params, prep.ActionHash); err != nil {
		t.Fatalf("Validate should succeed on matching commit: %v", err)
	}
}

func TestConfirmGuard_RejectMissingToken(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	if err := g.Validate("", "clean", "/tmp/cache", nil, ""); err == nil {
		t.Fatal("empty token must be rejected")
	}
}

func TestConfirmGuard_RejectUnknownToken(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	if err := g.Validate("deadbeef", "clean", "/tmp/cache", nil, ""); err == nil {
		t.Fatal("unknown token must be rejected")
	}
}

func TestConfirmGuard_RejectExpiredToken(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	prep, _ := g.Prepare("slay", "1234", nil, "", nil, "")

	clock = clock.Add(confirmTokenTTL + time.Second) // advance past expiry
	if err := g.Validate(prep.ConfirmToken, "slay", "1234", nil, prep.ActionHash); err == nil {
		t.Fatal("expired token must be rejected")
	}
}

func TestConfirmGuard_RejectMismatchedActionTarget(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	prep, _ := g.Prepare("clean", "/tmp/cache", nil, "", nil, "")

	// Same token, different target — must not authorize a different deletion.
	if err := g.Validate(prep.ConfirmToken, "clean", "/etc/passwd", nil, prep.ActionHash); err == nil {
		t.Fatal("token bound to /tmp/cache must not authorize /etc/passwd")
	}
}

func TestConfirmGuard_RejectMismatchedParams(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	prep, _ := g.Prepare("clean", "/tmp/cache", map[string]string{"force": "false"}, "", nil, "")

	// Commit with escalated params must be rejected (hash differs).
	if err := g.Validate(prep.ConfirmToken, "clean", "/tmp/cache", map[string]string{"force": "true"}, ""); err == nil {
		t.Fatal("commit with different params must be rejected")
	}
}

func TestConfirmGuard_RejectEchoedHashMismatch(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	prep, _ := g.Prepare("clean", "/tmp/cache", nil, "", nil, "")
	if err := g.Validate(prep.ConfirmToken, "clean", "/tmp/cache", nil, "tampered-hash"); err == nil {
		t.Fatal("mismatched echoed action_hash must be rejected")
	}
}

func TestConfirmGuard_RejectReusedToken(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	prep, _ := g.Prepare("vault/prune", "all", nil, "", nil, "")

	if err := g.Validate(prep.ConfirmToken, "vault/prune", "all", nil, prep.ActionHash); err != nil {
		t.Fatalf("first commit should succeed: %v", err)
	}
	// Replay the same token — single-use, must now fail.
	if err := g.Validate(prep.ConfirmToken, "vault/prune", "all", nil, prep.ActionHash); err == nil {
		t.Fatal("reused token must be rejected (single-use)")
	}
}

func TestActionHash_Deterministic(t *testing.T) {
	a := ActionHash("clean", "/tmp", map[string]string{"x": "1", "y": "2"})
	b := ActionHash("clean", "/tmp", map[string]string{"y": "2", "x": "1"}) // param order swapped
	if a != b {
		t.Error("ActionHash must be independent of param map ordering")
	}
	if c := ActionHash("clean", "/tmp", map[string]string{"x": "9", "y": "2"}); c == a {
		t.Error("ActionHash must change when a param value changes")
	}
}

func TestConfirmGuard_TokenGenFailure(t *testing.T) {
	clock := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g := newTestGuard(&clock)
	g.randFn = func([]byte) (int, error) { return 0, fmt.Errorf("entropy depleted") }
	if _, err := g.Prepare("clean", "/tmp", nil, "", nil, ""); err == nil {
		t.Fatal("Prepare must surface a token-generation failure, not proceed")
	}
}
