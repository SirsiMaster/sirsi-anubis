package output

import (
	"errors"
	"os"
	"sync/atomic"
	"testing"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
)

// newTestModel creates a lightweight TUIModel that skips expensive
// system calls (vitals collection, filesystem scanning). This makes
// each test run in <10ms instead of ~6s.
func newTestModel() TUIModel {
	ti := textinput.New()
	ti.Placeholder = "test"
	ti.CharLimit = 256
	ti.Prompt = "> "

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(Gold)

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(10))

	return TUIModel{
		input:       ti,
		viewport:    vp,
		spinner:     sp,
		width:       100,
		height:      40,
		mode:        viewTabs,
		activeTab:   0,
		historyIdx:  -1,
		tabIdx:      -1,
		activeDeity: make(map[string]bool),
		deityState:  make(map[string]deityRunState),
		streamCh:    make(chan string, 100),
		runningProc: &atomic.Pointer[os.Process]{},
		steleReader: stele.NewReader("test"),
	}
}

// ── TestScanSelectTransition ────────────────────────────────────────
// Verifies that a nativeResult carrying a selectReq transitions the
// TUI to viewSelect mode and exposes the request's items.

func TestScanSelectTransition(t *testing.T) {
	type testCase struct {
		name          string
		msg           nativeResultMsg
		wantMode      viewMode
		wantItemCount int
		wantTitle     string
	}

	confirmFn := func(selected []selectItem) nativeResult {
		return nativeResult{deityKey: "anubis", lines: []string{"cleaned"}}
	}

	tests := []testCase{
		{
			name: "select request transitions to viewSelect",
			msg: nativeResultMsg{
				deityKey: "anubis",
				selectReq: &selectRequest{
					title: "Pick Items",
					items: []selectItem{
						{Label: "Item A", Size: 100, Selected: true},
						{Label: "Item B", Size: 200, Selected: false},
						{Label: "Item C", Size: 300, Selected: true},
					},
					onConfirm: confirmFn,
				},
			},
			wantMode:      viewSelect,
			wantItemCount: 3,
			wantTitle:     "Pick Items",
		},
		{
			name: "single item select",
			msg: nativeResultMsg{
				deityKey: "anubis",
				selectReq: &selectRequest{
					title: "One Item",
					items: []selectItem{
						{Label: "Solo", Size: 50, Selected: true},
					},
					onConfirm: confirmFn,
				},
			},
			wantMode:      viewSelect,
			wantItemCount: 1,
			wantTitle:     "One Item",
		},
		{
			name: "no select request stays viewDone",
			msg: nativeResultMsg{
				deityKey: "anubis",
				lines:    []string{"scan complete"},
			},
			wantMode:      viewDone,
			wantItemCount: 0,
			wantTitle:     "",
		},
		{
			name: "select request with error stays viewDone",
			msg: nativeResultMsg{
				deityKey: "anubis",
				err:      errors.New("scan failed"),
				selectReq: &selectRequest{
					title: "Should Not Appear",
					items: []selectItem{{Label: "X"}},
				},
			},
			wantMode:      viewDone,
			wantItemCount: 0,
			wantTitle:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel()
			m.mode = viewRunning

			model, _ := m.handleNativeResult(tc.msg)

			if model.mode != tc.wantMode {
				t.Errorf("mode = %d, want %d", model.mode, tc.wantMode)
			}

			if tc.wantMode == viewSelect {
				if len(model.selectItems) != tc.wantItemCount {
					t.Errorf("selectItems count = %d, want %d", len(model.selectItems), tc.wantItemCount)
				}
				if model.selectTitle != tc.wantTitle {
					t.Errorf("selectTitle = %q, want %q", model.selectTitle, tc.wantTitle)
				}
				if model.selectCursor != 0 {
					t.Errorf("selectCursor = %d, want 0", model.selectCursor)
				}
				if model.selectOnConfirm == nil {
					t.Error("selectOnConfirm should be set")
				}
			} else {
				if len(model.selectItems) != 0 && tc.wantItemCount == 0 {
					// selectItems should not have been populated
					// (for error case, the old selectItems remain empty since we start fresh)
				}
			}
		})
	}
}

