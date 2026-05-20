// Package router — threads.go
//
// CTR thread registry. Tracks live agent threads/sessions (not just
// registered agents). Every open conversation, worker, or session that
// touches the router should register a thread, heartbeat while alive,
// and close when done. Horus reads this for the local-node live view.
//
// Schema is model-neutral: claude, codex, gemini, gemma, qwen, mcp,
// api, webhook, and future surfaces share the same shape.
package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	cryptorand "crypto/rand"
	"encoding/hex"
)

// DefaultThreadStaleAfter is the grace window before a thread without
// a recent heartbeat is considered stale.
const DefaultThreadStaleAfter = 5 * time.Minute

// ThreadStatus enumerates a thread's reported state.
type ThreadStatus string

const (
	ThreadStatusActive  ThreadStatus = "active"
	ThreadStatusIdle    ThreadStatus = "idle"
	ThreadStatusBlocked ThreadStatus = "blocked"
	ThreadStatusClosed  ThreadStatus = "closed"
)

// Thread is one live registration of an agent session.
type Thread struct {
	ThreadID      string       `json:"thread_id"`
	AgentID       string       `json:"agent_id"`
	Surface       string       `json:"surface"`
	Repo          string       `json:"repo,omitempty"`
	Workstream    string       `json:"workstream,omitempty"`
	StartedAt     time.Time    `json:"started_at"`
	LastSeenAt    time.Time    `json:"last_seen_at"`
	Status        ThreadStatus `json:"status"`
	Watches       []string     `json:"watches,omitempty"`
	WakeMechanism string       `json:"wake_mechanism,omitempty"`
	CurrentItem   string       `json:"current_item,omitempty"`
	LastError     string       `json:"last_error,omitempty"`
	PID           int          `json:"pid,omitempty"`
	Host          string       `json:"host,omitempty"`
}

// ThreadRegistry is the on-disk record of live threads.
type ThreadRegistry struct {
	Threads map[string]*Thread `json:"threads"`
}

const threadsFilename = "threads.json"

func threadsPath(routerRoot string) string {
	return filepath.Join(routerRoot, threadsFilename)
}

// LoadThreadRegistry reads threads.json. Missing file → empty registry.
func LoadThreadRegistry(routerRoot string) (*ThreadRegistry, error) {
	data, err := os.ReadFile(threadsPath(routerRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return &ThreadRegistry{Threads: map[string]*Thread{}}, nil
		}
		return nil, fmt.Errorf("read threads.json: %w", err)
	}
	var reg ThreadRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse threads.json: %w", err)
	}
	if reg.Threads == nil {
		reg.Threads = map[string]*Thread{}
	}
	for id, t := range reg.Threads {
		if t == nil {
			continue
		}
		if t.ThreadID == "" {
			t.ThreadID = id
		}
	}
	return &reg, nil
}

// SaveThreadRegistry writes threads.json atomically.
func SaveThreadRegistry(routerRoot string, reg *ThreadRegistry) error {
	if reg.Threads == nil {
		reg.Threads = map[string]*Thread{}
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal threads.json: %w", err)
	}
	tmp, err := os.CreateTemp(routerRoot, ".threads.json-*")
	if err != nil {
		return fmt.Errorf("create temp threads.json: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp threads.json: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp threads.json: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp threads.json: %w", err)
	}
	if err := os.Rename(tmpPath, threadsPath(routerRoot)); err != nil {
		return fmt.Errorf("replace threads.json: %w", err)
	}
	return nil
}

// NewThreadID returns a short opaque thread identifier.
func NewThreadID() string {
	var b [8]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		// Fallback to time-based ID if entropy is unavailable.
		return fmt.Sprintf("thr-%d", time.Now().UnixNano())
	}
	return "thr-" + hex.EncodeToString(b[:])
}

