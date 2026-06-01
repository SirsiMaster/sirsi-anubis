package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ProcessRole describes Pantheon's operational read of a process.
type ProcessRole string

const (
	ProcessRoleAgent    ProcessRole = "agent"
	ProcessRoleIDE      ProcessRole = "ide"
	ProcessRoleTerminal ProcessRole = "terminal"
	ProcessRoleSystem   ProcessRole = "system"
	ProcessRoleProcess  ProcessRole = "process"
)

// ProcessRecord is Pantheon's read-only inventory record for one visible PID.
type ProcessRecord struct {
	PID        int         `json:"pid"`
	PPID       int         `json:"ppid,omitempty"`
	Name       string      `json:"name"`
	Command    string      `json:"command"`
	User       string      `json:"user,omitempty"`
	RSS        int64       `json:"rss,omitempty"`
	VSZ        int64       `json:"vsz,omitempty"`
	CPUPercent float64     `json:"cpu_percent,omitempty"`
	Role       ProcessRole `json:"role"`
	Surface    string      `json:"surface,omitempty"`
	AgentID    string      `json:"agent_id,omitempty"`
	Repo       string      `json:"repo,omitempty"`
	Host       string      `json:"host,omitempty"`
	FirstSeen  time.Time   `json:"first_seen"`
	LastSeen   time.Time   `json:"last_seen"`
	Status     string      `json:"status"`
}

// ProcessRegistry is the durable process scout ledger.
type ProcessRegistry struct {
	GeneratedAt time.Time                 `json:"generated_at"`
	Host        string                    `json:"host"`
	Processes   map[string]*ProcessRecord `json:"processes"`
}

const processesFilename = "processes.json"

func processesPath(routerRoot string) string {
	return filepath.Join(routerRoot, processesFilename)
}

// LoadProcessRegistry reads processes.json. Missing file returns an empty registry.
func LoadProcessRegistry(routerRoot string) (*ProcessRegistry, error) {
	data, err := os.ReadFile(processesPath(routerRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return &ProcessRegistry{Processes: map[string]*ProcessRecord{}}, nil
		}
		return nil, fmt.Errorf("read processes.json: %w", err)
	}
	var reg ProcessRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse processes.json: %w", err)
	}
	if reg.Processes == nil {
		reg.Processes = map[string]*ProcessRecord{}
	}
	return &reg, nil
}

// SaveProcessRegistry writes processes.json atomically.
func SaveProcessRegistry(routerRoot string, reg *ProcessRegistry) error {
	if reg.Processes == nil {
		reg.Processes = map[string]*ProcessRecord{}
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal processes.json: %w", err)
	}
	tmp, err := os.CreateTemp(routerRoot, ".processes.json-*")
	if err != nil {
		return fmt.Errorf("create temp processes.json: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp processes.json: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp processes.json: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp processes.json: %w", err)
	}
	if err := os.Rename(tmpPath, processesPath(routerRoot)); err != nil {
		return fmt.Errorf("replace processes.json: %w", err)
	}
	return nil
}

// ReconcileProcessRegistry refreshes Pantheon's visible-PID inventory.
func ReconcileProcessRegistry(prev *ProcessRegistry, visible []ProcessRecord, host string, now time.Time) *ProcessRegistry {
	if prev == nil || prev.Processes == nil {
		prev = &ProcessRegistry{Processes: map[string]*ProcessRecord{}}
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	next := &ProcessRegistry{
		GeneratedAt: now,
		Host:        host,
		Processes:   map[string]*ProcessRecord{},
	}
	for _, p := range visible {
		if p.PID <= 0 {
			continue
		}
		key := processKey(host, p.PID)
		if old, ok := prev.Processes[key]; ok && !old.FirstSeen.IsZero() {
			p.FirstSeen = old.FirstSeen
		} else {
			p.FirstSeen = now
		}
		p.LastSeen = now
		p.Host = host
		p.Status = "visible"
		if p.Role == "" {
			p.Role = ClassifyProcessRole(p.Name, p.Command)
		}
		next.Processes[key] = &p
	}
	for key, old := range prev.Processes {
		if old == nil {
			continue
		}
		if _, ok := next.Processes[key]; ok {
			continue
		}
		cp := *old
		cp.Status = "gone"
		next.Processes[key] = &cp
	}
	return next
}

func processKey(host string, pid int) string {
	return fmt.Sprintf("%s:%d", host, pid)
}

// ClassifyProcessRole labels visible processes for operator scanning.
func ClassifyProcessRole(name, command string) ProcessRole {
	s := strings.ToLower(name + " " + command)
	switch {
	case containsAny(s, "claude", "codex", "gemini", "gemma", "qwen"):
		return ProcessRoleAgent
	case containsAny(s, "terminal.app", "iterm", "warp", "alacritty", "kitty", "wezterm", "zsh", "bash", "fish"):
		return ProcessRoleTerminal
	case containsAny(s, "visual studio code", "code helper", "cursor", "windsurf", "antigravity", "xcode", "zed"):
		return ProcessRoleIDE
	case containsAny(s, "kernel_task", "launchd", "windowserver", "sysmond", "distnoted", "cfprefsd", "runningboardd"):
		return ProcessRoleSystem
	default:
		return ProcessRoleProcess
	}
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

// SortedProcessRecords returns records sorted by role then descending RSS.
func (r *ProcessRegistry) SortedProcessRecords() []*ProcessRecord {
	if r == nil {
		return nil
	}
	out := make([]*ProcessRecord, 0, len(r.Processes))
	for _, p := range r.Processes {
		if p != nil {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Status != out[j].Status {
			return out[i].Status == "visible"
		}
		if out[i].Role != out[j].Role {
			return out[i].Role < out[j].Role
		}
		return out[i].RSS > out[j].RSS
	})
	return out
}