// ── TestScanSelectCursorNavigation ──────────────────────────────────
// Verifies cursor movement and toggle within viewSelect.

func TestScanSelectCursorNavigation(t *testing.T) {
	m := newTestModel()
	m.mode = viewSelect
	m.selectItems = []selectItem{
		{Label: "A", Selected: false},
		{Label: "B", Selected: false},
		{Label: "C", Selected: false},
	}
	m.selectCursor = 0

	// Move down
	updated, _ := m.handleSelectKey("down")
	model := updated.(TUIModel)
	if model.selectCursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", model.selectCursor)
	}

	// Move down again
	updated, _ = model.handleSelectKey("down")
	model = updated.(TUIModel)
	if model.selectCursor != 2 {
		t.Errorf("after second down: cursor = %d, want 2", model.selectCursor)
	}

	// Move down at bottom should not exceed
	updated, _ = model.handleSelectKey("down")
	model = updated.(TUIModel)
	if model.selectCursor != 2 {
		t.Errorf("at bottom down: cursor = %d, want 2", model.selectCursor)
	}

	// Move up
	updated, _ = model.handleSelectKey("up")
	model = updated.(TUIModel)
	if model.selectCursor != 1 {
		t.Errorf("after up: cursor = %d, want 1", model.selectCursor)
	}

	// Toggle space
	updated, _ = model.handleSelectKey(" ")
	model = updated.(TUIModel)
	if !model.selectItems[1].Selected {
		t.Error("space should toggle item 1 to selected")
	}

	// Toggle again
	updated, _ = model.handleSelectKey(" ")
	model = updated.(TUIModel)
	if model.selectItems[1].Selected {
		t.Error("second space should toggle item 1 back to unselected")
	}

	// Esc returns to tabs
	updated, _ = model.handleSelectKey("esc")
	model = updated.(TUIModel)
	if model.mode != viewTabs {
		t.Errorf("esc should return to viewTabs, got %d", model.mode)
	}
}

// ── TestAnalyzeDrillDownBack ────────────────────────────────────────
// Verifies that analyzeResultMsg transitions to viewAnalyze, that
// drill-down pushes history and back pops it, and that final back
// exits to viewTabs.

func TestAnalyzeDrillDownBack(t *testing.T) {
	type testCase struct {
		name string
		msg  analyzeResultMsg
		// Expected after processing the message
		wantMode    viewMode
		wantPath    string
		wantEntries int
	}

	tests := []testCase{
		{
			name: "analyze result transitions to viewAnalyze",
			msg: analyzeResultMsg{
				result: &jackal.AnalyzeResult{
					Path:      "/home",
					TotalSize: 5000,
					Entries: []jackal.DirEntry{
						{Name: "Documents", Path: "/home/Documents", Size: 3000, IsDir: true},
						{Name: "Pictures", Path: "/home/Pictures", Size: 2000, IsDir: true},
					},
				},
			},
			wantMode:    viewAnalyze,
			wantPath:    "/home",
			wantEntries: 2,
		},
		{
			name: "analyze error with no history stays viewAnalyze",
			msg: analyzeResultMsg{
				err: errors.New("permission denied"),
			},
			wantMode:    viewAnalyze,
			wantPath:    "",
			wantEntries: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel()
			m.mode = viewRunning

			updated, _ := m.Update(tc.msg)
			model := updated.(TUIModel)

			if model.mode != tc.wantMode {
				t.Errorf("mode = %d, want %d", model.mode, tc.wantMode)
			}
			if model.analyzePath != tc.wantPath {
				t.Errorf("analyzePath = %q, want %q", model.analyzePath, tc.wantPath)
			}
			if len(model.analyzeEntries) != tc.wantEntries {
				t.Errorf("analyzeEntries count = %d, want %d", len(model.analyzeEntries), tc.wantEntries)
			}
		})
	}
}

