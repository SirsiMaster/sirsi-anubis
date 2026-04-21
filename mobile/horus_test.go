package mobile

import (
	"encoding/json"
	"testing"
)

func TestHorusParseDir(t *testing.T) {
	tests := []struct {
		name   string
		root   string
		wantOK bool
	}{
		{
			name:   "parse internal/rtk (known Go package)",
			root:   "../internal/rtk",
			wantOK: true,
		},
		{
			name:   "empty root",
			root:   "",
			wantOK: false,
		},
		{
			// GoParser.ParseDir returns an empty graph (not an error) for
			// directories with no .go files, including nonexistent ones.
			name:   "nonexistent directory returns empty graph",
			root:   "/nonexistent/path/that/does/not/exist",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HorusParseDir(tt.root)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.OK != tt.wantOK {
				t.Errorf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}

			if resp.OK && tt.root == "../internal/rtk" {
				var graph struct {
					Root     string   `json:"root"`
					Packages []string `json:"packages"`
					Symbols  []struct {
						Name string `json:"name"`
						Kind string `json:"kind"`
					} `json:"symbols"`
					Stats struct {
						Files     int `json:"files"`
						Functions int `json:"functions"`
					} `json:"stats"`
				}
				if err := json.Unmarshal(resp.Data, &graph); err != nil {
					t.Fatalf("failed to parse graph: %v", err)
				}
				if len(graph.Symbols) == 0 {
					t.Error("expected at least one symbol")
				}
				if graph.Stats.Files == 0 {
					t.Error("expected at least one file in stats")
				}
			}
		})
	}
}

func TestHorusFileOutline(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		filePath string
		wantOK   bool
	}{
		{
			name:     "valid file outline",
			root:     "../internal/rtk",
			filePath: "../internal/rtk/rtk.go",
			wantOK:   true,
		},
		{
			name:     "missing root",
			root:     "",
			filePath: "test.go",
			wantOK:   false,
		},
		{
			name:     "missing filePath",
			root:     "../internal/rtk",
			filePath: "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HorusFileOutline(tt.root, tt.filePath)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.OK != tt.wantOK {
				t.Errorf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}

			if resp.OK {
				var r struct {
					Outline string `json:"outline"`
				}
				if err := json.Unmarshal(resp.Data, &r); err != nil {
					t.Fatalf("failed to parse outline: %v", err)
				}
				if r.Outline == "" {
					t.Error("expected non-empty outline")
				}
			}
		})
	}
}

func TestHorusContextFor(t *testing.T) {
	tests := []struct {
		name       string
		root       string
		symbolName string
		wantOK     bool
	}{
		{
			name:       "known symbol",
			root:       "../internal/rtk",
			symbolName: "DefaultConfig",
			wantOK:     true,
		},
		{
			name:       "missing root",
			root:       "",
			symbolName: "Test",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HorusContextFor(tt.root, tt.symbolName)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.OK != tt.wantOK {
				t.Errorf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}
		})
	}
}

func TestHorusMatchSymbols(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		pattern string
		wantOK  bool
	}{
		{
			name:    "wildcard match all",
			root:    "../internal/rtk",
			pattern: "*",
			wantOK:  true,
		},
		{
			name:    "specific pattern",
			root:    "../internal/rtk",
			pattern: "Filter*",
			wantOK:  true,
		},
		{
			name:    "missing root",
			root:    "",
			pattern: "*",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HorusMatchSymbols(tt.root, tt.pattern)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.OK != tt.wantOK {
				t.Errorf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}
		})
	}
}
