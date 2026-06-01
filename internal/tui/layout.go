package tui

// Layout model (docs/TUI_DESIGN_PROOF.md §1).
//
// Panes tile via a binary split tree (horizontal or vertical splits only),
// capped at depth 2 so the layout is always describable in one sentence — the
// fix for v0.22's arbitrary nesting. The operator never hand-tiles; each view
// declares one of three named layouts and the engine derives the rectangles.

// Minimum and reference cell budgets (§2.2). Below the minimum the console
// shows a single centered message rather than rendering broken.
const (
	MinWidth  = 80
	MinHeight = 24

	// minPaneW and minPaneH are the smallest interior a pane may receive; a
	// split that would violate them is rejected in favor of a single pane.
	minPaneW = 24
	minPaneH = 6
)

// Layout is one of the three named tilings (§1.3). Layout is chosen by the
// view, not the operator.
type Layout int

const (
	// LayoutSurvey is a single full-content table (scan, router list).
	LayoutSurvey Layout = iota
	// LayoutInspect is master+detail, a 60/40 vertical split (finding → detail).
	LayoutInspect
	// LayoutStream is table+log, a 70/30 horizontal split (ra deploy, monitor).
	LayoutStream
)

// String renders the layout's stable name.
func (l Layout) String() string {
	switch l {
	case LayoutSurvey:
		return "Survey"
	case LayoutInspect:
		return "Inspect"
	case LayoutStream:
		return "Stream"
	default:
		return "Unknown"
	}
}

// Rect is a cell rectangle (top-left origin).
type Rect struct {
	X, Y, W, H int
}

// PaneRect is a named region produced by tiling a layout.
type PaneRect struct {
	Name string
	Rect Rect
}

// Fits reports whether r can host a pane interior without violating the minimum.
func (r Rect) Fits() bool {
	return r.W >= minPaneW && r.H >= minPaneH
}

// Tile derives pane rectangles for the layout within content. The content rect
// excludes the title and status rows (those belong to the Frame). If a split
// would produce a sub-minimum pane, Tile degrades to a single pane — the layout
// never renders a broken sliver.
func Tile(l Layout, content Rect) []PaneRect {
	switch l {
	case LayoutInspect:
		leftW := content.W * 60 / 100
		rightW := content.W - leftW - 1 // 1 col gutter
		left := Rect{X: content.X, Y: content.Y, W: leftW, H: content.H}
		right := Rect{X: content.X + leftW + 1, Y: content.Y, W: rightW, H: content.H}
		if left.Fits() && right.Fits() {
			return []PaneRect{{"master", left}, {"detail", right}}
		}
	case LayoutStream:
		topH := content.H * 70 / 100
		botH := content.H - topH - 1 // 1 row gutter
		top := Rect{X: content.X, Y: content.Y, W: content.W, H: topH}
		bot := Rect{X: content.X, Y: content.Y + topH + 1, W: content.W, H: botH}
		if top.Fits() && bot.Fits() {
			return []PaneRect{{"table", top}, {"log", bot}}
		}
	}
	// LayoutSurvey, or any degraded split: a single full-content pane.
	return []PaneRect{{"main", content}}
}

// Depth returns the split depth of a layout (0 = single pane). v1 caps at 1
// real split; the constant documents the invariant the design forbids exceeding.
func (l Layout) Depth() int {
	switch l {
	case LayoutInspect, LayoutStream:
		return 1
	default:
		return 0
	}
}