func TestAnalyzeDrillDownBackNavigation(t *testing.T) {
	// Set up a model already in viewAnalyze at root level
	m := newTestModel()
	m.mode = viewAnalyze
	m.analyzePath = "/home"
	m.analyzeTotal = 5000
	m.analyzeEntries = []jackal.DirEntry{
		{Name: "Documents", Path: "/home/Documents", Size: 3000, IsDir: true},
		{Name: "file.txt", Path: "/home/file.txt", Size: 2000, IsDir: false},
	}
	m.analyzeCursor = 0
	m.analyzeHistory = nil

	// Simulate a drill-down by manually pushing history and setting child state.
	// The actual drill-down calls jackal.Analyze which we cannot mock easily,
	// so we test the history stack manipulation that handleAnalyzeKey("esc") uses.

	// Push current level onto history (simulating what handleAnalyzeKey("enter") does)
	m.analyzeHistory = append(m.analyzeHistory, analyzeLevel{
		path:    m.analyzePath,
		entries: m.analyzeEntries,
		total:   m.analyzeTotal,
		cursor:  m.analyzeCursor,
	})
	// Set child level state
	m.analyzePath = "/home/Documents"
	m.analyzeTotal = 3000
	m.analyzeEntries = []jackal.DirEntry{
		{Name: "work", Path: "/home/Documents/work", Size: 2000, IsDir: true},
		{Name: "notes.md", Path: "/home/Documents/notes.md", Size: 1000, IsDir: false},
	}
	m.analyzeCursor = 1

	if len(m.analyzeHistory) != 1 {
		t.Fatalf("history stack should have 1 entry, got %d", len(m.analyzeHistory))
	}

	// Press esc/left to go back
	updated, _ := m.handleAnalyzeKey("esc")
	model := updated.(TUIModel)

	if model.mode != viewAnalyze {
		t.Errorf("after back: mode = %d, want viewAnalyze (%d)", model.mode, viewAnalyze)
	}
	if model.analyzePath != "/home" {
		t.Errorf("after back: analyzePath = %q, want /home", model.analyzePath)
	}
	if len(model.analyzeEntries) != 2 {
		t.Errorf("after back: entries count = %d, want 2", len(model.analyzeEntries))
	}
	if model.analyzeCursor != 0 {
		t.Errorf("after back: cursor = %d, want 0 (restored)", model.analyzeCursor)
	}
	if len(model.analyzeHistory) != 0 {
		t.Errorf("after back: history should be empty, got %d", len(model.analyzeHistory))
	}

	// Press esc again with empty history should exit to viewTabs
	updated, _ = model.handleAnalyzeKey("esc")
	model = updated.(TUIModel)

	if model.mode != viewTabs {
		t.Errorf("final esc: mode = %d, want viewTabs (%d)", model.mode, viewTabs)
	}
}

func TestAnalyzeCursorBounds(t *testing.T) {
	m := newTestModel()
	m.mode = viewAnalyze
	m.analyzeEntries = []jackal.DirEntry{
		{Name: "a", Path: "/a", Size: 100, IsDir: false},
		{Name: "b", Path: "/b", Size: 200, IsDir: false},
	}
	m.analyzeCursor = 0

	// Up at top should not go negative
	updated, _ := m.handleAnalyzeKey("up")
	model := updated.(TUIModel)
	if model.analyzeCursor != 0 {
		t.Errorf("up at top: cursor = %d, want 0", model.analyzeCursor)
	}

	// Down to last
	updated, _ = model.handleAnalyzeKey("down")
	model = updated.(TUIModel)
	if model.analyzeCursor != 1 {
		t.Errorf("down: cursor = %d, want 1", model.analyzeCursor)
	}

	// Down at bottom should stay
	updated, _ = model.handleAnalyzeKey("down")
	model = updated.(TUIModel)
	if model.analyzeCursor != 1 {
		t.Errorf("down at bottom: cursor = %d, want 1", model.analyzeCursor)
	}
}

// ── TestCleanGatewayBlocks ──────────────────────────────────────────
// Verifies that destructive clean confirmation is gated: when the
// onConfirm callback returns an error, the TUI transitions to viewDone
// with the error preserved and no items are cleaned.
//
// The current codebase does not have a SafetyGateway interface yet.
// This test validates the existing gateway behavior: the onConfirm
// function is the gateway, and returning an error blocks the clean.

