package tui

import "testing"

// envMap builds an env lookup func from a map, for deterministic capability
// tests that never touch the real environment.
func envMap(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDetectCapabilities(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantColor  ColorDepth
		wantUni    bool
		wantAlt    bool
		wantMotion bool // ReducedMotion
	}{
		{
			name:      "truecolor xterm",
			env:       map[string]string{"TERM": "xterm-256color", "COLORTERM": "truecolor"},
			wantColor: ColorTrue, wantUni: true, wantAlt: true, wantMotion: false,
		},
		{
			name:      "NO_COLOR collapses to attribute-only and reduces motion",
			env:       map[string]string{"TERM": "xterm-256color", "NO_COLOR": "1"},
			wantColor: ColorNone, wantUni: true, wantAlt: true, wantMotion: true,
		},
		{
			name:      "dumb terminal takes the safe linear path",
			env:       map[string]string{"TERM": "dumb"},
			wantColor: ColorNone, wantUni: false, wantAlt: false, wantMotion: true,
		},
		{
			name:      "explicit no-altscreen selects linear renderer",
			env:       map[string]string{"TERM": "xterm-256color", "SIRSI_TUI_NO_ALTSCREEN": "1"},
			wantColor: Color16, wantUni: true, wantAlt: false, wantMotion: false,
		},
		{
			name:      "explicit reduce-motion",
			env:       map[string]string{"TERM": "xterm-256color", "SIRSI_TUI_REDUCE_MOTION": "1"},
			wantColor: Color16, wantUni: true, wantAlt: true, wantMotion: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			caps := DetectCapabilities(envMap(tc.env))
			if caps.Color != tc.wantColor {
				t.Errorf("Color = %v, want %v", caps.Color, tc.wantColor)
			}
			if caps.UnicodeLayout != tc.wantUni {
				t.Errorf("UnicodeLayout = %v, want %v", caps.UnicodeLayout, tc.wantUni)
			}
			if caps.AltScreen != tc.wantAlt {
				t.Errorf("AltScreen = %v, want %v", caps.AltScreen, tc.wantAlt)
			}
			if caps.ReducedMotion != tc.wantMotion {
				t.Errorf("ReducedMotion = %v, want %v", caps.ReducedMotion, tc.wantMotion)
			}
		})
	}
}

func TestSeverityLabelsAreColorIndependent(t *testing.T) {
	// Severity must carry a text token so color is never the sole signal (§5).
	want := map[Token]string{
		TokOK:     "PASS",
		TokWarn:   "WARN",
		TokDanger: "BLOCK",
		TokDim:    "INFO",
	}
	for tok, label := range want {
		if got := tok.SeverityLabel(); got != label {
			t.Errorf("Token(%d).SeverityLabel() = %q, want %q", tok, got, label)
		}
	}
}
