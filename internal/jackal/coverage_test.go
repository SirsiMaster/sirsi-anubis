package jackal

import (
	"testing"
)

func TestDefaultEngine(t *testing.T) {
	engine := DefaultEngine()
	if engine == nil {
		t.Fatal("DefaultEngine returned nil")
	}
	// DefaultEngine returns an empty engine — rules must be registered by caller.
	rules := engine.Rules()
	if len(rules) != 0 {
		t.Errorf("DefaultEngine should have 0 rules, got %d", len(rules))
	}
}
