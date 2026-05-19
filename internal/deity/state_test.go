package deity

import (
	"reflect"
	"testing"
)

func TestSaveLoadStatePreservesSessionSummary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	wantRecommendations := []string{"scan", "clean --dry-run", "diagnose"}
	if err := SaveState(PersistedState{
		DeityState: map[string]RunState{
			"anubis": StateHasData,
			"isis":   StateSucceeded,
		},
		LastCommand:         "scan",
		LastSummary:         "Completed",
		LastRecommendations: wantRecommendations,
	}); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	got, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if got.DeityState["anubis"] != StateHasData {
		t.Fatalf("anubis state = %v, want %v", got.DeityState["anubis"], StateHasData)
	}
	if got.LastCommand != "scan" {
		t.Fatalf("LastCommand = %q, want scan", got.LastCommand)
	}
	if got.LastSummary != "Completed" {
		t.Fatalf("LastSummary = %q, want Completed", got.LastSummary)
	}
	if !reflect.DeepEqual(got.LastRecommendations, wantRecommendations) {
		t.Fatalf("LastRecommendations = %#v, want %#v", got.LastRecommendations, wantRecommendations)
	}
	if got.LastUsed == "" {
		t.Fatal("LastUsed was not populated")
	}
}
