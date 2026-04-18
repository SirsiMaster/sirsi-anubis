package workstream

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/platform"
)

const inventoryVersion = 1

// ToolStatus represents a detected tool on the system.
type ToolStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"` // "ai" or "ide"
	Installed bool   `json:"installed"`
}

// GitRepo represents a discovered git repository.
type GitRepo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Inventory is the cached system snapshot.
type Inventory struct {
	Version   int          `json:"version"`
	ScannedAt time.Time    `json:"scanned_at"`
	OS        string       `json:"os"`
	Arch      string       `json:"arch"`
	Shell     string       `json:"shell"`
	HomeDir   string       `json:"home_dir"`
	Tools     []ToolStatus `json:"tools"`
	GitRepos  []GitRepo    `json:"git_repos"`
}

// InstalledAI returns tools where Kind=="ai" and Installed==true.
func (inv *Inventory) InstalledAI() []ToolStatus {
	return inv.filterTools("ai", true)
}

// InstalledIDEs returns tools where Kind=="ide" and Installed==true.
func (inv *Inventory) InstalledIDEs() []ToolStatus {
	return inv.filterTools("ide", true)
}

func (inv *Inventory) filterTools(kind string, installed bool) []ToolStatus {
	var out []ToolStatus
	for _, t := range inv.Tools {
		if t.Kind == kind && t.Installed == installed {
			out = append(out, t)
		}
	}
	return out
}

// IsStale returns true if the inventory is older than 7 days.
func (inv *Inventory) IsStale() bool {
	return time.Since(inv.ScannedAt) > 7*24*time.Hour
}

// Age returns how long since the inventory was scanned.
func (inv *Inventory) Age() time.Duration {
	return time.Since(inv.ScannedAt)
}

// InventoryPath returns the standard inventory cache location.
// It's a variable so tests can override it.
var InventoryPath = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".sirsi", "inventory.json")
}

// ScanInventory scans the system for installed tools and git repos.
func ScanInventory(p platform.Platform) *Inventory {
	home, _ := p.UserHomeDir()
	shell := p.Getenv("SHELL")
	if shell != "" {
		shell = filepath.Base(shell)
	}

	var tools []ToolStatus
	for _, l := range AllLaunchers() {
		tools = append(tools, ToolStatus{
			ID:        l.ID(),
			Name:      l.Name(),
			Kind:      l.Kind(),
			Installed: l.Installed(p),
		})
	}

	return &Inventory{
		Version:   inventoryVersion,
		ScannedAt: time.Now(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Shell:     shell,
		HomeDir:   home,
		Tools:     tools,
		GitRepos:  discoverGitRepos(home),
	}
}

// LoadInventory reads the cached inventory from disk.
// Returns os.ErrNotExist if this is a first run.
func LoadInventory() (*Inventory, error) {
	data, err := os.ReadFile(InventoryPath())
	if err != nil {
		return nil, err
	}
	var inv Inventory
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("parse inventory: %w", err)
	}
	return &inv, nil
}

// SaveInventory writes the inventory to disk atomically.
func SaveInventory(inv *Inventory) error {
	path := InventoryPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal inventory: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write inventory: %w", err)
	}
	return os.Rename(tmp, path)
}

// discoverGitRepos walks common development directories up to 2 levels deep
// looking for .git directories. Caps at 100 repos.
func discoverGitRepos(homeDir string) []GitRepo {
	devDirs := []string{
		"Development", "Projects", "src", "repos",
		"code", "work", "go/src",
	}

	var repos []GitRepo
	for _, base := range devDirs {
		root := filepath.Join(homeDir, base)
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || e.Name()[0] == '.' {
				continue
			}
			candidate := filepath.Join(root, e.Name())

			// Level 1: check for .git
			if isGitRepo(candidate) {
				repos = append(repos, GitRepo{Name: e.Name(), Path: candidate})
				if len(repos) >= 100 {
					return repos
				}
				continue
			}

			// Level 2: one directory deeper
			subEntries, err := os.ReadDir(candidate)
			if err != nil {
				continue
			}
			for _, se := range subEntries {
				if !se.IsDir() || se.Name()[0] == '.' {
					continue
				}
				sub := filepath.Join(candidate, se.Name())
				if isGitRepo(sub) {
					repos = append(repos, GitRepo{Name: se.Name(), Path: sub})
					if len(repos) >= 100 {
						return repos
					}
				}
			}
		}
	}
	return repos
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
