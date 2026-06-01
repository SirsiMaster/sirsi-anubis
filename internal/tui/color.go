package tui

import "os"

// Color & capability model (docs/TUI_DESIGN_PROOF.md §2.4, §5).
//
// Color is semantic, never decorative: every token means exactly one thing and
// degrades down a truecolor → 256 → 16 → attribute-only ladder. NO_COLOR
// collapses to attribute-only so meaning is never carried by color alone.

// ColorDepth is a rung on the degrade ladder.
type ColorDepth int

const (
	// ColorNone is attribute-only (bold/reverse/underline); meaning must still
	// survive. This is the NO_COLOR and linear-renderer default.
	ColorNone ColorDepth = iota
	// Color16 is the ANSI base palette.
	Color16
	// Color256 is the xterm 256-color cube.
	Color256
	// ColorTrue is 24-bit truecolor — full brand fidelity.
	ColorTrue
)

// Token is a semantic color role. The render layer maps a token to an actual
// color (or attribute) at the active ColorDepth; primitives only ever name the
// role, never a hex value, so the ladder is honored everywhere by construction.
type Token int

const (
	TokBrand  Token = iota // gold #C8A951 — identity, headers, selected counts
	TokAccent              // lapis #1A1A5E — mode chips, active focus border
	TokOK                  // green — pass, safe, healthy
	TokWarn                // amber — needs attention, reclaimable
	TokDanger              // red — destructive, error, protected-path block
	TokDim                 // gray — chrome, hints, truncation
)

// SeverityLabel is the text token a severity carries IN ADDITION to color, so
// colorblind and NO_COLOR operators lose no information (§5). A severity is
// never communicated by color alone.
func (t Token) SeverityLabel() string {
	switch t {
	case TokOK:
		return "PASS"
	case TokWarn:
		return "WARN"
	case TokDanger:
		return "BLOCK"
	case TokDim:
		return "INFO"
	default:
		return ""
	}
}

// Capabilities is the resolved terminal/environment profile that the renderer
// reads. It is intentionally a plain value so tests can construct any profile
// directly without touching the real environment.
type Capabilities struct {
	// Color is the deepest color rung this surface may use.
	Color ColorDepth
	// UnicodeLayout permits declared layout glyphs (glyph.go) to render as
	// their preferred codepoint instead of the ASCII fallback.
	UnicodeLayout bool
	// ReducedMotion disables spinners/slides/pulses (§5); progress and spinners
	// degrade to determinate, static forms.
	ReducedMotion bool
	// AltScreen indicates the fullscreen alternate buffer is in use. When
	// false, the linear renderer (the accessible, screen-reader-traversable
	// path) is selected — a first-class renderer, not an adapter.
	AltScreen bool
}

// DetectCapabilities resolves capabilities from the environment. It is
// deliberately conservative: anything it cannot positively confirm degrades to
// the safe rung. The capability probe for layout glyphs is modeled here as a
// TERM allowlist rather than a live cursor-advance measurement so the scaffold
// stays deterministic; the probe (§2.3 G3) is a Gate-3 concern.
func DetectCapabilities(env func(string) string) Capabilities {
	if env == nil {
		env = os.Getenv
	}
	caps := Capabilities{
		Color:         Color16,
		UnicodeLayout: true,
		ReducedMotion: false,
		AltScreen:     true,
	}

	// NO_COLOR (https://no-color.org) collapses to attribute-only.
	if env("NO_COLOR") != "" {
		caps.Color = ColorNone
		caps.ReducedMotion = true
	} else {
		switch env("COLORTERM") {
		case "truecolor", "24bit":
			caps.Color = ColorTrue
		}
	}

	switch term := env("TERM"); term {
	case "", "dumb":
		// No positive terminal signal: take the safest path, including the
		// linear renderer and ASCII glyphs.
		caps.Color = ColorNone
		caps.UnicodeLayout = false
		caps.AltScreen = false
		caps.ReducedMotion = true
	}

	// An explicit opt-out of the alternate screen selects the linear renderer
	// regardless of terminal (accessibility / screen-reader pairing).
	if env("SIRSI_TUI_NO_ALTSCREEN") != "" {
		caps.AltScreen = false
	}
	if env("SIRSI_TUI_REDUCE_MOTION") != "" {
		caps.ReducedMotion = true
	}

	return caps
}
