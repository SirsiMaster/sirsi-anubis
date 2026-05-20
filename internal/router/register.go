package router

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// RegisterAgent adds or updates one agent in agents.json.
func RegisterAgent(routerRoot string, cfg AgentConfig) error {
	if cfg.ID == "" {
		return fmt.Errorf("agent ID is required")
	}
	reg, err := LoadRegistry(routerRoot)
	if err != nil {
		return err
	}
	if reg.Agents == nil {
		reg.Agents = make(map[string]AgentConfig)
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	reg.Agents[cfg.ID] = cfg
	return SaveRegistry(routerRoot, reg)
}

// DetectInstalledAgents registers installed local AI CLIs without overwriting
// existing custom entries. It returns the IDs added.
func DetectInstalledAgents(repoRoot string) ([]string, error) {
	routerRoot := filepath.Join(repoRoot, ".agents", "idea-router")
	reg, err := LoadRegistry(routerRoot)
	if err != nil {
		return nil, err
	}
	if reg.Agents == nil {
		reg.Agents = make(map[string]AgentConfig)
	}

	candidates := []struct {
		id      string
		typ     string
		command []string
	}{
		{"claude-pantheon", "claude", []string{"claude", "--print", "--permission-mode", "auto"}},
		{"codex-pantheon", "codex", []string{"codex", "exec", "-C", repoRoot, "--sandbox", "workspace-write"}},
		{"gemini-pantheon", "gemini", []string{"gemini", "--prompt"}},
		{"qwen-pantheon", "qwen", []string{"qwen", "--prompt"}},
		{"kimi-pantheon", "kimi", []string{"kimi", "--prompt"}},
		{"gemma-pantheon", "gemma", []string{"gemma", "--prompt"}},
	}

	var added []string
	for _, c := range candidates {
		if _, exists := reg.Agents[c.id]; exists {
			continue
		}
		if _, err := exec.LookPath(c.command[0]); err != nil {
			continue
		}
		reg.Agents[c.id] = AgentConfig{
			ID:         c.id,
			Type:       c.typ,
			Command:    c.command,
			Cwd:        repoRoot,
			Workstream: "pantheon",
			Wake: WakeConfig{
				Mechanism:   WakeCLISpawn,
				HealthCheck: []string{c.command[0], "--version"},
			},
		}
		added = append(added, c.id)
	}
	if len(added) == 0 {
		return nil, nil
	}
	if err := SaveRegistry(routerRoot, reg); err != nil {
		return nil, err
	}
	return added, nil
}
