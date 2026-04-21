package mobile

import (
	"encoding/json"

	"github.com/SirsiMaster/sirsi-pantheon/internal/rtk"
)

// RtkFilter applies the RTK output filter to raw tool output with optional config override.
// configJSON accepts: {"strip_ansi": true, "dedup": true, "max_lines": 200, ...}
// Returns Response JSON with FilterResult data.
func RtkFilter(rawOutput string, configJSON string) string {
	cfg := rtk.DefaultConfig()

	if configJSON != "" {
		var override struct {
			StripANSI     *bool `json:"strip_ansi"`
			Dedup         *bool `json:"dedup"`
			DedupWindow   *int  `json:"dedup_window"`
			MaxLines      *int  `json:"max_lines"`
			MaxBytes      *int  `json:"max_bytes"`
			TailLines     *int  `json:"tail_lines"`
			CollapseBlank *bool `json:"collapse_blank"`
		}
		if err := json.Unmarshal([]byte(configJSON), &override); err != nil {
			return errorJSON("invalid config: " + err.Error())
		}
		if override.StripANSI != nil {
			cfg.StripANSI = *override.StripANSI
		}
		if override.Dedup != nil {
			cfg.Dedup = *override.Dedup
		}
		if override.DedupWindow != nil {
			cfg.DedupWindow = *override.DedupWindow
		}
		if override.MaxLines != nil {
			cfg.MaxLines = *override.MaxLines
		}
		if override.MaxBytes != nil {
			cfg.MaxBytes = *override.MaxBytes
		}
		if override.TailLines != nil {
			cfg.TailLines = *override.TailLines
		}
		if override.CollapseBlank != nil {
			cfg.CollapseBlank = *override.CollapseBlank
		}
	}

	filter := rtk.New(cfg)
	result := filter.Apply(rawOutput)

	return successJSON(struct {
		Output        string  `json:"output"`
		OriginalBytes int     `json:"original_bytes"`
		FilteredBytes int     `json:"filtered_bytes"`
		LinesRemoved  int     `json:"lines_removed"`
		DupsCollapsed int     `json:"dups_collapsed"`
		Truncated     bool    `json:"truncated"`
		Ratio         float64 `json:"ratio"`
	}{
		Output:        result.Output,
		OriginalBytes: result.OriginalBytes,
		FilteredBytes: result.FilteredBytes,
		LinesRemoved:  result.LinesRemoved,
		DupsCollapsed: result.DupsCollapsed,
		Truncated:     result.Truncated,
		Ratio:         result.Ratio,
	})
}

// RtkDefaultConfig returns the default RTK filter configuration as JSON.
// Returns Response JSON with FilterConfig data.
func RtkDefaultConfig() string {
	cfg := rtk.DefaultConfig()

	return successJSON(struct {
		StripANSI     bool `json:"strip_ansi"`
		Dedup         bool `json:"dedup"`
		DedupWindow   int  `json:"dedup_window"`
		MaxLines      int  `json:"max_lines"`
		MaxBytes      int  `json:"max_bytes"`
		TailLines     int  `json:"tail_lines"`
		CollapseBlank bool `json:"collapse_blank"`
	}{
		StripANSI:     cfg.StripANSI,
		Dedup:         cfg.Dedup,
		DedupWindow:   cfg.DedupWindow,
		MaxLines:      cfg.MaxLines,
		MaxBytes:      cfg.MaxBytes,
		TailLines:     cfg.TailLines,
		CollapseBlank: cfg.CollapseBlank,
	})
}
