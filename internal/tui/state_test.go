package tui

import "testing"

// Rule A1 alignment: a destructive command must NOT execute from a single
// keystroke. It arms a confirm modal; only a deliberate second confirmation
// resolves it. This is the reducer's most important governance property.
func TestDestructiveCommandArmsConfirmInsteadOfExecuting(t *testing.T) {
	clean := Command{ID: CmdClean, Key: "c", Destructive: true}

	st := ViewState{RowCount: 6, Selected: 2}
	st = Reduce(st, clean)
	if st.Confirm == nil {
		t.Fatal("destructive command did not arm a confirm modal")
	}
	if *st.Confirm != CmdClean {
		t.Errorf("confirm armed for %q, want %q", *st.Confirm, CmdClean)
	}

	// A second arbitrary keystroke while confirm is pending is ignored — the
	// modal captures the decision; nothing destructive runs.
	st2 := Reduce(st, Command{ID: CmdMoveDown, Key: "down"})
	if st2.Confirm == nil {
		t.Error("confirm modal was dismissed by an unrelated key")
	}

	// esc cancels: confirm clears, nothing executed.
	canceled := Reduce(st, Command{ID: CmdBack, Key: "esc"})
	if canceled.Confirm != nil {
		t.Error("esc did not cancel the confirm modal")
	}

	// enter confirms: modal resolves.
	confirmed := Reduce(st, Command{ID: CmdInspect, Key: "enter"})
	if confirmed.Confirm != nil {
		t.Error("enter did not resolve the confirm modal")
	}
}

func TestSelectionClamps(t *testing.T) {
	down := Command{ID: CmdMoveDown}
	up := Command{ID: CmdMoveUp}

	st := ViewState{RowCount: 3, Selected: 0}
	st = Reduce(st, up) // already at top
	if st.Selected != 0 {
		t.Errorf("moving up at top = %d, want 0", st.Selected)
	}
	for i := 0; i < 10; i++ {
		st = Reduce(st, down)
	}
	if st.Selected != 2 {
		t.Errorf("moving down past end = %d, want clamp at 2", st.Selected)
	}

	st = Reduce(st, Command{ID: CmdTop})
	if st.Selected != 0 {
		t.Errorf("CmdTop = %d, want 0", st.Selected)
	}
	st = Reduce(st, Command{ID: CmdBottom})
	if st.Selected != 2 {
		t.Errorf("CmdBottom = %d, want 2", st.Selected)
	}
}

func TestPaletteToggleAndBack(t *testing.T) {
	st := ViewState{RowCount: 1}
	st = Reduce(st, Command{ID: CmdPalette})
	if !st.PaletteOpen {
		t.Fatal("palette did not open")
	}
	st = Reduce(st, Command{ID: CmdBack})
	if st.PaletteOpen {
		t.Error("esc did not close the palette")
	}
}

func TestNewAppComposes(t *testing.T) {
	app, err := NewApp(Capabilities{AltScreen: true, UnicodeLayout: true})
	if err != nil {
		t.Fatalf("NewApp error = %v", err)
	}
	if len(app.Views) != 3 {
		t.Errorf("NewApp views = %d, want 3 proof screens", len(app.Views))
	}
	if app.ActiveView() == nil {
		t.Error("NewApp has no active view")
	}
	// Every proof view's hints must validate against the composed registry.
	for _, v := range app.Views {
		if err := ValidateView(app.Registry, v); err != nil {
			t.Errorf("view %q hints invalid: %v", v.Name(), err)
		}
	}
}
