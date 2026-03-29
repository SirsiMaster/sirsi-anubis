package neith

import (
	"testing"
)

// ── Neith Test Suite ────────────────────────────────────────────────
// Tests for the Net (Neith) Weaver module — plan alignment & tapestry.

func TestWeave_AssessLogs(t *testing.T) {
	t.Parallel()

	w := &Weave{
		SessionID: "test-001",
		Plan:      []string{"Build pulse engine", "Tag v1.0.0-rc1"},
	}

	score, err := w.AssessLogs("## Sprint 15: Built pulse engine")
	if err != nil {
		t.Fatalf("AssessLogs() error: %v", err)
	}
	if score < 0 || score > 1.0 {
		t.Errorf("AssessLogs() score = %f, want between 0.0 and 1.0", score)
	}
}

func TestWeave_EmptyPlan(t *testing.T) {
	t.Parallel()

	w := &Weave{}
	score, err := w.AssessLogs("")
	if err != nil {
		t.Fatalf("AssessLogs(empty) error: %v", err)
	}
	if score != 1.0 {
		t.Errorf("AssessLogs(empty) = %f, want 1.0 (no plan = no drift)", score)
	}
}

func TestTapestry_Align_Consistent(t *testing.T) {
	t.Parallel()

	tap := &Tapestry{
		MaatConsistent:  true,
		AnubisCorrect:   true,
		KaExtinguished:  true,
		ThothAccurate:   true,
		SekhmetHardened: true,
	}

	err := tap.Align()
	if err != nil {
		t.Errorf("Align() should pass when Ma'at is consistent, got: %v", err)
	}
}

func TestTapestry_Align_Inconsistent(t *testing.T) {
	t.Parallel()

	tap := &Tapestry{
		MaatConsistent:  false,
		AnubisCorrect:   true,
		KaExtinguished:  true,
		ThothAccurate:   true,
		SekhmetHardened: true,
	}

	err := tap.Align()
	if err == nil {
		t.Error("Align() should fail when Ma'at is inconsistent")
	}
}

func TestTapestry_Align_EmptyTapestry(t *testing.T) {
	t.Parallel()

	tap := &Tapestry{}
	err := tap.Align()
	if err == nil {
		t.Error("Align() should fail with zero-value tapestry (MaatConsistent=false)")
	}
}

func TestWeave_FieldAccess(t *testing.T) {
	t.Parallel()

	w := &Weave{
		SessionID:    "session-42",
		Plan:         []string{"A", "B", "C"},
		Achievements: []string{"A"},
		DriftFound:   true,
	}

	if w.SessionID != "session-42" {
		t.Errorf("SessionID = %q", w.SessionID)
	}
	if len(w.Plan) != 3 {
		t.Errorf("Plan len = %d, want 3", len(w.Plan))
	}
	if len(w.Achievements) != 1 {
		t.Errorf("Achievements len = %d, want 1", len(w.Achievements))
	}
	if !w.DriftFound {
		t.Error("DriftFound should be true")
	}
}
