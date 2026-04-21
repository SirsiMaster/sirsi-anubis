package mobile

import (
	"encoding/json"
	"testing"
)

func TestRtkDefaultConfig(t *testing.T) {
	result := RtkDefaultConfig()

	var resp Response
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got error: %s", resp.Error)
	}

	var cfg struct {
		StripANSI     bool `json:"strip_ansi"`
		Dedup         bool `json:"dedup"`
		DedupWindow   int  `json:"dedup_window"`
		MaxLines      int  `json:"max_lines"`
		MaxBytes      int  `json:"max_bytes"`
		TailLines     int  `json:"tail_lines"`
		CollapseBlank bool `json:"collapse_blank"`
	}
	if err := json.Unmarshal(resp.Data, &cfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if !cfg.StripANSI {
		t.Error("expected strip_ansi=true by default")
	}
	if !cfg.Dedup {
		t.Error("expected dedup=true by default")
	}
	if cfg.DedupWindow != 32 {
		t.Errorf("expected dedup_window=32, got %d", cfg.DedupWindow)
	}
	if cfg.TailLines != 20 {
		t.Errorf("expected tail_lines=20, got %d", cfg.TailLines)
	}
}

func TestRtkFilter_BasicText(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		config    string
		wantOK    bool
		checkFunc func(t *testing.T, data json.RawMessage)
	}{
		{
			name:   "empty input",
			input:  "",
			config: "",
			wantOK: true,
			checkFunc: func(t *testing.T, data json.RawMessage) {
				var r struct {
					Output        string  `json:"output"`
					OriginalBytes int     `json:"original_bytes"`
					FilteredBytes int     `json:"filtered_bytes"`
					Ratio         float64 `json:"ratio"`
				}
				if err := json.Unmarshal(data, &r); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if r.OriginalBytes != 0 {
					t.Errorf("expected 0 original bytes, got %d", r.OriginalBytes)
				}
			},
		},
		{
			name:   "duplicate lines collapsed",
			input:  "hello\nhello\nhello\nworld",
			config: "",
			wantOK: true,
			checkFunc: func(t *testing.T, data json.RawMessage) {
				var r struct {
					Output        string `json:"output"`
					DupsCollapsed int    `json:"dups_collapsed"`
				}
				if err := json.Unmarshal(data, &r); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if r.DupsCollapsed != 2 {
					t.Errorf("expected 2 dups collapsed, got %d", r.DupsCollapsed)
				}
			},
		},
		{
			name:   "with max_lines config override",
			input:  "a\nb\nc\nd\ne\nf",
			config: `{"max_lines": 3, "tail_lines": 0}`,
			wantOK: true,
			checkFunc: func(t *testing.T, data json.RawMessage) {
				var r struct {
					Truncated bool `json:"truncated"`
				}
				if err := json.Unmarshal(data, &r); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if !r.Truncated {
					t.Error("expected truncated=true with max_lines=3")
				}
			},
		},
		{
			name:   "invalid config JSON",
			input:  "test",
			config: "{bad json",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RtkFilter(tt.input, tt.config)

			var resp Response
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.OK != tt.wantOK {
				t.Errorf("expected ok=%v, got ok=%v (error: %s)", tt.wantOK, resp.OK, resp.Error)
			}

			if tt.checkFunc != nil && resp.OK {
				tt.checkFunc(t, resp.Data)
			}
		})
	}
}
