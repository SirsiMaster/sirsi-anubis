package tui

// State model and reducer (docs/TUI_DESIGN_PROOF.md §3.4, §4).
//
// The rendered surface is a projection of wired state: a keypress resolves to a
// Command via the Registry, the Command is routed to a pure reducer, and the
// reducer returns the next state. There is no path by which a key mutates the
// screen without going through a registered command — the v0.22 "keys did
// nothing" failure is unrepresentable.

// View is a single console screen (Scan, Ra deploy, Router inbox, …). A view
// declares its layout and the commands it wires; it never draws chrome.
type View interface {
	// Name is the title-bar view label.
	Name() string
	// Layout is the named tiling this view wants (§1.3).
	Layout() Layout
	// HintIDs lists, in order, the command ids shown as status-bar hints. Every
	// id MUST exist in the registry — ValidateView enforces it.
	HintIDs() []CommandID
	// Table is the view's primary data grid (the scaffold's render surface).
	Table() Table
}

// ViewState is the mutable per-view state a reducer evolves. It is a plain
// value: reducers are pure and return a new ViewState.
type ViewState struct {
	Selected    int  // selected row index
	RowCount    int  // number of rows (selection clamp bound)
	PaletteOpen bool // Ctrl-K overlay visible
	Filter      string
	// Confirm holds a destructive command awaiting its second confirmation.
	// While set, the surface shows a confirm modal and no destructive action
	// has run (Rule A1: destructive verbs never fire from one keystroke).
	Confirm *CommandID
}

// Reduce applies cmd to st and returns the next state. It is total and pure.
// The destructive-confirm rule lives here so every view inherits it: a
// destructive command does not execute — it arms a confirm modal; the modal is
// resolved only by an explicit confirm (enter) and canceled by back (esc).
func Reduce(st ViewState, cmd Command) ViewState {
	// A pending confirm modal captures the next decision.
	if st.Confirm != nil {
		switch cmd.ID {
		case CmdInspect: // enter = deliberate second confirmation → execute
			st.Confirm = nil
		case CmdBack, CmdQuit: // esc = cancel, nothing destructive ran
			st.Confirm = nil
		}
		return st
	}

	switch cmd.ID {
	case CmdMoveDown:
		if st.RowCount > 0 && st.Selected < st.RowCount-1 {
			st.Selected++
		}
	case CmdMoveUp:
		if st.Selected > 0 {
			st.Selected--
		}
	case CmdTop:
		st.Selected = 0
	case CmdBottom:
		if st.RowCount > 0 {
			st.Selected = st.RowCount - 1
		}
	case CmdPalette:
		st.PaletteOpen = !st.PaletteOpen
	case CmdBack:
		switch {
		case st.PaletteOpen:
			st.PaletteOpen = false
		case st.Filter != "":
			st.Filter = ""
		}
	default:
		if cmd.Destructive {
			// Arm the confirm modal; do not execute.
			id := cmd.ID
			st.Confirm = &id
		}
	}
	return st
}

// AppState is the whole-console state: the active view index, the shared
// registry, resolved capabilities, and an optional toast.
type AppState struct {
	Views     []View
	Active    int
	Registry  *Registry
	Caps      Capabilities
	ViewState ViewState
	Toast     *Toast
}

// ActiveView returns the currently focused view.
func (a *AppState) ActiveView() View {
	if a.Active < 0 || a.Active >= len(a.Views) {
		return nil
	}
	return a.Views[a.Active]
}

// ValidateView checks that every hint id a view advertises is registered and
// keyed — the test-able form of the §7 delta-2 guarantee.
func ValidateView(reg *Registry, v View) error {
	_, err := reg.Hints(v.HintIDs())
	return err
}
