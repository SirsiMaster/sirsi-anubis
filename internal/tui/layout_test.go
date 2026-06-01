package tui

import "testing"

func TestTilePaneCounts(t *testing.T) {
	content := Rect{X: 0, Y: 1, W: 120, H: 36}
	tests := []struct {
		layout    Layout
		wantPanes int
		wantNames []string
	}{
		{LayoutSurvey, 1, []string{"main"}},
		{LayoutInspect, 2, []string{"master", "detail"}},
		{LayoutStream, 2, []string{"table", "log"}},
	}
	for _, tc := range tests {
		t.Run(tc.layout.String(), func(t *testing.T) {
			panes := Tile(tc.layout, content)
			if len(panes) != tc.wantPanes {
				t.Fatalf("Tile(%s) = %d panes, want %d", tc.layout, len(panes), tc.wantPanes)
			}
			for i, want := range tc.wantNames {
				if panes[i].Name != want {
					t.Errorf("pane %d = %q, want %q", i, panes[i].Name, want)
				}
				if !panes[i].Rect.Fits() {
					t.Errorf("pane %q rect %+v does not meet the minimum interior", want, panes[i].Rect)
				}
			}
		})
	}
}

func TestTileDegradesWhenTooSmall(t *testing.T) {
	// A content region too small to host two minimum panes collapses to one,
	// never a broken sliver.
	tiny := Rect{X: 0, Y: 0, W: 40, H: 8}
	for _, l := range []Layout{LayoutInspect, LayoutStream} {
		panes := Tile(l, tiny)
		if len(panes) != 1 || panes[0].Name != "main" {
			t.Errorf("Tile(%s, tiny) = %+v, want single 'main' pane", l, panes)
		}
	}
}

func TestLayoutDepthCap(t *testing.T) {
	// v1 caps the split tree at depth 1 (one real split). No layout exceeds it.
	for _, l := range []Layout{LayoutSurvey, LayoutInspect, LayoutStream} {
		if d := l.Depth(); d > 1 {
			t.Errorf("Layout %s depth = %d, exceeds the v1 cap of 1", l, d)
		}
	}
}

func TestRectFits(t *testing.T) {
	if (Rect{W: minPaneW, H: minPaneH}).Fits() != true {
		t.Error("exact-minimum rect should fit")
	}
	if (Rect{W: minPaneW - 1, H: minPaneH}).Fits() != false {
		t.Error("sub-minimum width should not fit")
	}
}
