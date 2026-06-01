package tui

import (
	"testing"
	"unicode/utf8"
)

// Codex Gate-1 precondition: width/fallback discipline must cover EVERY
// non-ASCII layout-bearing glyph, not only hieroglyphs. This test iterates the
// closed glyph set and enforces the G1–G3 invariants for each entry, so adding
// a glyph without a safe width and fallback fails CI.
func TestLayoutGlyphsAreGridSafe(t *testing.T) {
	for _, g := range LayoutGlyphs() {
		t.Run(g.Name, func(t *testing.T) {
			// G3: every layout glyph is declared single-width. A multi-width or
			// ambiguous glyph in a fixed cell shifts the grid.
			if g.Width != 1 {
				t.Errorf("glyph %q (%q): width = %d, want 1 (layout-bearing glyphs must be single-width)", g.Name, string(g.R), g.Width)
			}
			// G1: no layout glyph may be an Egyptian hieroglyph.
			if IsHieroglyph(g.R) {
				t.Errorf("glyph %q is a hieroglyph (U+%X) — forbidden in layout-bearing cells", g.Name, g.R)
			}
			// G2: the fallback must be a non-empty, single-cell ASCII string.
			if g.ASCII == "" {
				t.Errorf("glyph %q has no ASCII fallback", g.Name)
			}
			if rc := utf8.RuneCountInString(g.ASCII); rc != 1 {
				t.Errorf("glyph %q fallback %q is %d cells, want 1", g.Name, g.ASCII, rc)
			}
			for _, fr := range g.ASCII {
				if fr > 127 {
					t.Errorf("glyph %q fallback %q is not ASCII", g.Name, g.ASCII)
				}
			}
		})
	}
}

func TestSigilUsesFallbackWithoutUnicode(t *testing.T) {
	unicodeCaps := Capabilities{UnicodeLayout: true}
	asciiCaps := Capabilities{UnicodeLayout: false}

	for _, g := range LayoutGlyphs() {
		if got := Sigil(g.Name, unicodeCaps); got != string(g.R) {
			t.Errorf("Sigil(%q, unicode) = %q, want %q", g.Name, got, string(g.R))
		}
		if got := Sigil(g.Name, asciiCaps); got != g.ASCII {
			t.Errorf("Sigil(%q, ascii) = %q, want fallback %q", g.Name, got, g.ASCII)
		}
	}
	// Unknown names degrade to a width-1 last resort, never a panic or wide rune.
	if got := Sigil("does-not-exist", unicodeCaps); got != "?" {
		t.Errorf("Sigil(unknown) = %q, want %q", got, "?")
	}
}

func TestIsLayoutSafe(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"ascii letter", 'A', true},
		{"ascii space", ' ', true},
		{"declared sigil", '◉', true},
		{"declared box", '│', true},
		{"hieroglyph horus", '𓂀', false},
		{"hieroglyph maat", '𓆄', false},
		{"undeclared emoji", '🐺', false},
		{"undeclared wide", '＊', false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsLayoutSafe(tc.r); got != tc.want {
				t.Errorf("IsLayoutSafe(%q) = %v, want %v", string(tc.r), got, tc.want)
			}
		})
	}
}
