// Package workstream implements workstream management for the Sirsi CLI.
// Manages development contexts across multiple AI assistants and IDEs.
package workstream

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WorkstreamStatus represents the lifecycle state of a workstream.
type WorkstreamStatus string

const (
	StatusActive   WorkstreamStatus = "active"
	StatusRetired  WorkstreamStatus = "retired"
	StatusArchived WorkstreamStatus = "archived"
)

// Workstream represents a named development context.
// The struct is backwards-compatible with the existing sw script's JSON schema;
// new fields use omitempty so the sw script ignores them.
type Workstream struct {
	Name      string           `json:"name"`
	Dir       string           `json:"dir"`
	Memory    string           `json:"memory,omitempty"`
	Status    WorkstreamStatus `json:"status"`
	AI        string           `json:"ai,omitempty"`
	IDE       string           `json:"ide,omitempty"`
	Tags      []string         `json:"tags,omitempty"`
	CreatedAt string           `json:"created_at,omitempty"`
	UpdatedAt string           `json:"updated_at,omitempty"`
	LastUsed  string           `json:"last_used,omitempty"`
}

// Store manages workstream persistence in a JSON file.
type Store struct {
	path  string
	items []Workstream
}

// NewStore creates a Store backed by the given JSON file path.
// If the file does not exist, it is created with an empty array.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create config dir: %w", err)
		}
		s.items = []Workstream{}
		if err := s.save(); err != nil {
			return nil, fmt.Errorf("init config: %w", err)
		}
		return s, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &s.items); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return s, nil
}

// save writes the current items to disk atomically.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return os.Rename(tmp, s.path)
}

// All returns every workstream regardless of status.
func (s *Store) All() []Workstream {
	return s.items
}

// Active returns only workstreams with status "active".
func (s *Store) Active() []Workstream {
	var out []Workstream
	for _, ws := range s.items {
		if ws.Status == StatusActive {
			out = append(out, ws)
		}
	}
	return out
}

// GetActive returns the active workstream at 1-based display index.
func (s *Store) GetActive(displayNum int) (Workstream, int, error) {
	active := s.Active()
	if displayNum < 1 || displayNum > len(active) {
		return Workstream{}, -1, fmt.Errorf("invalid workstream number: %d (have %d active)", displayNum, len(active))
	}
	ws := active[displayNum-1]
	// Find actual index in full items slice
	for i, item := range s.items {
		if item.Name == ws.Name && item.Dir == ws.Dir {
			return ws, i, nil
		}
	}
	return Workstream{}, -1, fmt.Errorf("workstream not found in store")
}

// Add appends a new active workstream. Returns error if name already exists.
func (s *Store) Add(name, dir string) error {
	for _, ws := range s.items {
		if ws.Name == name {
			return fmt.Errorf("workstream %q already exists", name)
		}
	}
	now := time.Now().Format(time.RFC3339)
	s.items = append(s.items, Workstream{
		Name:      name,
		Dir:       dir,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	})
	return s.save()
}

// Rename changes the name of the workstream at the given array index.
func (s *Store) Rename(idx int, newName string) error {
	if idx < 0 || idx >= len(s.items) {
		return fmt.Errorf("invalid index: %d", idx)
	}
	s.items[idx].Name = newName
	s.items[idx].UpdatedAt = time.Now().Format(time.RFC3339)
	return s.save()
}

// Retire sets the workstream status to retired.
func (s *Store) Retire(idx int) error {
	if idx < 0 || idx >= len(s.items) {
		return fmt.Errorf("invalid index: %d", idx)
	}
	s.items[idx].Status = StatusRetired
	s.items[idx].UpdatedAt = time.Now().Format(time.RFC3339)
	return s.save()
}

// Activate sets the workstream status to active.
func (s *Store) Activate(idx int) error {
	if idx < 0 || idx >= len(s.items) {
		return fmt.Errorf("invalid index: %d", idx)
	}
	s.items[idx].Status = StatusActive
	s.items[idx].UpdatedAt = time.Now().Format(time.RFC3339)
	return s.save()
}

// Delete permanently removes the workstream at the given array index.
func (s *Store) Delete(idx int) error {
	if idx < 0 || idx >= len(s.items) {
		return fmt.Errorf("invalid index: %d", idx)
	}
	s.items = append(s.items[:idx], s.items[idx+1:]...)
	return s.save()
}

// TouchLastUsed updates the LastUsed timestamp for the workstream at idx.
func (s *Store) TouchLastUsed(idx int) error {
	if idx < 0 || idx >= len(s.items) {
		return fmt.Errorf("invalid index: %d", idx)
	}
	s.items[idx].LastUsed = time.Now().Format(time.RFC3339)
	return s.save()
}

// SetAI sets the default AI tool for the workstream at idx.
func (s *Store) SetAI(idx int, ai string) error {
	if idx < 0 || idx >= len(s.items) {
		return fmt.Errorf("invalid index: %d", idx)
	}
	s.items[idx].AI = ai
	s.items[idx].UpdatedAt = time.Now().Format(time.RFC3339)
	return s.save()
}

// SetIDE sets the default IDE for the workstream at idx.
func (s *Store) SetIDE(idx int, ide string) error {
	if idx < 0 || idx >= len(s.items) {
		return fmt.Errorf("invalid index: %d", idx)
	}
	s.items[idx].IDE = ide
	s.items[idx].UpdatedAt = time.Now().Format(time.RFC3339)
	return s.save()
}

// DefaultConfigPath returns the standard workstreams.json location.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".sirsi", "workstreams.json")
}

// ExpandDir replaces a leading ~ with the user's home directory.
func ExpandDir(dir string) string {
	if strings.HasPrefix(dir, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, dir[2:])
	}
	if dir == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	return dir
}

// CompressDir replaces the user's home directory with ~.
func CompressDir(dir string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(dir, home) {
		return "~" + dir[len(home):]
	}
	return dir
}
