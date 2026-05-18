// Package router — registry.go
//
// Agent registry for the multi-agent work queue (Router v3).
// Agents are registered in .agents/idea-router/agents.json.
// Ra owns the router; Thoth preserves continuity; Ma'at validates governance.
package router

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AgentConfig defines a registered agent that can receive work.
type AgentConfig struct {
	// ID is the unique agent identifier (e.g., "claude-pantheon", "codex-pantheon")
	ID string `json:"id"`

	// Type is the agent platform (e.g., "claude", "codex", "gemini", "qwen")
	Type string `json:"type"`

	// Command is the launch command as an array (never shell strings).
	// The work prompt is appended as the final argument by the executor.
	Command []string `json:"command"`

	// Cwd is the working directory for the agent.
	Cwd string `json:"cwd"`

	// Env is optional environment variable overrides.
	Env map[string]string `json:"env,omitempty"`

	// Workstream is the default workstream this agent handles.
	Workstream string `json:"workstream,omitempty"`
}

// Registry holds all registered agent configurations.
type Registry struct {
	Agents map[string]AgentConfig `json:"agents"`
}

// LoadRegistry reads agents.json from the router directory.
func LoadRegistry(routerRoot string) (*Registry, error) {
	path := filepath.Join(routerRoot, "agents.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Agents: make(map[string]AgentConfig)}, nil
		}
		return nil, fmt.Errorf("read agents.json: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse agents.json: %w", err)
	}
	if reg.Agents == nil {
		reg.Agents = make(map[string]AgentConfig)
	}

	// Inject IDs from map keys
	for id, cfg := range reg.Agents {
		cfg.ID = id
		reg.Agents[id] = cfg
	}

	return &reg, nil
}

// SaveRegistry writes agents.json to the router directory.
func SaveRegistry(routerRoot string, reg *Registry) error {
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agents.json: %w", err)
	}
	return os.WriteFile(filepath.Join(routerRoot, "agents.json"), data, 0o644)
}

// Lookup returns the agent config for the given ID, or an error if not registered.
func (r *Registry) Lookup(agentID string) (*AgentConfig, error) {
	cfg, ok := r.Agents[agentID]
	if !ok {
		return nil, fmt.Errorf("agent %q not registered — add to .agents/idea-router/agents.json", agentID)
	}
	return &cfg, nil
}

// Validate checks that an agent config has the minimum required fields.
func (cfg *AgentConfig) Validate() error {
	if cfg.ID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if cfg.Type == "" {
		return fmt.Errorf("agent %q: type is required", cfg.ID)
	}
	if len(cfg.Command) == 0 {
		return fmt.Errorf("agent %q: command array is required (no shell strings)", cfg.ID)
	}
	if cfg.Cwd == "" {
		return fmt.Errorf("agent %q: cwd is required", cfg.ID)
	}
	return nil
}

// ValidateAll checks every registered agent.
func (r *Registry) ValidateAll() []error {
	var errs []error
	for _, cfg := range r.Agents {
		if err := cfg.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// IsRegistered returns true if the agent ID exists in the registry.
func (r *Registry) IsRegistered(agentID string) bool {
	_, ok := r.Agents[agentID]
	return ok
}
