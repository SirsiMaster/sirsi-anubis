package mobile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
)

// setupTestStele creates a temp Stele file with N entries and returns a cleanup func.
func setupTestStele(t *testing.T, count int) func() {
	t.Helper()
	tmpDir := t.TempDir()
	stelePath := filepath.Join(tmpDir, "test_stele.jsonl")

	os.Setenv("PANTHEON_STELE_PATH", stelePath)

	// Write N entries via the real Stele Ledger.
	ledger, err := stele.Open(stelePath)
	if err != nil {
		t.Fatalf("failed to open test stele: %v", err)
	}
	for i := 0; i < count; i++ {
		deities := []string{"thoth", "maat", "ra", "anubis"}
		types := []string{"commit", "governance", "tool_use", "deploy_start"}
		deity := deities[i%len(deities)]
		evType := types[i%len(types)]
		if err := ledger.Append(deity, evType, "test-repo", map[string]string{"idx": string(rune('0' + i))}); err != nil {
			t.Fatalf("failed to append entry %d: %v", i, err)
		}
	}

	return func() {
		os.Unsetenv("PANTHEON_STELE_PATH")
	}
}

func TestSteleReadRecent(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		count      int
		wantOK     bool
		wantLen    int
	}{
		{
			name:       "read last 3 of 5",
			totalItems: 5,
			count:      3,
			wantOK:     true,
			wantLen:    3,
		},
		{
			name:       "read more than available",
			totalItems: 2,
			count:      10,
			wantOK:     true,
			wantLen:    2,
		},
		{
			name:       "read exactly available",
			totalItems: 4,
			count:      4,
			wantOK:     true,
			wantLen:    4,
		},
		{
			name:       "zero count returns error",
			totalItems: 3,
			count:      0,
			wantOK:     false,
		},
		{
			name:       "negative count returns error",
			totalItems: 3,
			count:      -1,
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestStele(t, tt.totalItems)
			defer cleanup()

			result := SteleReadRecent(tt.count)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.OK != tt.wantOK {
				t.Errorf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}

			if resp.OK {
				var entries []stele.Entry
				if err := json.Unmarshal(resp.Data, &entries); err != nil {
					t.Fatalf("failed to parse entries: %v", err)
				}
				if len(entries) != tt.wantLen {
					t.Errorf("expected %d entries, got %d", tt.wantLen, len(entries))
				}
				// Verify newest-first ordering.
				if len(entries) > 1 {
					if entries[0].Seq < entries[1].Seq {
						t.Errorf("expected newest first: seq[0]=%d < seq[1]=%d", entries[0].Seq, entries[1].Seq)
					}
				}
			}
		})
	}
}

func TestSteleReadRecent_EmptyStele(t *testing.T) {
	tmpDir := t.TempDir()
	stelePath := filepath.Join(tmpDir, "empty.jsonl")
	os.WriteFile(stelePath, []byte(""), 0644)
	os.Setenv("PANTHEON_STELE_PATH", stelePath)
	defer os.Unsetenv("PANTHEON_STELE_PATH")

	result := SteleReadRecent(5)
	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true for empty stele, got error: %s", resp.Error)
	}
}

func TestSteleStats(t *testing.T) {
	tests := []struct {
		name        string
		totalItems  int
		wantTotal   int
		wantDeities int // number of distinct deities
		wantTypes   int // number of distinct types
		wantFirstTS bool
		wantLastTS  bool
	}{
		{
			name:        "stats for 8 entries",
			totalItems:  8,
			wantTotal:   8,
			wantDeities: 4, // thoth, maat, ra, anubis
			wantTypes:   4, // commit, governance, tool_use, deploy_start
			wantFirstTS: true,
			wantLastTS:  true,
		},
		{
			name:        "stats for single entry",
			totalItems:  1,
			wantTotal:   1,
			wantDeities: 1,
			wantTypes:   1,
			wantFirstTS: true,
			wantLastTS:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestStele(t, tt.totalItems)
			defer cleanup()

			result := SteleStats()

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if !resp.OK {
				t.Fatalf("stats failed: %s", resp.Error)
			}

			var stats SteleStatsResult
			if err := json.Unmarshal(resp.Data, &stats); err != nil {
				t.Fatalf("failed to parse stats: %v", err)
			}
			if stats.TotalEntries != tt.wantTotal {
				t.Errorf("expected %d total entries, got %d", tt.wantTotal, stats.TotalEntries)
			}
			if len(stats.DeityCounts) != tt.wantDeities {
				t.Errorf("expected %d deities, got %d", tt.wantDeities, len(stats.DeityCounts))
			}
			if len(stats.TypeCounts) != tt.wantTypes {
				t.Errorf("expected %d types, got %d", tt.wantTypes, len(stats.TypeCounts))
			}
			if tt.wantFirstTS && stats.FirstTS == "" {
				t.Error("expected non-empty firstTs")
			}
			if tt.wantLastTS && stats.LastTS == "" {
				t.Error("expected non-empty lastTs")
			}
		})
	}
}

