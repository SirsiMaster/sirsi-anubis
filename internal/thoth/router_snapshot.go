package thoth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func appendRouterSnapshot(repoRoot, summary string) string {
	snapshot, ok := buildRouterSnapshot(repoRoot)
	if !ok {
		return summary
	}
	if strings.TrimSpace(summary) == "" {
		return snapshot
	}
	return strings.TrimSpace(summary) + "\n\n" + snapshot
}

func buildRouterSnapshot(repoRoot string) (string, bool) {
	routerDir := filepath.Join(repoRoot, ".agents", "idea-router")
	statePath := filepath.Join(routerDir, "state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return "", false
	}

	var raw struct {
		ActiveTopics     []string            `json:"active_topics"`
		CompletedTopics  []string            `json:"completed_topics"`
		Pending          map[string][]string `json:"pending"`
		PendingForCodex  []string            `json:"pending_for_codex"`
		PendingForClaude []string            `json:"pending_for_claude"`
		LastCodexRead    string              `json:"last_codex_read"`
		LastClaudeRead   string              `json:"last_claude_read"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", false
	}

	pending := make(map[string][]string)
	for agent, ids := range raw.Pending {
		if len(ids) > 0 {
			pending[agent] = append([]string(nil), ids...)
		}
	}
	if len(raw.PendingForCodex) > 0 {
		pending["codex"] = append(pending["codex"], raw.PendingForCodex...)
	}
	if len(raw.PendingForClaude) > 0 {
		pending["claude"] = append(pending["claude"], raw.PendingForClaude...)
	}

	var lines []string
	lines = append(lines, "Router snapshot:")
	if len(raw.ActiveTopics) == 0 {
		lines = append(lines, "- active topics: none")
	} else {
		lines = append(lines, "- active topics: "+strings.Join(raw.ActiveTopics, ", "))
	}
	lines = append(lines, fmt.Sprintf("- completed topics: %d", len(raw.CompletedTopics)))
	if raw.LastCodexRead != "" {
		lines = append(lines, "- last Codex read: "+raw.LastCodexRead)
	}
	if raw.LastClaudeRead != "" {
		lines = append(lines, "- last Claude read: "+raw.LastClaudeRead)
	}

	if len(pending) == 0 {
		lines = append(lines, "- pending: none")
	} else {
		agents := make([]string, 0, len(pending))
		for agent := range pending {
			agents = append(agents, agent)
		}
		sort.Strings(agents)
		lines = append(lines, "- pending:")
		for _, agent := range agents {
			ids := append([]string(nil), pending[agent]...)
			sort.Strings(ids)
			lines = append(lines, fmt.Sprintf("  - %s: %s", agent, strings.Join(ids, ", ")))
		}
	}

	if ledgerInfo, ok := dispatchLedgerSummary(routerDir); ok {
		lines = append(lines, "- dispatch ledger: "+ledgerInfo)
	}

	return strings.Join(lines, "\n"), true
}

func dispatchLedgerSummary(routerDir string) (string, bool) {
	info, err := os.Stat(filepath.Join(routerDir, "dispatch-ledger.json"))
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("%d bytes, updated %s", info.Size(), info.ModTime().Format("2006-01-02 15:04:05")), true
}
