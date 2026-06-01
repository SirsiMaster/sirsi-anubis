package tui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// The three Gate-1 fixture sizes Codex named.
var fixtureSizes = []struct {
	name          string
	width, height int
}{
	{"80x24", 80, 24},
	{"100col", 100, 30},
	{"120x40", 120, 40},
}

// Codex Gate-1 precondition: fixture-based renderer tests for the three proof
// screens at 80x24, 100-column, and 120x40. Every rendered line must fit the
// width and contain only grid-safe glyphs — the structural proof that the grid
// cannot break, at every supported size, for every proof screen.
func TestProofScreensRenderWithinBudget(t *testing.T) {
	caps := Capabilities{Color: ColorTrue, UnicodeLayout: true, AltScreen: true}
	for _, sz := range fixtureSizes {
		for _, v := range ProofViews() {
			name := sz.name + "/" + v.Name()
			t.Run(name, func(t *testing.T) {
				app, err := NewApp(caps)
				if err != nil {
					t.Fatalf("NewApp error = %v", err)
				}
				app.Active = indexOfView(app, v.Name())
				lines := NewRenderer(caps).Render(app, sz.width, sz.height)

				if len(lines) > sz.height {
					t.Errorf("rendered %d lines, exceeds height %d", len(lines), sz.height)
				}
				for i, ln := range lines {
					if w := utf8.RuneCountInString(ln); w > sz.width {
						t.Errorf("line %d width %d exceeds %d: %q", i, w, sz.width, ln)
					}
					for _, r := range ln {
						if !IsLayoutSafe(r) {
							t.Errorf("line %d contains non-layout-safe rune %q (U+%X)", i, string(r), r)
						}
					}
				}
			})
		}
	}
}

// Codex Gate-1 precondition: no-altscreen fixtures. When the alternate screen
// is unavailable the linear renderer is selected and produces labeled,
// box-free, ASCII-safe output a screen reader can traverse.
func TestLinearRendererIsSelectedAndAccessible(t *testing.T) {
	caps := DetectCapabilities(envMap(map[string]string{
		"TERM": "xterm-256color", "SIRSI_TUI_NO_ALTSCREEN": "1", "NO_COLOR": "1",
	}))
	if _, ok := NewRenderer(caps).(LinearRenderer); !ok {
		t.Fatalf("no-altscreen caps did not select LinearRenderer")
	}
	app, err := NewApp(caps)
	if err != nil {
		t.Fatalf("NewApp error = %v", err)
	}
	lines := NewRenderer(caps).Render(app, 100, 40)

	joined := strings.Join(lines, "\n")
	// Linear mode prefixes each row with a semantic position label.
	if !strings.Contains(joined, "Row 1 of") {
		t.Error("linear output missing semantic row labels")
	}
	// No box-drawing chrome in linear mode, and never any wide/tofu rune.
	for _, r := range joined {
		if r > 127 {
			t.Errorf("linear output contains non-ASCII rune %q; want screen-reader-safe ASCII", string(r))
		}
	}
}

func TestTooSmallTakeover(t *testing.T) {
	caps := Capabilities{AltScreen: true, UnicodeLayout: true}
	app, _ := NewApp(caps)
	lines := AltScreenRenderer{}.Render(app, 40, 10)
	if len(lines) == 0 {
		t.Fatal("expected a takeover message")
	}
	if !strings.Contains(strings.Join(lines, ""), "Horus needs") {
		t.Errorf("small terminal did not show the size takeover: %v", lines)
	}
}

func TestStatusBarRefusesDeadHint(t *testing.T) {
	// A view whose hint references a palette-only command must make the status
	// bar build fail rather than render a dead key.
	caps := Capabilities{AltScreen: true, UnicodeLayout: true, Color: ColorTrue}
	app, _ := NewApp(caps)
	app.Views = []View{badHintView{}}
	app.Active = 0
	if _, err := statusBar(app, 120); err == nil {
		t.Error("statusBar accepted a dead hint; want error")
	}
}

// badHintView surfaces a keyless (palette-only) command as a status hint.
type badHintView struct{}

func (badHintView) Name() string         { return "Bad" }
func (badHintView) Layout() Layout       { return LayoutSurvey }
func (badHintView) HintIDs() []CommandID { return []CommandID{CmdScan} } // CmdScan has no key
func (badHintView) Table() Table         { return Table{} }

func indexOfView(app *AppState, name string) int {
	for i, v := range app.Views {
		if v.Name() == name {
			return i
		}
	}
	return 0
}