func TestSteleStats_EmptyStele(t *testing.T) {
	tmpDir := t.TempDir()
	stelePath := filepath.Join(tmpDir, "empty.jsonl")
	os.WriteFile(stelePath, []byte(""), 0644)
	os.Setenv("PANTHEON_STELE_PATH", stelePath)
	defer os.Unsetenv("PANTHEON_STELE_PATH")

	result := SteleStats()
	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	var stats SteleStatsResult
	if err := json.Unmarshal(resp.Data, &stats); err != nil {
		t.Fatalf("failed to parse stats: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("expected 0 entries, got %d", stats.TotalEntries)
	}
}

func TestSteleVerify(t *testing.T) {
	tests := []struct {
		name       string
		entries    int
		wantStatus string
		wantChain  int
	}{
		{
			name:       "valid 5-entry chain",
			entries:    5,
			wantStatus: "verified",
			wantChain:  5,
		},
		{
			name:       "valid single entry",
			entries:    1,
			wantStatus: "verified",
			wantChain:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestStele(t, tt.entries)
			defer cleanup()

			result := SteleVerify()

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if !resp.OK {
				t.Fatalf("verify failed: %s", resp.Error)
			}

			var vr SteleVerifyResult
			if err := json.Unmarshal(resp.Data, &vr); err != nil {
				t.Fatalf("failed to parse verify result: %v", err)
			}
			if vr.Status != tt.wantStatus {
				t.Errorf("expected status=%q, got %q", tt.wantStatus, vr.Status)
			}
			if vr.ChainLength != tt.wantChain {
				t.Errorf("expected chainLength=%d, got %d", tt.wantChain, vr.ChainLength)
			}
			if vr.TotalCount != tt.entries {
				t.Errorf("expected totalCount=%d, got %d", tt.entries, vr.TotalCount)
			}
			if vr.VerifiedAt == "" {
				t.Error("expected non-empty verifiedAt")
			}
			if vr.Status == "verified" && len(vr.Breaks) != 0 {
				t.Errorf("expected 0 breaks for verified chain, got %d", len(vr.Breaks))
			}
		})
	}
}

func TestSteleVerify_BrokenChain(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "broken.jsonl")
	os.Setenv("PANTHEON_STELE_PATH", path)
	defer os.Unsetenv("PANTHEON_STELE_PATH")

	// Write a valid entry, then a manually corrupted one.
	ledger, err := stele.Open(path)
	if err != nil {
		t.Fatalf("failed to open stele: %v", err)
	}
	if err := ledger.Append("thoth", "commit", "repo", nil); err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	// Append a raw corrupted line (wrong prev hash).
	f, _ := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)
	f.WriteString(`{"seq":1,"prev":"0000000000000000000000000000000000000000000000000000000000000000","deity":"ra","type":"deploy_start","scope":"","data":{},"ts":"2026-01-01T00:00:00Z","hash":"badhash"}` + "\n")
	f.Close()

	result := SteleVerify()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("verify call failed: %s", resp.Error)
	}

	var vr SteleVerifyResult
	if err := json.Unmarshal(resp.Data, &vr); err != nil {
		t.Fatalf("failed to parse verify result: %v", err)
	}
	if vr.Status != "broken" {
		t.Errorf("expected status=broken, got %q", vr.Status)
	}
	if len(vr.Breaks) == 0 {
		t.Error("expected at least one break in corrupted chain")
	}
}

func TestSteleVerify_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PANTHEON_STELE_PATH", filepath.Join(tmpDir, "nonexistent.jsonl"))
	defer os.Unsetenv("PANTHEON_STELE_PATH")

	result := SteleVerify()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	// Verify returns an error for missing file since stele.Verify needs data to read.
	// The readAllEntries returns nil for missing files, but stele.Verify returns an error.
	// Either way, the response should be parseable.
}
