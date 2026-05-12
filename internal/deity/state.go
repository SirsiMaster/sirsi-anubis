package deity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// RunState tracks the last-run outcome for a deity.
type RunState int

const (
	StateNeverRun  RunState = iota
	StateSucceeded          // last run completed successfully
	StateFailed             // last run had an error
	StateHasData            // has actionable data (e.g. Anubis findings)
)

// PersistedState is the JSON-serializable deity state shared across TUI and menubar.
type PersistedState struct {
	DeityState map[string]RunState `json:"deity_state"`
	LastUsed   string              `json:"last_used"`
}

// StatePath returns the filesystem path to the shared deity state file.
func StatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pantheon", "tui-state.json")
}

// LoadState reads the shared deity state from disk.
func LoadState() (PersistedState, error) {
	data, err := os.ReadFile(StatePath())
	if err != nil {
		return PersistedState{DeityState: make(map[string]RunState)}, err
	}
	var state PersistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return PersistedState{DeityState: make(map[string]RunState)}, err
	}
	if state.DeityState == nil {
		state.DeityState = make(map[string]RunState)
	}
	return state, nil
}

// SaveState writes the shared deity state to disk.
func SaveState(state PersistedState) error {
	state.LastUsed = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(StatePath())
	_ = os.MkdirAll(dir, 0755)
	return os.WriteFile(StatePath(), data, 0644)
}
