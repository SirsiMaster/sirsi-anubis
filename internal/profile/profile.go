// Package profile manages developer profiles and user configuration.
// Profiles define which scan categories are active and custom settings.
package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile defines a named scan configuration.
type Profile struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	Categories   []string          `yaml:"categories"`
	MinAgeDays   int               `yaml:"min_age_days,omitempty"`
	ExcludeRules []string          `yaml:"exclude_rules,omitempty"`
	Settings     map[string]string `yaml:"settings,omitempty"`
}

// Config represents the user's global Anubis configuration.
type Config struct {
	ActiveProfile string            `yaml:"active_profile"`
	DryRunDefault bool              `yaml:"dry_run_default"`
	JSONOutput    bool              `yaml:"json_output"`
	UseTrash      bool              `yaml:"use_trash"`
	Settings      map[string]string `yaml:"settings,omitempty"`
}

// DefaultProfiles returns all built-in profiles.
func DefaultProfiles() []Profile {
	return []Profile{
		{
			Name:        "general",
			Description: "General cleanup — system caches, logs, crash reports, downloads junk",
			Categories:  []string{"general"},
			MinAgeDays:  7,
		},
		{
			Name:        "developer",
			Description: "Full developer workstation — frameworks, IDEs, cloud, and general cleanup",
			Categories:  []string{"general", "dev", "ides", "cloud"},
			MinAgeDays:  14,
		},
		{
			Name:        "ai-engineer",
			Description: "AI/ML workstation — model caches, framework build outputs, GPU artifacts",
			Categories:  []string{"general", "dev", "ai", "ides"},
			MinAgeDays:  7,
		},
		{
			Name:        "devops",
			Description: "DevOps/infrastructure — containers, cloud CLIs, Kubernetes, Terraform",
			Categories:  []string{"general", "dev", "cloud", "storage"},
			MinAgeDays:  14,
		},
	}
}

// DefaultConfig returns the default global configuration.
func DefaultConfig() Config {
	return Config{
		ActiveProfile: "developer",
		DryRunDefault: true,
		JSONOutput:    false,
		UseTrash:      true,
	}
}

// ConfigDir returns the path to ~/.config/anubis/
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "anubis")
}

// LoadConfig loads the global config from ~/.config/anubis/config.yaml
func LoadConfig() (*Config, error) {
	path := filepath.Join(ConfigDir(), "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes global config to disk.
func SaveConfig(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path := filepath.Join(dir, "config.yaml")
	return os.WriteFile(path, data, 0644)
}

// LoadProfile loads a named profile from ~/.config/anubis/profiles/<name>.yaml
func LoadProfile(name string) (*Profile, error) {
	// Check built-in profiles first
	for _, p := range DefaultProfiles() {
		if p.Name == name {
			return &p, nil
		}
	}

	// Check user profiles
	path := filepath.Join(ConfigDir(), "profiles", name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profile %q not found (checked built-in and %s)", name, path)
	}

	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profile %q: %w", name, err)
	}
	return &p, nil
}

// SaveProfile writes a profile to the user profiles directory.
func SaveProfile(p *Profile) error {
	dir := filepath.Join(ConfigDir(), "profiles")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create profiles dir: %w", err)
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}

	path := filepath.Join(dir, p.Name+".yaml")
	return os.WriteFile(path, data, 0644)
}

// ListProfiles returns all available profiles (built-in + user).
func ListProfiles() ([]Profile, error) {
	profiles := DefaultProfiles()

	// Scan user profile directory
	dir := filepath.Join(ConfigDir(), "profiles")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return profiles, nil // No user profiles yet
		}
		return nil, err
	}

	builtInNames := make(map[string]bool)
	for _, p := range profiles {
		builtInNames[p.Name] = true
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isYAMLFile(name) {
			continue
		}
		profileName := name[:len(name)-5] // Strip .yaml
		if builtInNames[profileName] {
			continue // Skip — built-in takes precedence
		}

		p, err := LoadProfile(profileName)
		if err != nil {
			continue
		}
		profiles = append(profiles, *p)
	}

	return profiles, nil
}

func isYAMLFile(name string) bool {
	return filepath.Ext(name) == ".yaml" || filepath.Ext(name) == ".yml"
}
