package mobile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
)

// stelePath returns the canonical Stele ledger path.
// Honors PANTHEON_STELE_PATH env var for testing.
func stelePath() string {
	if p := os.Getenv("PANTHEON_STELE_PATH"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sirsi", "stele", "events.jsonl")
}

// readAllEntries reads every entry from the Stele JSONL file.
func readAllEntries(path string) ([]stele.Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []stele.Entry
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e stele.Entry
		if err := json.Unmarshal([]byte(line), &e); err == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

// SteleReadRecent returns the last N entries from the Stele as a JSON array.
// Returns a Response JSON envelope with an array of Entry objects.
func SteleReadRecent(count int) string {
	if count <= 0 {
		return errorJSON("count must be positive")
	}

	entries, err := readAllEntries(stelePath())
	if err != nil {
		return errorJSON("stele read: " + err.Error())
	}

	// Return the last N entries, newest first.
	if len(entries) > count {
		entries = entries[len(entries)-count:]
	}

	// Reverse so newest is first.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return successJSON(entries)
}

// SteleStatsResult holds aggregate statistics about the Stele ledger.
type SteleStatsResult struct {
	TotalEntries int            `json:"totalEntries"`
	DeityCounts  map[string]int `json:"deityCounts"`
	TypeCounts   map[string]int `json:"typeCounts"`
	FirstTS      string         `json:"firstTs,omitempty"`
	LastTS       string         `json:"lastTs,omitempty"`
}

// SteleStats returns aggregate statistics about the Stele ledger.
// Returns a Response JSON envelope with SteleStatsResult data.
func SteleStats() string {
	entries, err := readAllEntries(stelePath())
	if err != nil {
		return errorJSON("stele stats: " + err.Error())
	}

	stats := SteleStatsResult{
		TotalEntries: len(entries),
		DeityCounts:  make(map[string]int),
		TypeCounts:   make(map[string]int),
	}

	for i, e := range entries {
		stats.DeityCounts[e.Deity]++
		stats.TypeCounts[e.Type]++
		if i == 0 {
			stats.FirstTS = e.TS
		}
		if i == len(entries)-1 {
			stats.LastTS = e.TS
		}
	}

	return successJSON(stats)
}

// SteleVerifyResult holds the result of a hash chain verification.
type SteleVerifyResult struct {
	Status      string   `json:"status"`      // "verified" or "broken"
	ChainLength int      `json:"chainLength"` // number of valid entries
	TotalCount  int      `json:"totalCount"`  // total entries in file
	Breaks      []string `json:"breaks"`      // human-readable break descriptions
	VerifiedAt  string   `json:"verifiedAt"`  // ISO-8601 timestamp of verification
}

// SteleVerify verifies the hash chain integrity of the Stele ledger.
// Returns a Response JSON envelope with SteleVerifyResult data.
func SteleVerify() string {
	path := stelePath()

	// Count total entries first.
	entries, err := readAllEntries(path)
	if err != nil {
		return errorJSON("stele verify: " + err.Error())
	}

	var valid int
	var errs []error
	func() {
		defer func() {
			if r := recover(); r != nil {
				errs = []error{fmt.Errorf("verify panic: %v", r)}
			}
		}()
		valid, errs = stele.Verify(path)
	}()

	result := SteleVerifyResult{
		ChainLength: valid,
		TotalCount:  len(entries),
		VerifiedAt:  time.Now().Format(time.RFC3339),
	}

	if len(errs) == 0 {
		result.Status = "verified"
	} else {
		result.Status = "broken"
		result.Breaks = make([]string, len(errs))
		for i, e := range errs {
			result.Breaks[i] = e.Error()
		}
	}

	return successJSON(result)
}
