package tui

// Fixture views (docs/TUI_DESIGN_PROOF.md §6). These are the three canonical
// proof screens — Scan, Ra deploy, Router inbox — expressed as data. They are
// the visual proof the gate accepted; here they double as deterministic render
// fixtures for the test suite. They carry no behavior beyond the View contract.

// fixtureView is a static View backed by a fixed table and hint set.
type fixtureView struct {
	name    string
	layout  Layout
	hintIDs []CommandID
	table   Table
}

func (v fixtureView) Name() string         { return v.name }
func (v fixtureView) Layout() Layout       { return v.layout }
func (v fixtureView) HintIDs() []CommandID { return v.hintIDs }
func (v fixtureView) Table() Table         { return v.table }

// ScanView is §6.1 — scan results, Layout A "Survey" (Jackal / Anubis).
func ScanView() View {
	return fixtureView{
		name:    "Scan",
		layout:  LayoutSurvey,
		hintIDs: []CommandID{CmdMoveUp, CmdInspect, CmdClean, CmdFilter, CmdRefresh},
		table: Table{
			Columns: []Column{
				{Title: "RULE", Width: 22, Align: AlignLeft},
				{Title: "FINDINGS", Width: 8, Align: AlignRight},
				{Title: "RECLAIMABLE", Width: 11, Align: AlignRight},
				{Title: "SEVERITY", Width: 8, Align: AlignLeft},
				{Title: "LAST SEEN", Width: 9, Align: AlignLeft},
			},
			Rows: [][]string{
				{"parallels-remnants", "12", "4.2 GB", "WARN", "12s ago"},
				{"docker-dangling", "8", "1.1 GB", "WARN", "12s ago"},
				{"ghost-apps", "5", "318 MB", "INFO", "12s ago"},
				{"node-modules-orphan", "31", "9.7 GB", "WARN", "12s ago"},
				{"xcode-derived-data", "3", "14.0 GB", "WARN", "12s ago"},
				{"spotlight-corrupt", "1", "-", "BLOCK", "12s ago"},
			},
			Selected: 0,
		},
	}
}

// RaView is §6.2 — Ra deployment status, Layout C "Stream" (fleet).
func RaView() View {
	return fixtureView{
		name:    "Ra Fleet",
		layout:  LayoutStream,
		hintIDs: []CommandID{CmdMoveUp, CmdInspect, CmdRefresh, CmdBack, CmdFilter},
		table: Table{
			Columns: []Column{
				{Title: "NODE", Width: 16, Align: AlignLeft},
				{Title: "SURFACE", Width: 8, Align: AlignLeft},
				{Title: "STATE", Width: 8, Align: AlignLeft},
				{Title: "STEP", Width: 22, Align: AlignLeft},
				{Title: "ELAPSED", Width: 8, Align: AlignRight},
			},
			Rows: [][]string{
				{"mac-studio-01", "claude", "RUNNING", "3/7 weighing findings", "00:42"},
				{"mac-mini-02", "codex", "RUNNING", "5/7 purging", "01:07"},
				{"vm-ubuntu-03", "claude", "WAIT", "2/7 awaiting confirm", "00:51"},
				{"vm-ubuntu-04", "claude", "FAIL", "0/7 health-check failed", "00:03"},
			},
			Selected: 0,
		},
	}
}

// InboxView is §6.3 — router inbox, Layout B "Inspect" (idea-router).
func InboxView() View {
	return fixtureView{
		name:    "Router Inbox",
		layout:  LayoutInspect,
		hintIDs: []CommandID{CmdMoveUp, CmdInspect, CmdRouterAck, CmdFilter, CmdRefresh},
		table: Table{
			Columns: []Column{
				{Title: "FROM -> TO", Width: 18, Align: AlignLeft},
				{Title: "LANE / TITLE", Width: 24, Align: AlignLeft},
				{Title: "STATE", Width: 6, Align: AlignLeft},
			},
			Rows: [][]string{
				{"codex -> claude", "B canon-correction v2", "CLOSED"},
				{"codex -> claude", "A dispatch concurrency", "OPEN"},
				{"claude -> codex", "B TUI design proof", "DRAFT"},
			},
			Selected: 0,
		},
	}
}

// ProofViews returns the three canonical fixture screens in nav order.
func ProofViews() []View {
	return []View{ScanView(), RaView(), InboxView()}
}

// NewApp assembles a console AppState over the proof views with the canonical
// registry and the given capabilities. It is the scaffold's single composition
// root; nothing here launches anything — it only builds renderable state.
func NewApp(caps Capabilities) (*AppState, error) {
	reg, err := DefaultRegistry()
	if err != nil {
		return nil, err
	}
	views := ProofViews()
	app := &AppState{
		Views:     views,
		Active:    0,
		Registry:  reg,
		Caps:      caps,
		ViewState: ViewState{RowCount: len(views[0].Table().Rows), Selected: 0},
	}
	return app, nil
}
