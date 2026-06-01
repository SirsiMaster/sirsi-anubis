package tui

import "strings"

// The five layout primitives (docs/TUI_DESIGN_PROOF.md §1.1). Anything not
// expressible in these is out of scope for v1 — constraint is the point.
//
// Frame   — fixed outer chrome (title row, status row, one content region)
// Pane    — a bordered, focusable rectangle inside the content region
// Table   — a column-aligned, no-wrap data grid inside a pane
// Palette — a modal overlay (command palette / pick-list); one open at a time
// Toast   — a transient banner above the status bar; never steals focus

// Align is a column's horizontal alignment.
type Align int

const (
	AlignLeft Align = iota
	AlignRight
)

// Column describes one table column. Width is the fixed cell budget; values
// wider than Width are truncated with an ellipsis, never wrapped (§2.2).
type Column struct {
	Title string
	Width int
	Align Align
}

// Table is a column-aligned grid. It owns its selection; scroll and sort are
// Gate-3 concerns and intentionally absent from the scaffold.
type Table struct {
	Columns  []Column
	Rows     [][]string
	Selected int // index into Rows; -1 for none
}

// Frame is the fixed outer chrome. Content is the already-rendered pane region.
type Frame struct {
	AppName    string
	View       string
	Breadcrumb string
	RightMeta  string // e.g. "⏱ 1.2s" or "4 nodes"
	Content    []string
	Hints      []Hint
	Mode       string // status-bar mode chip, e.g. "SCAN"
	Counts     string // status-bar gold counts, e.g. "29.3 GB"
}

// Pane is a bordered region. Focused panes carry the focus marker in their
// title and (in the styled renderer) an accent border — focus is never color
// alone (§5).
type Pane struct {
	Title   string
	Focused bool
	Body    []string
	Rect    Rect
}

// Toast is a transient banner; Token carries its severity color and text label.
type Toast struct {
	Text  string
	Token Token
}

// Palette is the modal command overlay. Query is the current fuzzy filter.
type Palette struct {
	Query    string
	Items    []string
	Selected int
}

// truncate fits s into width cells. Because the glyph policy (glyph.go) admits
// only width-1 non-ASCII glyphs, rune count equals display width, so truncation
// is exact. An over-long value loses its tail to a single ellipsis sigil.
func truncate(s string, width int, caps Capabilities) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	ell := Sigil("ellipsis", caps)
	if width == 1 {
		return ell
	}
	return string(r[:width-1]) + ell
}

// pad left/right-aligns s to width cells.
func pad(s string, width int, align Align) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	gap := strings.Repeat(" ", width-len(r))
	if align == AlignRight {
		return gap + s
	}
	return s + gap
}

// cell renders one table cell: truncate to the column budget, then pad to it.
func (c Column) cell(value string, caps Capabilities) string {
	return pad(truncate(value, c.Width, caps), c.Width, c.Align)
}

// Render returns the table as plain, cell-aligned lines (no ANSI). This is the
// linear/no-altscreen rendering — the base case the styled renderer decorates,
// never a separate adapter (Codex Gate-1 precondition). The focused-row marker
// is a width-1 sigil so the grid holds with or without Unicode.
func (t Table) Render(caps Capabilities) []string {
	lines := make([]string, 0, len(t.Rows)+2)

	// Header.
	header := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		header[i] = c.cell(c.Title, caps)
	}
	lines = append(lines, "  "+strings.Join(header, "  "))

	// Rule.
	width := len([]rune(lines[0]))
	lines = append(lines, strings.Repeat(Sigil("box-h", caps), width))

	// Rows.
	marker := Sigil("focus-marker", caps)
	for ri, row := range t.Rows {
		cells := make([]string, len(t.Columns))
		for ci, c := range t.Columns {
			val := ""
			if ci < len(row) {
				val = row[ci]
			}
			cells[ci] = c.cell(val, caps)
		}
		prefix := "  "
		if ri == t.Selected {
			prefix = marker + " "
		}
		lines = append(lines, prefix+strings.Join(cells, "  "))
	}
	return lines
}
