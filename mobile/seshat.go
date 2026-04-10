package mobile

import (
	"encoding/json"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/seshat"
)

// SeshatIngest runs knowledge extraction from available sources.
// optionsJSON accepts: {"sources": ["chrome","apple_notes"], "since_days": 7}
// Returns Response JSON with ingestion summary.
func SeshatIngest(optionsJSON string) string {
	var req struct {
		Sources   []string `json:"sources"`
		SinceDays int      `json:"since_days"`
	}
	if optionsJSON != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &req); err != nil {
			return errorJSON("invalid options: " + err.Error())
		}
	}

	since := time.Now().AddDate(0, 0, -7)
	if req.SinceDays > 0 {
		since = time.Now().AddDate(0, 0, -req.SinceDays)
	}

	registry := seshat.DefaultRegistry()
	filter := seshat.DefaultFilter()
	wrapped := seshat.WrapRegistry(registry, filter)

	var results []map[string]any

	for name, src := range wrapped.Sources {
		if len(req.Sources) > 0 && !containsStr(req.Sources, name) {
			continue
		}

		items, err := src.Ingest(since)
		if err != nil {
			results = append(results, map[string]any{
				"source": name,
				"error":  err.Error(),
				"count":  0,
			})
			continue
		}

		results = append(results, map[string]any{
			"source": name,
			"count":  len(items),
		})
	}

	return successJSON(results)
}

// SeshatListSources returns available knowledge sources.
// Returns Response JSON with source descriptors.
func SeshatListSources() string {
	registry := seshat.DefaultRegistry()

	var list []map[string]string
	for name, src := range registry.Sources {
		list = append(list, map[string]string{
			"name":        name,
			"description": src.Description(),
		})
	}

	return successJSON(list)
}

// SeshatListTargets returns available export targets.
// Returns Response JSON with target descriptors.
func SeshatListTargets() string {
	registry := seshat.DefaultRegistry()

	var list []map[string]string
	for name, tgt := range registry.Targets {
		list = append(list, map[string]string{
			"name":        name,
			"description": tgt.Description(),
		})
	}

	return successJSON(list)
}

// SeshatListKnowledgeItems returns all stored knowledge items.
// Returns Response JSON with []string of item names.
func SeshatListKnowledgeItems() string {
	paths := seshat.DefaultPaths()
	items, err := seshat.ListKnowledgeItems(paths)
	if err != nil {
		return errorJSON("list failed: " + err.Error())
	}

	return successJSON(items)
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