// RegisterThread upserts a thread record. If t.ThreadID is empty, a new ID
// is generated and stored on t before saving.
func RegisterThread(routerRoot string, t *Thread) (*Thread, error) {
	if t == nil {
		return nil, fmt.Errorf("thread is nil")
	}
	if t.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if t.Surface == "" {
		return nil, fmt.Errorf("surface is required")
	}
	now := time.Now().UTC()
	if t.ThreadID == "" {
		t.ThreadID = NewThreadID()
	}
	if t.StartedAt.IsZero() {
		t.StartedAt = now
	}
	t.LastSeenAt = now
	if t.Status == "" {
		t.Status = ThreadStatusActive
	}
	if len(t.Watches) == 0 {
		t.Watches = []string{t.AgentID}
	}

	reg, err := LoadThreadRegistry(routerRoot)
	if err != nil {
		return nil, err
	}
	reg.Threads[t.ThreadID] = t
	if err := SaveThreadRegistry(routerRoot, reg); err != nil {
		return nil, err
	}
	return t, nil
}

// HeartbeatThread updates LastSeenAt and optionally status/current_item/last_error.
type HeartbeatUpdate struct {
	Status      ThreadStatus
	CurrentItem *string
	LastError   *string
}

// Heartbeat updates a thread's last_seen_at and optional fields.
// Returns the updated thread, or an error if the thread is unknown.
func Heartbeat(routerRoot, threadID string, upd HeartbeatUpdate) (*Thread, error) {
	if threadID == "" {
		return nil, fmt.Errorf("thread_id is required")
	}
	reg, err := LoadThreadRegistry(routerRoot)
	if err != nil {
		return nil, err
	}
	t, ok := reg.Threads[threadID]
	if !ok {
		return nil, fmt.Errorf("thread %q not registered", threadID)
	}
	t.LastSeenAt = time.Now().UTC()
	if upd.Status != "" {
		t.Status = upd.Status
	}
	if upd.CurrentItem != nil {
		t.CurrentItem = *upd.CurrentItem
	}
	if upd.LastError != nil {
		t.LastError = *upd.LastError
	}
	if err := SaveThreadRegistry(routerRoot, reg); err != nil {
		return nil, err
	}
	return t, nil
}

// CloseThread marks a thread closed (does not delete it; callers may prune).
func CloseThread(routerRoot, threadID string) (*Thread, error) {
	reg, err := LoadThreadRegistry(routerRoot)
	if err != nil {
		return nil, err
	}
	t, ok := reg.Threads[threadID]
	if !ok {
		return nil, fmt.Errorf("thread %q not registered", threadID)
	}
	t.Status = ThreadStatusClosed
	t.LastSeenAt = time.Now().UTC()
	if err := SaveThreadRegistry(routerRoot, reg); err != nil {
		return nil, err
	}
	return t, nil
}

// IsStale reports whether a thread should be considered stale given now and
// the configured stale-after window. Closed threads are not stale.
func (t *Thread) IsStale(now time.Time, staleAfter time.Duration) bool {
	if t == nil || t.Status == ThreadStatusClosed {
		return false
	}
	if staleAfter <= 0 {
		staleAfter = DefaultThreadStaleAfter
	}
	return now.Sub(t.LastSeenAt) > staleAfter
}

// SortedThreads returns thread records sorted by LastSeenAt descending.
func (r *ThreadRegistry) SortedThreads() []*Thread {
	out := make([]*Thread, 0, len(r.Threads))
	for _, t := range r.Threads {
		if t == nil {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastSeenAt.After(out[j].LastSeenAt)
	})
	return out
}

// PruneClosed removes closed threads older than maxAge. Returns the count removed.
func (r *ThreadRegistry) PruneClosed(now time.Time, maxAge time.Duration) int {
	if maxAge <= 0 {
		return 0
	}
	removed := 0
	for id, t := range r.Threads {
		if t == nil {
			delete(r.Threads, id)
			continue
		}
		if t.Status == ThreadStatusClosed && now.Sub(t.LastSeenAt) > maxAge {
			delete(r.Threads, id)
			removed++
		}
	}
	return removed
}
