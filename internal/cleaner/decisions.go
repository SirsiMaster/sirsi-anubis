package cleaner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Decision records what happened to a file and why.
type Decision struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	Action     string    `json:"action"`   // trash, delete, keep, skip
	Reason     string    `json:"reason"`   // why this action was taken
	DupGroupID string    `json:"group_id"` // which duplicate group this belongs to
	SHA256     string    `json:"sha256"`   // file hash for verification
	Timestamp  time.Time `json:"timestamp"`
	Reversible bool      `json:"reversible"` // true if in trash, false if permanently deleted
}

// DecisionLog tracks all cleaning decisions for rollback.
type DecisionLog struct {
	SessionID  string     `json:"session_id"`
	StartTime  time.Time  `json:"start_time"`
	Decisions  []Decision `json:"decisions"`
	TotalFreed int64      `json:"total_freed"`
	path       string     // file path for persistence
}

// NewDecisionLog creates a new decision log for this cleaning session.
// The log is persisted to ~/.config/anubis/mirror/decisions/ so it
// survives across sessions and can be used for rollback.
func NewDecisionLog() (*DecisionLog, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(home, ".config", "anubis", "mirror", "decisions")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("create decision log dir: %w", err)
	}

	sessionID := time.Now().Format("20060102-150405")
	logPath := filepath.Join(logDir, fmt.Sprintf("session-%s.json", sessionID))

	return &DecisionLog{
		SessionID: sessionID,
		StartTime: time.Now(),
		path:      logPath,
	}, nil
}

// Record adds a decision to the log and persists it.
func (dl *DecisionLog) Record(d Decision) error {
	d.Timestamp = time.Now()
	dl.Decisions = append(dl.Decisions, d)

	if d.Action == "trash" || d.Action == "delete" {
		dl.TotalFreed += d.Size
	}

	return dl.save()
}

// save persists the decision log to disk.
func (dl *DecisionLog) save() error {
	data, err := json.MarshalIndent(dl, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dl.path, data, 0600)
}

// LoadDecisionLog reads a previous decision log for review or rollback.
func LoadDecisionLog(path string) (*DecisionLog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var log DecisionLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, err
	}
	log.path = path
	return &log, nil
}

// ListDecisionLogs returns all decision log files.
func ListDecisionLogs() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	logDir := filepath.Join(home, ".config", "anubis", "mirror", "decisions")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if !e.IsDir() {
			paths = append(paths, filepath.Join(logDir, e.Name()))
		}
	}
	return paths, nil
}
