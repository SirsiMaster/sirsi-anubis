package mobile

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
)

// AnubisScan runs the Jackal scan engine and returns findings as JSON.
// optionsJSON accepts: {"categories": ["dev","ai"], "max_depth": 5}
// Returns Response JSON with ScanResult data.
func AnubisScan(rootPath string, optionsJSON string) string {
	var req struct {
		Categories []string `json:"categories"`
		MaxDepth   int      `json:"max_depth"`
	}
	if optionsJSON != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &req); err != nil {
			return errorJSON("invalid options: " + err.Error())
		}
	}

	engine := jackal.NewEngine()

	opts := jackal.ScanOptions{
		HomeDir: rootPath,
	}

	if len(req.Categories) > 0 {
		cats := make([]jackal.Category, len(req.Categories))
		for i, c := range req.Categories {
			cats[i] = jackal.Category(c)
		}
		opts.Categories = cats
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := engine.Scan(ctx, opts)
	if err != nil {
		return errorJSON("scan failed: " + err.Error())
	}

	return successJSON(result)
}

// AnubisCategories returns available scan categories as JSON.
func AnubisCategories() string {
	cats := []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	}{
		{"general", "General System"},
		{"dev", "Development Tools"},
		{"ai", "AI & ML Models"},
		{"vms", "Virtual Machines"},
		{"ides", "IDEs & Editors"},
		{"cloud", "Cloud Credentials"},
		{"storage", "Storage & Large Files"},
	}
	return successJSON(cats)
}
