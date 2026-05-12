// Package output — suggestions.go
//
// Command completion support for the power-user prompt (: key).
// The tab-based TUI doesn't need inline predictions, but the
// prompt mode uses this for basic completion.
package output

import (
	"sort"
	"strings"
)

// topLevelCommands are available at the root prompt.
var topLevelCommands = []string{
	"scan", "ghosts", "dedup", "guard", "doctor", "mcp", "version", "help",
	"ra", "net", "thoth", "maat", "isis", "seshat",
	"anubis", "seba", "osiris",
}

// deduplicateHistory extracts unique command strings from history entries,
// preserving the original casing, most recent occurrence wins.
func deduplicateHistory(history []historyEntry) []string {
	seen := make(map[string]bool)
	var result []string
	for i := len(history) - 1; i >= 0; i-- {
		cmd := history[i].command
		lower := strings.ToLower(cmd)
		if cmd != "" && !seen[lower] {
			seen[lower] = true
			result = append(result, cmd)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return i > j })
	return result
}