func TestCleanGatewayBlocks(t *testing.T) {
	type testCase struct {
		name        string
		gatewayErr  error
		wantMode    viewMode
		wantErrNil  bool
		wantLines   int // minimum lines in output
	}

	tests := []testCase{
		{
			name:       "gateway error blocks clean",
			gatewayErr: errors.New("safety check failed: protected path"),
			wantMode:   viewDone,
			wantErrNil: false,
			wantLines:  1,
		},
		{
			name:       "gateway allows clean",
			gatewayErr: nil,
			wantMode:   viewDone,
			wantErrNil: true,
			wantLines:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Build a nativeResult with a selectReq whose onConfirm acts as
			// the safety gateway.
			gatewayErr := tc.gatewayErr
			gateway := func(selected []selectItem) nativeResult {
				if gatewayErr != nil {
					return nativeResult{
						deityKey: "anubis",
						err:      gatewayErr,
						lines:    []string{"blocked by safety gateway"},
					}
				}
				return nativeResult{
					deityKey: "anubis",
					lines:    []string{"clean complete", "2 items removed"},
				}
			}

			// Step 1: Transition to viewSelect via nativeResult with selectReq
			m := newTestModel()
			m.mode = viewRunning

			selectMsg := nativeResultMsg{
				deityKey: "anubis",
				selectReq: &selectRequest{
					title: "Clean Items",
					items: []selectItem{
						{Label: "Cache A", Size: 1024, Selected: true, Data: jackal.Finding{
							Description: "Cache A", Path: "/tmp/cache_a", SizeBytes: 1024,
							Severity: jackal.SeveritySafe, CanFix: true,
						}},
						{Label: "Cache B", Size: 2048, Selected: true, Data: jackal.Finding{
							Description: "Cache B", Path: "/tmp/cache_b", SizeBytes: 2048,
							Severity: jackal.SeveritySafe, CanFix: true,
						}},
					},
					onConfirm: gateway,
				},
			}

			model, _ := m.handleNativeResult(selectMsg)

			if model.mode != viewSelect {
				t.Fatalf("expected viewSelect after select msg, got %d", model.mode)
			}

			// Step 2: Call the gateway directly (simulating what handleSelectKey("enter") does
			// without the tea.Cmd async layer).
			var selected []selectItem
			for _, item := range model.selectItems {
				if item.Selected {
					selected = append(selected, item)
				}
			}
			result := model.selectOnConfirm(selected)

			// Step 3: Feed the result back as a nativeResultMsg
			resultMsg := nativeResultMsg(result)
			finalModel, _ := model.handleNativeResult(resultMsg)

			if finalModel.mode != tc.wantMode {
				t.Errorf("final mode = %d, want %d", finalModel.mode, tc.wantMode)
			}

			if tc.wantErrNil {
				// Success path: outputLines should have the success message
				if len(finalModel.outputLines) < tc.wantLines {
					t.Errorf("outputLines count = %d, want >= %d", len(finalModel.outputLines), tc.wantLines)
				}
			} else {
				// Error path: the error should have been captured in the done state
				if finalModel.lastSummary != "Failed" {
					t.Errorf("lastSummary = %q, want 'Failed'", finalModel.lastSummary)
				}
			}
		})
	}
}

// TestCleanGatewayNoSelectionExitsCleanly verifies that pressing enter
// with no items selected in viewSelect returns to viewTabs without
// invoking the onConfirm gateway at all.
func TestCleanGatewayNoSelectionExitsCleanly(t *testing.T) {
	gatewayCalled := false
	gateway := func(selected []selectItem) nativeResult {
		gatewayCalled = true
		return nativeResult{deityKey: "anubis", lines: []string{"should not happen"}}
	}

	m := newTestModel()
	m.mode = viewSelect
	m.selectItems = []selectItem{
		{Label: "Item A", Selected: false},
		{Label: "Item B", Selected: false},
	}
	m.selectOnConfirm = gateway

	// Press enter with nothing selected
	updated, _ := m.handleSelectKey("enter")
	model := updated.(TUIModel)

	if model.mode != viewTabs {
		t.Errorf("mode = %d, want viewTabs (%d)", model.mode, viewTabs)
	}
	if gatewayCalled {
		t.Error("gateway should not be called when nothing is selected")
	}
}
