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
	// ThreadStatusReaped is terminal: the recorded PID was confirmed gone or
	// defunct (Z) against the live OS process table. A reaped record MUST NOT
	// be revived by a late heartbeat — the only way back is re-registration.
	ThreadStatusReaped ThreadStatus = "reaped"
	// ThreadStatusStale marks a thread whose PID is alive but whose heartbeat
	// loop has gone quiet past the stale window — live-but-silent, not dead.
	ThreadStatusStale ThreadStatus = "stale-heartbeat"
)

// IsTerminal reports whether a status is a final resting state that a heartbeat
// must never resurrect. Closed (operator/agent ended it) and Reaped (OS truth
// says the PID is gone/defunct) are both terminal; everything else is live.
func (s ThreadStatus) IsTerminal() bool {
	return s == ThreadStatusClosed || s == ThreadStatusReaped
}

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

	reg, err := LoadThreadRegistry(routerRoot)
	if err != nil {
		return nil, err
	}

	// Idempotent registration: if the caller did not pin a ThreadID but this
	// (agent_id, pid) already has a LIVE (non-terminal) record, reuse it instead
	// of minting a new thread + heartbeat loop. Without this, every register/
	// discover call for the same session spawned a duplicate record and a
	// duplicate caffeinate loop — 150+ loops for ~10 live PIDs, all waking each
	// minute. One live session → one thread.
	if t.ThreadID == "" && t.PID > 0 {
		for id, existing := range reg.Threads {
			if existing == nil {
				continue
			}
			if existing.AgentID == t.AgentID && existing.PID == t.PID && !existing.Status.IsTerminal() {
				existing.LastSeenAt = now
				if t.CurrentItem != "" {
					existing.CurrentItem = t.CurrentItem
				}
				if err := SaveThreadRegistry(routerRoot, reg); err != nil {
					return nil, err
				}
				_ = id
				return existing, nil
			}
		}
	}

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
	// reaped-is-terminal: a closed/reaped record must never be revived by a
	// late heartbeat. Refusing the write here is what stops a dead PID from
	// reappearing as `active` with a fresh last_seen_at while still carrying
	// `last_error: reaped`. Reopening requires a new registration (new ID).
	if t.Status.IsTerminal() {
		return nil, fmt.Errorf("thread %q is %s and cannot be revived by heartbeat (last_error=%q); register a new thread to resume", threadID, t.Status, t.LastError)
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

// ReapedThread records one thread the reaper retired against OS truth.
type ReapedThread struct {
	ThreadID string
	AgentID  string
	PID      int
	State    PIDState // gone | defunct
}

// ReapDeadThreads retires non-terminal threads whose recorded PID is confirmed
// dead by OS truth — gone, or defunct (zombie Z). Such records are set to
// ThreadStatusReaped with a descriptive last_error; the Heartbeat guard then
// refuses to revive them. This is what stops a dead PID from re-presenting as
// `active` after a late heartbeat.
//
// host scopes the sweep to threads on THIS machine: a thread whose Host differs
// (or is empty) is left untouched, because we cannot observe another host's
// process table. Threads without a PID are also skipped (unverifiable).
//
// Returns the reaped records (empty if none). The registry is saved only when
// at least one thread was reaped.
func ReapDeadThreads(routerRoot, host string) ([]ReapedThread, error) {
	reg, err := LoadThreadRegistry(routerRoot)
	if err != nil {
		return nil, err
	}
	var reaped []ReapedThread
	now := time.Now().UTC()
	for _, t := range reg.Threads {
		if t == nil || t.Status.IsTerminal() {
			continue
		}
		if t.PID <= 0 || (host != "" && t.Host != host) {
			continue // unverifiable PID or a different host's process table
		}
		state := PIDStateOf(t.PID)
		if !DeadByOSTruth(state) {
			continue
		}
		t.Status = ThreadStatusReaped
		t.LastSeenAt = now
		t.LastError = fmt.Sprintf("reaped: PID %d %s per OS truth at %s", t.PID, state, now.Format(time.RFC3339))
		reaped = append(reaped, ReapedThread{ThreadID: t.ThreadID, AgentID: t.AgentID, PID: t.PID, State: state})
	}
	if len(reaped) > 0 {
		if err := SaveThreadRegistry(routerRoot, reg); err != nil {
			return reaped, err
		}
	}
	return reaped, nil
}

// IsStale reports whether a thread should be considered stale given now and
// the configured stale-after window. Closed threads are not stale.
func (t *Thread) IsStale(now time.Time, staleAfter time.Duration) bool {
	if t == nil || t.Status.IsTerminal() {
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

// PruneClosed removes terminal threads (closed or reaped) older than maxAge.
// Returns the count removed.
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
		if t.Status.IsTerminal() && now.Sub(t.LastSeenAt) > maxAge {
			delete(r.Threads, id)
			removed++
		}
	}
	return removed
}
