// Package router — workitem.go
//
// Work item status tracking for the multi-agent work queue (Router v3).
// Each dispatch has durable status: pending → dispatched → started →
// working → completed | failed | blocked.
package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WorkStatus represents the lifecycle state of a work item dispatch.
type WorkStatus string

const (
	StatusPending    WorkStatus = "pending"
	StatusDispatched WorkStatus = "dispatched"
	StatusStarted    WorkStatus = "started"
	StatusWorking    WorkStatus = "working"
	StatusCompleted  WorkStatus = "completed"
	StatusFailed     WorkStatus = "failed"
	StatusBlocked    WorkStatus = "blocked"
)

// WorkItem tracks a single piece of work addressed to an agent.
type WorkItem struct {
	// Identity
	ID    string `json:"id"`
	DocID string `json:"doc_id"`
	Topic string `json:"topic,omitempty"`
	Goal  string `json:"goal,omitempty"`

	// Routing
	TargetAgentID string `json:"target_agent_id"`
	SourceAgentID string `json:"source_agent_id,omitempty"`

	// Status
	Status    WorkStatus `json:"status"`
	Attempts  []Attempt  `json:"attempts,omitempty"`
	LastError string     `json:"last_error,omitempty"`

	// Expected writeback
	ExpectedArtifact    string `json:"expected_artifact,omitempty"` // e.g., "review"
	ExpectedStateUpdate bool   `json:"expected_state_update"`

	// Timestamps
	CreatedAt    time.Time `json:"created_at"`
	DispatchedAt time.Time `json:"dispatched_at,omitempty"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
}

// Attempt records a single dispatch attempt.
type Attempt struct {
	At       time.Time `json:"at"`
	ExitCode int       `json:"exit_code,omitempty"`
	Error    string    `json:"error,omitempty"`
	Stderr   string    `json:"stderr,omitempty"`
}

// WorkQueue manages work items in a persistent JSON file.
type WorkQueue struct {
	path  string
	Items []WorkItem `json:"items"`
}

// LoadWorkQueue reads or creates the work queue file.
func LoadWorkQueue(routerRoot string) (*WorkQueue, error) {
	path := filepath.Join(routerRoot, "work-queue.json")
	wq := &WorkQueue{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return wq, nil
		}
		return nil, fmt.Errorf("read work-queue.json: %w", err)
	}

	if err := json.Unmarshal(data, wq); err != nil {
		return nil, fmt.Errorf("parse work-queue.json: %w", err)
	}
	return wq, nil
}

// Save persists the work queue to disk.
func (wq *WorkQueue) Save() error {
	data, err := json.MarshalIndent(wq, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(wq.path, data, 0o644)
}

// AddItem creates a new pending work item.
func (wq *WorkQueue) AddItem(docID, targetAgentID, sourceAgentID, topic string) *WorkItem {
	if existing := wq.Find(fmt.Sprintf("%s:%s", targetAgentID, docID)); existing != nil {
		return existing
	}
	item := WorkItem{
		ID:                  fmt.Sprintf("%s:%s", targetAgentID, docID),
		DocID:               docID,
		TargetAgentID:       targetAgentID,
		SourceAgentID:       sourceAgentID,
		Topic:               topic,
		Status:              StatusPending,
		CreatedAt:           time.Now(),
		ExpectedStateUpdate: true,
	}
	wq.Items = append(wq.Items, item)
	return &wq.Items[len(wq.Items)-1]
}

// Find returns a work item by ID.
func (wq *WorkQueue) Find(itemID string) *WorkItem {
	for i := range wq.Items {
		if wq.Items[i].ID == itemID {
			return &wq.Items[i]
		}
	}
	return nil
}

// PendingFor returns all pending work items for the given agent ID.
func (wq *WorkQueue) PendingFor(agentID string) []WorkItem {
	var result []WorkItem
	for _, item := range wq.Items {
		if item.TargetAgentID == agentID && item.Status == StatusPending {
			result = append(result, item)
		}
	}
	return result
}

// AllPending returns all pending work items across all agents.
func (wq *WorkQueue) AllPending() []WorkItem {
	var result []WorkItem
	for _, item := range wq.Items {
		if item.Status == StatusPending {
			result = append(result, item)
		}
	}
	return result
}

// UpdateStatus transitions a work item to a new status.
func (wq *WorkQueue) UpdateStatus(itemID string, status WorkStatus, err string) bool {
	for i := range wq.Items {
		if wq.Items[i].ID == itemID {
			wq.Items[i].Status = status
			if err != "" {
				wq.Items[i].LastError = err
			}
			switch status {
			case StatusDispatched:
				wq.Items[i].DispatchedAt = time.Now()
			case StatusCompleted, StatusFailed, StatusBlocked:
				wq.Items[i].CompletedAt = time.Now()
			}
			return true
		}
	}
	return false
}

// RecordAttempt adds a dispatch attempt to a work item.
func (wq *WorkQueue) RecordAttempt(itemID string, exitCode int, errMsg, stderr string) {
	for i := range wq.Items {
		if wq.Items[i].ID == itemID {
			wq.Items[i].Attempts = append(wq.Items[i].Attempts, Attempt{
				At:       time.Now(),
				ExitCode: exitCode,
				Error:    errMsg,
				Stderr:   stderr,
			})
			return
		}
	}
}
