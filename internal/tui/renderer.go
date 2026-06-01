package tui

import (
	"fmt"
	"strings"
)

// Renderer contract (docs/TUI_DESIGN_PROOF.md §1, §5).
//
// Two renderers implement the same contract. The LinearRenderer (no-altscreen)
// is the FIRST-CLASS accessible path — screen-reader-traversable, line-oriented
// — not an adapter bolted onto the fullscreen renderer (Codex Gate-1
// precondition). The AltScreenRenderer adds chrome on top of the same data.
// Both clamp every emitted line to the terminal width so a stray wide glyph can
// never push the grid past the edge.

// Renderer turns app state into terminal lines.
type Renderer interface {
	Render(app *AppState, width, height int) []string
}

// NewRenderer selects the renderer from capabilities: linear when the alternate
// screen is unavailable or declined (accessibility), fullscreen otherwise.
func NewRenderer(caps Capabilities) Renderer {
	if !caps.AltScreen {
		return LinearRenderer{}
	}
	return AltScreenRenderer{}
}

// clampLine guarantees a line never exceeds width cells. Because every glyph the
// renderer places is width 1 (glyph.go), rune length equals display width and
// the clamp is exact.
func clampLine(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	return string(r[:width])
}

// tooSmall renders the single centered "needs ≥ 80×24" takeover (§2.2).
func tooSmall(width, height int) []string {
	msg := fmt.Sprintf("Horus needs >= %dx%d", MinWidth, MinHeight)
	if width < len(msg) {
		return []string{clampLine(msg, width)}
	}
	line := pad(strings.Repeat(" ", (width-len(msg))/2)+msg, width, AlignLeft)
	out := make([]string, 0, height)
	for i := 0; i < height; i++ {
		if i == height/2 {
			out = append(out, clampLine(line, width))
		} else {
			out = append(out, "")
		}
	}
	return out
}

// titleBar builds row 0: identity sigil, app name, active view, breadcrumb, and
// a right-anchored meta field.
func titleBar(app *AppState, width int) string {
	v := app.ActiveView()
	bullet := Sigil("bullet", app.Caps)
	left := fmt.Sprintf("%s Horus %s %s", Sigil("horus", app.Caps), bullet, v.Name())
	return clampLine(pad(left, width, AlignLeft), width)
}

// statusBar builds the last row from the active view's registered hints. If any
// hint references an unwired command the registry refuses to build it, so a dead
// hint can never reach the screen (§7 delta 2).
func statusBar(app *AppState, width int) (string, error) {
	v := app.ActiveView()
	hints, err := app.Registry.Hints(v.HintIDs())
	if err != nil {
		return "", err
	}
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = fmt.Sprintf("%s %s", h.Key, h.Label)
	}
	sep := " " + Sigil("bullet", app.Caps) + " "
	return clampLine(pad(strings.Join(parts, sep), width, AlignLeft), width), nil
}

// AltScreenRenderer draws the fullscreen Frame: title row, a single boxed
// content region holding the active view's table, and the status row.
type AltScreenRenderer struct{}

// Render implements Renderer.
func (AltScreenRenderer) Render(app *AppState, width, height int) []string {
	if width < MinWidth || height < MinHeight {
		return tooSmall(width, height)
	}
	caps := app.Caps
	out := make([]string, 0, height)
	out = append(out, titleBar(app, width))

	// Boxed content region between title and status rows.
	interiorH := height - 2 - 2 // minus title+status, minus box top/bottom
	body := app.ActiveView().Table().Render(caps)
	out = append(out, boxTop(width, caps))
	for i := 0; i < interiorH; i++ {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		v := Sigil("box-v", caps)
		inner := pad(clampLine(line, width-4), width-4, AlignLeft)
		out = append(out, fmt.Sprintf("%s %s %s", v, inner, v))
	}
	out = append(out, boxBottom(width, caps))

	if sb, err := statusBar(app, width); err == nil {
		out = append(out, sb)
	} else {
		out = append(out, clampLine("status unavailable", width))
	}
	return out
}

func boxTop(width int, caps Capabilities) string {
	h := Sigil("box-h", caps)
	return Sigil("box-tl", caps) + strings.Repeat(h, width-2) + Sigil("box-tr", caps)
}

func boxBottom(width int, caps Capabilities) string {
	h := Sigil("box-h", caps)
	return Sigil("box-bl", caps) + strings.Repeat(h, width-2) + Sigil("box-br", caps)
}

// LinearRenderer is the accessible, no-altscreen path: labeled, scrolling,
// line-oriented output. Each data row is prefixed with its semantic position so
// a screen reader can announce "Row 3 of 14". No box drawing, no chrome that
// only makes sense visually.
type LinearRenderer struct{}

// Render implements Renderer.
func (LinearRenderer) Render(app *AppState, width, height int) []string {
	v := app.ActiveView()
	out := []string{
		clampLine(fmt.Sprintf("Horus - %s", v.Name()), width),
	}

	tbl := v.Table()
	n := len(tbl.Rows)
	for i, row := range tbl.Rows {
		label := fmt.Sprintf("Row %d of %d: %s", i+1, n, strings.Join(row, " | "))
		out = append(out, clampLine(label, width))
	}

	hints, err := app.Registry.Hints(v.HintIDs())
	if err == nil {
		parts := make([]string, len(hints))
		for i, h := range hints {
			parts[i] = fmt.Sprintf("%s %s", h.Key, h.Label)
		}
		out = append(out, clampLine("Actions: "+strings.Join(parts, ", "), width))
	}
	return out
}
