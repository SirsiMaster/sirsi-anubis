package output

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/vitals"
)

// nativeResult is returned by native deity calls.
type nativeResult struct {
	lines      []string // rendered output lines
	deityKey   string   // which deity ran
	fixCmds    []string // actionable fix commands (override suggest engine)
	err        error
	selectReq  *selectRequest        // if non-nil, enter viewSelect mode instead of viewDone
	analyzeRes *jackal.AnalyzeResult // if non-nil, enter viewAnalyze mode
}

type nativeResultMsg nativeResult

type analyzeResultMsg struct {
	result *jackal.AnalyzeResult
	err    error
}

// ── View Mode ────────────────────────────────────────────────────────

type viewMode int

const (
	viewTabs    viewMode = iota // Showing a tab landing page
	viewRunning                 // Command executing
	viewDone                    // Command finished, output + next actions
	viewPrompt                  // Power-user command prompt (: key)
	viewSelect                  // Interactive checkbox selection
	viewAnalyze                 // Disk space analyzer with drill-down
)

// ── Selection Types ──────────────────────────────────────────────────

type selectItem struct {
	Label    string
	Detail   string // secondary line (path, size, etc.)
	Size     int64  // for size display
	Selected bool
	Data     interface{} // opaque payload for the confirm handler
}

type selectRequest struct {
	title     string
	items     []selectItem
	onConfirm func(selected []selectItem) nativeResult
}

type analyzeLevel struct {
	path    string
	entries []jackal.DirEntry
	total   int64
	cursor  int
}

type systemVitals = vitals.Snapshot

type deityRunState = deity.RunState

const (
	stateNeverRun  = deity.StateNeverRun
	stateSucceeded = deity.StateSucceeded
	stateFailed    = deity.StateFailed
	stateHasData   = deity.StateHasData
)

type historyEntry struct {
	deity, command, output string
}

// ── Messages ─────────────────────────────────────────────────────────

type refreshMsg time.Time
type elapsedTickMsg time.Time
type liveTickMsg time.Time

type streamLineMsg struct {
	line string
	done bool
	err  error
}

func refreshTick() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

func liveTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return liveTickMsg(t) })
}

func elapsedTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return elapsedTickMsg(t) })
}

func waitForStreamLine(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return streamLineMsg{done: true}
		}
		return streamLineMsg{line: line}
	}
}

func waitForStreamLineResult(ch <-chan string, errCh <-chan error) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if ok {
			return streamLineMsg{line: line}
		}
		var err error
		if errCh != nil {
			err = <-errCh
		}
		return streamLineMsg{done: true, err: err}
	}
}
