// Package output — Pantheon TUI
//
// Tab-based interface inspired by Mole (mole.fit). Each deity group gets
// its own page, its own job. No split panes, no REPL. Guided navigation
// with numbered actions. Press a number to act. Press esc to go back.
//
// Layout:
//   tui.go                     — Model, constructor, Init, Update, launchers, persistence
//   tui_keys.go                — Key handling dispatch per view mode
//   tui_view.go                — View rendering, tab bar, status dashboard, layout helpers
//   tui_runner.go              — Command execution (native + subprocess streaming)
//   tui_actions.go             — Tab definitions, native deity functions
//   tui_messages.go            — Message types, view modes, tick functions
//   tui_render*.go             — Render primitives and detail renderers
package output

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/SirsiMaster/sirsi-pantheon/internal/deity"
	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
	"github.com/SirsiMaster/sirsi-pantheon/internal/notify"
	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
	"github.com/SirsiMaster/sirsi-pantheon/internal/suggest"
)

// ── Model ────────────────────────────────────────────────────────────

type TUIModel struct {
	width  int
	height int

	activeTab int      // 0-4 index into tabs
	mode      viewMode // current view state

	// Checkbox selection state
	selectItems     []selectItem
	selectCursor    int
	selectTitle     string
	selectOnConfirm func(selected []selectItem) nativeResult

	// Command execution
	input        textinput.Model
	viewport     viewport.Model
	outputLines  []string
	runningDeity string
	runningCmd   string
	runningArgs  []string
	cmdStartTime time.Time
	spinner      spinner.Model
	streamCh     chan string
	runningProc  *atomic.Pointer[os.Process]

	// Post-run suggestions
	postRunCmds    []string
	postRunActions []suggest.Action // full actions with descriptions
	tabIdx         int
	lastDeity      string // deity key preserved for done view
	lastCommand    string
	lastSummary    string

	// History
	history    []historyEntry
	cmdHistory []string
	historyIdx int

	// State
	activeDeity map[string]bool
	deityState  map[string]deityRunState
	steleReader *stele.Reader
	quitting    bool

	// Notifications
	notifyStore       *notify.Store
	recentNotify      []notify.Notification
	notifyRefreshTime time.Time

	// System vitals
	vitals systemVitals

	// Live dashboard history (ring buffers for sparklines)
	cpuHistory  []float64 // last 60 samples
	memHistory  []float64 // last 60 samples
	netDownHist []float64 // last 60 samples
	netUpHist   []float64 // last 60 samples

	// Safety gateway for destructive actions
	safetyGateway SafetyGateway

	// Disk analyzer state
	analyzePath string
	analyzeEntries []jackal.DirEntry
	analyzeCursor  int
	analyzeTotal   int64
	analyzeHistory []analyzeLevel
}

// ── Constructor ──────────────────────────────────────────────────────

func NewTUIModel() TUIModel {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.CharLimit = 256
	ti.Prompt = "𓉴 "
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Gold).Bold(true)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(White)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	ti.SetStyles(styles)

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(Gold)

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(10))

	m := TUIModel{
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
		steleReader: stele.NewReader("tui"),
	}
	m.safetyGateway = &defaultSafetyGateway{}
	setCleanGateway(m.safetyGateway)
	m.refreshActive()
	return m
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(refreshTick())
}

// ── Update ───────────────────────────────────────────────────────────

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcViewport()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case nativeResultMsg:
		return m.handleNativeResult(msg)

	case analyzeResultMsg:
		if msg.err != nil {
			if len(m.analyzeHistory) > 0 {
				last := m.analyzeHistory[len(m.analyzeHistory)-1]
				m.analyzeHistory = m.analyzeHistory[:len(m.analyzeHistory)-1]
				m.analyzePath = last.path
				m.analyzeEntries = last.entries
				m.analyzeTotal = last.total
				m.analyzeCursor = last.cursor
			}
			m.mode = viewAnalyze
			return m, nil
		}
		m.mode = viewAnalyze
		m.analyzePath = msg.result.Path
		m.analyzeEntries = msg.result.Entries
		m.analyzeTotal = msg.result.TotalSize
		m.analyzeCursor = 0
		m.runningDeity = ""
		m.runningCmd = ""
		m.runningArgs = nil
		return m, nil

	case streamLineMsg:
		return m.handleStreamLine(msg)

	case liveTickMsg:
		if m.mode == viewTabs && m.activeTab == 4 {
			m.refreshVitals()
			m.appendHistory()
			return m, liveTick()
		}
		return m, nil

	case elapsedTickMsg:
		if m.mode == viewRunning {
			return m, elapsedTick()
		}
		return m, nil

	case refreshMsg:
		m.refreshActive()
		return m, refreshTick()

	case spinner.TickMsg:
		if m.mode == viewRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Pass through to active component
	if m.mode == viewPrompt {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// ── Background Refresh ───────────────────────────────────────────────

func (m *TUIModel) refreshActive() {
	for k := range m.activeDeity {
		delete(m.activeDeity, k)
	}
	entries, _ := m.steleReader.ReadNew()
	now := time.Now()
	for _, e := range entries {
		ts, err := time.Parse(time.RFC3339, e.TS)
		if err != nil {
			continue
		}
		if now.Sub(ts) < 5*time.Minute {
			dKey := strings.ToLower(e.Deity)
			if !strings.Contains(dKey, ":") {
				m.activeDeity[dKey] = true
			}
		}
	}
	m.refreshNotifications()
	m.refreshVitals()

	home, _ := os.UserHomeDir()
	pidDir := filepath.Join(home, ".config", "ra", "pids")
	pidEntries, _ := os.ReadDir(pidDir)
	for _, f := range pidEntries {
		if f.IsDir() {
			continue
		}
		name := strings.TrimSuffix(f.Name(), ".pid")
		for _, d := range deity.Roster {
			if strings.Contains(strings.ToLower(name), d.Key) {
				m.activeDeity[d.Key] = true
			}
		}
	}
}

// ── Launcher ─────────────────────────────────────────────────────────

func LaunchTUI() error {
	return LaunchTUIWithNotify(nil)
}

// LaunchTUIOnTab opens the TUI directly on a specific tab (0-indexed).
func LaunchTUIOnTab(tab int) error {
	m := NewTUIModel()
	if tab >= 0 && tab < len(tabs) {
		m.activeTab = tab
	}
	m.loadPersistedState()
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func LaunchTUIWithNotify(store *notify.Store) error {
	m := NewTUIModel()
	m.notifyStore = store
	m.refreshNotifications()
	m.loadPersistedState()
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// ── Persistent State ─────────────────────────────────────────────────

func (m *TUIModel) loadPersistedState() {
	state, err := deity.LoadState()
	if err != nil {
		return
	}
	for k, v := range state.DeityState {
		m.deityState[k] = v
	}
	m.lastCommand = state.LastCommand
	m.lastSummary = state.LastSummary
	m.postRunCmds = append([]string(nil), state.LastRecommendations...)
}

func (m *TUIModel) savePersistedState() {
	_ = deity.SaveState(deity.PersistedState{
		DeityState:          m.deityState,
		LastCommand:         m.lastCommand,
		LastSummary:         m.lastSummary,
		LastRecommendations: append([]string(nil), m.postRunCmds...),
	})
}

// ── Unused but required by tests ─────────────────────────────────────
// These are no-ops preserved for test compilation. The old REPL functions
// (showFindings, showHelp, renderRosterColumns, etc.) are removed.

var _ = color.RGBA{} // keep image/color import for lipgloss
