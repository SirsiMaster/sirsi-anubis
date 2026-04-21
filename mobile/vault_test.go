package mobile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestVault(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_vault.db")

	// Reset singleton and point to temp path.
	resetVault()
	os.Setenv("PANTHEON_VAULT_PATH", dbPath)

	return func() {
		resetVault()
		os.Unsetenv("PANTHEON_VAULT_PATH")
	}
}

func TestVaultStore(t *testing.T) {
	cleanup := setupTestVault(t)
	defer cleanup()

	tests := []struct {
		name    string
		source  string
		tag     string
		content string
		tokens  int
		wantOK  bool
	}{
		{
			name:    "store basic entry",
			source:  "go build",
			tag:     "build_output",
			content: "PASS ok github.com/example 0.5s",
			tokens:  12,
			wantOK:  true,
		},
		{
			name:    "store large content",
			source:  "test_runner",
			tag:     "test_output",
			content: "line1\nline2\nline3\nline4\nline5",
			tokens:  50,
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VaultStore(tt.source, tt.tag, tt.content, tt.tokens)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.OK != tt.wantOK {
				t.Errorf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}

			if resp.OK {
				var entry struct {
					ID     int64  `json:"id"`
					Source string `json:"source"`
					Tag    string `json:"tag"`
					Tokens int    `json:"tokens"`
				}
				if err := json.Unmarshal(resp.Data, &entry); err != nil {
					t.Fatalf("failed to parse entry: %v", err)
				}
				if entry.ID <= 0 {
					t.Error("expected positive entry ID")
				}
				if entry.Source != tt.source {
					t.Errorf("expected source=%q, got %q", tt.source, entry.Source)
				}
			}
		})
	}
}

func TestVaultSearchAndGet(t *testing.T) {
	cleanup := setupTestVault(t)
	defer cleanup()

	// Store test data.
	VaultStore("compiler", "build", "error: undefined reference to main", 10)
	VaultStore("linter", "lint", "warning: unused variable in main.go", 8)

	// Search for "error".
	result := VaultSearch("error", 10)
	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse search response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("search failed: %s", resp.Error)
	}

	var searchResult struct {
		Query     string `json:"query"`
		TotalHits int    `json:"totalHits"`
		Entries   []struct {
			ID int64 `json:"id"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(resp.Data, &searchResult); err != nil {
		t.Fatalf("failed to parse search result: %v", err)
	}
	if searchResult.TotalHits == 0 {
		t.Error("expected at least one search hit")
	}

	// Get the first entry by ID.
	if len(searchResult.Entries) > 0 {
		getResult := VaultGet(searchResult.Entries[0].ID)
		var getResp Response
		if err := json.Unmarshal([]byte(getResult), &getResp); err != nil {
			t.Fatalf("failed to parse get response: %v", err)
		}
		if !getResp.OK {
			t.Fatalf("get failed: %s", getResp.Error)
		}
	}
}

func TestVaultStats(t *testing.T) {
	cleanup := setupTestVault(t)
	defer cleanup()

	// Store a couple entries.
	VaultStore("test", "tag1", "content one", 5)
	VaultStore("test", "tag2", "content two", 10)

	result := VaultStats()
	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse stats response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("stats failed: %s", resp.Error)
	}

	var stats struct {
		TotalEntries int            `json:"totalEntries"`
		TotalTokens  int64          `json:"totalTokens"`
		TagCounts    map[string]int `json:"tagCounts"`
	}
	if err := json.Unmarshal(resp.Data, &stats); err != nil {
		t.Fatalf("failed to parse stats: %v", err)
	}
	if stats.TotalEntries != 2 {
		t.Errorf("expected 2 entries, got %d", stats.TotalEntries)
	}
	if stats.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", stats.TotalTokens)
	}
}

func TestVaultPrune(t *testing.T) {
	cleanup := setupTestVault(t)
	defer cleanup()

	VaultStore("test", "old", "old content", 5)

	// Pruning with 9999 hours should remove nothing (entries are brand new).
	result := VaultPrune(9999)
	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse prune response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("prune failed: %s", resp.Error)
	}

	var pruneResult struct {
		Pruned int `json:"pruned"`
	}
	if err := json.Unmarshal(resp.Data, &pruneResult); err != nil {
		t.Fatalf("failed to parse prune result: %v", err)
	}
	if pruneResult.Pruned != 0 {
		t.Errorf("expected 0 pruned (entries are new), got %d", pruneResult.Pruned)
	}
}

func TestVaultPrune_InvalidHours(t *testing.T) {
	cleanup := setupTestVault(t)
	defer cleanup()

	result := VaultPrune(0)
	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false for olderThanHours=0")
	}
}

func TestVaultSearch_EmptyQuery(t *testing.T) {
	cleanup := setupTestVault(t)
	defer cleanup()

	result := VaultSearch("", 10)
	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false for empty query")
	}
}
